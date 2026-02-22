# Code Review Area 6: Git & Merge Operations

## Why This Matters

Branch management, worktrees, and merge conflict resolution are complex and failure-prone. Dangling branches and worktrees waste disk space and can confuse the system. The merge flow is multi-step with several failure points.

## Files to Review

- `internal/git/git.go` - Core git operations (worktrees, branches, merge, status)
- `internal/merge/merge.go` + `queue.go` - Merge logic and FIFO queue
- `internal/resolver/resolver.go` - Conflict resolution agent spawning

## What to Check

- Worktree cleanup on all failure paths (spawn failure, worker death, merge failure)
- Branch cleanup: are branches deleted when workers finish or die?
- `git add -A` usage: could this stage sensitive files (`.env`, credentials)?
- Merge conflict detection accuracy: parsing git output for conflict file list
- Resolver completion detection: checking for `<<<<<<<` markers
  - What about binary file conflicts? Files with literal `<<<<<<<` in content?
- What happens if the base branch moves forward during a long-running merge?
- Force delete (`-D`) on branch cleanup: is this always safe?
- `DeleteWorktree` uses `--force`: what state could be lost?
- Queue ordering: filesystem timestamp granularity and sort correctness
- Merge queue persistence: what if queue file is corrupt?

## Severity Estimate

**MEDIUM** - Git operations are generally well-understood, but edge cases in conflict resolution and cleanup could cause state drift.
