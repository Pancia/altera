package daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/anthropics/altera/internal/agent"
	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/events"
	"github.com/anthropics/altera/internal/message"
	"github.com/anthropics/altera/internal/task"
)

// setupE2EProject creates a full project directory backed by a real git
// repo. The repo has an initial commit so branches can be created from it.
// Returns the project root (which is also the git repo root).
func setupE2EProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	// Initialize a real git repo at root.
	gitCmd(t, root, "init")
	gitCmd(t, root, "config", "user.name", "test")
	gitCmd(t, root, "config", "user.email", "test@test.local")

	// Create a small Go program as test rig content.
	writeTestFile(t, root, "main.go", `package main

import "fmt"

func main() {
	fmt.Println("hello from test rig")
}
`)
	writeTestFile(t, root, "go.mod", "module testrig\n\ngo 1.25\n")

	gitCmd(t, root, "add", "-A")
	gitCmd(t, root, "commit", "-m", "initial commit")

	// Create .alt/ structure.
	altDir := filepath.Join(root, ".alt")
	for _, sub := range []string{
		"agents", "tasks", "messages", "messages/archive",
		"merge-queue", "rigs", "worktrees",
	} {
		if err := os.MkdirAll(filepath.Join(altDir, sub), 0o755); err != nil {
			t.Fatalf("setup: mkdir %s: %v", sub, err)
		}
	}

	// Write default config with known constraints.
	cfg := config.NewConfig()
	data, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(filepath.Join(altDir, "config.json"), data, 0o644); err != nil {
		t.Fatalf("setup: write config: %v", err)
	}

	return root
}

// gitCmd runs a git command in the given directory.
func gitCmd(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %s: %v", args, out, err)
	}
	return string(out)
}

// writeTestFile creates or overwrites a file relative to dir.
func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// createFeatureBranch creates a branch from the current HEAD with the given
// file changes, then switches back to the original branch.
func createFeatureBranch(t *testing.T, repo, branch string, files map[string]string) {
	t.Helper()
	original := gitCmd(t, repo, "rev-parse", "--abbrev-ref", "HEAD")
	// Trim newline.
	for len(original) > 0 && (original[len(original)-1] == '\n' || original[len(original)-1] == '\r') {
		original = original[:len(original)-1]
	}

	gitCmd(t, repo, "checkout", "-b", branch)
	for name, content := range files {
		writeTestFile(t, repo, name, content)
	}
	gitCmd(t, repo, "add", "-A")
	gitCmd(t, repo, "commit", "-m", "changes on "+branch)
	gitCmd(t, repo, "checkout", original)
}

// simulateWorker creates a branch with commits, an agent record, and a task
// assignment — simulating what the daemon's assignTasks + a real worker
// would produce, without needing tmux.
func simulateWorker(t *testing.T, d *Daemon, taskID, branchName, agentID string, files map[string]string) {
	t.Helper()
	createFeatureBranch(t, d.rootDir, branchName, files)

	// Create agent record.
	a := &agent.Agent{
		ID:          agentID,
		Role:        agent.RoleWorker,
		Status:      agent.StatusActive,
		CurrentTask: taskID,
		PID:         os.Getpid(), // Our PID so liveness checks pass.
		Heartbeat:   time.Now(),
		StartedAt:   time.Now(),
	}
	if err := d.agents.Create(a); err != nil {
		t.Fatalf("create agent %s: %v", agentID, err)
	}

	// Update task to assigned+in_progress with branch and agent.
	tk, err := d.tasks.Get(taskID)
	if err != nil {
		t.Fatalf("get task %s: %v", taskID, err)
	}
	tk.Status = task.StatusInProgress
	tk.AssignedTo = agentID
	tk.Branch = branchName
	if err := d.tasks.ForceWrite(tk); err != nil {
		t.Fatalf("force write task %s: %v", taskID, err)
	}
}

// eventsOfType filters events by type.
func eventsOfType(evts []events.Event, typ events.Type) []events.Event {
	var out []events.Event
	for _, e := range evts {
		if e.Type == typ {
			out = append(out, e)
		}
	}
	return out
}

// readAllEvents reads all events from the daemon's event log.
func readAllEvents(t *testing.T, d *Daemon) []events.Event {
	t.Helper()
	evts, err := d.evReader.ReadAll()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read events: %v", err)
	}
	return evts
}

// --- Test 1: Full Pipeline ---
//
// Task created → worker assigned (simulated) → worker commits → task_done
// message → daemon processes message (queues merge) → daemon merges → done.

func TestE2E_FullPipeline(t *testing.T) {
	root := setupE2EProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Mock gitLogTimestamp so checkProgress doesn't fail.
	origGitLog := gitLogTimestamp
	defer func() { gitLogTimestamp = origGitLog }()
	gitLogTimestamp = func(worktree string) (string, error) {
		return fmt.Sprintf("%d", time.Now().Unix()), nil
	}

	// Step 1: Create a task.
	tk := &task.Task{
		ID:          "t-pipe01",
		Title:       "Add greeting feature",
		Description: "Add a greeting function to the test rig",
		Status:      task.StatusOpen,
		Rig:         "test-rig",
	}
	if err := d.tasks.Create(tk); err != nil {
		t.Fatalf("create task: %v", err)
	}

	// Step 2: Simulate a worker that implements the feature.
	simulateWorker(t, d, "t-pipe01", "worker/w-pipe01", "w-pipe01", map[string]string{
		"greet.go": "package main\n\nfunc greet(name string) string {\n\treturn \"Hello, \" + name\n}\n",
	})

	// Step 3: Worker signals completion via task_done message.
	_, err = d.messages.Create(
		message.TypeTaskDone,
		"w-pipe01",
		"daemon",
		"t-pipe01",
		map[string]any{"result": "implemented greeting feature"},
	)
	if err != nil {
		t.Fatalf("create task_done message: %v", err)
	}

	// Step 4: Daemon processes messages → task marked done, queued for merge.
	var tickEvents []events.Event
	d.processMessages(&tickEvents)

	updatedTask, err := d.tasks.Get("t-pipe01")
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if updatedTask.Status != task.StatusDone {
		t.Errorf("task status = %q, want %q", updatedTask.Status, task.StatusDone)
	}
	if updatedTask.Result != "implemented greeting feature" {
		t.Errorf("task result = %q, want %q", updatedTask.Result, "implemented greeting feature")
	}

	// Verify TaskDone event was emitted.
	doneEvents := eventsOfType(tickEvents, events.TaskDone)
	if len(doneEvents) != 1 {
		t.Fatalf("expected 1 TaskDone event, got %d", len(doneEvents))
	}

	// Verify merge queue has an entry.
	queueDir := filepath.Join(d.altDir, "merge-queue")
	entries, _ := os.ReadDir(queueDir)
	queueCount := 0
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".json" {
			queueCount++
		}
	}
	if queueCount != 1 {
		t.Fatalf("merge queue items = %d, want 1", queueCount)
	}

	// Step 5: Daemon processes merge queue → branch merged into main.
	tickEvents = nil
	d.processMergeQueue(&tickEvents)

	// Verify merge events: should have MergeStarted + MergeSuccess.
	startedEvents := eventsOfType(tickEvents, events.MergeStarted)
	successEvents := eventsOfType(tickEvents, events.MergeSuccess)
	if len(startedEvents) != 1 {
		t.Errorf("expected 1 MergeStarted event, got %d", len(startedEvents))
	}
	if len(successEvents) != 1 {
		t.Errorf("expected 1 MergeSuccess event, got %d", len(successEvents))
	}

	// Verify merge queue is now empty.
	entries, _ = os.ReadDir(queueDir)
	queueCount = 0
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".json" {
			queueCount++
		}
	}
	if queueCount != 0 {
		t.Errorf("merge queue items after merge = %d, want 0", queueCount)
	}

	// Verify the file exists on main (the merge actually happened in git).
	greetPath := filepath.Join(root, "greet.go")
	if _, err := os.Stat(greetPath); err != nil {
		t.Errorf("greet.go not found on main after merge: %v", err)
	}

	// Verify merge result message was sent to the worker.
	msgs, err := d.messages.ListPending("w-pipe01")
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 merge_result message, got %d", len(msgs))
	}
	if msgs[0].Type != message.TypeMergeResult {
		t.Errorf("message type = %q, want %q", msgs[0].Type, message.TypeMergeResult)
	}
	if msgs[0].Payload["success"] != true {
		t.Errorf("merge result success = %v, want true", msgs[0].Payload["success"])
	}
}

// --- Test 2: Multi-Worker Parallel ---
//
// Three tasks run in parallel on separate branches modifying different files.
// All merge successfully with no conflicts.

func TestE2E_MultiWorkerParallel(t *testing.T) {
	root := setupE2EProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	origGitLog := gitLogTimestamp
	defer func() { gitLogTimestamp = origGitLog }()
	gitLogTimestamp = func(worktree string) (string, error) {
		return fmt.Sprintf("%d", time.Now().Unix()), nil
	}

	// Create 3 tasks + simulate 3 workers, each modifying different files.
	workers := []struct {
		taskID   string
		agentID  string
		branch   string
		filename string
		content  string
	}{
		{"t-par001", "w-par001", "worker/w-par001", "feature_a.go", "package main\n\nfunc featureA() {}\n"},
		{"t-par002", "w-par002", "worker/w-par002", "feature_b.go", "package main\n\nfunc featureB() {}\n"},
		{"t-par003", "w-par003", "worker/w-par003", "feature_c.go", "package main\n\nfunc featureC() {}\n"},
	}

	for _, w := range workers {
		tk := &task.Task{
			ID:     w.taskID,
			Title:  "Task " + w.taskID,
			Status: task.StatusOpen,
		}
		if err := d.tasks.Create(tk); err != nil {
			t.Fatalf("create task %s: %v", w.taskID, err)
		}
		simulateWorker(t, d, w.taskID, w.branch, w.agentID, map[string]string{
			w.filename: w.content,
		})
	}

	// All 3 workers signal completion.
	for _, w := range workers {
		_, err := d.messages.Create(
			message.TypeTaskDone, w.agentID, "daemon", w.taskID,
			map[string]any{"result": "done"},
		)
		if err != nil {
			t.Fatalf("create task_done for %s: %v", w.taskID, err)
		}
	}

	// Daemon processes messages (queues all 3 for merge).
	var tickEvents []events.Event
	d.processMessages(&tickEvents)

	// Verify all 3 tasks are done and queued.
	doneEvents := eventsOfType(tickEvents, events.TaskDone)
	if len(doneEvents) != 3 {
		t.Errorf("expected 3 TaskDone events, got %d", len(doneEvents))
	}

	// Daemon processes merge queue (should merge all 3 sequentially).
	tickEvents = nil
	d.processMergeQueue(&tickEvents)

	successEvents := eventsOfType(tickEvents, events.MergeSuccess)
	if len(successEvents) != 3 {
		t.Errorf("expected 3 MergeSuccess events, got %d", len(successEvents))
	}

	// Verify all 3 files exist on main.
	for _, w := range workers {
		path := filepath.Join(root, w.filename)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("%s not found on main after merge: %v", w.filename, err)
		}
	}

	// Verify merge queue is empty.
	queueDir := filepath.Join(d.altDir, "merge-queue")
	entries, _ := os.ReadDir(queueDir)
	count := 0
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".json" {
			count++
		}
	}
	if count != 0 {
		t.Errorf("merge queue items after all merges = %d, want 0", count)
	}
}

// --- Test 3: Dependency Tracking ---
//
// Task B depends on Task A. Daemon should only assign A initially, then
// assign B after A completes.

func TestE2E_DependencyTracking(t *testing.T) {
	root := setupE2EProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Create Task A (no deps).
	taskA := &task.Task{
		ID:     "t-depA01",
		Title:  "Task A - base implementation",
		Status: task.StatusOpen,
	}
	if err := d.tasks.Create(taskA); err != nil {
		t.Fatalf("create task A: %v", err)
	}

	// Create Task B (depends on A).
	taskB := &task.Task{
		ID:     "t-depB01",
		Title:  "Task B - depends on A",
		Status: task.StatusOpen,
		Deps:   []string{"t-depA01"},
	}
	if err := d.tasks.Create(taskB); err != nil {
		t.Fatalf("create task B: %v", err)
	}

	// FindReady should only return Task A.
	ready, err := d.tasks.FindReady()
	if err != nil {
		t.Fatalf("FindReady: %v", err)
	}
	if len(ready) != 1 {
		t.Fatalf("ready tasks = %d, want 1 (only A)", len(ready))
	}
	if ready[0].ID != "t-depA01" {
		t.Errorf("ready task = %q, want %q", ready[0].ID, "t-depA01")
	}

	// Complete Task A.
	taskA.Status = task.StatusDone
	taskA.UpdatedAt = time.Now().UTC()
	if err := d.tasks.ForceWrite(taskA); err != nil {
		t.Fatalf("force write task A: %v", err)
	}

	// Now FindReady should return Task B.
	ready, err = d.tasks.FindReady()
	if err != nil {
		t.Fatalf("FindReady after A done: %v", err)
	}
	if len(ready) != 1 {
		t.Fatalf("ready tasks after A done = %d, want 1 (only B)", len(ready))
	}
	if ready[0].ID != "t-depB01" {
		t.Errorf("ready task after A done = %q, want %q", ready[0].ID, "t-depB01")
	}
}

// --- Test 4: Worker Crash Recovery ---
//
// A worker dies (process gone, stale heartbeat). The daemon detects this via
// checkAgentLiveness, marks the agent dead, and reclaims the task so a new
// worker can be spawned.

func TestE2E_WorkerCrashRecovery(t *testing.T) {
	root := setupE2EProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Create a task assigned to a worker.
	tk := &task.Task{
		ID:         "t-crash1",
		Title:      "Task that will be reclaimed",
		Status:     task.StatusInProgress,
		AssignedTo: "w-crash1",
		Branch:     "worker/w-crash1",
	}
	if err := d.tasks.Create(tk); err != nil {
		t.Fatalf("create task: %v", err)
	}
	// ForceWrite to bypass transition validation (we set it directly to in_progress).
	if err := d.tasks.ForceWrite(tk); err != nil {
		t.Fatalf("force write task: %v", err)
	}

	// Create a "dead" worker: non-existent PID + stale heartbeat.
	deadAgent := &agent.Agent{
		ID:          "w-crash1",
		Role:        agent.RoleWorker,
		Status:      agent.StatusActive,
		CurrentTask: "t-crash1",
		PID:         9999999, // Non-existent process.
		Heartbeat:   time.Now().Add(-5 * time.Minute),
		StartedAt:   time.Now().Add(-10 * time.Minute),
	}
	if err := d.agents.Create(deadAgent); err != nil {
		t.Fatalf("create dead agent: %v", err)
	}

	// Run liveness check.
	var tickEvents []events.Event
	d.checkAgentLiveness(&tickEvents)

	// Agent should be marked dead.
	updated, err := d.agents.Get("w-crash1")
	if err != nil {
		t.Fatalf("get agent: %v", err)
	}
	if updated.Status != agent.StatusDead {
		t.Errorf("agent status = %q, want %q", updated.Status, agent.StatusDead)
	}

	// Task should be reclaimed (back to open, unassigned).
	updatedTask, err := d.tasks.Get("t-crash1")
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if updatedTask.Status != task.StatusOpen {
		t.Errorf("task status = %q, want %q", updatedTask.Status, task.StatusOpen)
	}
	if updatedTask.AssignedTo != "" {
		t.Errorf("task assigned_to = %q, want empty", updatedTask.AssignedTo)
	}

	// Should have AgentDied event.
	diedEvents := eventsOfType(tickEvents, events.AgentDied)
	if len(diedEvents) != 1 {
		t.Fatalf("expected 1 AgentDied event, got %d", len(diedEvents))
	}
	if diedEvents[0].AgentID != "w-crash1" {
		t.Errorf("died event agent_id = %q, want %q", diedEvents[0].AgentID, "w-crash1")
	}

	// Now the task is open again — FindReady should return it.
	ready, err := d.tasks.FindReady()
	if err != nil {
		t.Fatalf("FindReady: %v", err)
	}
	found := false
	for _, r := range ready {
		if r.ID == "t-crash1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("reclaimed task t-crash1 not found in ready tasks")
	}
}

// --- Test 5: Merge Conflict Scenario ---
//
// Two workers modify the same file. The first merge succeeds. The second
// merge detects a conflict, aborts, and emits a conflict event.

func TestE2E_MergeConflict(t *testing.T) {
	root := setupE2EProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	origGitLog := gitLogTimestamp
	defer func() { gitLogTimestamp = origGitLog }()
	gitLogTimestamp = func(worktree string) (string, error) {
		return fmt.Sprintf("%d", time.Now().Unix()), nil
	}

	// Worker 1: modifies main.go one way.
	tk1 := &task.Task{ID: "t-conf01", Title: "Worker 1 changes", Status: task.StatusOpen}
	if err := d.tasks.Create(tk1); err != nil {
		t.Fatalf("create task 1: %v", err)
	}
	simulateWorker(t, d, "t-conf01", "worker/w-conf01", "w-conf01", map[string]string{
		"main.go": "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"version A\")\n}\n",
	})

	// Worker 2: modifies main.go a different way.
	tk2 := &task.Task{ID: "t-conf02", Title: "Worker 2 changes", Status: task.StatusOpen}
	if err := d.tasks.Create(tk2); err != nil {
		t.Fatalf("create task 2: %v", err)
	}
	simulateWorker(t, d, "t-conf02", "worker/w-conf02", "w-conf02", map[string]string{
		"main.go": "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"version B\")\n}\n",
	})

	// Both signal completion.
	for _, w := range []struct{ agent, task string }{
		{"w-conf01", "t-conf01"},
		{"w-conf02", "t-conf02"},
	} {
		_, err := d.messages.Create(
			message.TypeTaskDone, w.agent, "daemon", w.task,
			map[string]any{"result": "done"},
		)
		if err != nil {
			t.Fatalf("create task_done for %s: %v", w.task, err)
		}
	}

	// Daemon processes messages → both queued for merge.
	var tickEvents []events.Event
	d.processMessages(&tickEvents)

	// Daemon processes merge queue.
	tickEvents = nil
	d.processMergeQueue(&tickEvents)

	// First merge should succeed, second should conflict.
	successEvents := eventsOfType(tickEvents, events.MergeSuccess)
	conflictEvents := eventsOfType(tickEvents, events.MergeConflict)

	if len(successEvents) != 1 {
		t.Errorf("expected 1 MergeSuccess event, got %d", len(successEvents))
	}
	if len(conflictEvents) != 1 {
		t.Errorf("expected 1 MergeConflict event, got %d", len(conflictEvents))
	}

	// The repo should still be in a clean state (abort was called).
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = root
	statusOut, err := cmd.Output()
	if err != nil {
		t.Fatalf("git status: %v", err)
	}
	if len(statusOut) != 0 {
		t.Errorf("repo is dirty after conflict handling: %s", statusOut)
	}
}

// --- Test 6: Budget Enforcement ---
//
// Set a low budget ceiling. Write events with token_cost that exceed it.
// The daemon's checkConstraints should emit a BudgetExceeded event, and
// CanSpawnWorker should return false.

func TestE2E_BudgetEnforcement(t *testing.T) {
	root := setupE2EProject(t)

	// Override config with a very low budget.
	altDir := filepath.Join(root, ".alt")
	cfg := config.Config{
		Rigs: make(map[string]config.RigConfig),
		Constraints: config.Constraints{
			BudgetCeiling: 5.0, // Very low.
			MaxWorkers:    4,
			MaxQueueDepth: 10,
		},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(filepath.Join(altDir, "config.json"), data, 0o644)

	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Write events that exceed the budget (6 events × 1.0 = 6.0 > 5.0).
	for i := 0; i < 6; i++ {
		err := d.events.Append(events.Event{
			Timestamp: time.Now(),
			Type:      events.TaskDone,
			AgentID:   fmt.Sprintf("w-%d", i),
			TaskID:    fmt.Sprintf("t-%d", i),
			Data:      map[string]any{"token_cost": 1.0},
		})
		if err != nil {
			t.Fatalf("append event: %v", err)
		}
	}

	// checkConstraints should emit BudgetExceeded.
	var tickEvents []events.Event
	d.checkConstraints(&tickEvents)

	budgetEvents := eventsOfType(tickEvents, events.BudgetExceeded)
	if len(budgetEvents) != 1 {
		t.Fatalf("expected 1 BudgetExceeded event, got %d", len(budgetEvents))
	}

	// CanSpawnWorker should return false.
	ok, reason := d.checker.CanSpawnWorker()
	if ok {
		t.Error("CanSpawnWorker should return false when budget exceeded")
	}
	if reason == "" {
		t.Error("expected a reason string when budget exceeded")
	}
}

// --- Test 7: Stalled Worker Detection ---
//
// A worker has been active for a long time with no recent commits.
// checkProgress detects this and sends a help message to the liaison.

func TestE2E_StalledWorkerDetection(t *testing.T) {
	root := setupE2EProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	worktreeDir := t.TempDir()

	// Create a worker agent with a worktree.
	stalledWorker := &agent.Agent{
		ID:          "w-stall1",
		Role:        agent.RoleWorker,
		Status:      agent.StatusActive,
		CurrentTask: "t-stall1",
		Worktree:    worktreeDir,
		PID:         os.Getpid(),
		Heartbeat:   time.Now(),
		StartedAt:   time.Now().Add(-2 * time.Hour),
	}
	if err := d.agents.Create(stalledWorker); err != nil {
		t.Fatalf("create stalled worker: %v", err)
	}

	// Create a liaison to receive the stall notification.
	liaison := &agent.Agent{
		ID:        "liaison-01",
		Role:      agent.RoleLiaison,
		Status:    agent.StatusActive,
		PID:       os.Getpid(),
		Heartbeat: time.Now(),
		StartedAt: time.Now(),
	}
	if err := d.agents.Create(liaison); err != nil {
		t.Fatalf("create liaison: %v", err)
	}

	// Mock gitLogTimestamp to return a stale time (45 minutes ago,
	// well beyond the 30-minute StalledThreshold).
	origGitLog := gitLogTimestamp
	defer func() { gitLogTimestamp = origGitLog }()
	stalledTime := time.Now().Add(-45 * time.Minute)
	gitLogTimestamp = func(worktree string) (string, error) {
		return fmt.Sprintf("%d", stalledTime.Unix()), nil
	}

	// Run progress check.
	var tickEvents []events.Event
	d.checkProgress(&tickEvents)

	// Should have WorkerStalled event.
	stalledEvents := eventsOfType(tickEvents, events.WorkerStalled)
	if len(stalledEvents) != 1 {
		t.Fatalf("expected 1 WorkerStalled event, got %d", len(stalledEvents))
	}
	if stalledEvents[0].AgentID != "w-stall1" {
		t.Errorf("stalled event agent_id = %q, want %q", stalledEvents[0].AgentID, "w-stall1")
	}

	// Liaison should have received a help message.
	msgs, err := d.messages.ListPending("liaison-01")
	if err != nil {
		t.Fatalf("list pending for liaison: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 help message to liaison, got %d", len(msgs))
	}
	if msgs[0].Type != message.TypeHelp {
		t.Errorf("message type = %q, want %q", msgs[0].Type, message.TypeHelp)
	}
	// The message payload should identify the stalled worker.
	if msgs[0].Payload["worker_id"] != "w-stall1" {
		t.Errorf("help message worker_id = %v, want %q", msgs[0].Payload["worker_id"], "w-stall1")
	}
}

// --- Test: Full Tick Integration ---
//
// Run a complete daemon tick with real git operations to verify all seven
// steps work together without interfering.

func TestE2E_FullTick(t *testing.T) {
	root := setupE2EProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Mock gitLogTimestamp so checkProgress doesn't shell out.
	origGitLog := gitLogTimestamp
	defer func() { gitLogTimestamp = origGitLog }()
	gitLogTimestamp = func(worktree string) (string, error) {
		return fmt.Sprintf("%d", time.Now().Unix()), nil
	}

	// Set up scenario: one completed task ready to merge, one open task
	// with a dependency, and a dead agent to reclaim.

	// Task 1: completed, ready to merge.
	tk1 := &task.Task{
		ID:         "t-tick01",
		Title:      "Completed task",
		Status:     task.StatusInProgress,
		AssignedTo: "w-tick01",
		Branch:     "worker/w-tick01",
	}
	if err := d.tasks.Create(tk1); err != nil {
		t.Fatalf("create task 1: %v", err)
	}
	if err := d.tasks.ForceWrite(tk1); err != nil {
		t.Fatalf("force write task 1: %v", err)
	}

	// Create the branch with actual changes.
	createFeatureBranch(t, root, "worker/w-tick01", map[string]string{
		"tick_feature.go": "package main\n\nfunc tickFeature() {}\n",
	})

	// Agent for task 1 (alive).
	w1 := &agent.Agent{
		ID: "w-tick01", Role: agent.RoleWorker, Status: agent.StatusActive,
		CurrentTask: "t-tick01", PID: os.Getpid(),
		Heartbeat: time.Now(), StartedAt: time.Now(),
	}
	if err := d.agents.Create(w1); err != nil {
		t.Fatalf("create agent 1: %v", err)
	}

	// Task_done message for task 1.
	_, err = d.messages.Create(
		message.TypeTaskDone, "w-tick01", "daemon", "t-tick01",
		map[string]any{"result": "done"},
	)
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	// Task 2: depends on task 1.
	tk2 := &task.Task{
		ID:    "t-tick02",
		Title: "Dependent task",
		Deps:  []string{"t-tick01"},
	}
	if err := d.tasks.Create(tk2); err != nil {
		t.Fatalf("create task 2: %v", err)
	}

	// Dead agent for cleanup.
	deadA := &agent.Agent{
		ID: "w-dead01", Role: agent.RoleWorker, Status: agent.StatusActive,
		CurrentTask: "", PID: 9999999,
		Heartbeat: time.Now().Add(-5 * time.Minute), StartedAt: time.Now().Add(-10 * time.Minute),
	}
	if err := d.agents.Create(deadA); err != nil {
		t.Fatalf("create dead agent: %v", err)
	}

	// Run a full tick.
	d.tick()

	// Verify results:
	// 1. Dead agent marked dead.
	updated, err := d.agents.Get("w-dead01")
	if err != nil {
		t.Fatalf("get dead agent: %v", err)
	}
	if updated.Status != agent.StatusDead {
		t.Errorf("dead agent status = %q, want %q", updated.Status, agent.StatusDead)
	}

	// 2. Task 1 should be done.
	updatedTk1, err := d.tasks.Get("t-tick01")
	if err != nil {
		t.Fatalf("get task 1: %v", err)
	}
	if updatedTk1.Status != task.StatusDone {
		t.Errorf("task 1 status = %q, want %q", updatedTk1.Status, task.StatusDone)
	}

	// 3. The merge should have happened (tick_feature.go on main).
	if _, err := os.Stat(filepath.Join(root, "tick_feature.go")); err != nil {
		t.Errorf("tick_feature.go not found on main after tick: %v", err)
	}

	// 4. Events should have been written.
	allEvents := readAllEvents(t, d)
	if len(allEvents) == 0 {
		t.Error("expected events to be written after tick")
	}

	// Check specific event types were emitted.
	hasAgentDied := false
	hasTaskDone := false
	hasMergeStarted := false
	for _, ev := range allEvents {
		switch ev.Type {
		case events.AgentDied:
			hasAgentDied = true
		case events.TaskDone:
			hasTaskDone = true
		case events.MergeStarted:
			hasMergeStarted = true
		}
	}
	if !hasAgentDied {
		t.Error("expected AgentDied event in tick output")
	}
	if !hasTaskDone {
		t.Error("expected TaskDone event in tick output")
	}
	if !hasMergeStarted {
		t.Error("expected MergeStarted event in tick output")
	}
}

