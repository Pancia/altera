// Package git provides operations on git repositories by shelling out to
// the git binary. It supports worktree management, branch operations, merging,
// status checks, author configuration, and commit/push workflows.
package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Sentinel errors for common git failure modes.
var (
	ErrNotRepo     = errors.New("not a git repository")
	ErrBranchExists = errors.New("branch already exists")
	ErrNotClean    = errors.New("working tree is not clean")
	ErrConflict    = errors.New("merge conflict")
)

// run executes a git command in the given directory and returns its
// combined stdout. If the command fails, the error includes stderr.
func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %s: %w", strings.Join(args, " "), strings.TrimSpace(stderr.String()), err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// --- Worktree Operations ---

// CreateWorktree creates a new git worktree at path for the given branch.
// The branch must already exist.
func CreateWorktree(repo, branch, path string) error {
	_, err := run(repo, "worktree", "add", path, branch)
	if err != nil {
		return fmt.Errorf("creating worktree: %w", err)
	}
	return nil
}

// DeleteWorktree removes a git worktree at the given path. It calls
// 'git worktree remove --force' from the worktree's main repository.
func DeleteWorktree(repo, path string) error {
	_, err := run(repo, "worktree", "remove", "--force", path)
	if err != nil {
		return fmt.Errorf("deleting worktree: %w", err)
	}
	return nil
}

// --- Branch Operations ---

// CreateBranch creates a new branch pointing at base. If base is empty,
// it defaults to HEAD.
func CreateBranch(repo, name, base string) error {
	args := []string{"branch", name}
	if base != "" {
		args = append(args, base)
	}
	_, err := run(repo, args...)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("creating branch %q: %w", name, ErrBranchExists)
		}
		return fmt.Errorf("creating branch %q: %w", name, err)
	}
	return nil
}

// DeleteBranch deletes a local branch. It uses -D (force delete) so
// unmerged branches can also be removed.
func DeleteBranch(repo, name string) error {
	_, err := run(repo, "branch", "-D", name)
	if err != nil {
		return fmt.Errorf("deleting branch %q: %w", name, err)
	}
	return nil
}

// CurrentBranch returns the name of the currently checked-out branch
// in the given working directory.
func CurrentBranch(path string) (string, error) {
	out, err := run(path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("getting current branch: %w", err)
	}
	return out, nil
}

// --- Merge ---

// MergeResult holds the outcome of a merge operation.
type MergeResult struct {
	Clean     bool
	Conflicts []string
}

// Merge merges the given branch into the currently checked-out branch in
// the working directory at path. It returns whether the merge was clean
// and any conflicting file paths.
func Merge(path, branch string) (MergeResult, error) {
	_, err := run(path, "merge", "--no-edit", branch)
	if err == nil {
		return MergeResult{Clean: true}, nil
	}

	// Check if we have conflicts vs a hard failure.
	statusOut, statusErr := run(path, "diff", "--name-only", "--diff-filter=U")
	if statusErr != nil {
		return MergeResult{}, fmt.Errorf("merging branch %q: %w", branch, err)
	}
	if statusOut == "" {
		// No unmerged files means this was a non-conflict error.
		return MergeResult{}, fmt.Errorf("merging branch %q: %w", branch, err)
	}

	conflicts := strings.Split(statusOut, "\n")
	return MergeResult{
		Clean:     false,
		Conflicts: conflicts,
	}, nil
}

// AbortMerge aborts an in-progress merge.
func AbortMerge(path string) error {
	_, err := run(path, "merge", "--abort")
	if err != nil {
		return fmt.Errorf("aborting merge: %w", err)
	}
	return nil
}

// --- Status ---

// IsClean returns true if the working tree and index have no modifications,
// no untracked files, and no staged changes.
func IsClean(path string) (bool, error) {
	out, err := run(path, "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("checking clean status: %w", err)
	}
	return out == "", nil
}

// HasUncommittedChanges returns true if there are staged or unstaged changes
// (but ignores untracked files).
func HasUncommittedChanges(path string) (bool, error) {
	// Check for staged changes.
	_, stagedErr := run(path, "diff", "--cached", "--quiet")
	if stagedErr != nil {
		return true, nil
	}
	// Check for unstaged changes.
	_, unstagedErr := run(path, "diff", "--quiet")
	if unstagedErr != nil {
		return true, nil
	}
	return false, nil
}

// --- Author Identity ---

// SetAuthor configures the git user.name and user.email for the repository
// at the given path using local (per-repo) config.
func SetAuthor(path, name, email string) error {
	if _, err := run(path, "config", "user.name", name); err != nil {
		return fmt.Errorf("setting author name: %w", err)
	}
	if _, err := run(path, "config", "user.email", email); err != nil {
		return fmt.Errorf("setting author email: %w", err)
	}
	return nil
}

// --- Commit Operations ---

// Add stages the given files for commit. If files is empty, it stages
// all changes (git add -A).
func Add(path string, files []string) error {
	args := []string{"add"}
	if len(files) == 0 {
		args = append(args, "-A")
	} else {
		args = append(args, "--")
		args = append(args, files...)
	}
	_, err := run(path, args...)
	if err != nil {
		return fmt.Errorf("staging files: %w", err)
	}
	return nil
}

// Commit creates a commit with the given message. The working directory
// must have staged changes.
func Commit(path, message string) error {
	_, err := run(path, "commit", "-m", message)
	if err != nil {
		return fmt.Errorf("committing: %w", err)
	}
	return nil
}

// Push pushes the given branch to the remote.
func Push(path, remote, branch string) error {
	_, err := run(path, "push", remote, branch)
	if err != nil {
		return fmt.Errorf("pushing: %w", err)
	}
	return nil
}

// --- Utility ---

// Init initializes a new git repository at path.
func Init(path string) error {
	_, err := run(path, "init")
	if err != nil {
		return fmt.Errorf("initializing repository: %w", err)
	}
	return nil
}

// Checkout switches the working tree at path to the given branch.
func Checkout(path, branch string) error {
	_, err := run(path, "checkout", branch)
	if err != nil {
		return fmt.Errorf("checking out %q: %w", branch, err)
	}
	return nil
}

// Log returns the one-line log of the last n commits in the repo at path.
func Log(path string, n int) (string, error) {
	out, err := run(path, "log", "--oneline", fmt.Sprintf("-n%d", n))
	if err != nil {
		return "", fmt.Errorf("reading log: %w", err)
	}
	return out, nil
}

// Rev returns the full commit hash of the given revision.
func Rev(path, rev string) (string, error) {
	out, err := run(path, "rev-parse", rev)
	if err != nil {
		return "", fmt.Errorf("resolving rev %q: %w", rev, err)
	}
	return out, nil
}
