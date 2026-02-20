package merge

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/events"
	"github.com/anthropics/altera/internal/git"
	"github.com/anthropics/altera/internal/message"
	"github.com/anthropics/altera/internal/task"
)

// --- Test helpers ---

// initRepo creates a temporary git repo with an initial commit.
func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := git.Init(dir); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := git.SetAuthor(dir, "Test", "test@example.com"); err != nil {
		t.Fatalf("SetAuthor: %v", err)
	}
	writeFile(t, dir, "README.md", "# test\n")
	if err := git.Add(dir, nil); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := git.Commit(dir, "initial commit"); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	return dir
}

// writeFile creates or overwrites a file relative to dir.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// defaultBranch returns the main branch name of the repo.
func defaultBranch(t *testing.T, repo string) string {
	t.Helper()
	br, err := git.CurrentBranch(repo)
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	return br
}

// runGit runs a git command in the given directory.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %s: %v", args, out, err)
	}
}

// initRepoWithRemote creates a repo and a bare-like remote, returning both paths.
func initRepoWithRemote(t *testing.T) (repo string, mainBranch string) {
	t.Helper()
	repo = initRepo(t)
	mainBranch = defaultBranch(t, repo)

	remote := t.TempDir()
	git.Init(remote)
	runGit(t, remote, "config", "receive.denyCurrentBranch", "ignore")
	runGit(t, repo, "remote", "add", "origin", remote)
	runGit(t, repo, "push", "-u", "origin", mainBranch)
	return repo, mainBranch
}

// testPipeline creates a Pipeline with temp-backed dependencies for testing.
func testPipeline(t *testing.T, root string) (*Pipeline, *events.Reader) {
	t.Helper()
	altDir := filepath.Join(root, ".alt")
	os.MkdirAll(altDir, 0o755)

	taskStore, err := task.NewStore(root)
	if err != nil {
		t.Fatalf("NewStore(task): %v", err)
	}

	evPath := filepath.Join(altDir, "events.jsonl")
	evWriter := events.NewWriter(evPath)
	evReader := events.NewReader(evPath)

	msgDir := filepath.Join(altDir, "messages")
	msgStore, err := message.NewStore(msgDir)
	if err != nil {
		t.Fatalf("NewStore(message): %v", err)
	}

	queueDir := filepath.Join(altDir, "merge-queue")
	q, err := NewQueue(queueDir)
	if err != nil {
		t.Fatalf("NewQueue: %v", err)
	}

	p := NewPipeline(taskStore, evWriter, msgStore, q)
	return p, evReader
}

// createTask creates a task with the given ID, branch, and assignee in the store.
func createTask(t *testing.T, root, taskID, branch, assignee string) {
	t.Helper()
	store, err := task.NewStore(root)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	tk := &task.Task{
		ID:         taskID,
		Title:      "Test task " + taskID,
		Branch:     branch,
		AssignedTo: assignee,
	}
	if err := store.Create(tk); err != nil {
		t.Fatalf("Create task: %v", err)
	}
}

// --- ExtractConflicts ---

func TestExtractConflicts_SingleConflict(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "conflict.txt")
	content := `line 1
line 2
<<<<<<< HEAD
our content
=======
their content
>>>>>>> feature
line 8
`
	os.WriteFile(path, []byte(content), 0o644)

	info := ExtractConflicts(path)
	if len(info.Markers) != 1 {
		t.Fatalf("expected 1 marker, got %d", len(info.Markers))
	}

	m := info.Markers[0]
	if m.OursStart != 3 {
		t.Errorf("OursStart = %d, want 3", m.OursStart)
	}
	if m.OursEnd != 5 {
		t.Errorf("OursEnd = %d, want 5", m.OursEnd)
	}
	if m.TheirsStart != 5 {
		t.Errorf("TheirsStart = %d, want 5", m.TheirsStart)
	}
	if m.TheirsEnd != 7 {
		t.Errorf("TheirsEnd = %d, want 7", m.TheirsEnd)
	}
}

func TestExtractConflicts_MultipleConflicts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "multi.txt")
	content := `<<<<<<< HEAD
ours 1
=======
theirs 1
>>>>>>> branch
middle
<<<<<<< HEAD
ours 2
=======
theirs 2
>>>>>>> branch
`
	os.WriteFile(path, []byte(content), 0o644)

	info := ExtractConflicts(path)
	if len(info.Markers) != 2 {
		t.Fatalf("expected 2 markers, got %d", len(info.Markers))
	}

	if info.Markers[0].OursStart != 1 {
		t.Errorf("first conflict OursStart = %d, want 1", info.Markers[0].OursStart)
	}
	if info.Markers[1].OursStart != 7 {
		t.Errorf("second conflict OursStart = %d, want 7", info.Markers[1].OursStart)
	}
}

func TestExtractConflicts_NoConflicts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "clean.txt")
	os.WriteFile(path, []byte("no conflicts here\n"), 0o644)

	info := ExtractConflicts(path)
	if len(info.Markers) != 0 {
		t.Errorf("expected 0 markers, got %d", len(info.Markers))
	}
}

func TestExtractConflicts_NonexistentFile(t *testing.T) {
	info := ExtractConflicts("/nonexistent/path/file.txt")
	if len(info.Markers) != 0 {
		t.Errorf("expected 0 markers for nonexistent file, got %d", len(info.Markers))
	}
}

func TestExtractConflicts_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	os.WriteFile(path, []byte(""), 0o644)

	info := ExtractConflicts(path)
	if len(info.Markers) != 0 {
		t.Errorf("expected 0 markers for empty file, got %d", len(info.Markers))
	}
}

// --- Pipeline ---

func TestNewPipeline(t *testing.T) {
	root := t.TempDir()
	p, _ := testPipeline(t, root)
	if p == nil {
		t.Fatal("expected non-nil Pipeline")
	}
}

// --- AttemptMerge: success ---

func TestAttemptMerge_Success(t *testing.T) {
	repo, mainBranch := initRepoWithRemote(t)

	// Create feature branch with a new file.
	git.CreateBranch(repo, "feature", "")
	git.Checkout(repo, "feature")
	writeFile(t, repo, "feature.go", "package feature\n")
	git.Add(repo, nil)
	git.Commit(repo, "add feature")
	git.Checkout(repo, mainBranch)

	root := t.TempDir()
	p, evReader := testPipeline(t, root)
	createTask(t, root, "t-succ01", "feature", "worker-1")

	rig := config.RigConfig{
		RepoPath:      repo,
		DefaultBranch: mainBranch,
		TestCommand:   "true", // always passes
	}

	result, err := p.AttemptMerge("t-succ01", rig, repo)
	if err != nil {
		t.Fatalf("AttemptMerge: %v", err)
	}
	if result.Outcome != OutcomeSuccess {
		t.Errorf("Outcome = %q, want %q", result.Outcome, OutcomeSuccess)
	}

	// Verify events were emitted.
	all, err := evReader.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll events: %v", err)
	}
	hasStarted := false
	hasSuccess := false
	for _, ev := range all {
		if ev.Type == events.MergeStarted {
			hasStarted = true
		}
		if ev.Type == events.MergeSuccess {
			hasSuccess = true
		}
	}
	if !hasStarted {
		t.Error("expected merge_started event")
	}
	if !hasSuccess {
		t.Error("expected merge_success event")
	}
}

// --- AttemptMerge: conflict ---

func TestAttemptMerge_Conflict(t *testing.T) {
	repo := initRepo(t)
	mainBranch := defaultBranch(t, repo)

	// Create feature branch that conflicts.
	git.CreateBranch(repo, "conflict-br", "")
	git.Checkout(repo, "conflict-br")
	writeFile(t, repo, "README.md", "conflict branch content\n")
	git.Add(repo, nil)
	git.Commit(repo, "conflict change")

	// Make conflicting change on main.
	git.Checkout(repo, mainBranch)
	writeFile(t, repo, "README.md", "main branch content\n")
	git.Add(repo, nil)
	git.Commit(repo, "main change")

	root := t.TempDir()
	p, evReader := testPipeline(t, root)
	createTask(t, root, "t-conf01", "conflict-br", "worker-1")

	rig := config.RigConfig{
		RepoPath:      repo,
		DefaultBranch: mainBranch,
	}

	result, err := p.AttemptMerge("t-conf01", rig, repo)
	if err != nil {
		t.Fatalf("AttemptMerge: %v", err)
	}
	if result.Outcome != OutcomeConflict {
		t.Errorf("Outcome = %q, want %q", result.Outcome, OutcomeConflict)
	}
	if len(result.Conflicts) == 0 {
		t.Error("expected at least one conflict")
	}

	// Verify merge_conflict event was emitted.
	all, _ := evReader.ReadAll()
	hasConflict := false
	for _, ev := range all {
		if ev.Type == events.MergeConflict {
			hasConflict = true
		}
	}
	if !hasConflict {
		t.Error("expected merge_conflict event")
	}

	// Worktree should be clean after abort.
	clean, err := git.IsClean(repo)
	if err != nil {
		t.Fatalf("IsClean: %v", err)
	}
	if !clean {
		t.Error("expected clean worktree after conflict abort")
	}
}

// --- AttemptMerge: test failure ---

func TestAttemptMerge_TestFailure(t *testing.T) {
	repo, mainBranch := initRepoWithRemote(t)

	// Create a clean feature branch.
	git.CreateBranch(repo, "test-fail-br", "")
	git.Checkout(repo, "test-fail-br")
	writeFile(t, repo, "new.txt", "new content\n")
	git.Add(repo, nil)
	git.Commit(repo, "add new file")
	git.Checkout(repo, mainBranch)

	root := t.TempDir()
	p, evReader := testPipeline(t, root)
	createTask(t, root, "t-tfail1", "test-fail-br", "worker-1")

	rig := config.RigConfig{
		RepoPath:      repo,
		DefaultBranch: mainBranch,
		TestCommand:   "exit 1", // always fails
	}

	result, err := p.AttemptMerge("t-tfail1", rig, repo)
	if err != nil {
		t.Fatalf("AttemptMerge: %v", err)
	}
	if result.Outcome != OutcomeTestFailure {
		t.Errorf("Outcome = %q, want %q", result.Outcome, OutcomeTestFailure)
	}

	// Verify merge_failed event was emitted.
	all, _ := evReader.ReadAll()
	hasFailed := false
	for _, ev := range all {
		if ev.Type == events.MergeFailed {
			hasFailed = true
		}
	}
	if !hasFailed {
		t.Error("expected merge_failed event")
	}
}

// --- AttemptMerge: no branch ---

func TestAttemptMerge_NoBranch(t *testing.T) {
	root := t.TempDir()
	p, _ := testPipeline(t, root)

	// Create task with no branch.
	store, _ := task.NewStore(root)
	store.Create(&task.Task{ID: "t-nobr01", Title: "no branch"})

	rig := config.RigConfig{DefaultBranch: "main"}

	_, err := p.AttemptMerge("t-nobr01", rig, t.TempDir())
	if err == nil {
		t.Fatal("expected error for task with no branch")
	}
}

// --- AttemptMerge: nonexistent task ---

func TestAttemptMerge_NonexistentTask(t *testing.T) {
	root := t.TempDir()
	p, _ := testPipeline(t, root)

	rig := config.RigConfig{DefaultBranch: "main"}

	_, err := p.AttemptMerge("t-nonex1", rig, t.TempDir())
	if err == nil {
		t.Fatal("expected error for nonexistent task")
	}
}

// --- AttemptMerge: no test command (skip tests) ---

func TestAttemptMerge_NoTestCommand(t *testing.T) {
	repo, mainBranch := initRepoWithRemote(t)

	// Create feature branch.
	git.CreateBranch(repo, "no-test-br", "")
	git.Checkout(repo, "no-test-br")
	writeFile(t, repo, "feature.go", "package feature\n")
	git.Add(repo, nil)
	git.Commit(repo, "add feature")
	git.Checkout(repo, mainBranch)

	root := t.TempDir()
	p, _ := testPipeline(t, root)
	createTask(t, root, "t-notest", "no-test-br", "worker-1")

	rig := config.RigConfig{
		RepoPath:      repo,
		DefaultBranch: mainBranch,
		TestCommand:   "", // no test command
	}

	result, err := p.AttemptMerge("t-notest", rig, repo)
	if err != nil {
		t.Fatalf("AttemptMerge: %v", err)
	}
	if result.Outcome != OutcomeSuccess {
		t.Errorf("Outcome = %q, want %q (should succeed without tests)", result.Outcome, OutcomeSuccess)
	}
}

// --- AttemptMerge: messages ---

func TestAttemptMerge_SendsMessage(t *testing.T) {
	repo, mainBranch := initRepoWithRemote(t)

	git.CreateBranch(repo, "msg-br", "")
	git.Checkout(repo, "msg-br")
	writeFile(t, repo, "msg.txt", "content\n")
	git.Add(repo, nil)
	git.Commit(repo, "add msg file")
	git.Checkout(repo, mainBranch)

	root := t.TempDir()
	p, _ := testPipeline(t, root)
	createTask(t, root, "t-msg001", "msg-br", "worker-1")

	rig := config.RigConfig{
		RepoPath:      repo,
		DefaultBranch: mainBranch,
		TestCommand:   "true",
	}

	p.AttemptMerge("t-msg001", rig, repo)

	// Check that a merge_result message was sent to the worker.
	msgDir := filepath.Join(root, ".alt", "messages")
	msgStore, _ := message.NewStore(msgDir)
	msgs, err := msgStore.ListPending("worker-1")
	if err != nil {
		t.Fatalf("ListPending: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message to worker-1, got %d", len(msgs))
	}
	if msgs[0].Type != message.TypeMergeResult {
		t.Errorf("message type = %q, want %q", msgs[0].Type, message.TypeMergeResult)
	}
	if msgs[0].Payload["outcome"] != string(OutcomeSuccess) {
		t.Errorf("payload outcome = %v, want %q", msgs[0].Payload["outcome"], OutcomeSuccess)
	}
}

// --- runTests helper ---

func TestRunTests_Pass(t *testing.T) {
	dir := t.TempDir()
	output, err := runTests(dir, "echo ok")
	if err != nil {
		t.Fatalf("expected tests to pass, got error: %v", err)
	}
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestRunTests_Fail(t *testing.T) {
	dir := t.TempDir()
	_, err := runTests(dir, "exit 1")
	if err == nil {
		t.Fatal("expected tests to fail")
	}
}

func TestRunTests_CapturesOutput(t *testing.T) {
	dir := t.TempDir()
	output, _ := runTests(dir, "echo 'test output' && exit 1")
	if output == "" {
		t.Error("expected output to be captured even on failure")
	}
}

// --- Outcome constants ---

func TestOutcomeValues(t *testing.T) {
	if OutcomeSuccess != "success" {
		t.Errorf("OutcomeSuccess = %q", OutcomeSuccess)
	}
	if OutcomeConflict != "conflict" {
		t.Errorf("OutcomeConflict = %q", OutcomeConflict)
	}
	if OutcomeTestFailure != "test_failure" {
		t.Errorf("OutcomeTestFailure = %q", OutcomeTestFailure)
	}
}

// --- Integration: queue + pipeline ---

func TestQueueThenMerge(t *testing.T) {
	repo, mainBranch := initRepoWithRemote(t)

	git.CreateBranch(repo, "queued-br", "")
	git.Checkout(repo, "queued-br")
	writeFile(t, repo, "queued.txt", "content\n")
	git.Add(repo, nil)
	git.Commit(repo, "queued change")
	git.Checkout(repo, mainBranch)

	root := t.TempDir()
	p, _ := testPipeline(t, root)
	createTask(t, root, "t-queued", "queued-br", "worker-1")

	// Enqueue, then dequeue and merge.
	p.queue.Enqueue("t-queued")
	time.Sleep(time.Millisecond)
	taskID, _ := p.queue.Dequeue()

	rig := config.RigConfig{
		RepoPath:      repo,
		DefaultBranch: mainBranch,
		TestCommand:   "true",
	}

	result, err := p.AttemptMerge(taskID, rig, repo)
	if err != nil {
		t.Fatalf("AttemptMerge: %v", err)
	}
	if result.Outcome != OutcomeSuccess {
		t.Errorf("Outcome = %q, want %q", result.Outcome, OutcomeSuccess)
	}
}
