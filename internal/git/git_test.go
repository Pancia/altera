package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// initRepo creates a temporary bare-bones git repo with an initial commit
// and returns its path.
func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := Init(dir); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if err := SetAuthor(dir, "Test", "test@example.com"); err != nil {
		t.Fatalf("SetAuthor: %v", err)
	}
	// Create an initial commit so HEAD exists.
	writeFile(t, dir, "README.md", "# test\n")
	if err := Add(dir, nil); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := Commit(dir, "initial commit"); err != nil {
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

// --- Init ---

func TestInit(t *testing.T) {
	dir := t.TempDir()
	if err := Init(dir); err != nil {
		t.Fatalf("Init: %v", err)
	}
	// .git directory should exist.
	info, err := os.Stat(filepath.Join(dir, ".git"))
	if err != nil {
		t.Fatalf(".git not found: %v", err)
	}
	if !info.IsDir() {
		t.Fatal(".git is not a directory")
	}
}

// --- Branch Operations ---

func TestCreateBranch(t *testing.T) {
	repo := initRepo(t)

	if err := CreateBranch(repo, "feature", ""); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}
	// Verify branch exists by checking it out.
	if err := Checkout(repo, "feature"); err != nil {
		t.Fatalf("Checkout feature: %v", err)
	}
	br, err := CurrentBranch(repo)
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	if br != "feature" {
		t.Errorf("branch = %q, want %q", br, "feature")
	}
}

func TestCreateBranch_FromBase(t *testing.T) {
	repo := initRepo(t)

	// Create a second commit on main so we can verify base works.
	writeFile(t, repo, "second.txt", "second")
	Add(repo, nil)
	Commit(repo, "second commit")

	headRev, _ := Rev(repo, "HEAD")
	firstRev, _ := Rev(repo, "HEAD~1")

	// Create branch from the first commit.
	if err := CreateBranch(repo, "from-first", "HEAD~1"); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}
	branchRev, _ := Rev(repo, "from-first")
	if branchRev != firstRev {
		t.Errorf("branch points at %s, want %s (not %s)", branchRev[:8], firstRev[:8], headRev[:8])
	}
}

func TestCreateBranch_Duplicate(t *testing.T) {
	repo := initRepo(t)

	if err := CreateBranch(repo, "dup", ""); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}
	err := CreateBranch(repo, "dup", "")
	if err == nil {
		t.Fatal("expected error for duplicate branch")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists', got: %v", err)
	}
}

func TestDeleteBranch(t *testing.T) {
	repo := initRepo(t)

	CreateBranch(repo, "to-delete", "")
	if err := DeleteBranch(repo, "to-delete"); err != nil {
		t.Fatalf("DeleteBranch: %v", err)
	}
	// Trying to checkout should fail.
	err := Checkout(repo, "to-delete")
	if err == nil {
		t.Fatal("expected error checking out deleted branch")
	}
}

func TestDeleteBranch_NonExistent(t *testing.T) {
	repo := initRepo(t)
	err := DeleteBranch(repo, "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent branch")
	}
}

func TestCurrentBranch(t *testing.T) {
	repo := initRepo(t)
	br, err := CurrentBranch(repo)
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	// Default branch name varies; just ensure non-empty.
	if br == "" {
		t.Fatal("expected non-empty branch name")
	}
}

// --- Worktree Operations ---

func TestCreateAndDeleteWorktree(t *testing.T) {
	repo := initRepo(t)

	// Create a branch for the worktree.
	if err := CreateBranch(repo, "wt-branch", ""); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}

	wtPath := filepath.Join(t.TempDir(), "worktree")
	if err := CreateWorktree(repo, "wt-branch", wtPath); err != nil {
		t.Fatalf("CreateWorktree: %v", err)
	}

	// Verify the worktree is functional.
	br, err := CurrentBranch(wtPath)
	if err != nil {
		t.Fatalf("CurrentBranch in worktree: %v", err)
	}
	if br != "wt-branch" {
		t.Errorf("worktree branch = %q, want %q", br, "wt-branch")
	}

	// Files from main should be present.
	if _, err := os.Stat(filepath.Join(wtPath, "README.md")); err != nil {
		t.Error("expected README.md in worktree")
	}

	// Delete worktree.
	if err := DeleteWorktree(repo, wtPath); err != nil {
		t.Fatalf("DeleteWorktree: %v", err)
	}

	// Path should no longer exist.
	if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
		t.Error("expected worktree path to be removed")
	}
}

// --- Status ---

func TestIsClean(t *testing.T) {
	repo := initRepo(t)

	clean, err := IsClean(repo)
	if err != nil {
		t.Fatalf("IsClean: %v", err)
	}
	if !clean {
		t.Error("expected clean repo after initial commit")
	}

	// Make it dirty.
	writeFile(t, repo, "dirty.txt", "dirty")
	clean, err = IsClean(repo)
	if err != nil {
		t.Fatalf("IsClean: %v", err)
	}
	if clean {
		t.Error("expected dirty repo after adding file")
	}
}

func TestIsClean_StagedChanges(t *testing.T) {
	repo := initRepo(t)

	writeFile(t, repo, "staged.txt", "staged")
	Add(repo, []string{"staged.txt"})

	clean, err := IsClean(repo)
	if err != nil {
		t.Fatalf("IsClean: %v", err)
	}
	if clean {
		t.Error("expected dirty repo with staged changes")
	}
}

func TestHasUncommittedChanges(t *testing.T) {
	repo := initRepo(t)

	has, err := HasUncommittedChanges(repo)
	if err != nil {
		t.Fatalf("HasUncommittedChanges: %v", err)
	}
	if has {
		t.Error("expected no uncommitted changes after initial commit")
	}

	// Modify tracked file.
	writeFile(t, repo, "README.md", "modified\n")
	has, err = HasUncommittedChanges(repo)
	if err != nil {
		t.Fatalf("HasUncommittedChanges: %v", err)
	}
	if !has {
		t.Error("expected uncommitted changes after modifying tracked file")
	}
}

func TestHasUncommittedChanges_Staged(t *testing.T) {
	repo := initRepo(t)

	writeFile(t, repo, "new.txt", "new")
	Add(repo, []string{"new.txt"})

	has, err := HasUncommittedChanges(repo)
	if err != nil {
		t.Fatalf("HasUncommittedChanges: %v", err)
	}
	if !has {
		t.Error("expected uncommitted changes with staged files")
	}
}

func TestHasUncommittedChanges_UntrackedOnly(t *testing.T) {
	repo := initRepo(t)

	// Untracked files should not count as uncommitted changes.
	writeFile(t, repo, "untracked.txt", "untracked")
	has, err := HasUncommittedChanges(repo)
	if err != nil {
		t.Fatalf("HasUncommittedChanges: %v", err)
	}
	if has {
		t.Error("expected no uncommitted changes with only untracked files")
	}
}

// --- Author Identity ---

func TestSetAuthor(t *testing.T) {
	repo := initRepo(t)

	if err := SetAuthor(repo, "Alice", "alice@example.com"); err != nil {
		t.Fatalf("SetAuthor: %v", err)
	}

	name, err := run(repo, "config", "user.name")
	if err != nil {
		t.Fatalf("getting user.name: %v", err)
	}
	if name != "Alice" {
		t.Errorf("user.name = %q, want %q", name, "Alice")
	}

	email, err := run(repo, "config", "user.email")
	if err != nil {
		t.Fatalf("getting user.email: %v", err)
	}
	if email != "alice@example.com" {
		t.Errorf("user.email = %q, want %q", email, "alice@example.com")
	}
}

// --- Commit Operations ---

func TestAddAndCommit(t *testing.T) {
	repo := initRepo(t)

	writeFile(t, repo, "feature.go", "package feature\n")
	if err := Add(repo, []string{"feature.go"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := Commit(repo, "add feature"); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	clean, _ := IsClean(repo)
	if !clean {
		t.Error("expected clean after commit")
	}

	logOut, err := Log(repo, 1)
	if err != nil {
		t.Fatalf("Log: %v", err)
	}
	if !strings.Contains(logOut, "add feature") {
		t.Errorf("log should contain commit message, got %q", logOut)
	}
}

func TestAdd_AllFiles(t *testing.T) {
	repo := initRepo(t)

	writeFile(t, repo, "a.txt", "a")
	writeFile(t, repo, "b.txt", "b")
	// Add with nil stages everything.
	if err := Add(repo, nil); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := Commit(repo, "add all"); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	clean, _ := IsClean(repo)
	if !clean {
		t.Error("expected clean after adding all and committing")
	}
}

func TestCommit_NothingStaged(t *testing.T) {
	repo := initRepo(t)
	err := Commit(repo, "nothing")
	if err == nil {
		t.Fatal("expected error when committing with nothing staged")
	}
}

// --- Merge ---

func TestMerge_CleanMerge(t *testing.T) {
	repo := initRepo(t)

	// Create and switch to feature branch.
	CreateBranch(repo, "feature", "")
	Checkout(repo, "feature")
	writeFile(t, repo, "feature.txt", "feature content")
	Add(repo, nil)
	Commit(repo, "add feature file")

	// Switch back to main branch and merge.
	mainBranch := defaultBranch(t, repo)
	Checkout(repo, mainBranch)

	result, err := Merge(repo, "feature")
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if !result.Clean {
		t.Errorf("expected clean merge, got conflicts: %v", result.Conflicts)
	}

	// Feature file should now exist on main.
	if _, err := os.Stat(filepath.Join(repo, "feature.txt")); err != nil {
		t.Error("expected feature.txt after merge")
	}
}

func TestMerge_Conflict(t *testing.T) {
	repo := initRepo(t)
	mainBranch := defaultBranch(t, repo)

	// Create feature branch and modify README.
	CreateBranch(repo, "conflict", "")
	Checkout(repo, "conflict")
	writeFile(t, repo, "README.md", "conflict branch content\n")
	Add(repo, nil)
	Commit(repo, "conflict change")

	// Switch to main and make a conflicting change.
	Checkout(repo, mainBranch)
	writeFile(t, repo, "README.md", "main branch content\n")
	Add(repo, nil)
	Commit(repo, "main change")

	result, err := Merge(repo, "conflict")
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if result.Clean {
		t.Error("expected conflict, got clean merge")
	}
	if len(result.Conflicts) == 0 {
		t.Error("expected at least one conflict file")
	}

	found := false
	for _, f := range result.Conflicts {
		if f == "README.md" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected README.md in conflicts, got %v", result.Conflicts)
	}

	// Clean up the merge state.
	if err := AbortMerge(repo); err != nil {
		t.Fatalf("AbortMerge: %v", err)
	}
}

// --- Rev ---

func TestRev(t *testing.T) {
	repo := initRepo(t)

	rev, err := Rev(repo, "HEAD")
	if err != nil {
		t.Fatalf("Rev: %v", err)
	}
	// Full SHA-1 is 40 hex chars.
	if len(rev) != 40 {
		t.Errorf("expected 40-char SHA, got %d chars: %q", len(rev), rev)
	}
}

func TestRev_InvalidRef(t *testing.T) {
	repo := initRepo(t)
	_, err := Rev(repo, "nonexistent-ref-xyz")
	if err == nil {
		t.Fatal("expected error for invalid ref")
	}
}

// --- Log ---

func TestLog(t *testing.T) {
	repo := initRepo(t)

	writeFile(t, repo, "a.txt", "a")
	Add(repo, nil)
	Commit(repo, "second commit")

	out, err := Log(repo, 2)
	if err != nil {
		t.Fatalf("Log: %v", err)
	}
	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 log lines, got %d: %q", len(lines), out)
	}
}

// --- Checkout ---

func TestCheckout(t *testing.T) {
	repo := initRepo(t)

	CreateBranch(repo, "other", "")
	if err := Checkout(repo, "other"); err != nil {
		t.Fatalf("Checkout: %v", err)
	}
	br, _ := CurrentBranch(repo)
	if br != "other" {
		t.Errorf("branch = %q, want %q", br, "other")
	}
}

func TestCheckout_NonExistent(t *testing.T) {
	repo := initRepo(t)
	err := Checkout(repo, "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent branch")
	}
}

// --- Helpers ---

// defaultBranch returns the current branch right after initRepo, which
// may be "main" or "master" depending on git config.
func defaultBranch(t *testing.T, repo string) string {
	t.Helper()
	br, err := CurrentBranch(repo)
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	return br
}
