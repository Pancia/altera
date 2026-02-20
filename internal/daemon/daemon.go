// Package daemon implements the main orchestration loop for the Altera
// multi-agent system. It runs a 60-second tick cycle that checks agent
// liveness, monitors progress, assigns tasks, processes messages, manages
// the merge queue, enforces constraints, and emits events.
package daemon

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/anthropics/altera/internal/agent"
	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/constraints"
	"github.com/anthropics/altera/internal/events"
	"github.com/anthropics/altera/internal/git"
	"github.com/anthropics/altera/internal/message"
	"github.com/anthropics/altera/internal/task"
	"github.com/anthropics/altera/internal/tmux"
)

// TickInterval is the duration between daemon ticks.
const TickInterval = 60 * time.Second

// StalledThreshold is the duration after which a worker with no recent
// commits is considered stalled.
const StalledThreshold = 30 * time.Minute

// Daemon is the main orchestration process. It coordinates agent
// lifecycle, task assignment, merge queue processing, and event logging.
type Daemon struct {
	altDir  string // path to .alt/ directory
	rootDir string // project root (parent of .alt/)

	cfg       config.Config
	agents    *agent.Store
	tasks     *task.Store
	messages  *message.Store
	events    *events.Writer
	evReader  *events.Reader
	checker   *constraints.Checker

	pidFile  string   // path to .alt/daemon.pid
	lockFile *os.File // held flock on pid file

	logger   *log.Logger
	shutdown chan struct{} // closed to signal shutdown
}

// New creates a Daemon rooted at the given project directory. The .alt/
// directory must already exist. Call Run to start the tick loop.
func New(rootDir string) (*Daemon, error) {
	altDir := filepath.Join(rootDir, config.DirName)
	if _, err := os.Stat(altDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("daemon: %s directory not found in %s", config.DirName, rootDir)
	}

	cfg, err := config.Load(altDir)
	if err != nil {
		return nil, fmt.Errorf("daemon: load config: %w", err)
	}

	agentStore, err := agent.NewStore(filepath.Join(altDir, "agents"))
	if err != nil {
		return nil, fmt.Errorf("daemon: create agent store: %w", err)
	}

	taskStore, err := task.NewStore(rootDir)
	if err != nil {
		return nil, fmt.Errorf("daemon: create task store: %w", err)
	}

	msgStore, err := message.NewStore(filepath.Join(altDir, "messages"))
	if err != nil {
		return nil, fmt.Errorf("daemon: create message store: %w", err)
	}

	evPath := filepath.Join(altDir, "events.jsonl")
	evWriter := events.NewWriter(evPath)
	evReader := events.NewReader(evPath)

	mergeQueueDir := filepath.Join(altDir, "merge-queue")
	if err := os.MkdirAll(mergeQueueDir, 0o755); err != nil {
		return nil, fmt.Errorf("daemon: create merge-queue dir: %w", err)
	}

	checker := constraints.NewChecker(cfg.Constraints, agentStore, evReader, mergeQueueDir)

	logger := log.New(os.Stderr, "daemon: ", log.LstdFlags|log.Lmsgprefix)

	return &Daemon{
		altDir:   altDir,
		rootDir:  rootDir,
		cfg:      cfg,
		agents:   agentStore,
		tasks:    taskStore,
		messages: msgStore,
		events:   evWriter,
		evReader: evReader,
		checker:  checker,
		pidFile:  filepath.Join(altDir, "daemon.pid"),
		logger:   logger,
		shutdown: make(chan struct{}),
	}, nil
}

// Run starts the daemon tick loop. It acquires a PID file lock, installs
// signal handlers, and loops until signaled to stop. Run blocks until
// the daemon shuts down.
func (d *Daemon) Run() error {
	if err := d.acquireLock(); err != nil {
		return err
	}
	defer d.releaseLock()

	d.installSignalHandler()
	d.logger.Println("started")

	// Run one tick immediately, then loop on the interval.
	d.tick()

	ticker := time.NewTicker(TickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-d.shutdown:
			d.logger.Println("shutting down gracefully")
			return nil
		case <-ticker.C:
			d.tick()
		}
	}
}

// Stop signals the daemon to shut down after finishing the current tick.
func (d *Daemon) Stop() {
	select {
	case <-d.shutdown:
		// already closed
	default:
		close(d.shutdown)
	}
}

// tick runs all seven daemon steps in sequence.
func (d *Daemon) tick() {
	d.logger.Println("tick start")
	start := time.Now()

	var tickEvents []events.Event

	d.checkAgentLiveness(&tickEvents)
	d.checkProgress(&tickEvents)
	d.assignTasks(&tickEvents)
	d.processMessages(&tickEvents)
	d.processMergeQueue(&tickEvents)
	d.checkConstraints(&tickEvents)
	d.emitEvents(tickEvents)

	d.logger.Printf("tick complete (%s)", time.Since(start).Round(time.Millisecond))
}

// --- Step 1: CheckAgentLiveness ---

// checkAgentLiveness checks heartbeat timestamps and OS process existence
// for all active agents. Dead agents have their tasks reclaimed.
func (d *Daemon) checkAgentLiveness(tickEvents *[]events.Event) {
	active, err := d.agents.ListByStatus(agent.StatusActive)
	if err != nil {
		d.logger.Printf("liveness: list active agents: %v", err)
		return
	}

	for _, a := range active {
		if agent.CheckLiveness(a) {
			continue
		}

		d.logger.Printf("liveness: agent %s is dead (role=%s, task=%s)", a.ID, a.Role, a.CurrentTask)

		// Mark agent as dead.
		a.Status = agent.StatusDead
		if err := d.agents.Update(a); err != nil {
			d.logger.Printf("liveness: update agent %s: %v", a.ID, err)
			continue
		}

		*tickEvents = append(*tickEvents, events.Event{
			Timestamp: time.Now(),
			Type:      events.AgentDied,
			AgentID:   a.ID,
			TaskID:    a.CurrentTask,
		})

		// Reclaim task: force status back to open, clear assigned_to.
		// This bypasses normal transition validation since it's a
		// recovery operation for dead agents.
		if a.CurrentTask != "" {
			if err := d.reclaimTask(a.CurrentTask); err != nil {
				d.logger.Printf("liveness: reclaim task %s: %v", a.CurrentTask, err)
			} else {
				d.logger.Printf("liveness: reclaimed task %s", a.CurrentTask)
			}
		}

		// Clean up tmux session if it exists.
		if a.TmuxSession != "" && tmux.SessionExists(a.TmuxSession) {
			if err := tmux.KillSession(a.TmuxSession); err != nil {
				d.logger.Printf("liveness: kill tmux session %s: %v", a.TmuxSession, err)
			}
		}
	}
}

// reclaimTask forces a task back to open status, bypassing normal transition
// validation. This is used when a dead agent's task needs to be reassigned.
func (d *Daemon) reclaimTask(taskID string) error {
	t, err := d.tasks.Get(taskID)
	if err != nil {
		return err
	}
	t.Status = task.StatusOpen
	t.AssignedTo = ""
	t.UpdatedAt = time.Now().UTC()
	return d.tasks.ForceWrite(t)
}

// --- Step 2: CheckProgress ---

// checkProgress checks last commit time in each worker's worktree. If a
// worker has been stalled for longer than StalledThreshold, a help message
// is sent to the liaison.
func (d *Daemon) checkProgress(tickEvents *[]events.Event) {
	workers, err := d.agents.ListByRole(agent.RoleWorker)
	if err != nil {
		d.logger.Printf("progress: list workers: %v", err)
		return
	}

	for _, w := range workers {
		if w.Status != agent.StatusActive || w.Worktree == "" {
			continue
		}

		lastCommitTime, err := d.lastCommitTime(w.Worktree)
		if err != nil {
			// No commits yet or error reading - skip silently.
			continue
		}

		if time.Since(lastCommitTime) > StalledThreshold {
			d.logger.Printf("progress: worker %s stalled (last commit %s ago)", w.ID, time.Since(lastCommitTime).Round(time.Second))

			*tickEvents = append(*tickEvents, events.Event{
				Timestamp: time.Now(),
				Type:      events.WorkerStalled,
				AgentID:   w.ID,
				TaskID:    w.CurrentTask,
				Data: map[string]any{
					"stalled_since": lastCommitTime.Format(time.RFC3339),
				},
			})

			// Send help message to liaison.
			liaisons, err := d.agents.ListByRole(agent.RoleLiaison)
			if err != nil || len(liaisons) == 0 {
				d.logger.Printf("progress: no liaison to notify about stalled worker %s", w.ID)
				continue
			}
			_, err = d.messages.Create(
				message.TypeHelp,
				"daemon",
				liaisons[0].ID,
				w.CurrentTask,
				map[string]any{
					"worker_id":    w.ID,
					"stalled_since": lastCommitTime.Format(time.RFC3339),
					"message":       fmt.Sprintf("worker %s stalled for %s", w.ID, time.Since(lastCommitTime).Round(time.Minute)),
				},
			)
			if err != nil {
				d.logger.Printf("progress: send help message: %v", err)
			}
		}
	}
}

// lastCommitTime returns the author timestamp of the most recent commit
// in the given worktree path.
func (d *Daemon) lastCommitTime(worktree string) (time.Time, error) {
	// Use git log with a format that gives us the unix timestamp.
	out, err := gitLogTimestamp(worktree)
	if err != nil {
		return time.Time{}, err
	}
	if out == "" {
		return time.Time{}, fmt.Errorf("no commits in %s", worktree)
	}
	ts, err := strconv.ParseInt(out, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse commit timestamp %q: %w", out, err)
	}
	return time.Unix(ts, 0), nil
}

// --- Step 3: AssignTasks ---

// assignTasks finds open tasks with resolved dependencies, checks
// constraints, and spawns worker agents for each assignable task.
func (d *Daemon) assignTasks(tickEvents *[]events.Event) {
	ready, err := d.tasks.FindReady()
	if err != nil {
		d.logger.Printf("assign: find ready tasks: %v", err)
		return
	}

	for _, t := range ready {
		ok, reason := d.checker.CanSpawnWorker()
		if !ok {
			d.logger.Printf("assign: cannot spawn worker: %s", reason)
			break // constraints apply globally, no point continuing
		}

		agentID, err := d.spawnWorker(t)
		if err != nil {
			d.logger.Printf("assign: spawn worker for task %s: %v", t.ID, err)
			continue
		}

		*tickEvents = append(*tickEvents, events.Event{
			Timestamp: time.Now(),
			Type:      events.TaskAssigned,
			AgentID:   agentID,
			TaskID:    t.ID,
		})
		*tickEvents = append(*tickEvents, events.Event{
			Timestamp: time.Now(),
			Type:      events.AgentSpawned,
			AgentID:   agentID,
			TaskID:    t.ID,
		})
	}
}

// spawnWorker creates a new worker agent for the given task. It creates a
// git branch and worktree, a tmux session, assigns the task, and registers
// the agent.
func (d *Daemon) spawnWorker(t *task.Task) (string, error) {
	agentID, err := generateAgentID()
	if err != nil {
		return "", fmt.Errorf("generate agent id: %w", err)
	}

	branchName := "worker/" + agentID
	worktreePath := filepath.Join(d.rootDir, ".alt", "worktrees", agentID)
	sessionName := tmux.SessionName("worker", agentID)

	// Create branch from default branch (main).
	if err := git.CreateBranch(d.rootDir, branchName, "HEAD"); err != nil {
		return "", fmt.Errorf("create branch: %w", err)
	}

	// Create worktree.
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0o755); err != nil {
		return "", fmt.Errorf("create worktree parent: %w", err)
	}
	if err := git.CreateWorktree(d.rootDir, branchName, worktreePath); err != nil {
		// Clean up branch on failure.
		git.DeleteBranch(d.rootDir, branchName)
		return "", fmt.Errorf("create worktree: %w", err)
	}

	// Create tmux session.
	if err := tmux.CreateSession(sessionName); err != nil {
		// Clean up worktree and branch on failure.
		git.DeleteWorktree(d.rootDir, worktreePath)
		git.DeleteBranch(d.rootDir, branchName)
		return "", fmt.Errorf("create tmux session: %w", err)
	}

	// Register the agent.
	a := &agent.Agent{
		ID:          agentID,
		Role:        agent.RoleWorker,
		Status:      agent.StatusActive,
		CurrentTask: t.ID,
		Worktree:    worktreePath,
		TmuxSession: sessionName,
		Heartbeat:   time.Now(),
		StartedAt:   time.Now(),
	}
	if err := d.agents.Create(a); err != nil {
		tmux.KillSession(sessionName)
		git.DeleteWorktree(d.rootDir, worktreePath)
		git.DeleteBranch(d.rootDir, branchName)
		return "", fmt.Errorf("register agent: %w", err)
	}

	// Assign the task to the agent.
	if err := d.tasks.Update(t.ID, func(t *task.Task) error {
		t.Status = task.StatusAssigned
		t.AssignedTo = agentID
		t.Branch = branchName
		return nil
	}); err != nil {
		d.agents.Delete(agentID)
		tmux.KillSession(sessionName)
		git.DeleteWorktree(d.rootDir, worktreePath)
		git.DeleteBranch(d.rootDir, branchName)
		return "", fmt.Errorf("assign task: %w", err)
	}

	d.logger.Printf("assign: spawned worker %s for task %s", agentID, t.ID)
	return agentID, nil
}

// --- Step 4: ProcessMessages ---

// processMessages reads pending messages addressed to the daemon and
// dispatches them by type.
func (d *Daemon) processMessages(tickEvents *[]events.Event) {
	msgs, err := d.messages.ListPending("daemon")
	if err != nil {
		d.logger.Printf("messages: list pending: %v", err)
		return
	}

	for _, msg := range msgs {
		switch msg.Type {
		case message.TypeTaskDone:
			d.handleTaskDone(msg, tickEvents)
		case message.TypeHelp:
			d.handleHelp(msg)
		default:
			d.logger.Printf("messages: unhandled type %s from %s", msg.Type, msg.From)
		}

		// Archive processed message.
		if err := d.messages.Archive(msg.ID); err != nil {
			d.logger.Printf("messages: archive %s: %v", msg.ID, err)
		}
	}
}

// handleTaskDone processes a task_done message by marking the task as done
// and adding it to the merge queue.
func (d *Daemon) handleTaskDone(msg *message.Message, tickEvents *[]events.Event) {
	taskID := msg.TaskID
	if taskID == "" {
		d.logger.Printf("messages: task_done without task_id from %s", msg.From)
		return
	}

	// Mark task as done.
	if err := d.tasks.Update(taskID, func(t *task.Task) error {
		t.Status = task.StatusDone
		if result, ok := msg.Payload["result"].(string); ok {
			t.Result = result
		}
		return nil
	}); err != nil {
		d.logger.Printf("messages: mark task %s done: %v", taskID, err)
		return
	}

	*tickEvents = append(*tickEvents, events.Event{
		Timestamp: time.Now(),
		Type:      events.TaskDone,
		AgentID:   msg.From,
		TaskID:    taskID,
	})

	// Add to merge queue.
	t, err := d.tasks.Get(taskID)
	if err != nil {
		d.logger.Printf("messages: get task %s for merge queue: %v", taskID, err)
		return
	}
	if err := d.addToMergeQueue(t); err != nil {
		d.logger.Printf("messages: add task %s to merge queue: %v", taskID, err)
	}
}

// handleHelp forwards a help message to the first available liaison.
func (d *Daemon) handleHelp(msg *message.Message) {
	liaisons, err := d.agents.ListByRole(agent.RoleLiaison)
	if err != nil || len(liaisons) == 0 {
		d.logger.Printf("messages: no liaison available to handle help from %s", msg.From)
		return
	}

	_, err = d.messages.Create(
		message.TypeHelp,
		"daemon",
		liaisons[0].ID,
		msg.TaskID,
		msg.Payload,
	)
	if err != nil {
		d.logger.Printf("messages: forward help to liaison: %v", err)
	}
}

// --- Step 5: ProcessMergeQueue ---

// MergeItem represents a task waiting to be merged in the queue.
type MergeItem struct {
	TaskID   string    `json:"task_id"`
	Branch   string    `json:"branch"`
	AgentID  string    `json:"agent_id"`
	QueuedAt time.Time `json:"queued_at"`
}

// processMergeQueue processes items in the merge queue FIFO. Each item
// is a task branch that needs to be merged into the default branch.
func (d *Daemon) processMergeQueue(tickEvents *[]events.Event) {
	queueDir := filepath.Join(d.altDir, "merge-queue")
	entries, err := os.ReadDir(queueDir)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		d.logger.Printf("merge: read queue dir: %v", err)
		return
	}

	// Process items in filename order (FIFO by timestamp prefix).
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}

		// Check for shutdown between merge attempts.
		select {
		case <-d.shutdown:
			d.logger.Println("merge: shutdown during queue processing")
			return
		default:
		}

		itemPath := filepath.Join(queueDir, e.Name())
		data, err := os.ReadFile(itemPath)
		if err != nil {
			d.logger.Printf("merge: read queue item %s: %v", e.Name(), err)
			continue
		}

		var item MergeItem
		if err := json.Unmarshal(data, &item); err != nil {
			d.logger.Printf("merge: parse queue item %s: %v", e.Name(), err)
			continue
		}

		*tickEvents = append(*tickEvents, events.Event{
			Timestamp: time.Now(),
			Type:      events.MergeStarted,
			AgentID:   item.AgentID,
			TaskID:    item.TaskID,
		})

		result, err := git.Merge(d.rootDir, item.Branch)
		if err != nil {
			d.logger.Printf("merge: merge branch %s: %v", item.Branch, err)
			*tickEvents = append(*tickEvents, events.Event{
				Timestamp: time.Now(),
				Type:      events.MergeFailed,
				AgentID:   item.AgentID,
				TaskID:    item.TaskID,
				Data:      map[string]any{"error": err.Error()},
			})
			continue
		}

		if !result.Clean {
			d.logger.Printf("merge: conflict merging %s: %v", item.Branch, result.Conflicts)
			git.AbortMerge(d.rootDir)
			*tickEvents = append(*tickEvents, events.Event{
				Timestamp: time.Now(),
				Type:      events.MergeConflict,
				AgentID:   item.AgentID,
				TaskID:    item.TaskID,
				Data:      map[string]any{"conflicts": result.Conflicts},
			})

			// Send merge result back to the agent.
			d.messages.Create(
				message.TypeMergeResult,
				"daemon",
				item.AgentID,
				item.TaskID,
				map[string]any{
					"success":   false,
					"conflicts": result.Conflicts,
				},
			)
			// Remove failed item from queue.
			os.Remove(itemPath)
			continue
		}

		d.logger.Printf("merge: successfully merged %s", item.Branch)
		*tickEvents = append(*tickEvents, events.Event{
			Timestamp: time.Now(),
			Type:      events.MergeSuccess,
			AgentID:   item.AgentID,
			TaskID:    item.TaskID,
		})

		// Remove merged item from queue.
		os.Remove(itemPath)

		// Send success notification.
		d.messages.Create(
			message.TypeMergeResult,
			"daemon",
			item.AgentID,
			item.TaskID,
			map[string]any{"success": true},
		)
	}
}

// addToMergeQueue writes a merge item to the queue directory.
func (d *Daemon) addToMergeQueue(t *task.Task) error {
	item := MergeItem{
		TaskID:   t.ID,
		Branch:   t.Branch,
		AgentID:  t.AssignedTo,
		QueuedAt: time.Now(),
	}
	data, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal merge item: %w", err)
	}
	data = append(data, '\n')

	filename := fmt.Sprintf("%d-%s.json", time.Now().UnixNano(), t.ID)
	path := filepath.Join(d.altDir, "merge-queue", filename)

	return atomicWrite(path, data)
}

// --- Step 6: CheckConstraints ---

// checkConstraints checks budget ceiling, max workers, and queue depth.
// If any constraint is violated, it emits an event.
func (d *Daemon) checkConstraints(tickEvents *[]events.Event) {
	if ok, reason, err := d.checker.CheckBudget(); err != nil {
		d.logger.Printf("constraints: budget check error: %v", err)
	} else if !ok {
		d.logger.Printf("constraints: %s", reason)
		*tickEvents = append(*tickEvents, events.Event{
			Timestamp: time.Now(),
			Type:      events.BudgetExceeded,
			Data:      map[string]any{"reason": reason},
		})
	}
}

// --- Step 7: EmitEvents ---

// emitEvents appends all accumulated tick events to the event log.
func (d *Daemon) emitEvents(tickEvents []events.Event) {
	if len(tickEvents) == 0 {
		return
	}
	if err := d.events.Append(tickEvents...); err != nil {
		d.logger.Printf("events: append: %v", err)
	}
}

// --- PID File & Lock Management ---

// acquireLock writes the daemon PID to .alt/daemon.pid and acquires
// an exclusive flock on it to prevent double-start.
func (d *Daemon) acquireLock() error {
	f, err := os.OpenFile(d.pidFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("daemon: open pid file: %w", err)
	}

	// Try non-blocking exclusive lock.
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return fmt.Errorf("daemon: already running (flock on %s failed): %w", d.pidFile, err)
	}

	// Write our PID.
	pid := fmt.Sprintf("%d\n", os.Getpid())
	if _, err := f.WriteString(pid); err != nil {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
		return fmt.Errorf("daemon: write pid: %w", err)
	}

	d.lockFile = f
	return nil
}

// releaseLock releases the flock and removes the PID file.
func (d *Daemon) releaseLock() {
	if d.lockFile == nil {
		return
	}
	syscall.Flock(int(d.lockFile.Fd()), syscall.LOCK_UN)
	d.lockFile.Close()
	os.Remove(d.pidFile)
	d.lockFile = nil
}

// installSignalHandler sets up SIGTERM and SIGINT to trigger graceful
// shutdown. The handler finishes the current tick before exiting.
func (d *Daemon) installSignalHandler() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		d.logger.Printf("received signal: %s", sig)
		d.Stop()
	}()
}

// --- Status ---

// Status represents the running state of the daemon.
type Status struct {
	Running bool  `json:"running"`
	PID     int   `json:"pid,omitempty"`
}

// ReadStatus checks whether a daemon is running by reading the PID file
// and checking the process. This is a static function that does not require
// a running Daemon instance.
func ReadStatus(altDir string) Status {
	pidFile := filepath.Join(altDir, "daemon.pid")
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return Status{Running: false}
	}

	pidStr := string(data)
	// Trim newline.
	if len(pidStr) > 0 && pidStr[len(pidStr)-1] == '\n' {
		pidStr = pidStr[:len(pidStr)-1]
	}
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return Status{Running: false}
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return Status{Running: false}
	}
	if proc.Signal(syscall.Signal(0)) != nil {
		return Status{Running: false}
	}

	return Status{Running: true, PID: pid}
}

// SendStop sends SIGTERM to a running daemon identified by its PID file.
func SendStop(altDir string) error {
	st := ReadStatus(altDir)
	if !st.Running {
		return fmt.Errorf("daemon is not running")
	}

	proc, err := os.FindProcess(st.PID)
	if err != nil {
		return fmt.Errorf("find daemon process: %w", err)
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("send SIGTERM to daemon (pid %d): %w", st.PID, err)
	}
	return nil
}

// --- Helpers ---

// generateAgentID creates a random agent ID in the form w-{6 hex chars}.
func generateAgentID() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate agent id: %w", err)
	}
	return "w-" + hex.EncodeToString(b), nil
}

// gitLogTimestamp runs git log to get the unix timestamp of the latest commit.
var gitLogTimestamp = func(worktree string) (string, error) {
	return gitLogFormat(worktree, "%at")
}

// gitLogFormat runs git log with a custom format string and returns the output
// for the most recent commit.
func gitLogFormat(path, format string) (string, error) {
	out, err := runGit(path, "log", "-1", "--format="+format)
	if err != nil {
		return "", err
	}
	return out, nil
}

// runGit executes a git command in the given directory. This is a thin wrapper
// around exec.Command to allow testing.
var runGit = func(dir string, args ...string) (string, error) {
	cmd := execCommand("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %v: %w", args, err)
	}
	return trimSpace(string(out)), nil
}

// execCommand is a variable for testing.
var execCommand = exec.Command

// trimSpace removes leading/trailing whitespace.
func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '\n' || s[0] == '\r') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t' || s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

// atomicWrite writes data to path via temp file + rename.
func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-daemon-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}
