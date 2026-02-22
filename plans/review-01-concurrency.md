# Code Review Area 1: Concurrency & Data Integrity

## Why This Matters

The system has a single daemon but CLI commands (heartbeat, checkpoint, task-done) run concurrently from worker tmux sessions. All state lives on the filesystem with atomic writes (temp + rename) but no per-file locking on tasks or agents.

## Files to Review

- `internal/task/task.go` - Read-modify-write in `Update()` without locking
- `internal/agent/agent.go` - `TouchHeartbeat()` has same read-modify-write pattern
- `internal/message/message.go` - `ListPending()` + `Archive()` can race on directory listing
- `internal/events/writer.go` - Has flock, but review scope and granularity of lock
- `internal/daemon/daemon.go` - PID lock, signal handling, tick loop, forced-tick channel

## What to Check

- TOCTOU (time-of-check-time-of-use) windows in all store operations
- Whether CLI commands (heartbeat, checkpoint, task-done) can corrupt state the daemon is reading/writing
- Signal handler goroutine correctness and shutdown sequencing
- Buffered channel for forced ticks - can signals be lost?
- Whether atomic writes are sufficient or if file-level locking is needed for read-modify-write cycles
- Event log append contention under load (multiple workers heartbeating simultaneously)

## Preliminary Findings

- `task.go:Update()` reads task, applies mutation, validates, writes - no lock between read and write
- `agent.go:TouchHeartbeat()` same pattern - concurrent heartbeat + daemon status update could conflict
- `message.go:ListPending()` scans directory while `Archive()` moves files - iterator invalidation risk
- Signal handler uses `select` with buffered channel `make(chan struct{}, 1)` - at most one forced tick queued

## Severity Estimate

**HIGH** - Race conditions between CLI commands and daemon could cause data loss or corruption.
