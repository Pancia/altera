# Code Review Area 2: Error Handling & Recovery

## Why This Matters

Several cleanup paths silently discard errors. Failed worker spawns could leave partial state (tmux session without agent record, worktree without tmux session, etc.). Reconciliation at startup is the safety net - but does it cover all failure modes?

## Files to Review

- `internal/daemon/daemon.go` - Ignored `os.Remove` errors, cleanup `_` assignments in `spawnWorker()`
- `internal/worker/worker.go` - Silent cleanup failures in teardown
- `internal/tmux/tmux.go` - Ignored tmux config errors (mouse mode, history limit)
- `internal/git/git.go` - No timeouts on subprocess execution, error wrapping quality
- `internal/resolver/resolver.go` - Resolver spawn failure cleanup

## What to Check

- Every `_ =` assignment - is the error actually ignorable?
- Partial state from failed `spawnWorker()`: what cleanup runs on each failure path?
  - tmux session created but agent record not written
  - worktree created but tmux session fails
  - agent registered but task assignment fails
- Reconciliation logic completeness - does startup cleanup cover all partial-state scenarios?
- Whether `os.Remove` failures in merge queue can cause repeated message processing
- Error wrapping consistency - are all errors wrapped with sufficient context?
- Whether errors from deferred cleanup functions are ever surfaced

## Preliminary Findings

- `daemon.go:277` - `os.Remove(path)` in `reconcileMergeQueue()` ignores error
- `daemon.go:933, 979, 992` - `os.Remove(itemPath)` in `processMergeQueue()` ignores error
- `daemon.go:680, 687, 715-717` - Multiple cleanup calls use `_` in `spawnWorker()`
- `worker.go:116, 122-123, 172, 196` - Cleanup and logging errors discarded
- `tmux.go:41, 44` - Mouse mode and history limit errors ignored

## Severity Estimate

**MEDIUM** - Most critical operations are handled, but edge cases could accumulate orphaned resources over time.
