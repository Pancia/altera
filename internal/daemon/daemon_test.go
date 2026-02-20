package daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anthropics/altera/internal/agent"
	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/events"
	"github.com/anthropics/altera/internal/merge"
	"github.com/anthropics/altera/internal/message"
	"github.com/anthropics/altera/internal/task"
)

// setupTestProject creates a minimal project directory with .alt/ structure
// and returns the project root and cleanup function.
func setupTestProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	altDir := filepath.Join(root, ".alt")
	for _, sub := range []string{"agents", "tasks", "messages", "messages/archive", "merge-queue", "rigs"} {
		if err := os.MkdirAll(filepath.Join(altDir, sub), 0o755); err != nil {
			t.Fatalf("setup: mkdir %s: %v", sub, err)
		}
	}

	// Write default config.
	cfg := config.NewConfig()
	data, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(filepath.Join(altDir, "config.json"), data, 0o644); err != nil {
		t.Fatalf("setup: write config: %v", err)
	}

	return root
}

func TestNew(t *testing.T) {
	root := setupTestProject(t)

	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if d.altDir != filepath.Join(root, ".alt") {
		t.Errorf("altDir = %q, want %q", d.altDir, filepath.Join(root, ".alt"))
	}
	if d.rootDir != root {
		t.Errorf("rootDir = %q, want %q", d.rootDir, root)
	}
}

func TestNew_NoAltDir(t *testing.T) {
	root := t.TempDir()
	_, err := New(root)
	if err == nil {
		t.Fatal("New: expected error for missing .alt dir")
	}
}

func TestAcquireReleaseLock(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Acquire lock should succeed.
	if err := d.acquireLock(); err != nil {
		t.Fatalf("acquireLock: %v", err)
	}

	// PID file should exist with our PID.
	data, err := os.ReadFile(d.pidFile)
	if err != nil {
		t.Fatalf("read pid file: %v", err)
	}
	expectedPID := fmt.Sprintf("%d\n", os.Getpid())
	if string(data) != expectedPID {
		t.Errorf("pid file = %q, want %q", string(data), expectedPID)
	}

	// Second acquire should fail (double-start prevention).
	d2, _ := New(root)
	if err := d2.acquireLock(); err == nil {
		t.Fatal("acquireLock: expected error for double-start")
		d2.releaseLock()
	}

	// Release should remove the PID file.
	d.releaseLock()
	if _, err := os.Stat(d.pidFile); !os.IsNotExist(err) {
		t.Error("pid file should be removed after releaseLock")
	}
}

func TestReadStatus_NotRunning(t *testing.T) {
	root := setupTestProject(t)
	altDir := filepath.Join(root, ".alt")

	st := ReadStatus(altDir)
	if st.Running {
		t.Error("ReadStatus: expected not running when no pid file")
	}
}

func TestReadStatus_Running(t *testing.T) {
	root := setupTestProject(t)
	altDir := filepath.Join(root, ".alt")

	// Write our own PID to simulate a running daemon.
	pidFile := filepath.Join(altDir, "daemon.pid")
	pid := fmt.Sprintf("%d\n", os.Getpid())
	if err := os.WriteFile(pidFile, []byte(pid), 0o644); err != nil {
		t.Fatalf("write pid file: %v", err)
	}

	st := ReadStatus(altDir)
	if !st.Running {
		t.Error("ReadStatus: expected running")
	}
	if st.PID != os.Getpid() {
		t.Errorf("ReadStatus PID = %d, want %d", st.PID, os.Getpid())
	}
}

func TestReadStatus_StalePID(t *testing.T) {
	root := setupTestProject(t)
	altDir := filepath.Join(root, ".alt")

	// Write a PID that doesn't exist (use a very large PID).
	pidFile := filepath.Join(altDir, "daemon.pid")
	if err := os.WriteFile(pidFile, []byte("9999999\n"), 0o644); err != nil {
		t.Fatalf("write pid file: %v", err)
	}

	st := ReadStatus(altDir)
	if st.Running {
		t.Error("ReadStatus: expected not running for stale PID")
	}
}

func TestCheckAgentLiveness(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Create an agent with a stale heartbeat.
	a := &agent.Agent{
		ID:          "test-dead",
		Role:        agent.RoleWorker,
		Status:      agent.StatusActive,
		CurrentTask: "",
		PID:         9999999, // non-existent process
		Heartbeat:   time.Now().Add(-5 * time.Minute),
		StartedAt:   time.Now().Add(-10 * time.Minute),
	}
	if err := d.agents.Create(a); err != nil {
		t.Fatalf("create agent: %v", err)
	}

	var tickEvents []events.Event
	d.checkAgentLiveness(&tickEvents)

	// Agent should be marked dead.
	updated, err := d.agents.Get("test-dead")
	if err != nil {
		t.Fatalf("get agent: %v", err)
	}
	if updated.Status != agent.StatusDead {
		t.Errorf("agent status = %q, want %q", updated.Status, agent.StatusDead)
	}

	// Should have emitted an AgentDied event.
	if len(tickEvents) != 1 {
		t.Fatalf("tick events = %d, want 1", len(tickEvents))
	}
	if tickEvents[0].Type != events.AgentDied {
		t.Errorf("event type = %q, want %q", tickEvents[0].Type, events.AgentDied)
	}
}

func TestCheckAgentLiveness_WithTaskReclaim(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Create a task that's assigned with a branch.
	tk := &task.Task{
		Title:      "test task",
		Status:     task.StatusAssigned,
		AssignedTo: "test-dead-2",
		Branch:     "worker/test-dead-2",
	}
	if err := d.tasks.Create(tk); err != nil {
		t.Fatalf("create task: %v", err)
	}

	// Create a dead agent with that task.
	a := &agent.Agent{
		ID:          "test-dead-2",
		Role:        agent.RoleWorker,
		Status:      agent.StatusActive,
		CurrentTask: tk.ID,
		PID:         9999999,
		Heartbeat:   time.Now().Add(-5 * time.Minute),
		StartedAt:   time.Now().Add(-10 * time.Minute),
	}
	if err := d.agents.Create(a); err != nil {
		t.Fatalf("create agent: %v", err)
	}

	var tickEvents []events.Event
	d.checkAgentLiveness(&tickEvents)

	// Task should be reclaimed (status back to open).
	updatedTask, err := d.tasks.Get(tk.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if updatedTask.Status != task.StatusOpen {
		t.Errorf("task status = %q, want %q", updatedTask.Status, task.StatusOpen)
	}
	if updatedTask.AssignedTo != "" {
		t.Errorf("task assigned_to = %q, want empty", updatedTask.AssignedTo)
	}
	if updatedTask.Branch != "" {
		t.Errorf("task branch = %q, want empty", updatedTask.Branch)
	}
}

func TestCheckAgentLiveness_AliveAgent(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Create an agent with a fresh heartbeat and our own PID.
	a := &agent.Agent{
		ID:        "test-alive",
		Role:      agent.RoleWorker,
		Status:    agent.StatusActive,
		PID:       os.Getpid(),
		Heartbeat: time.Now(),
		StartedAt: time.Now(),
	}
	if err := d.agents.Create(a); err != nil {
		t.Fatalf("create agent: %v", err)
	}

	var tickEvents []events.Event
	d.checkAgentLiveness(&tickEvents)

	// Agent should remain active.
	updated, err := d.agents.Get("test-alive")
	if err != nil {
		t.Fatalf("get agent: %v", err)
	}
	if updated.Status != agent.StatusActive {
		t.Errorf("agent status = %q, want %q", updated.Status, agent.StatusActive)
	}
	if len(tickEvents) != 0 {
		t.Errorf("tick events = %d, want 0", len(tickEvents))
	}
}

func TestCheckProgress_StalledWorker(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	worktreeDir := t.TempDir()

	// Create a worker agent with a worktree.
	a := &agent.Agent{
		ID:          "stalled-worker",
		Role:        agent.RoleWorker,
		Status:      agent.StatusActive,
		CurrentTask: "t-abc123",
		Worktree:    worktreeDir,
		PID:         os.Getpid(),
		Heartbeat:   time.Now(),
		StartedAt:   time.Now().Add(-2 * time.Hour),
	}
	if err := d.agents.Create(a); err != nil {
		t.Fatalf("create agent: %v", err)
	}

	// Create a liaison to receive stall notification.
	liaison := &agent.Agent{
		ID:        "liaison-1",
		Role:      agent.RoleLiaison,
		Status:    agent.StatusActive,
		PID:       os.Getpid(),
		Heartbeat: time.Now(),
		StartedAt: time.Now(),
	}
	if err := d.agents.Create(liaison); err != nil {
		t.Fatalf("create liaison: %v", err)
	}

	// Mock gitLogTimestamp to return a stale time (2 hours ago).
	origGitLog := gitLogTimestamp
	defer func() { gitLogTimestamp = origGitLog }()
	gitLogTimestamp = func(worktree string) (string, error) {
		return fmt.Sprintf("%d", time.Now().Add(-2*time.Hour).Unix()), nil
	}

	var tickEvents []events.Event
	d.checkProgress(&tickEvents)

	// Should have emitted a WorkerStalled event.
	if len(tickEvents) != 1 {
		t.Fatalf("tick events = %d, want 1", len(tickEvents))
	}
	if tickEvents[0].Type != events.WorkerStalled {
		t.Errorf("event type = %q, want %q", tickEvents[0].Type, events.WorkerStalled)
	}
	if tickEvents[0].AgentID != "stalled-worker" {
		t.Errorf("event agent_id = %q, want %q", tickEvents[0].AgentID, "stalled-worker")
	}

	// Check that a help message was sent to the liaison.
	msgs, err := d.messages.ListPending(liaison.ID)
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("liaison messages = %d, want 1", len(msgs))
	}
	if msgs[0].Type != message.TypeHelp {
		t.Errorf("message type = %q, want %q", msgs[0].Type, message.TypeHelp)
	}
}

func TestCheckProgress_ActiveWorker(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	worktreeDir := t.TempDir()

	a := &agent.Agent{
		ID:          "active-worker",
		Role:        agent.RoleWorker,
		Status:      agent.StatusActive,
		CurrentTask: "t-abc123",
		Worktree:    worktreeDir,
		PID:         os.Getpid(),
		Heartbeat:   time.Now(),
		StartedAt:   time.Now(),
	}
	if err := d.agents.Create(a); err != nil {
		t.Fatalf("create agent: %v", err)
	}

	// Mock gitLogTimestamp to return a recent time.
	origGitLog := gitLogTimestamp
	defer func() { gitLogTimestamp = origGitLog }()
	gitLogTimestamp = func(worktree string) (string, error) {
		return fmt.Sprintf("%d", time.Now().Add(-5*time.Minute).Unix()), nil
	}

	var tickEvents []events.Event
	d.checkProgress(&tickEvents)

	// No events should be emitted for an active worker.
	if len(tickEvents) != 0 {
		t.Errorf("tick events = %d, want 0", len(tickEvents))
	}
}

func TestProcessMessages_TaskDone(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Create a task in in_progress state.
	tk := &task.Task{
		Title:      "test task",
		Status:     task.StatusInProgress,
		AssignedTo: "worker-1",
		Branch:     "worker/w-abc123",
	}
	if err := d.tasks.Create(tk); err != nil {
		t.Fatalf("create task: %v", err)
	}

	// Create a task_done message.
	_, err = d.messages.Create(
		message.TypeTaskDone,
		"worker-1",
		"daemon",
		tk.ID,
		map[string]any{"result": "completed successfully"},
	)
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	var tickEvents []events.Event
	d.processMessages(&tickEvents)

	// Task should be marked done.
	updated, err := d.tasks.Get(tk.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if updated.Status != task.StatusDone {
		t.Errorf("task status = %q, want %q", updated.Status, task.StatusDone)
	}
	if updated.Result != "completed successfully" {
		t.Errorf("task result = %q, want %q", updated.Result, "completed successfully")
	}

	// Should have a TaskDone event.
	foundDone := false
	for _, ev := range tickEvents {
		if ev.Type == events.TaskDone {
			foundDone = true
			break
		}
	}
	if !foundDone {
		t.Error("expected TaskDone event")
	}

	// Merge queue should have an item.
	queueDir := filepath.Join(d.altDir, "merge-queue")
	entries, err := os.ReadDir(queueDir)
	if err != nil {
		t.Fatalf("read queue dir: %v", err)
	}
	found := 0
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".json" {
			found++
		}
	}
	if found != 1 {
		t.Errorf("merge queue items = %d, want 1", found)
	}

	// Message should be archived.
	pending, err := d.messages.ListPending("daemon")
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(pending) != 0 {
		t.Errorf("pending daemon messages = %d, want 0", len(pending))
	}
}

func TestProcessMessages_Help(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Create a liaison.
	liaison := &agent.Agent{
		ID:        "liaison-1",
		Role:      agent.RoleLiaison,
		Status:    agent.StatusActive,
		PID:       os.Getpid(),
		Heartbeat: time.Now(),
		StartedAt: time.Now(),
	}
	if err := d.agents.Create(liaison); err != nil {
		t.Fatalf("create liaison: %v", err)
	}

	// Create a help message to daemon.
	_, err = d.messages.Create(
		message.TypeHelp,
		"worker-1",
		"daemon",
		"t-abc123",
		map[string]any{"message": "I'm stuck"},
	)
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	var tickEvents []events.Event
	d.processMessages(&tickEvents)

	// Help message should be forwarded to liaison.
	msgs, err := d.messages.ListPending(liaison.ID)
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("liaison messages = %d, want 1", len(msgs))
	}
	if msgs[0].Type != message.TypeHelp {
		t.Errorf("message type = %q, want %q", msgs[0].Type, message.TypeHelp)
	}
}

func TestProcessMergeQueue(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Write a merge queue item.
	item := MergeItem{
		TaskID:   "t-abc123",
		Branch:   "worker/test-branch",
		AgentID:  "w-abc123",
		QueuedAt: time.Now(),
	}
	data, _ := json.MarshalIndent(item, "", "  ")
	data = append(data, '\n')
	queuePath := filepath.Join(d.altDir, "merge-queue", fmt.Sprintf("%d-%s.json", time.Now().UnixNano(), item.TaskID))
	if err := os.WriteFile(queuePath, data, 0o644); err != nil {
		t.Fatalf("write queue item: %v", err)
	}

	var tickEvents []events.Event
	d.processMergeQueue(&tickEvents)

	// Should have MergeStarted event (the merge itself will fail since
	// we don't have a real git repo, but that's expected).
	if len(tickEvents) < 1 {
		t.Fatal("expected at least one merge event")
	}
	if tickEvents[0].Type != events.MergeStarted {
		t.Errorf("first event type = %q, want %q", tickEvents[0].Type, events.MergeStarted)
	}
}

func TestCheckConstraints_BudgetOK(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	var tickEvents []events.Event
	d.checkConstraints(&tickEvents)

	// No events when budget is fine.
	if len(tickEvents) != 0 {
		t.Errorf("tick events = %d, want 0", len(tickEvents))
	}
}

func TestCheckConstraints_BudgetExceeded(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Write events that exceed the budget (default ceiling is 100.0).
	evWriter := events.NewWriter(filepath.Join(d.altDir, "events.jsonl"))
	for i := 0; i < 20; i++ {
		err := evWriter.Append(events.Event{
			Timestamp: time.Now(),
			Type:      events.TaskDone,
			AgentID:   "test",
			TaskID:    fmt.Sprintf("t-%d", i),
			Data:      map[string]any{"token_cost": 10.0},
		})
		if err != nil {
			t.Fatalf("append event: %v", err)
		}
	}

	var tickEvents []events.Event
	d.checkConstraints(&tickEvents)

	// Should have a BudgetExceeded event.
	if len(tickEvents) != 1 {
		t.Fatalf("tick events = %d, want 1", len(tickEvents))
	}
	if tickEvents[0].Type != events.BudgetExceeded {
		t.Errorf("event type = %q, want %q", tickEvents[0].Type, events.BudgetExceeded)
	}
}

func TestEmitEvents(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	tickEvents := []events.Event{
		{
			Timestamp: time.Now(),
			Type:      events.AgentDied,
			AgentID:   "test-agent",
			TaskID:    "t-abc123",
		},
		{
			Timestamp: time.Now(),
			Type:      events.TaskDone,
			AgentID:   "test-agent",
			TaskID:    "t-abc123",
		},
	}

	d.emitEvents(tickEvents)

	// Verify events were written.
	reader := events.NewReader(filepath.Join(d.altDir, "events.jsonl"))
	all, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("read events: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("events written = %d, want 2", len(all))
	}
}

func TestEmitEvents_Empty(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Empty events should not create the file.
	d.emitEvents(nil)

	_, err = os.Stat(filepath.Join(d.altDir, "events.jsonl"))
	if err == nil {
		t.Error("events file should not exist for empty events")
	}
}

func TestStop(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Stop should close the shutdown channel.
	d.Stop()

	select {
	case <-d.shutdown:
		// ok
	default:
		t.Error("shutdown channel should be closed after Stop()")
	}

	// Double stop should not panic.
	d.Stop()
}

func TestAddToMergeQueue(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	tk := &task.Task{
		ID:         "t-abc123",
		Title:      "test task",
		Branch:     "worker/w-abc123",
		AssignedTo: "w-abc123",
	}

	if err := d.addToMergeQueue(tk); err != nil {
		t.Fatalf("addToMergeQueue: %v", err)
	}

	// Verify the queue item was written.
	entries, err := os.ReadDir(filepath.Join(d.altDir, "merge-queue"))
	if err != nil {
		t.Fatalf("read merge-queue: %v", err)
	}
	found := 0
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".json" {
			found++
		}
	}
	if found != 1 {
		t.Errorf("merge queue items = %d, want 1", found)
	}
}

func TestGenerateAgentID(t *testing.T) {
	id, err := generateAgentID()
	if err != nil {
		t.Fatalf("generateAgentID: %v", err)
	}
	if len(id) != 8 { // "w-" + 6 hex chars
		t.Errorf("agent id length = %d, want 8", len(id))
	}
	if id[:2] != "w-" {
		t.Errorf("agent id prefix = %q, want %q", id[:2], "w-")
	}

	// Generate multiple IDs to check uniqueness.
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id, err := generateAgentID()
		if err != nil {
			t.Fatalf("generateAgentID: %v", err)
		}
		if seen[id] {
			t.Errorf("duplicate agent id: %s", id)
		}
		seen[id] = true
	}
}

func TestAtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	data := []byte(`{"key": "value"}`)

	if err := atomicWrite(path, data); err != nil {
		t.Fatalf("atomicWrite: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("file content = %q, want %q", string(got), string(data))
	}
}

func TestTick(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Mock gitLogTimestamp so progress check doesn't shell out.
	origGitLog := gitLogTimestamp
	defer func() { gitLogTimestamp = origGitLog }()
	gitLogTimestamp = func(worktree string) (string, error) {
		return fmt.Sprintf("%d", time.Now().Unix()), nil
	}

	// tick should not panic with an empty project.
	d.tick()
}

// --- Reconciliation tests ---

func TestReconcileAgents_StaleMarkedDead(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Create a stale active agent (dead process).
	a := &agent.Agent{
		ID:        "stale-agent",
		Role:      agent.RoleWorker,
		Status:    agent.StatusActive,
		PID:       9999999,
		Heartbeat: time.Now().Add(-5 * time.Minute),
		StartedAt: time.Now().Add(-10 * time.Minute),
	}
	if err := d.agents.Create(a); err != nil {
		t.Fatalf("create agent: %v", err)
	}

	d.reconcileAgents()

	updated, err := d.agents.Get("stale-agent")
	if err != nil {
		t.Fatalf("get agent: %v", err)
	}
	if updated.Status != agent.StatusDead {
		t.Errorf("agent status = %q, want %q", updated.Status, agent.StatusDead)
	}
}

func TestReconcileAgents_LivePreserved(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Create a live agent (our own PID, fresh heartbeat).
	a := &agent.Agent{
		ID:        "live-agent",
		Role:      agent.RoleWorker,
		Status:    agent.StatusActive,
		PID:       os.Getpid(),
		Heartbeat: time.Now(),
		StartedAt: time.Now(),
	}
	if err := d.agents.Create(a); err != nil {
		t.Fatalf("create agent: %v", err)
	}

	d.reconcileAgents()

	updated, err := d.agents.Get("live-agent")
	if err != nil {
		t.Fatalf("get agent: %v", err)
	}
	if updated.Status != agent.StatusActive {
		t.Errorf("agent status = %q, want %q", updated.Status, agent.StatusActive)
	}
}

func TestReconcileTasks_ReclaimFromDeadAgent(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Create a dead agent.
	a := &agent.Agent{
		ID:        "dead-worker",
		Role:      agent.RoleWorker,
		Status:    agent.StatusDead,
		PID:       9999999,
		Heartbeat: time.Now().Add(-5 * time.Minute),
		StartedAt: time.Now().Add(-10 * time.Minute),
	}
	if err := d.agents.Create(a); err != nil {
		t.Fatalf("create agent: %v", err)
	}

	// Create a task assigned to the dead agent.
	tk := &task.Task{
		Title:      "orphaned task",
		Status:     task.StatusAssigned,
		AssignedTo: "dead-worker",
		Branch:     "worker/dead-worker",
	}
	if err := d.tasks.Create(tk); err != nil {
		t.Fatalf("create task: %v", err)
	}

	d.reconcileTasks()

	updated, err := d.tasks.Get(tk.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if updated.Status != task.StatusOpen {
		t.Errorf("task status = %q, want %q", updated.Status, task.StatusOpen)
	}
	if updated.AssignedTo != "" {
		t.Errorf("task assigned_to = %q, want empty", updated.AssignedTo)
	}
	if updated.Branch != "" {
		t.Errorf("task branch = %q, want empty", updated.Branch)
	}
}

func TestReconcileTasks_PreserveLiveAgentTasks(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Create a live agent.
	a := &agent.Agent{
		ID:        "live-worker",
		Role:      agent.RoleWorker,
		Status:    agent.StatusActive,
		PID:       os.Getpid(),
		Heartbeat: time.Now(),
		StartedAt: time.Now(),
	}
	if err := d.agents.Create(a); err != nil {
		t.Fatalf("create agent: %v", err)
	}

	// Create a task assigned to the live agent.
	tk := &task.Task{
		Title:      "live task",
		Status:     task.StatusAssigned,
		AssignedTo: "live-worker",
		Branch:     "worker/live-worker",
	}
	if err := d.tasks.Create(tk); err != nil {
		t.Fatalf("create task: %v", err)
	}

	d.reconcileTasks()

	updated, err := d.tasks.Get(tk.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if updated.Status != task.StatusAssigned {
		t.Errorf("task status = %q, want %q", updated.Status, task.StatusAssigned)
	}
	if updated.AssignedTo != "live-worker" {
		t.Errorf("task assigned_to = %q, want %q", updated.AssignedTo, "live-worker")
	}
}

func TestReconcileMergeQueue_CleansOrphanedTempFiles(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	queueDir := filepath.Join(d.altDir, "merge-queue")

	// Create an orphaned temp file.
	tmpFile := filepath.Join(queueDir, ".tmp-daemon-12345")
	if err := os.WriteFile(tmpFile, []byte("partial"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	// Create a normal queue item (should be preserved).
	normalFile := filepath.Join(queueDir, "123-t-abc.json")
	if err := os.WriteFile(normalFile, []byte(`{"task_id":"t-abc"}`), 0o644); err != nil {
		t.Fatalf("write normal file: %v", err)
	}

	d.reconcileMergeQueue()

	// Temp file should be removed.
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Error("orphaned temp file should be removed")
	}
	// Normal file should still exist.
	if _, err := os.Stat(normalFile); err != nil {
		t.Error("normal queue item should be preserved")
	}
}

// --- Stall notification throttling ---

func TestCheckProgress_StallNotificationThrottled(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	worktreeDir := t.TempDir()

	// Create a worker that was recently notified about stalling.
	a := &agent.Agent{
		ID:                "throttled-worker",
		Role:              agent.RoleWorker,
		Status:            agent.StatusActive,
		CurrentTask:       "t-abc123",
		Worktree:          worktreeDir,
		PID:               os.Getpid(),
		Heartbeat:         time.Now(),
		StartedAt:         time.Now().Add(-2 * time.Hour),
		LastStallNotified: time.Now().Add(-10 * time.Minute), // recent
	}
	if err := d.agents.Create(a); err != nil {
		t.Fatalf("create agent: %v", err)
	}

	// Create a liaison.
	liaison := &agent.Agent{
		ID:        "liaison-throttle",
		Role:      agent.RoleLiaison,
		Status:    agent.StatusActive,
		PID:       os.Getpid(),
		Heartbeat: time.Now(),
		StartedAt: time.Now(),
	}
	if err := d.agents.Create(liaison); err != nil {
		t.Fatalf("create liaison: %v", err)
	}

	// Mock stale commit time.
	origGitLog := gitLogTimestamp
	defer func() { gitLogTimestamp = origGitLog }()
	gitLogTimestamp = func(worktree string) (string, error) {
		return fmt.Sprintf("%d", time.Now().Add(-2*time.Hour).Unix()), nil
	}

	var tickEvents []events.Event
	d.checkProgress(&tickEvents)

	// Should NOT emit a stall event because notification was recent.
	if len(tickEvents) != 0 {
		t.Errorf("tick events = %d, want 0 (throttled)", len(tickEvents))
	}
}

func TestCheckProgress_StallNotificationAfterThreshold(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	worktreeDir := t.TempDir()

	// Create a worker whose last stall notification was beyond the threshold.
	a := &agent.Agent{
		ID:                "renotify-worker",
		Role:              agent.RoleWorker,
		Status:            agent.StatusActive,
		CurrentTask:       "t-abc123",
		Worktree:          worktreeDir,
		PID:               os.Getpid(),
		Heartbeat:         time.Now(),
		StartedAt:         time.Now().Add(-3 * time.Hour),
		LastStallNotified: time.Now().Add(-45 * time.Minute), // older than StalledThreshold
	}
	if err := d.agents.Create(a); err != nil {
		t.Fatalf("create agent: %v", err)
	}

	liaison := &agent.Agent{
		ID:        "liaison-renotify",
		Role:      agent.RoleLiaison,
		Status:    agent.StatusActive,
		PID:       os.Getpid(),
		Heartbeat: time.Now(),
		StartedAt: time.Now(),
	}
	if err := d.agents.Create(liaison); err != nil {
		t.Fatalf("create liaison: %v", err)
	}

	origGitLog := gitLogTimestamp
	defer func() { gitLogTimestamp = origGitLog }()
	gitLogTimestamp = func(worktree string) (string, error) {
		return fmt.Sprintf("%d", time.Now().Add(-2*time.Hour).Unix()), nil
	}

	var tickEvents []events.Event
	d.checkProgress(&tickEvents)

	// Should emit a stall event since threshold elapsed.
	if len(tickEvents) != 1 {
		t.Fatalf("tick events = %d, want 1", len(tickEvents))
	}
	if tickEvents[0].Type != events.WorkerStalled {
		t.Errorf("event type = %q, want %q", tickEvents[0].Type, events.WorkerStalled)
	}

	// Verify LastStallNotified was updated.
	updated, err := d.agents.Get("renotify-worker")
	if err != nil {
		t.Fatalf("get agent: %v", err)
	}
	if time.Since(updated.LastStallNotified) > 5*time.Second {
		t.Error("LastStallNotified should have been updated to ~now")
	}
}

// --- Merge conflict handling ---

func TestProcessMergeQueue_FailedItemRemoved(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Write a merge queue item that will fail (no git repo).
	item := MergeItem{
		TaskID:   "t-fail",
		Branch:   "worker/nonexistent",
		AgentID:  "w-fail",
		QueuedAt: time.Now(),
	}
	data, _ := json.MarshalIndent(item, "", "  ")
	data = append(data, '\n')
	queuePath := filepath.Join(d.altDir, "merge-queue", fmt.Sprintf("%d-%s.json", time.Now().UnixNano(), item.TaskID))
	if err := os.WriteFile(queuePath, data, 0o644); err != nil {
		t.Fatalf("write queue item: %v", err)
	}

	var tickEvents []events.Event
	d.processMergeQueue(&tickEvents)

	// The failed item should have been removed.
	if _, err := os.Stat(queuePath); !os.IsNotExist(err) {
		t.Error("failed merge queue item should be removed to prevent infinite retry")
	}
}

func TestProcessMergeQueue_SkipsTmpFiles(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Write a .tmp file that should be skipped.
	tmpPath := filepath.Join(d.altDir, "merge-queue", ".tmp-daemon-orphan.json")
	if err := os.WriteFile(tmpPath, []byte(`{"task_id":"t-tmp"}`), 0o644); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}

	var tickEvents []events.Event
	d.processMergeQueue(&tickEvents)

	// No events should be emitted (the .tmp file is skipped).
	if len(tickEvents) != 0 {
		t.Errorf("tick events = %d, want 0", len(tickEvents))
	}
}

// --- Concurrent atomic write safety ---

func TestAtomicWrite_ConcurrentSafety(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "concurrent.json")

	const numGoroutines = 10
	const writesPerGoroutine = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < writesPerGoroutine; i++ {
				data := []byte(fmt.Sprintf(`{"writer":%d,"seq":%d}`, id, i))
				if err := atomicWrite(path, data); err != nil {
					t.Errorf("goroutine %d write %d: %v", id, i, err)
					return
				}
			}
		}(g)
	}

	wg.Wait()

	// Verify the file contains valid JSON (no partial reads).
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("file content is not valid JSON after concurrent writes: %v (content: %q)", err, string(data))
	}

	// Verify no temp files left behind.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp-") {
			t.Errorf("leftover temp file: %s", e.Name())
		}
	}
}

// --- Resolver wiring ---

func TestBuildConflictContext(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Create a task with rig and description.
	tk := &task.Task{
		ID:          "t-ctx01",
		Title:       "Test task",
		Description: "Implement feature X",
		Status:      task.StatusDone,
		Rig:         "test-rig",
		Branch:      "worker/w-ctx01",
		AssignedTo:  "w-ctx01",
	}
	if err := d.tasks.Create(tk); err != nil {
		t.Fatalf("create task: %v", err)
	}

	item := MergeItem{
		TaskID:  "t-ctx01",
		Branch:  "worker/w-ctx01",
		AgentID: "w-ctx01",
	}

	conflicts := []merge.ConflictInfo{
		{Path: "main.go", Markers: []merge.ConflictMarker{{OursStart: 5, OursEnd: 8, TheirsStart: 8, TheirsEnd: 11}}},
	}

	ctx := d.buildConflictContext(item, conflicts)

	if ctx.TaskID != "t-ctx01" {
		t.Errorf("TaskID = %q, want %q", ctx.TaskID, "t-ctx01")
	}
	if ctx.Branch != "worker/w-ctx01" {
		t.Errorf("Branch = %q, want %q", ctx.Branch, "worker/w-ctx01")
	}
	if ctx.RigName != "test-rig" {
		t.Errorf("RigName = %q, want %q", ctx.RigName, "test-rig")
	}
	if ctx.TaskDescription != "Implement feature X" {
		t.Errorf("TaskDescription = %q, want %q", ctx.TaskDescription, "Implement feature X")
	}
	if len(ctx.Conflicts) != 1 || ctx.Conflicts[0].Path != "main.go" {
		t.Errorf("Conflicts = %v, want 1 conflict for main.go", ctx.Conflicts)
	}
}

func TestBuildConflictContext_MissingTask(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	item := MergeItem{
		TaskID:  "t-missing",
		Branch:  "worker/w-missing",
		AgentID: "w-missing",
	}

	ctx := d.buildConflictContext(item, nil)

	// Should still have item fields, but empty rig/description.
	if ctx.TaskID != "t-missing" {
		t.Errorf("TaskID = %q, want %q", ctx.TaskID, "t-missing")
	}
	if ctx.RigName != "" {
		t.Errorf("RigName = %q, want empty", ctx.RigName)
	}
	if ctx.TaskDescription != "" {
		t.Errorf("TaskDescription = %q, want empty", ctx.TaskDescription)
	}
}

func TestLoadConflictContext(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Create a fake worktree directory with a conflict-context.json.
	worktreeDir := t.TempDir()
	ctxData := `{
  "task_id": "t-load01",
  "branch": "worker/w-load01",
  "conflicts": [
    {"path": "file1.go", "markers": [{"OursStart": 1, "OursEnd": 5, "TheirsStart": 5, "TheirsEnd": 9}]},
    {"path": "file2.go", "markers": []}
  ]
}`
	if err := os.WriteFile(filepath.Join(worktreeDir, "conflict-context.json"), []byte(ctxData), 0o644); err != nil {
		t.Fatalf("write context: %v", err)
	}

	a := &agent.Agent{
		ID:       "resolver-01",
		Worktree: worktreeDir,
	}

	conflicts, err := d.loadConflictContext(a)
	if err != nil {
		t.Fatalf("loadConflictContext: %v", err)
	}
	if len(conflicts) != 2 {
		t.Fatalf("conflicts = %d, want 2", len(conflicts))
	}
	if conflicts[0].Path != "file1.go" {
		t.Errorf("conflict[0].Path = %q, want %q", conflicts[0].Path, "file1.go")
	}
}

func TestLoadConflictContext_MissingFile(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	a := &agent.Agent{
		ID:       "resolver-01",
		Worktree: t.TempDir(), // Empty dir, no conflict-context.json.
	}

	_, err = d.loadConflictContext(a)
	if err == nil {
		t.Fatal("expected error for missing conflict-context.json")
	}
}

func TestCheckResolvers_NoResolvers(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Should not panic or error with no resolvers.
	var tickEvents []events.Event
	d.checkResolvers(&tickEvents)

	if len(tickEvents) != 0 {
		t.Errorf("tick events = %d, want 0", len(tickEvents))
	}
}

func TestCheckResolvers_SkipsDeadResolvers(t *testing.T) {
	root := setupTestProject(t)
	d, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Create a dead resolver agent.
	a := &agent.Agent{
		ID:          "resolver-01",
		Role:        agent.RoleResolver,
		Status:      agent.StatusDead,
		CurrentTask: "t-dead",
		Worktree:    t.TempDir(),
		Heartbeat:   time.Now(),
		StartedAt:   time.Now(),
	}
	if err := d.agents.Create(a); err != nil {
		t.Fatalf("create agent: %v", err)
	}

	var tickEvents []events.Event
	d.checkResolvers(&tickEvents)

	// Dead resolver should be skipped, no events.
	if len(tickEvents) != 0 {
		t.Errorf("tick events = %d, want 0", len(tickEvents))
	}
}

// --- Budget validation ---

func TestNew_InvalidConstraints(t *testing.T) {
	root := setupTestProject(t)
	altDir := filepath.Join(root, ".alt")

	// Write a config with invalid constraints (negative budget).
	cfg := config.NewConfig()
	cfg.Constraints.BudgetCeiling = -1
	data, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(filepath.Join(altDir, "config.json"), data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := New(root)
	if err == nil {
		t.Fatal("New: expected error for negative budget ceiling")
	}
}

func TestNew_ZeroWorkers(t *testing.T) {
	root := setupTestProject(t)
	altDir := filepath.Join(root, ".alt")

	cfg := config.NewConfig()
	cfg.Constraints.MaxWorkers = 0
	data, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(filepath.Join(altDir, "config.json"), data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := New(root)
	if err == nil {
		t.Fatal("New: expected error for zero max workers")
	}
}

func TestSendStop_NotRunning(t *testing.T) {
	root := setupTestProject(t)
	altDir := filepath.Join(root, ".alt")

	err := SendStop(altDir)
	if err == nil {
		t.Fatal("SendStop: expected error when daemon not running")
	}
}
