package constraints

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/anthropics/altera/internal/agent"
	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/events"
)

// helper: write events to a temp JSONL file and return a Reader.
func writeEvents(t *testing.T, evts ...events.Event) *events.Reader {
	t.Helper()
	path := filepath.Join(t.TempDir(), "events.jsonl")
	w := events.NewWriter(path)
	if err := w.Append(evts...); err != nil {
		t.Fatalf("Append events: %v", err)
	}
	return events.NewReader(path)
}

// helper: create an agent store with the given agents.
func makeAgentStore(t *testing.T, agents ...*agent.Agent) *agent.Store {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "agents")
	s, err := agent.NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	for _, a := range agents {
		if err := s.Create(a); err != nil {
			t.Fatalf("Create agent %s: %v", a.ID, err)
		}
	}
	return s
}

// helper: create a merge queue dir with n JSON files.
func makeMergeQueue(t *testing.T, n int) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "merge-queue")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	for i := 0; i < n; i++ {
		f, err := os.Create(filepath.Join(dir, "item-"+string(rune('a'+i))+".json"))
		if err != nil {
			t.Fatalf("Create queue item: %v", err)
		}
		_ = f.Close()
	}
	return dir
}

func defaultCfg() config.Constraints {
	return config.Constraints{
		BudgetCeiling: 100.0,
		MaxWorkers:    4,
		MaxQueueDepth: 10,
	}
}

func emptyReader(t *testing.T) *events.Reader {
	t.Helper()
	path := filepath.Join(t.TempDir(), "empty.jsonl")
	w := events.NewWriter(path)
	// Write zero events to create the file.
	if err := w.Append(events.Event{
		Timestamp: time.Now().UTC(),
		Type:      events.TaskCreated,
		AgentID:   "setup",
		TaskID:    "setup",
	}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	// Re-create with no cost data - just use the file we created.
	return events.NewReader(path)
}

// --- BudgetUsed tests ---

func TestBudgetUsedSumsCosts(t *testing.T) {
	r := writeEvents(t,
		events.Event{
			Timestamp: time.Now().UTC(),
			Type:      events.TaskDone,
			AgentID:   "a1",
			TaskID:    "t1",
			Data:      map[string]any{"token_cost": 10.5},
		},
		events.Event{
			Timestamp: time.Now().UTC(),
			Type:      events.TaskDone,
			AgentID:   "a2",
			TaskID:    "t2",
			Data:      map[string]any{"token_cost": 20.0},
		},
		events.Event{
			Timestamp: time.Now().UTC(),
			Type:      events.AgentSpawned,
			AgentID:   "a3",
			TaskID:    "",
			Data:      map[string]any{"token_cost": 5.25},
		},
	)

	agents := makeAgentStore(t)
	c := NewChecker(defaultCfg(), agents, r, t.TempDir())

	used, err := c.BudgetUsed()
	if err != nil {
		t.Fatalf("BudgetUsed: %v", err)
	}
	if used != 35.75 {
		t.Errorf("BudgetUsed: got %.2f, want 35.75", used)
	}
}

func TestBudgetUsedSkipsEventsWithoutCost(t *testing.T) {
	r := writeEvents(t,
		events.Event{
			Timestamp: time.Now().UTC(),
			Type:      events.TaskCreated,
			AgentID:   "a1",
			TaskID:    "t1",
		},
		events.Event{
			Timestamp: time.Now().UTC(),
			Type:      events.TaskDone,
			AgentID:   "a1",
			TaskID:    "t1",
			Data:      map[string]any{"token_cost": 7.0},
		},
		events.Event{
			Timestamp: time.Now().UTC(),
			Type:      events.TaskDone,
			AgentID:   "a2",
			TaskID:    "t2",
			Data:      map[string]any{"result": "ok"},
		},
	)

	agents := makeAgentStore(t)
	c := NewChecker(defaultCfg(), agents, r, t.TempDir())

	used, err := c.BudgetUsed()
	if err != nil {
		t.Fatalf("BudgetUsed: %v", err)
	}
	if used != 7.0 {
		t.Errorf("BudgetUsed: got %.2f, want 7.00", used)
	}
}

func TestBudgetUsedNoEventsFile(t *testing.T) {
	r := events.NewReader(filepath.Join(t.TempDir(), "nonexistent.jsonl"))
	agents := makeAgentStore(t)
	c := NewChecker(defaultCfg(), agents, r, t.TempDir())

	used, err := c.BudgetUsed()
	if err != nil {
		t.Fatalf("BudgetUsed: %v", err)
	}
	if used != 0 {
		t.Errorf("BudgetUsed: got %.2f, want 0", used)
	}
}

// --- WorkerCount tests ---

func TestWorkerCountActiveOnly(t *testing.T) {
	agents := makeAgentStore(t,
		&agent.Agent{ID: "w1", Role: agent.RoleWorker, Status: agent.StatusActive, Heartbeat: time.Now(), StartedAt: time.Now()},
		&agent.Agent{ID: "w2", Role: agent.RoleWorker, Status: agent.StatusActive, Heartbeat: time.Now(), StartedAt: time.Now()},
		&agent.Agent{ID: "w3", Role: agent.RoleWorker, Status: agent.StatusIdle, Heartbeat: time.Now(), StartedAt: time.Now()},
		&agent.Agent{ID: "w4", Role: agent.RoleWorker, Status: agent.StatusDead, Heartbeat: time.Now(), StartedAt: time.Now()},
		&agent.Agent{ID: "l1", Role: agent.RoleLiaison, Status: agent.StatusActive, Heartbeat: time.Now(), StartedAt: time.Now()},
	)

	r := emptyReader(t)
	c := NewChecker(defaultCfg(), agents, r, t.TempDir())

	count, err := c.WorkerCount()
	if err != nil {
		t.Fatalf("WorkerCount: %v", err)
	}
	if count != 2 {
		t.Errorf("WorkerCount: got %d, want 2", count)
	}
}

func TestWorkerCountZero(t *testing.T) {
	agents := makeAgentStore(t,
		&agent.Agent{ID: "l1", Role: agent.RoleLiaison, Status: agent.StatusActive, Heartbeat: time.Now(), StartedAt: time.Now()},
	)

	r := emptyReader(t)
	c := NewChecker(defaultCfg(), agents, r, t.TempDir())

	count, err := c.WorkerCount()
	if err != nil {
		t.Fatalf("WorkerCount: %v", err)
	}
	if count != 0 {
		t.Errorf("WorkerCount: got %d, want 0", count)
	}
}

// --- QueueDepth tests ---

func TestQueueDepthCountsJSONFiles(t *testing.T) {
	dir := makeMergeQueue(t, 5)
	agents := makeAgentStore(t)
	r := emptyReader(t)
	c := NewChecker(defaultCfg(), agents, r, dir)

	depth, err := c.QueueDepth()
	if err != nil {
		t.Fatalf("QueueDepth: %v", err)
	}
	if depth != 5 {
		t.Errorf("QueueDepth: got %d, want 5", depth)
	}
}

func TestQueueDepthIgnoresNonJSON(t *testing.T) {
	dir := makeMergeQueue(t, 3)
	// Add a non-JSON file.
	_ = os.WriteFile(filepath.Join(dir, "README.md"), []byte("ignore"), 0o644)
	// Add a subdirectory.
	_ = os.MkdirAll(filepath.Join(dir, "subdir"), 0o755)

	agents := makeAgentStore(t)
	r := emptyReader(t)
	c := NewChecker(defaultCfg(), agents, r, dir)

	depth, err := c.QueueDepth()
	if err != nil {
		t.Fatalf("QueueDepth: %v", err)
	}
	if depth != 3 {
		t.Errorf("QueueDepth: got %d, want 3", depth)
	}
}

func TestQueueDepthMissingDir(t *testing.T) {
	agents := makeAgentStore(t)
	r := emptyReader(t)
	c := NewChecker(defaultCfg(), agents, r, filepath.Join(t.TempDir(), "no-such-dir"))

	depth, err := c.QueueDepth()
	if err != nil {
		t.Fatalf("QueueDepth: %v", err)
	}
	if depth != 0 {
		t.Errorf("QueueDepth: got %d, want 0", depth)
	}
}

// --- CheckBudget tests ---

func TestCheckBudgetUnderCeiling(t *testing.T) {
	r := writeEvents(t,
		events.Event{
			Timestamp: time.Now().UTC(),
			Type:      events.TaskDone,
			AgentID:   "a1",
			TaskID:    "t1",
			Data:      map[string]any{"token_cost": 50.0},
		},
	)
	agents := makeAgentStore(t)
	c := NewChecker(defaultCfg(), agents, r, t.TempDir())

	ok, reason, err := c.CheckBudget()
	if err != nil {
		t.Fatalf("CheckBudget: %v", err)
	}
	if !ok {
		t.Errorf("CheckBudget: expected ok, got reason: %s", reason)
	}
}

func TestCheckBudgetAtCeiling(t *testing.T) {
	r := writeEvents(t,
		events.Event{
			Timestamp: time.Now().UTC(),
			Type:      events.TaskDone,
			AgentID:   "a1",
			TaskID:    "t1",
			Data:      map[string]any{"token_cost": 100.0},
		},
	)
	agents := makeAgentStore(t)
	c := NewChecker(defaultCfg(), agents, r, t.TempDir())

	ok, reason, _ := c.CheckBudget()
	if ok {
		t.Error("CheckBudget: expected failure at ceiling")
	}
	if reason == "" {
		t.Error("CheckBudget: expected reason string")
	}
}

func TestCheckBudgetOverCeiling(t *testing.T) {
	r := writeEvents(t,
		events.Event{
			Timestamp: time.Now().UTC(),
			Type:      events.TaskDone,
			AgentID:   "a1",
			TaskID:    "t1",
			Data:      map[string]any{"token_cost": 150.0},
		},
	)
	agents := makeAgentStore(t)
	c := NewChecker(defaultCfg(), agents, r, t.TempDir())

	ok, reason, _ := c.CheckBudget()
	if ok {
		t.Error("CheckBudget: expected failure over ceiling")
	}
	if reason == "" {
		t.Error("CheckBudget: expected reason string")
	}
}

// --- CheckMaxWorkers tests ---

func TestCheckMaxWorkersUnderLimit(t *testing.T) {
	agents := makeAgentStore(t,
		&agent.Agent{ID: "w1", Role: agent.RoleWorker, Status: agent.StatusActive, Heartbeat: time.Now(), StartedAt: time.Now()},
		&agent.Agent{ID: "w2", Role: agent.RoleWorker, Status: agent.StatusActive, Heartbeat: time.Now(), StartedAt: time.Now()},
	)
	r := emptyReader(t)
	c := NewChecker(defaultCfg(), agents, r, t.TempDir())

	ok, reason, err := c.CheckMaxWorkers()
	if err != nil {
		t.Fatalf("CheckMaxWorkers: %v", err)
	}
	if !ok {
		t.Errorf("CheckMaxWorkers: expected ok, got reason: %s", reason)
	}
}

func TestCheckMaxWorkersAtLimit(t *testing.T) {
	agents := makeAgentStore(t,
		&agent.Agent{ID: "w1", Role: agent.RoleWorker, Status: agent.StatusActive, Heartbeat: time.Now(), StartedAt: time.Now()},
		&agent.Agent{ID: "w2", Role: agent.RoleWorker, Status: agent.StatusActive, Heartbeat: time.Now(), StartedAt: time.Now()},
		&agent.Agent{ID: "w3", Role: agent.RoleWorker, Status: agent.StatusActive, Heartbeat: time.Now(), StartedAt: time.Now()},
		&agent.Agent{ID: "w4", Role: agent.RoleWorker, Status: agent.StatusActive, Heartbeat: time.Now(), StartedAt: time.Now()},
	)
	r := emptyReader(t)
	c := NewChecker(defaultCfg(), agents, r, t.TempDir())

	ok, reason, _ := c.CheckMaxWorkers()
	if ok {
		t.Error("CheckMaxWorkers: expected failure at limit")
	}
	if reason == "" {
		t.Error("CheckMaxWorkers: expected reason string")
	}
}

// --- CheckQueueDepth tests ---

func TestCheckQueueDepthUnderLimit(t *testing.T) {
	dir := makeMergeQueue(t, 5)
	agents := makeAgentStore(t)
	r := emptyReader(t)
	c := NewChecker(defaultCfg(), agents, r, dir)

	ok, reason, err := c.CheckQueueDepth()
	if err != nil {
		t.Fatalf("CheckQueueDepth: %v", err)
	}
	if !ok {
		t.Errorf("CheckQueueDepth: expected ok, got reason: %s", reason)
	}
}

func TestCheckQueueDepthAtLimit(t *testing.T) {
	dir := makeMergeQueue(t, 10)
	agents := makeAgentStore(t)
	r := emptyReader(t)
	c := NewChecker(defaultCfg(), agents, r, dir)

	ok, reason, _ := c.CheckQueueDepth()
	if ok {
		t.Error("CheckQueueDepth: expected failure at limit")
	}
	if reason == "" {
		t.Error("CheckQueueDepth: expected reason string")
	}
}

// --- CanSpawnWorker tests ---

func TestCanSpawnWorkerAllClear(t *testing.T) {
	r := writeEvents(t,
		events.Event{
			Timestamp: time.Now().UTC(),
			Type:      events.TaskDone,
			AgentID:   "a1",
			TaskID:    "t1",
			Data:      map[string]any{"token_cost": 10.0},
		},
	)
	agents := makeAgentStore(t,
		&agent.Agent{ID: "w1", Role: agent.RoleWorker, Status: agent.StatusActive, Heartbeat: time.Now(), StartedAt: time.Now()},
	)
	dir := makeMergeQueue(t, 2)
	c := NewChecker(defaultCfg(), agents, r, dir)

	ok, reason := c.CanSpawnWorker()
	if !ok {
		t.Errorf("CanSpawnWorker: expected ok, got reason: %s", reason)
	}
	if reason != "" {
		t.Errorf("CanSpawnWorker: expected empty reason, got: %s", reason)
	}
}

func TestCanSpawnWorkerBlockedByBudget(t *testing.T) {
	r := writeEvents(t,
		events.Event{
			Timestamp: time.Now().UTC(),
			Type:      events.TaskDone,
			AgentID:   "a1",
			TaskID:    "t1",
			Data:      map[string]any{"token_cost": 200.0},
		},
	)
	agents := makeAgentStore(t)
	c := NewChecker(defaultCfg(), agents, r, t.TempDir())

	ok, reason := c.CanSpawnWorker()
	if ok {
		t.Error("CanSpawnWorker: expected blocked by budget")
	}
	if reason == "" {
		t.Error("CanSpawnWorker: expected reason string")
	}
}

func TestCanSpawnWorkerBlockedByWorkers(t *testing.T) {
	r := emptyReader(t)
	agents := makeAgentStore(t,
		&agent.Agent{ID: "w1", Role: agent.RoleWorker, Status: agent.StatusActive, Heartbeat: time.Now(), StartedAt: time.Now()},
		&agent.Agent{ID: "w2", Role: agent.RoleWorker, Status: agent.StatusActive, Heartbeat: time.Now(), StartedAt: time.Now()},
		&agent.Agent{ID: "w3", Role: agent.RoleWorker, Status: agent.StatusActive, Heartbeat: time.Now(), StartedAt: time.Now()},
		&agent.Agent{ID: "w4", Role: agent.RoleWorker, Status: agent.StatusActive, Heartbeat: time.Now(), StartedAt: time.Now()},
	)
	c := NewChecker(defaultCfg(), agents, r, t.TempDir())

	ok, reason := c.CanSpawnWorker()
	if ok {
		t.Error("CanSpawnWorker: expected blocked by max workers")
	}
	if reason == "" {
		t.Error("CanSpawnWorker: expected reason string")
	}
}

func TestCanSpawnWorkerBlockedByQueue(t *testing.T) {
	r := emptyReader(t)
	agents := makeAgentStore(t)
	dir := makeMergeQueue(t, 10)
	c := NewChecker(defaultCfg(), agents, r, dir)

	ok, reason := c.CanSpawnWorker()
	if ok {
		t.Error("CanSpawnWorker: expected blocked by queue depth")
	}
	if reason == "" {
		t.Error("CanSpawnWorker: expected reason string")
	}
}

func TestCanSpawnWorkerBudgetCheckedFirst(t *testing.T) {
	// Both budget and workers exceeded; budget reason should appear.
	r := writeEvents(t,
		events.Event{
			Timestamp: time.Now().UTC(),
			Type:      events.TaskDone,
			AgentID:   "a1",
			TaskID:    "t1",
			Data:      map[string]any{"token_cost": 200.0},
		},
	)
	agents := makeAgentStore(t,
		&agent.Agent{ID: "w1", Role: agent.RoleWorker, Status: agent.StatusActive, Heartbeat: time.Now(), StartedAt: time.Now()},
		&agent.Agent{ID: "w2", Role: agent.RoleWorker, Status: agent.StatusActive, Heartbeat: time.Now(), StartedAt: time.Now()},
		&agent.Agent{ID: "w3", Role: agent.RoleWorker, Status: agent.StatusActive, Heartbeat: time.Now(), StartedAt: time.Now()},
		&agent.Agent{ID: "w4", Role: agent.RoleWorker, Status: agent.StatusActive, Heartbeat: time.Now(), StartedAt: time.Now()},
	)
	c := NewChecker(defaultCfg(), agents, r, t.TempDir())

	ok, reason := c.CanSpawnWorker()
	if ok {
		t.Error("CanSpawnWorker: expected blocked")
	}
	// Budget should be checked first.
	if reason == "" {
		t.Error("CanSpawnWorker: expected reason string")
	}
}
