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
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/anthropics/altera/internal/agent"
	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/constraints"
	"github.com/anthropics/altera/internal/events"
	"github.com/anthropics/altera/internal/git"
	"github.com/anthropics/altera/internal/merge"
	"github.com/anthropics/altera/internal/message"
	"github.com/anthropics/altera/internal/resolver"
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

	cfg         config.Config
	agents      *agent.Store
	tasks       *task.Store
	messages    *message.Store
	events      *events.Writer
	evReader    *events.Reader
	checker     *constraints.Checker
	resolverMgr *resolver.Manager

	pidFile  string   // path to .alt/daemon.pid
	lockFile *os.File // held flock on pid file

	logger   *slog.Logger
	tickNum  int64
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

	if err := cfg.Constraints.Validate(); err != nil {
		return nil, fmt.Errorf("daemon: invalid constraints: %w", err)
	}

	checker := constraints.NewChecker(cfg.Constraints, agentStore, evReader, mergeQueueDir)

	resolverMgr := resolver.NewManager(rootDir, agentStore, evWriter)

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil)).With("component", "daemon")

	return &Daemon{
		altDir:      altDir,
		rootDir:     rootDir,
		cfg:         cfg,
		agents:      agentStore,
		tasks:       taskStore,
		messages:    msgStore,
		events:      evWriter,
		evReader:    evReader,
		checker:     checker,
		resolverMgr: resolverMgr,
		pidFile:     filepath.Join(altDir, "daemon.pid"),
		logger:      logger,
		shutdown:    make(chan struct{}),
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

	// Reconcile stale state from prior runs before first tick.
	d.reconcile()

	d.events.Append(events.Event{
		Timestamp: time.Now(),
		Type:      events.DaemonStarted,
	})
	d.logger.Info("started")

	// Run one tick immediately, then loop on the interval.
	d.tick()

	ticker := time.NewTicker(TickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-d.shutdown:
			d.logger.Info("shutting down gracefully")
			d.events.Append(events.Event{
				Timestamp: time.Now(),
				Type:      events.DaemonShutdown,
			})
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

// reconcile cleans up stale state left by a prior daemon run that may
// have crashed or been killed. It is called once at startup before the
// first tick.
func (d *Daemon) reconcile() {
	d.logger.Info("reconcile", "step", "start")
	d.reconcileAgents()
	d.reconcileTasks()
	d.reconcileMergeQueue()
	d.logger.Info("reconcile", "step", "complete")
}

// reconcileAgents checks all active agents for liveness, marking dead
// ones and killing orphaned tmux sessions.
func (d *Daemon) reconcileAgents() {
	active, err := d.agents.ListByStatus(agent.StatusActive)
	if err != nil {
		d.logger.Error("reconcile agents: list active", "error", err)
		return
	}
	for _, a := range active {
		// Skip liaison — it has no PID and is managed interactively.
		if a.Role == agent.RoleLiaison {
			continue
		}
		if agent.CheckLiveness(a) {
			continue
		}
		d.logger.Info("reconcile agents: marking dead", "agent", a.ID, "role", a.Role)
		a.Status = agent.StatusDead
		if err := d.agents.Update(a); err != nil {
			d.logger.Error("reconcile agents: update", "agent", a.ID, "error", err)
			continue
		}
		if a.TmuxSession != "" && tmux.SessionExists(a.TmuxSession) {
			if err := tmux.KillSession(a.TmuxSession); err != nil {
				d.logger.Error("reconcile agents: kill tmux", "session", a.TmuxSession, "error", err)
			}
		}
	}
}

// reconcileTasks finds tasks assigned to dead or missing agents and
// resets them to open status.
func (d *Daemon) reconcileTasks() {
	for _, status := range []task.Status{task.StatusAssigned, task.StatusInProgress} {
		tasks, err := d.tasks.List(task.Filter{Status: status})
		if err != nil {
			d.logger.Error("reconcile tasks: list", "status", status, "error", err)
			continue
		}
		for _, t := range tasks {
			if t.AssignedTo == "" {
				continue
			}
			a, err := d.agents.Get(t.AssignedTo)
			if err != nil || a.Status == agent.StatusDead {
				d.logger.Info("reconcile tasks: reclaiming", "task", t.ID, "agent", t.AssignedTo)
				branch := t.Branch
				if err := d.reclaimTask(t.ID); err != nil {
					d.logger.Error("reconcile tasks: reclaim", "task", t.ID, "error", err)
					continue
				}
				// Clean up git resources from the dead worker.
				if branch != "" {
					d.cleanupBranch(branch)
				}
			}
		}
	}
}

// reconcileMergeQueue removes orphaned .tmp-* files from the merge-queue
// directory left by interrupted atomic writes.
func (d *Daemon) reconcileMergeQueue() {
	queueDir := filepath.Join(d.altDir, "merge-queue")
	entries, err := os.ReadDir(queueDir)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		d.logger.Error("reconcile merge queue: read dir", "error", err)
		return
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp-") {
			path := filepath.Join(queueDir, e.Name())
			d.logger.Info("reconcile merge queue: removing temp file", "file", e.Name())
			os.Remove(path)
		}
	}
}

// cleanupBranch deletes a worktree (if it exists) and its branch.
func (d *Daemon) cleanupBranch(branch string) {
	// Derive worktree path from branch name: "worker/{id}" -> ".alt/worktrees/{id}"
	parts := strings.SplitN(branch, "/", 2)
	if len(parts) == 2 {
		worktreePath := filepath.Join(d.rootDir, ".alt", "worktrees", parts[1])
		if _, err := os.Stat(worktreePath); err == nil {
			if err := git.DeleteWorktree(d.rootDir, worktreePath); err != nil {
				d.logger.Error("cleanup: delete worktree", "path", worktreePath, "error", err)
			}
		}
	}
	if err := git.DeleteBranch(d.rootDir, branch); err != nil {
		d.logger.Error("cleanup: delete branch", "branch", branch, "error", err)
	}
}

// tick runs all seven daemon steps in sequence.
func (d *Daemon) tick() {
	d.tickNum++
	d.logger.Info("tick start", "tick", d.tickNum)
	start := time.Now()

	// Reload config from disk so runtime changes (e.g. max_workers) take effect.
	if cfg, err := config.Load(d.altDir); err == nil {
		d.cfg = cfg
		d.checker.UpdateConstraints(cfg.Constraints)
	} else {
		d.logger.Error("tick: reload config", "error", err)
	}

	var tickEvents []events.Event

	d.checkAgentLiveness(&tickEvents)
	d.checkProgress(&tickEvents)
	d.assignTasks(&tickEvents)
	d.processMessages(&tickEvents)
	d.processMergeQueue(&tickEvents)
	d.checkResolvers(&tickEvents)
	d.checkConstraints(&tickEvents)
	d.emitEvents(tickEvents)

	d.logger.Info("tick complete", "tick", d.tickNum, "duration", time.Since(start).Round(time.Millisecond))
}

// --- Step 1: CheckAgentLiveness ---

// checkAgentLiveness implements 3-stage heartbeat escalation for active agents:
//
//	Warning  – heartbeat stale > 3 min, PID alive: log, nudge worker, set escalation.
//	Critical – heartbeat stale > 6 min, PID alive: notify liaison, set escalation.
//	Dead     – heartbeat stale > 10 min (PID alive) OR PID missing: kill, reclaim.
//
// If the heartbeat becomes fresh again the escalation level is cleared.
func (d *Daemon) checkAgentLiveness(tickEvents *[]events.Event) {
	active, err := d.agents.ListByStatus(agent.StatusActive)
	if err != nil {
		d.logger.Error("liveness: list active agents", "error", err)
		return
	}

	for _, a := range active {
		// Skip liaison — it has no PID and is managed interactively.
		if a.Role == agent.RoleLiaison {
			continue
		}

		pidAlive := agent.CheckPID(a)
		staleness := agent.HeartbeatStaleness(a)

		// PID missing = immediate death (process crashed), skip escalation.
		if !pidAlive {
			d.markAgentDead(a, tickEvents)
			continue
		}

		// PID alive, heartbeat fresh — clear any prior escalation.
		if staleness <= agent.HeartbeatWarnTimeout {
			if a.EscalationLevel != "" {
				d.logger.Info("liveness: heartbeat recovered", "agent", a.ID, "was", a.EscalationLevel)
				a.EscalationLevel = ""
				a.LastEscalation = time.Time{}
				if err := d.agents.Update(a); err != nil {
					d.logger.Error("liveness: clear escalation", "agent", a.ID, "error", err)
				}
			}
			continue
		}

		// PID alive, heartbeat stale > DeadTimeout → dead.
		if staleness > agent.HeartbeatDeadTimeout {
			d.markAgentDead(a, tickEvents)
			continue
		}

		// PID alive, heartbeat stale > CriticalTimeout → critical.
		if staleness > agent.HeartbeatCriticalTimeout {
			if a.EscalationLevel != agent.EscalationCritical {
				d.escalateCritical(a, tickEvents)
			}
			continue
		}

		// PID alive, heartbeat stale > WarnTimeout → warning.
		if a.EscalationLevel == "" {
			d.escalateWarning(a, tickEvents)
		}
	}
}

// markAgentDead marks an agent as dead, emits an AgentDied event, reclaims
// its task, and tears down its tmux session and git resources.
func (d *Daemon) markAgentDead(a *agent.Agent, tickEvents *[]events.Event) {
	d.logger.Info("liveness: agent dead", "agent", a.ID, "role", a.Role, "task", a.CurrentTask)

	a.Status = agent.StatusDead
	if err := d.agents.Update(a); err != nil {
		d.logger.Error("liveness: update agent", "agent", a.ID, "error", err)
		return
	}

	*tickEvents = append(*tickEvents, events.Event{
		Timestamp: time.Now(),
		Type:      events.AgentDied,
		AgentID:   a.ID,
		TaskID:    a.CurrentTask,
	})

	if a.CurrentTask != "" {
		t, _ := d.tasks.Get(a.CurrentTask)
		var branch string
		if t != nil {
			branch = t.Branch
		}
		if err := d.reclaimTask(a.CurrentTask); err != nil {
			d.logger.Error("liveness: reclaim task", "task", a.CurrentTask, "error", err)
		} else {
			d.logger.Info("liveness: reclaimed task", "task", a.CurrentTask)
			if branch != "" {
				d.cleanupBranch(branch)
			}
		}
	}

	if a.TmuxSession != "" && tmux.SessionExists(a.TmuxSession) {
		if err := tmux.KillSession(a.TmuxSession); err != nil {
			d.logger.Error("liveness: kill tmux session", "session", a.TmuxSession, "error", err)
		}
	}
}

// escalateWarning sets the agent's escalation level to warning, writes
// a .alt-nudge file to its worktree, and emits an AgentWarning event.
func (d *Daemon) escalateWarning(a *agent.Agent, tickEvents *[]events.Event) {
	d.logger.Warn("liveness: escalating to warning", "agent", a.ID,
		"staleness", agent.HeartbeatStaleness(a).Round(time.Second))

	a.EscalationLevel = agent.EscalationWarning
	a.LastEscalation = time.Now()
	if err := d.agents.Update(a); err != nil {
		d.logger.Error("liveness: update escalation", "agent", a.ID, "error", err)
		return
	}

	// Nudge worker by writing a .alt-nudge file to its worktree.
	if a.Worktree != "" {
		nudgePath := filepath.Join(a.Worktree, ".alt-nudge")
		os.WriteFile(nudgePath, []byte(time.Now().Format(time.RFC3339)+"\n"), 0o644)
	}

	*tickEvents = append(*tickEvents, events.Event{
		Timestamp: time.Now(),
		Type:      events.AgentWarning,
		AgentID:   a.ID,
		TaskID:    a.CurrentTask,
		Data: map[string]any{
			"escalation_level":  agent.EscalationWarning,
			"staleness_seconds": int(agent.HeartbeatStaleness(a).Seconds()),
		},
	})
}

// escalateCritical sets the agent's escalation level to critical, notifies
// the liaison that the worker appears unresponsive, and emits an AgentCritical event.
func (d *Daemon) escalateCritical(a *agent.Agent, tickEvents *[]events.Event) {
	d.logger.Warn("liveness: escalating to critical", "agent", a.ID,
		"staleness", agent.HeartbeatStaleness(a).Round(time.Second))

	a.EscalationLevel = agent.EscalationCritical
	a.LastEscalation = time.Now()
	if err := d.agents.Update(a); err != nil {
		d.logger.Error("liveness: update escalation", "agent", a.ID, "error", err)
		return
	}

	// Notify liaison that worker appears unresponsive.
	liaisons, err := d.agents.ListByRole(agent.RoleLiaison)
	if err == nil && len(liaisons) > 0 {
		_, err := d.messages.Create(
			message.TypeHelp,
			"daemon",
			liaisons[0].ID,
			a.CurrentTask,
			map[string]any{
				"worker_id":        a.ID,
				"escalation_level": agent.EscalationCritical,
				"message": fmt.Sprintf("worker %s appears unresponsive (heartbeat stale for %s)",
					a.ID, agent.HeartbeatStaleness(a).Round(time.Second)),
			},
		)
		if err != nil {
			d.logger.Error("liveness: notify liaison", "error", err)
		}
	} else {
		d.logger.Info("liveness: no liaison to notify", "agent", a.ID)
	}

	*tickEvents = append(*tickEvents, events.Event{
		Timestamp: time.Now(),
		Type:      events.AgentCritical,
		AgentID:   a.ID,
		TaskID:    a.CurrentTask,
		Data: map[string]any{
			"escalation_level":  agent.EscalationCritical,
			"staleness_seconds": int(agent.HeartbeatStaleness(a).Seconds()),
		},
	})
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
	t.Branch = ""
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
		d.logger.Error("progress: list workers", "error", err)
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
			// Throttle stall notifications: skip if already notified within threshold.
			if !w.LastStallNotified.IsZero() && time.Since(w.LastStallNotified) < StalledThreshold {
				continue
			}

			d.logger.Info("progress: worker stalled", "agent", w.ID, "last_commit_ago", time.Since(lastCommitTime).Round(time.Second))

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
				d.logger.Info("progress: no liaison to notify", "agent", w.ID)
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
				d.logger.Error("progress: send help message", "error", err)
			}

			// Update LastStallNotified and persist.
			w.LastStallNotified = time.Now()
			if err := d.agents.Update(w); err != nil {
				d.logger.Error("progress: update agent stall time", "agent", w.ID, "error", err)
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
		d.logger.Error("assign: find ready tasks", "error", err)
		return
	}

	for _, t := range ready {
		ok, reason := d.checker.CanSpawnWorker()
		if !ok {
			d.logger.Info("assign: cannot spawn worker", "reason", reason)
			break // constraints apply globally, no point continuing
		}

		agentID, err := d.spawnWorker(t)
		if err != nil {
			d.logger.Error("assign: spawn worker", "task", t.ID, "error", err)
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
// git branch and worktree, writes task.json / CLAUDE.md / .claude/settings.json,
// starts Claude Code in a tmux session, assigns the task, and registers the agent.
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

	// Helper to clean up git resources on failure.
	cleanupGit := func() {
		git.DeleteWorktree(d.rootDir, worktreePath)
		git.DeleteBranch(d.rootDir, branchName)
	}

	// Write task.json to worktree root.
	if err := writeWorkerTaskJSON(worktreePath, t); err != nil {
		cleanupGit()
		return "", fmt.Errorf("write task.json: %w", err)
	}

	// Write CLAUDE.md with worker system prompt.
	if err := writeWorkerClaudeMD(worktreePath, t, agentID); err != nil {
		cleanupGit()
		return "", fmt.Errorf("write CLAUDE.md: %w", err)
	}

	// Write .claude/settings.json with heartbeat hooks.
	if err := writeWorkerClaudeSettings(worktreePath, agentID); err != nil {
		cleanupGit()
		return "", fmt.Errorf("write .claude/settings.json: %w", err)
	}

	// Create tmux session.
	if err := tmux.CreateSession(sessionName); err != nil {
		cleanupGit()
		return "", fmt.Errorf("create tmux session: %w", err)
	}

	// Start Claude Code in the tmux session with the task as the initial prompt.
	// The positional argument starts an interactive session with that first message,
	// so Claude begins working immediately instead of sitting idle.
	initialPrompt := fmt.Sprintf(
		"Read CLAUDE.md and task.json, then implement the task. When finished, run: alt task-done %s %s",
		t.ID, agentID,
	)
	claudeCmd := fmt.Sprintf("cd %s && claude --dangerously-skip-permissions %q", worktreePath, initialPrompt)
	if err := tmux.SendKeys(sessionName, claudeCmd); err != nil {
		tmux.KillSession(sessionName)
		cleanupGit()
		return "", fmt.Errorf("start claude in worker: %w", err)
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
		cleanupGit()
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
		cleanupGit()
		return "", fmt.Errorf("assign task: %w", err)
	}

	d.logger.Info("assign: spawned worker", "agent", agentID, "task", t.ID)
	return agentID, nil
}

// --- Step 4: ProcessMessages ---

// processMessages reads pending messages addressed to the daemon and
// dispatches them by type.
func (d *Daemon) processMessages(tickEvents *[]events.Event) {
	msgs, err := d.messages.ListPending("daemon")
	if err != nil {
		d.logger.Error("messages: list pending", "error", err)
		return
	}

	for _, msg := range msgs {
		switch msg.Type {
		case message.TypeTaskDone:
			d.handleTaskDone(msg, tickEvents)
		case message.TypeHelp:
			d.handleHelp(msg)
		default:
			d.logger.Info("messages: unhandled type", "type", msg.Type, "from", msg.From)
		}

		// Archive processed message.
		if err := d.messages.Archive(msg.ID); err != nil {
			d.logger.Error("messages: archive", "message", msg.ID, "error", err)
		}
	}
}

// handleTaskDone processes a task_done message by marking the task as done
// and adding it to the merge queue.
func (d *Daemon) handleTaskDone(msg *message.Message, tickEvents *[]events.Event) {
	taskID := msg.TaskID
	if taskID == "" {
		d.logger.Info("messages: task_done without task_id", "from", msg.From)
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
		d.logger.Error("messages: mark task done", "task", taskID, "error", err)
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
		d.logger.Error("messages: get task for merge queue", "task", taskID, "error", err)
		return
	}
	if err := d.addToMergeQueue(t); err != nil {
		d.logger.Error("messages: add task to merge queue", "task", taskID, "error", err)
	}
}

// handleHelp forwards a help message to the first available liaison.
func (d *Daemon) handleHelp(msg *message.Message) {
	liaisons, err := d.agents.ListByRole(agent.RoleLiaison)
	if err != nil || len(liaisons) == 0 {
		d.logger.Info("messages: no liaison available", "from", msg.From)
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
		d.logger.Error("messages: forward help to liaison", "error", err)
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
		d.logger.Error("merge: read queue dir", "error", err)
		return
	}

	// Process items in filename order (FIFO by timestamp prefix).
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" || strings.HasPrefix(e.Name(), ".tmp-") {
			continue
		}

		// Check for shutdown between merge attempts.
		select {
		case <-d.shutdown:
			d.logger.Info("merge: shutdown during queue processing")
			return
		default:
		}

		itemPath := filepath.Join(queueDir, e.Name())
		data, err := os.ReadFile(itemPath)
		if err != nil {
			d.logger.Error("merge: read queue item", "file", e.Name(), "error", err)
			continue
		}

		var item MergeItem
		if err := json.Unmarshal(data, &item); err != nil {
			d.logger.Error("merge: parse queue item", "file", e.Name(), "error", err)
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
			d.logger.Error("merge: merge branch", "branch", item.Branch, "error", err)
			*tickEvents = append(*tickEvents, events.Event{
				Timestamp: time.Now(),
				Type:      events.MergeFailed,
				AgentID:   item.AgentID,
				TaskID:    item.TaskID,
				Data:      map[string]any{"error": err.Error()},
			})
			// Remove failed item to prevent infinite retry.
			os.Remove(itemPath)
			continue
		}

		if !result.Clean {
			d.logger.Info("merge: conflict", "branch", item.Branch, "conflicts", result.Conflicts)
			git.AbortMerge(d.rootDir)

			// Extract structured conflict info for each file.
			conflicts := make([]merge.ConflictInfo, 0, len(result.Conflicts))
			for _, path := range result.Conflicts {
				fullPath := filepath.Join(d.rootDir, path)
				info := merge.ExtractConflicts(fullPath)
				info.Path = path
				conflicts = append(conflicts, info)
			}

			*tickEvents = append(*tickEvents, events.Event{
				Timestamp: time.Now(),
				Type:      events.MergeConflict,
				AgentID:   item.AgentID,
				TaskID:    item.TaskID,
				Data:      map[string]any{"conflicts": result.Conflicts},
			})

			// Build conflict context and spawn a resolver agent.
			ctx := d.buildConflictContext(item, conflicts)
			resolverAgent, err := d.resolverMgr.SpawnResolver(ctx)
			if err != nil {
				d.logger.Error("merge: spawn resolver", "task", item.TaskID, "error", err)
				// Fall back to notifying the original agent.
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
			} else {
				d.logger.Info("merge: spawned resolver", "resolver", resolverAgent.ID, "task", item.TaskID)
			}

			// Remove conflicting item from queue.
			os.Remove(itemPath)
			continue
		}

		d.logger.Info("merge: success", "branch", item.Branch)
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

// buildConflictContext creates a resolver.ConflictContext from a merge item
// and the extracted conflict info.
func (d *Daemon) buildConflictContext(item MergeItem, conflicts []merge.ConflictInfo) resolver.ConflictContext {
	ctx := resolver.ConflictContext{
		TaskID:    item.TaskID,
		Branch:    item.Branch,
		Conflicts: conflicts,
	}

	// Look up the task for rig name and description.
	t, err := d.tasks.Get(item.TaskID)
	if err == nil {
		ctx.RigName = t.Rig
		ctx.TaskDescription = t.Description
	}

	return ctx
}

// --- Step 5b: CheckResolvers ---

// checkResolvers checks active resolver agents for completed resolutions.
// When a resolver has finished (no conflict markers, clean tree), it is
// cleaned up and the task is re-queued for merge.
func (d *Daemon) checkResolvers(tickEvents *[]events.Event) {
	resolvers, err := d.resolverMgr.ListResolvers()
	if err != nil {
		d.logger.Error("resolvers: list", "error", err)
		return
	}

	for _, r := range resolvers {
		if r.Status != agent.StatusActive {
			continue
		}

		// Load the conflict context to know which files to check.
		conflicts, err := d.loadConflictContext(r)
		if err != nil {
			d.logger.Error("resolvers: load conflict context", "resolver", r.ID, "error", err)
			continue
		}

		resolved, err := resolver.DetectResolution(r, conflicts)
		if err != nil {
			d.logger.Error("resolvers: detect resolution", "resolver", r.ID, "error", err)
			continue
		}

		if !resolved {
			continue
		}

		d.logger.Info("resolvers: resolution detected", "resolver", r.ID, "task", r.CurrentTask)

		// Clean up the resolver agent.
		if err := d.resolverMgr.CleanupResolver(r); err != nil {
			d.logger.Error("resolvers: cleanup", "resolver", r.ID, "error", err)
			continue
		}

		// Re-queue the task for merge.
		t, err := d.tasks.Get(r.CurrentTask)
		if err != nil {
			d.logger.Error("resolvers: get task for re-queue", "task", r.CurrentTask, "error", err)
			continue
		}
		if err := d.addToMergeQueue(t); err != nil {
			d.logger.Error("resolvers: re-queue task", "task", r.CurrentTask, "error", err)
			continue
		}

		d.logger.Info("resolvers: re-queued task for merge", "task", r.CurrentTask)
	}
}

// loadConflictContext reads the conflict-context.json from a resolver's
// worktree and returns the conflict info needed for resolution detection.
func (d *Daemon) loadConflictContext(r *agent.Agent) ([]merge.ConflictInfo, error) {
	ctxPath := filepath.Join(r.Worktree, "conflict-context.json")
	data, err := os.ReadFile(ctxPath)
	if err != nil {
		return nil, fmt.Errorf("read conflict-context.json: %w", err)
	}
	var ctx resolver.ConflictContext
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil, fmt.Errorf("parse conflict-context.json: %w", err)
	}
	return ctx.Conflicts, nil
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
		d.logger.Error("constraints: budget check", "error", err)
	} else if !ok {
		d.logger.Info("constraints: budget exceeded", "reason", reason)
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
		d.logger.Error("events: append", "error", err)
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
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		d.logger.Info("received signal, shutting down gracefully", "signal", sig)
		d.Stop()

		// Second signal forces immediate exit.
		sig = <-sigCh
		d.logger.Error("received second signal, forcing exit", "signal", sig)
		os.Exit(1)
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

// --- Worker file generation ---

// writeWorkerTaskJSON writes task details as JSON to {worktree}/task.json.
func writeWorkerTaskJSON(worktreePath string, t *task.Task) error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal task: %w", err)
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(worktreePath, "task.json"), data, 0o644)
}

// writeWorkerClaudeMD writes CLAUDE.md with a minimal worker bootstrap prompt.
func writeWorkerClaudeMD(worktreePath string, t *task.Task, agentID string) error {
	prompt := fmt.Sprintf(`# Worker Agent: %s

- **Task ID**: %s
- **Title**: %s

Run `+"`alt help worker startup`"+` for full instructions.
`, agentID, t.ID, t.Title)

	return os.WriteFile(filepath.Join(worktreePath, "CLAUDE.md"), []byte(prompt), 0o644)
}

// writeWorkerClaudeSettings writes .claude/settings.json with heartbeat hooks.
func writeWorkerClaudeSettings(worktreePath, agentID string) error {
	claudeDir := filepath.Join(worktreePath, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		return fmt.Errorf("create .claude dir: %w", err)
	}

	type hookCmd struct {
		Type    string `json:"type"`
		Command string `json:"command"`
	}
	type hookGroup struct {
		Matcher string    `json:"matcher"`
		Hooks   []hookCmd `json:"hooks"`
	}
	type claudeSettings struct {
		Hooks map[string][]hookGroup `json:"hooks"`
	}

	settings := claudeSettings{
		Hooks: map[string][]hookGroup{
			"PreToolUse": {
				{
					Matcher: "",
					Hooks:   []hookCmd{{Type: "command", Command: fmt.Sprintf("alt heartbeat %s", agentID)}},
				},
			},
			"Stop": {
				{
					Matcher: "",
					Hooks:   []hookCmd{{Type: "command", Command: fmt.Sprintf("alt checkpoint %s", agentID)}},
				},
			},
		},
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0o644)
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
