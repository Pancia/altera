package resolver

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
	"github.com/anthropics/altera/internal/merge"
	"github.com/anthropics/altera/internal/tmux"
)

// setupProject creates a minimal project directory with .alt/, a git repo
// as the rig's repo, a rig config, and a conflicting branch for testing.
func setupProject(t *testing.T) (projectRoot, rigRepo, conflictBranch string) {
	t.Helper()

	projectRoot = t.TempDir()
	altDir := filepath.Join(projectRoot, config.DirName)

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

	// Create a repo with an initial commit.
	rigRepo = filepath.Join(t.TempDir(), "rig-repo")
	if err := os.MkdirAll(rigRepo, 0o755); err != nil {
		t.Fatalf("mkdir rig repo: %v", err)
	}
	if err := git.Init(rigRepo); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := git.SetAuthor(rigRepo, "test", "test@test.local"); err != nil {
		t.Fatalf("set author: %v", err)
	}

	// Create initial file.
	if err := os.WriteFile(filepath.Join(rigRepo, "hello.txt"), []byte("line 1\nline 2\nline 3\n"), 0o644); err != nil {
		t.Fatalf("write hello.txt: %v", err)
	}
	if err := git.Add(rigRepo, nil); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := git.Commit(rigRepo, "initial commit"); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	// Create a branch with conflicting changes.
	conflictBranch = "feature-conflict"
	if err := git.CreateBranch(rigRepo, conflictBranch, "main"); err != nil {
		t.Fatalf("create branch: %v", err)
	}
	if err := git.Checkout(rigRepo, conflictBranch); err != nil {
		t.Fatalf("checkout conflict branch: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rigRepo, "hello.txt"), []byte("line 1\ntheir change\nline 3\n"), 0o644); err != nil {
		t.Fatalf("write conflict: %v", err)
	}
	if err := git.Add(rigRepo, nil); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := git.Commit(rigRepo, "their change"); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	// Switch back to main and make a conflicting change.
	if err := git.Checkout(rigRepo, "main"); err != nil {
		t.Fatalf("checkout main: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rigRepo, "hello.txt"), []byte("line 1\nour change\nline 3\n"), 0o644); err != nil {
		t.Fatalf("write our change: %v", err)
	}
	if err := git.Add(rigRepo, nil); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := git.Commit(rigRepo, "our change"); err != nil {
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

	return projectRoot, rigRepo, conflictBranch
}

func sampleConflictContext(branch string) ConflictContext {
	return ConflictContext{
		TaskID:     "t-abc123",
		Branch:     branch,
		BaseBranch: "main",
		RigName:    "test-rig",
		Conflicts: []merge.ConflictInfo{
			{
				Path: "hello.txt",
				Markers: []merge.ConflictMarker{
					{OursStart: 2, OursEnd: 3, TheirsStart: 3, TheirsEnd: 4},
				},
			},
		},
		TaskDescription: "Implement the widget feature",
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

func TestParseResolverNum(t *testing.T) {
	tests := []struct {
		id   string
		want int
	}{
		{"resolver-01", 1},
		{"resolver-02", 2},
		{"resolver-10", 10},
		{"resolver-99", 99},
		{"resolver-", 0},
		{"not-a-resolver", 0},
		{"resolver-abc", 0},
		{"worker-01", 0},
		{"", 0},
	}
	for _, tc := range tests {
		got := parseResolverNum(tc.id)
		if got != tc.want {
			t.Errorf("parseResolverNum(%q) = %d, want %d", tc.id, got, tc.want)
		}
	}
}

func TestResolverID(t *testing.T) {
	tests := []struct {
		num  int
		want string
	}{
		{1, "resolver-01"},
		{2, "resolver-02"},
		{10, "resolver-10"},
		{99, "resolver-99"},
	}
	for _, tc := range tests {
		got := resolverID(tc.num)
		if got != tc.want {
			t.Errorf("resolverID(%d) = %q, want %q", tc.num, got, tc.want)
		}
	}
}

func TestNextResolverNum_Empty(t *testing.T) {
	projectRoot := t.TempDir()
	m := newTestManager(t, projectRoot)

	num, err := m.nextResolverNum()
	if err != nil {
		t.Fatalf("nextResolverNum: %v", err)
	}
	if num != 1 {
		t.Errorf("nextResolverNum() = %d, want 1", num)
	}
}

func TestNextResolverNum_WithExisting(t *testing.T) {
	projectRoot := t.TempDir()
	m := newTestManager(t, projectRoot)

	for _, id := range []string{"resolver-01", "resolver-03"} {
		a := &agent.Agent{
			ID:        id,
			Role:      agent.RoleResolver,
			Status:    agent.StatusActive,
			Heartbeat: time.Now(),
			StartedAt: time.Now(),
		}
		if err := m.agents.Create(a); err != nil {
			t.Fatalf("Create %s: %v", id, err)
		}
	}

	num, err := m.nextResolverNum()
	if err != nil {
		t.Fatalf("nextResolverNum: %v", err)
	}
	if num != 4 {
		t.Errorf("nextResolverNum() = %d, want 4", num)
	}
}

func TestWriteConflictContext(t *testing.T) {
	dir := t.TempDir()
	ctx := sampleConflictContext("feature-conflict")

	if err := writeConflictContext(dir, ctx); err != nil {
		t.Fatalf("writeConflictContext: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "conflict-context.json"))
	if err != nil {
		t.Fatalf("read conflict-context.json: %v", err)
	}

	var got ConflictContext
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal conflict-context.json: %v", err)
	}
	if got.TaskID != ctx.TaskID {
		t.Errorf("TaskID = %q, want %q", got.TaskID, ctx.TaskID)
	}
	if got.Branch != ctx.Branch {
		t.Errorf("Branch = %q, want %q", got.Branch, ctx.Branch)
	}
	if len(got.Conflicts) != 1 {
		t.Fatalf("Conflicts count = %d, want 1", len(got.Conflicts))
	}
	if got.Conflicts[0].Path != "hello.txt" {
		t.Errorf("Conflicts[0].Path = %q, want %q", got.Conflicts[0].Path, "hello.txt")
	}
}

func TestWriteClaudeSettings(t *testing.T) {
	dir := t.TempDir()
	agentID := "resolver-01"

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

	pre, ok := settings.Hooks["PreToolUse"]
	if !ok || len(pre) == 0 {
		t.Fatal("missing PreToolUse hook")
	}
	if pre[0].Command != "alt heartbeat resolver-01" {
		t.Errorf("PreToolUse command = %q, want %q", pre[0].Command, "alt heartbeat resolver-01")
	}

	stop, ok := settings.Hooks["Stop"]
	if !ok || len(stop) == 0 {
		t.Fatal("missing Stop hook")
	}
	if stop[0].Command != "alt checkpoint resolver-01" {
		t.Errorf("Stop command = %q, want %q", stop[0].Command, "alt checkpoint resolver-01")
	}
}

func TestResolverPrompt(t *testing.T) {
	ctx := sampleConflictContext("feature-conflict")
	prompt := ResolverPrompt(ctx, "resolver-05")

	for _, want := range []string{
		"resolver-05",
		"t-abc123",
		"feature-conflict",
		"main",
		"test-rig",
		"Implement the widget feature",
		"hello.txt",
		"resolve merge conflicts",
	} {
		if !contains(prompt, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
}

func TestWriteClaudeMD(t *testing.T) {
	dir := t.TempDir()
	ctx := sampleConflictContext("feature-conflict")

	if err := writeClaudeMD(dir, ctx, "resolver-01"); err != nil {
		t.Fatalf("writeClaudeMD: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}

	content := string(data)
	for _, want := range []string{
		"resolver-01",
		"t-abc123",
		"test-rig",
		"Implement the widget feature",
	} {
		if !contains(content, want) {
			t.Errorf("CLAUDE.md missing %q", want)
		}
	}
}

func TestHasConflictMarkers(t *testing.T) {
	dir := t.TempDir()

	// File with conflict markers.
	conflicted := filepath.Join(dir, "conflicted.txt")
	os.WriteFile(conflicted, []byte("before\n<<<<<<< HEAD\nours\n=======\ntheirs\n>>>>>>> branch\nafter\n"), 0o644)
	if !hasConflictMarkers(conflicted) {
		t.Error("expected conflict markers in conflicted.txt")
	}

	// File without conflict markers.
	clean := filepath.Join(dir, "clean.txt")
	os.WriteFile(clean, []byte("just some normal content\n"), 0o644)
	if hasConflictMarkers(clean) {
		t.Error("unexpected conflict markers in clean.txt")
	}

	// Nonexistent file.
	if hasConflictMarkers(filepath.Join(dir, "nonexistent.txt")) {
		t.Error("unexpected conflict markers in nonexistent file")
	}
}

func TestDetectResolution_WithMarkers(t *testing.T) {
	dir := t.TempDir()

	// Set up a git repo so IsClean works.
	if err := git.Init(dir); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := git.SetAuthor(dir, "test", "test@test.local"); err != nil {
		t.Fatalf("set author: %v", err)
	}

	// Write a file with conflict markers.
	os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("<<<<<<< HEAD\nours\n=======\ntheirs\n>>>>>>> branch\n"), 0o644)

	a := &agent.Agent{
		ID:       "resolver-01",
		Worktree: dir,
	}
	conflicts := []merge.ConflictInfo{
		{Path: "hello.txt"},
	}

	resolved, err := DetectResolution(a, conflicts)
	if err != nil {
		t.Fatalf("DetectResolution: %v", err)
	}
	if resolved {
		t.Error("expected not resolved when conflict markers present")
	}
}

func TestDetectResolution_CleanResolution(t *testing.T) {
	dir := t.TempDir()

	// Set up a git repo with a clean committed file.
	if err := git.Init(dir); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := git.SetAuthor(dir, "test", "test@test.local"); err != nil {
		t.Fatalf("set author: %v", err)
	}
	os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("resolved content\n"), 0o644)
	if err := git.Add(dir, nil); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := git.Commit(dir, "resolved conflicts"); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	a := &agent.Agent{
		ID:       "resolver-01",
		Worktree: dir,
	}
	conflicts := []merge.ConflictInfo{
		{Path: "hello.txt"},
	}

	resolved, err := DetectResolution(a, conflicts)
	if err != nil {
		t.Fatalf("DetectResolution: %v", err)
	}
	if !resolved {
		t.Error("expected resolved when markers gone and tree clean")
	}
}

func TestDetectResolution_UncommittedChanges(t *testing.T) {
	dir := t.TempDir()

	if err := git.Init(dir); err != nil {
		t.Fatalf("git init: %v", err)
	}
	if err := git.SetAuthor(dir, "test", "test@test.local"); err != nil {
		t.Fatalf("set author: %v", err)
	}
	// Initial commit.
	os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("initial\n"), 0o644)
	if err := git.Add(dir, nil); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := git.Commit(dir, "initial"); err != nil {
		t.Fatalf("git commit: %v", err)
	}

	// Modify without committing.
	os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("resolved content\n"), 0o644)

	a := &agent.Agent{
		ID:       "resolver-01",
		Worktree: dir,
	}
	conflicts := []merge.ConflictInfo{
		{Path: "hello.txt"},
	}

	resolved, err := DetectResolution(a, conflicts)
	if err != nil {
		t.Fatalf("DetectResolution: %v", err)
	}
	if resolved {
		t.Error("expected not resolved when uncommitted changes exist")
	}
}

func TestDetectResolution_NoWorktree(t *testing.T) {
	a := &agent.Agent{
		ID:       "resolver-01",
		Worktree: "",
	}

	_, err := DetectResolution(a, nil)
	if err == nil {
		t.Fatal("expected error for agent with no worktree")
	}
}

func TestSpawnResolver(t *testing.T) {
	if _, err := tmux.ListSessions(); err != nil {
		t.Skip("tmux not available")
	}

	projectRoot, _, conflictBranch := setupProject(t)
	m := newTestManager(t, projectRoot)
	ctx := sampleConflictContext(conflictBranch)

	a, err := m.SpawnResolver(ctx)
	if err != nil {
		t.Fatalf("SpawnResolver: %v", err)
	}
	t.Cleanup(func() {
		_ = m.CleanupResolver(a)
	})

	// Verify agent record.
	if a.ID != "resolver-01" {
		t.Errorf("agent ID = %q, want %q", a.ID, "resolver-01")
	}
	if a.Role != agent.RoleResolver {
		t.Errorf("agent Role = %q, want %q", a.Role, agent.RoleResolver)
	}
	if a.Status != agent.StatusActive {
		t.Errorf("agent Status = %q, want %q", a.Status, agent.StatusActive)
	}
	if a.CurrentTask != ctx.TaskID {
		t.Errorf("agent CurrentTask = %q, want %q", a.CurrentTask, ctx.TaskID)
	}
	if a.Rig != "test-rig" {
		t.Errorf("agent Rig = %q, want %q", a.Rig, "test-rig")
	}

	// Verify worktree exists.
	if _, err := os.Stat(a.Worktree); err != nil {
		t.Errorf("worktree does not exist: %v", err)
	}

	// Verify conflict-context.json in worktree.
	ctxData, err := os.ReadFile(filepath.Join(a.Worktree, "conflict-context.json"))
	if err != nil {
		t.Errorf("conflict-context.json not found: %v", err)
	} else {
		var gotCtx ConflictContext
		if err := json.Unmarshal(ctxData, &gotCtx); err != nil {
			t.Errorf("invalid conflict-context.json: %v", err)
		} else if gotCtx.TaskID != ctx.TaskID {
			t.Errorf("conflict-context.json TaskID = %q, want %q", gotCtx.TaskID, ctx.TaskID)
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

	// Verify CLAUDE.md in worktree.
	if _, err := os.Stat(filepath.Join(a.Worktree, "CLAUDE.md")); err != nil {
		t.Errorf("CLAUDE.md not found: %v", err)
	}

	// Verify conflict markers are present in worktree (merge was not aborted).
	helloData, err := os.ReadFile(filepath.Join(a.Worktree, "hello.txt"))
	if err != nil {
		t.Errorf("hello.txt not found: %v", err)
	} else {
		if !contains(string(helloData), "<<<<<<<") {
			t.Error("expected conflict markers in hello.txt")
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

func TestSpawnResolver_NoConflict(t *testing.T) {
	projectRoot, rigRepo, _ := setupProject(t)
	m := newTestManager(t, projectRoot)

	// Create a branch that merges cleanly.
	if err := git.CreateBranch(rigRepo, "clean-branch", "main"); err != nil {
		t.Fatalf("create branch: %v", err)
	}
	if err := git.Checkout(rigRepo, "clean-branch"); err != nil {
		t.Fatalf("checkout: %v", err)
	}
	os.WriteFile(filepath.Join(rigRepo, "new-file.txt"), []byte("new content\n"), 0o644)
	if err := git.Add(rigRepo, nil); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := git.Commit(rigRepo, "add new file"); err != nil {
		t.Fatalf("git commit: %v", err)
	}
	if err := git.Checkout(rigRepo, "main"); err != nil {
		t.Fatalf("checkout main: %v", err)
	}

	ctx := ConflictContext{
		TaskID:     "t-clean",
		Branch:     "clean-branch",
		BaseBranch: "main",
		RigName:    "test-rig",
	}

	_, err := m.SpawnResolver(ctx)
	if err == nil {
		t.Fatal("expected error when no conflicts exist")
	}
	if !contains(err.Error(), "no conflicts found") {
		t.Errorf("error = %q, want 'no conflicts found'", err.Error())
	}
}

func TestSpawnResolver_BadRig(t *testing.T) {
	projectRoot := t.TempDir()
	os.MkdirAll(filepath.Join(projectRoot, config.DirName, "agents"), 0o755)
	m := newTestManager(t, projectRoot)

	ctx := ConflictContext{
		TaskID:  "t-bad",
		RigName: "nonexistent-rig",
	}

	_, err := m.SpawnResolver(ctx)
	if err == nil {
		t.Fatal("expected error for nonexistent rig")
	}
}

func TestCleanupResolver(t *testing.T) {
	if _, err := tmux.ListSessions(); err != nil {
		t.Skip("tmux not available")
	}

	projectRoot, _, conflictBranch := setupProject(t)
	m := newTestManager(t, projectRoot)
	ctx := sampleConflictContext(conflictBranch)

	a, err := m.SpawnResolver(ctx)
	if err != nil {
		t.Fatalf("SpawnResolver: %v", err)
	}

	worktreePath := a.Worktree
	sessionName := a.TmuxSession

	if err := m.CleanupResolver(a); err != nil {
		t.Fatalf("CleanupResolver: %v", err)
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

func TestListResolvers(t *testing.T) {
	projectRoot := t.TempDir()
	m := newTestManager(t, projectRoot)

	agents := []struct {
		id   string
		role agent.Role
	}{
		{"resolver-02", agent.RoleResolver},
		{"resolver-01", agent.RoleResolver},
		{"worker-01", agent.RoleWorker},
		{"resolver-03", agent.RoleResolver},
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

	resolvers, err := m.ListResolvers()
	if err != nil {
		t.Fatalf("ListResolvers: %v", err)
	}
	if len(resolvers) != 3 {
		t.Fatalf("ListResolvers = %d agents, want 3", len(resolvers))
	}

	expectedIDs := []string{"resolver-01", "resolver-02", "resolver-03"}
	for i, r := range resolvers {
		if r.ID != expectedIDs[i] {
			t.Errorf("resolver[%d].ID = %q, want %q", i, r.ID, expectedIDs[i])
		}
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
