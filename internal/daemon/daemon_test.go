package daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/anthropics/altera/internal/agent"
	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/events"
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

	// Create a task that's assigned.
	tk := &task.Task{
		Title:      "test task",
		Status:     task.StatusAssigned,
		AssignedTo: "test-dead-2",
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

func TestSendStop_NotRunning(t *testing.T) {
	root := setupTestProject(t)
	altDir := filepath.Join(root, ".alt")

	err := SendStop(altDir)
	if err == nil {
		t.Fatal("SendStop: expected error when daemon not running")
	}
}
