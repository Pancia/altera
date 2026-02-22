# Code Review Area 3: Process Management & Subprocess Safety

## Why This Matters

The system shells out extensively to `git` and `tmux` with no timeouts. If either hangs, the daemon tick loop stalls. PID tracking goes through tmux which adds indirection. Process cleanup on daemon crash relies on reconciliation.

## Files to Review

- `internal/git/git.go` - All git operations via `exec.Command`
- `internal/tmux/tmux.go` - All tmux operations via `exec.Command`
- `internal/daemon/daemon.go` - Worker spawning, PID tracking, process cleanup
- `internal/worker/worker.go` - Claude Code process lifecycle
- `internal/resolver/resolver.go` - Resolver spawning and cleanup

## What to Check

- Missing `context.Context` / timeouts on all `exec.Command` calls
- PID tracking reliability: tmux pane PID vs actual Claude Code process PID
- What happens when `tmux` or `git` is not installed or not on PATH
- Process cleanup on daemon crash: orphaned tmux sessions, dangling worktrees
- Whether `KillSession` reliably terminates all child processes in the session
- `exec.Command` stderr handling - are error messages captured and logged?
- Whether `WaitForSessionReady` polling can loop indefinitely
- Signal propagation through tmux to child processes

## Preliminary Findings

- `git.go` uses `exec.Command` without `context.WithTimeout` throughout
- `tmux.go` same - all operations are blocking with no timeout
- `daemon.go:724` - `panePID` read from tmux might fail, logged as warning but continues with PID=0
- Worker cleanup relies on `tmux.KillSession()` to terminate Claude - no direct SIGTERM
- `tmux.WaitForSessionReady()` has a timeout parameter but unclear if all callers use it

## Severity Estimate

**MEDIUM** - A hung git or tmux command could stall the entire daemon. Unlikely in practice but no protection exists.
