package worker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/anthropics/altera/internal/agent"
	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/events"
	"github.com/anthropics/altera/internal/git"
	"github.com/anthropics/altera/internal/task"
	"github.com/anthropics/altera/internal/tmux"
)

// setupProject creates a minimal project directory with .alt/, a bare git
// repo as the rig's repo, and a rig config pointing to it.
func setupProject(t *testing.T) (projectRoot string, rigRepo string) {
	t.Helper()

	projectRoot = t.TempDir()
	altDir := filepath.Join(projectRoot, config.DirName)

	// Create .alt/ subdirectories.
	for _, d := range []string{
		altDir,
		filepath.Join(altDir, "agents"),
		filepath.Join(altDir, "rigs", "test-rig"),
		filepath.Join(projectRoot, "worktrees"),
	} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	// Create a bare repo to serve as the rig's repo, with an initial commit
	// so branches can be created from it.
	rigRepo = filepath.Join(t.TempDir(), "rig-repo")
	if err := os.MkdirAll(rigRepo, 0o755); err != nil {
		t.Fatalf("mkdir rig repo: %v", err)
	}
	if err := git.Init(rigRepo); err != nil {
		t.Fatalf("git init: %v", err)
	}
	// Need at least one commit for branches to work.
	if err := os.WriteFile(filepath.Join(rigRepo, "README.md"), []byte("# test\n"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}
	if err := git.Add(rigRepo, nil); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := git.SetAuthor(rigRepo, "test", "test@test.local"); err != nil {
		t.Fatalf("set author: %v", err)
	}
	if err := git.Commit(rigRepo, "initial commit"); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	// Save rig config.
	rc := config.RigConfig{
		RepoPath:      rigRepo,
		DefaultBranch: "main",
		TestCommand:   "echo ok",
	}
	if err := config.SaveRig(altDir, "test-rig", rc); err != nil {
		t.Fatalf("save rig config: %v", err)
	}

	return projectRoot, rigRepo
}

func sampleTask() *task.Task {
	return &task.Task{
		ID:          "t-abc123",
		Title:       "Test task",
		Description: "Implement the widget feature",
		Status:      task.StatusAssigned,
		Rig:         "test-rig",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
}

func newTestManager(t *testing.T, projectRoot string) *Manager {
	t.Helper()
	agentDir := filepath.Join(projectRoot, config.DirName, "agents")
	agents, err := agent.NewStore(agentDir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	evPath := filepath.Join(projectRoot, config.DirName, "events.jsonl")
	ew := events.NewWriter(evPath)
	return NewManager(projectRoot, agents, ew)
}

func TestParseWorkerNum(t *testing.T) {
	tests := []struct {
		id   string
		want int
	}{
		{"worker-01", 1},
		{"worker-02", 2},
		{"worker-10", 10},
		{"worker-99", 99},
		{"worker-", 0},
		{"not-a-worker", 0},
		{"worker-abc", 0},
		{"", 0},
	}
	for _, tc := range tests {
		got := parseWorkerNum(tc.id)
		if got != tc.want {
			t.Errorf("parseWorkerNum(%q) = %d, want %d", tc.id, got, tc.want)
		}
	}
}

func TestWorkerID(t *testing.T) {
	tests := []struct {
		num  int
		want string
	}{
		{1, "worker-01"},
		{2, "worker-02"},
		{10, "worker-10"},
		{99, "worker-99"},
	}
	for _, tc := range tests {
		got := workerID(tc.num)
		if got != tc.want {
			t.Errorf("workerID(%d) = %q, want %q", tc.num, got, tc.want)
		}
	}
}

func TestNextWorkerNum_Empty(t *testing.T) {
	projectRoot := t.TempDir()
	m := newTestManager(t, projectRoot)

	num, err := m.nextWorkerNum()
	if err != nil {
		t.Fatalf("nextWorkerNum: %v", err)
	}
	if num != 1 {
		t.Errorf("nextWorkerNum() = %d, want 1", num)
	}
}

func TestNextWorkerNum_WithExisting(t *testing.T) {
	projectRoot := t.TempDir()
	m := newTestManager(t, projectRoot)

	// Create a couple of existing workers.
	for _, id := range []string{"worker-01", "worker-03"} {
		a := &agent.Agent{
			ID:        id,
			Role:      agent.RoleWorker,
			Status:    agent.StatusActive,
			Heartbeat: time.Now(),
			StartedAt: time.Now(),
		}
		if err := m.agents.Create(a); err != nil {
			t.Fatalf("Create %s: %v", id, err)
		}
	}

	num, err := m.nextWorkerNum()
	if err != nil {
		t.Fatalf("nextWorkerNum: %v", err)
	}
	if num != 4 {
		t.Errorf("nextWorkerNum() = %d, want 4", num)
	}
}

func TestWriteTaskJSON(t *testing.T) {
	dir := t.TempDir()
	tk := sampleTask()

	if err := writeTaskJSON(dir, tk); err != nil {
		t.Fatalf("writeTaskJSON: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "task.json"))
	if err != nil {
		t.Fatalf("read task.json: %v", err)
	}

	var got task.Task
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal task.json: %v", err)
	}
	if got.ID != tk.ID {
		t.Errorf("task ID = %q, want %q", got.ID, tk.ID)
	}
	if got.Title != tk.Title {
		t.Errorf("task Title = %q, want %q", got.Title, tk.Title)
	}
}

func TestWriteClaudeSettings(t *testing.T) {
	dir := t.TempDir()
	agentID := "worker-01"

	if err := writeClaudeSettings(dir, agentID); err != nil {
		t.Fatalf("writeClaudeSettings: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}

	var settings ClaudeSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("unmarshal settings: %v", err)
	}

	// Check PreToolUse hook.
	pre, ok := settings.Hooks["PreToolUse"]
	if !ok || len(pre) == 0 {
		t.Fatal("missing PreToolUse hook")
	}
	if len(pre[0].Hooks) == 0 {
		t.Fatal("PreToolUse hook group has no hooks")
	}
	if pre[0].Hooks[0].Command != "alt heartbeat worker-01" {
		t.Errorf("PreToolUse command = %q, want %q", pre[0].Hooks[0].Command, "alt heartbeat worker-01")
	}

	// Check Stop hook.
	stop, ok := settings.Hooks["Stop"]
	if !ok || len(stop) == 0 {
		t.Fatal("missing Stop hook")
	}
	if len(stop[0].Hooks) == 0 {
		t.Fatal("Stop hook group has no hooks")
	}
	if stop[0].Hooks[0].Command != "alt checkpoint worker-01" {
		t.Errorf("Stop command = %q, want %q", stop[0].Hooks[0].Command, "alt checkpoint worker-01")
	}
}

func TestSpawnWorker(t *testing.T) {
	if _, err := tmux.ListSessions(); err != nil {
		t.Skip("tmux not available")
	}

	projectRoot, _ := setupProject(t)
	m := newTestManager(t, projectRoot)
	tk := sampleTask()

	a, err := m.SpawnWorker(tk, "test-rig")
	if err != nil {
		t.Fatalf("SpawnWorker: %v", err)
	}
	t.Cleanup(func() {
		_ = m.CleanupWorker(a)
	})

	// Verify agent record.
	if a.ID != "worker-01" {
		t.Errorf("agent ID = %q, want %q", a.ID, "worker-01")
	}
	if a.Role != agent.RoleWorker {
		t.Errorf("agent Role = %q, want %q", a.Role, agent.RoleWorker)
	}
	if a.Status != agent.StatusActive {
		t.Errorf("agent Status = %q, want %q", a.Status, agent.StatusActive)
	}
	if a.CurrentTask != tk.ID {
		t.Errorf("agent CurrentTask = %q, want %q", a.CurrentTask, tk.ID)
	}
	if a.Rig != "test-rig" {
		t.Errorf("agent Rig = %q, want %q", a.Rig, "test-rig")
	}

	// Verify worktree exists.
	if _, err := os.Stat(a.Worktree); err != nil {
		t.Errorf("worktree does not exist: %v", err)
	}

	// Verify task.json in worktree.
	taskData, err := os.ReadFile(filepath.Join(a.Worktree, "task.json"))
	if err != nil {
		t.Errorf("task.json not found: %v", err)
	} else {
		var gotTask task.Task
		if err := json.Unmarshal(taskData, &gotTask); err != nil {
			t.Errorf("invalid task.json: %v", err)
		} else if gotTask.ID != tk.ID {
			t.Errorf("task.json ID = %q, want %q", gotTask.ID, tk.ID)
		}
	}

	// Verify .claude/settings.json in worktree.
	settingsData, err := os.ReadFile(filepath.Join(a.Worktree, ".claude", "settings.json"))
	if err != nil {
		t.Errorf("settings.json not found: %v", err)
	} else {
		var settings ClaudeSettings
		if err := json.Unmarshal(settingsData, &settings); err != nil {
			t.Errorf("invalid settings.json: %v", err)
		}
	}

	// Verify tmux session exists.
	if !tmux.SessionExists(a.TmuxSession) {
		t.Error("tmux session does not exist")
	}

	// Verify agent is persisted.
	got, err := m.agents.Get(a.ID)
	if err != nil {
		t.Errorf("agent not persisted: %v", err)
	} else if got.ID != a.ID {
		t.Errorf("persisted agent ID = %q, want %q", got.ID, a.ID)
	}

	// Verify event was emitted.
	er := events.NewReader(filepath.Join(projectRoot, config.DirName, "events.jsonl"))
	evts, err := er.ReadAll()
	if err != nil {
		t.Errorf("reading events: %v", err)
	} else {
		found := false
		for _, e := range evts {
			if e.Type == events.AgentSpawned && e.AgentID == a.ID {
				found = true
				break
			}
		}
		if !found {
			t.Error("agent_spawned event not found")
		}
	}
}

func TestSpawnWorker_SequentialNaming(t *testing.T) {
	if _, err := tmux.ListSessions(); err != nil {
		t.Skip("tmux not available")
	}

	projectRoot, _ := setupProject(t)
	m := newTestManager(t, projectRoot)

	// Spawn first worker.
	t1 := sampleTask()
	a1, err := m.SpawnWorker(t1, "test-rig")
	if err != nil {
		t.Fatalf("SpawnWorker 1: %v", err)
	}
	t.Cleanup(func() { _ = m.CleanupWorker(a1) })

	if a1.ID != "worker-01" {
		t.Errorf("first worker ID = %q, want %q", a1.ID, "worker-01")
	}

	// Spawn second worker (different task ID to avoid branch conflict).
	t2 := &task.Task{
		ID:          "t-def456",
		Title:       "Second task",
		Description: "Another task",
		Status:      task.StatusAssigned,
		Rig:         "test-rig",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	a2, err := m.SpawnWorker(t2, "test-rig")
	if err != nil {
		t.Fatalf("SpawnWorker 2: %v", err)
	}
	t.Cleanup(func() { _ = m.CleanupWorker(a2) })

	if a2.ID != "worker-02" {
		t.Errorf("second worker ID = %q, want %q", a2.ID, "worker-02")
	}
}

func TestCleanupWorker(t *testing.T) {
	if _, err := tmux.ListSessions(); err != nil {
		t.Skip("tmux not available")
	}

	projectRoot, _ := setupProject(t)
	m := newTestManager(t, projectRoot)
	tk := sampleTask()

	a, err := m.SpawnWorker(tk, "test-rig")
	if err != nil {
		t.Fatalf("SpawnWorker: %v", err)
	}

	worktreePath := a.Worktree
	sessionName := a.TmuxSession

	if err := m.CleanupWorker(a); err != nil {
		t.Fatalf("CleanupWorker: %v", err)
	}

	// Verify tmux session is gone.
	if tmux.SessionExists(sessionName) {
		t.Error("tmux session still exists after cleanup")
	}

	// Verify worktree is gone.
	if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
		t.Error("worktree still exists after cleanup")
	}

	// Verify agent status is dead.
	got, err := m.agents.Get(a.ID)
	if err != nil {
		t.Fatalf("Get agent: %v", err)
	}
	if got.Status != agent.StatusDead {
		t.Errorf("agent status = %q, want %q", got.Status, agent.StatusDead)
	}

	// Verify agent_died event was emitted.
	er := events.NewReader(filepath.Join(projectRoot, config.DirName, "events.jsonl"))
	evts, err := er.ReadAll()
	if err != nil {
		t.Errorf("reading events: %v", err)
	} else {
		found := false
		for _, e := range evts {
			if e.Type == events.AgentDied && e.AgentID == a.ID {
				found = true
				break
			}
		}
		if !found {
			t.Error("agent_died event not found")
		}
	}
}

func TestListWorkers(t *testing.T) {
	projectRoot := t.TempDir()
	m := newTestManager(t, projectRoot)

	// Create some agents with mixed roles.
	agents := []struct {
		id   string
		role agent.Role
	}{
		{"worker-02", agent.RoleWorker},
		{"worker-01", agent.RoleWorker},
		{"liaison-01", agent.RoleLiaison},
		{"worker-03", agent.RoleWorker},
	}
	for _, tc := range agents {
		a := &agent.Agent{
			ID:        tc.id,
			Role:      tc.role,
			Status:    agent.StatusActive,
			Heartbeat: time.Now(),
			StartedAt: time.Now(),
		}
		if err := m.agents.Create(a); err != nil {
			t.Fatalf("Create %s: %v", tc.id, err)
		}
	}

	workers, err := m.ListWorkers()
	if err != nil {
		t.Fatalf("ListWorkers: %v", err)
	}
	if len(workers) != 3 {
		t.Fatalf("ListWorkers = %d agents, want 3", len(workers))
	}

	// Verify sorted order.
	expectedIDs := []string{"worker-01", "worker-02", "worker-03"}
	for i, w := range workers {
		if w.ID != expectedIDs[i] {
			t.Errorf("worker[%d].ID = %q, want %q", i, w.ID, expectedIDs[i])
		}
	}
}

func TestSpawnWorker_BadRig(t *testing.T) {
	projectRoot := t.TempDir()
	os.MkdirAll(filepath.Join(projectRoot, config.DirName, "agents"), 0o755)
	m := newTestManager(t, projectRoot)
	tk := sampleTask()

	_, err := m.SpawnWorker(tk, "nonexistent-rig")
	if err == nil {
		t.Fatal("expected error for nonexistent rig")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
