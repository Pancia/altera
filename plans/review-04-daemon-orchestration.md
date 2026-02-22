# Code Review Area 4: Daemon Tick Loop & Orchestration Logic

## Why This Matters

The daemon is the core of the system (~1400 lines). Its 7-step tick cycle orchestrates everything: liveness checks, task assignment, message processing, merge queue, resolver management, and constraint enforcement. Correctness here is critical.

## Files to Review

- `internal/daemon/daemon.go` - The entire tick cycle and all helper methods

## What to Check

### Heartbeat Escalation (Step 1)
- 3-stage escalation: warning (3min) → critical (6min) → dead (10min or PID missing)
- Are transitions correct? Can an agent get stuck in a state?
- Does clearing escalation work properly when agent recovers?
- Is the `.alt-nudge` file mechanism reliable?

### Progress Detection (Step 2)
- Stalled worker detection via last commit time in worktree
- 30-minute threshold with throttled help messages
- Can this misfire on workers doing long research without commits?

### Task Assignment (Step 3)
- Dependency resolution: does `FindReady()` detect cycles?
- What happens if all tasks have unresolvable dependencies?
- Constraint checking before spawn - is the check-then-act atomic?

### Message Processing (Step 4)
- `task_done` → mark done + enqueue merge
- `help` → forward to liaison
- What happens with unknown message types?
- Message idempotency - can a message be processed twice?

### Merge Queue (Step 5)
- FIFO ordering via filesystem timestamps
- Clean merge → success event → notify worker
- Conflict → extract conflict info → spawn resolver
- What if merge fails for reasons other than conflict?

### Resolver Checking (Step 6)
- Polls for clean working tree (no `<<<<<<<` markers)
- On resolution: cleanup resolver, re-queue task
- What if resolver commits but conflict markers remain in other files?

### Constraint Checking (Step 7)
- Budget from event log summation
- Max workers, max queue depth
- Are violations just logged or do they trigger corrective action?

### Cross-Cutting
- Can a tick partially complete and leave inconsistent state?
- What happens if the daemon crashes mid-tick?
- Is the tick idempotent (safe to re-run after crash)?

## Severity Estimate

**HIGH** - This is the system's brain. Logic errors here affect everything.
