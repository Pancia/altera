# 11. Resolver Integration — Bug Fixes

Status: **Do Now** | Priority: 1 (critical bugs; system is broken without fixes)

## Problem

The resolver loop IS wired end-to-end, but has three critical bugs and two medium issues. The original plan doc was written before the code was completed and overstates the gap.

---

## Critical Bug A: Resolver PID not captured

`resolver.go:204` — Agent record has `PID: 0`. `agent.CheckPID()` returns false for PID <= 0. On first daemon tick, `checkAgentLiveness` at `daemon.go:353` calls `markAgentDead` immediately.

**Fix** (`internal/resolver/resolver.go`, after line 215):
- Add a 500ms sleep + `tmux.PanePID(sessionName)` call, same pattern as `daemon.spawnWorker` (daemon.go:722-727)
- Set `a.PID = panePID` before `m.agents.Create(a)`

---

## Critical Bug B: Re-merge after resolution merges the wrong branch

`daemon.go:1067-1072` — After resolution detected, `checkResolvers` calls `addToMergeQueue(t)`. `addToMergeQueue` at line 1100 uses `t.Branch` (original worker branch like `worker/w-abc123`). But the resolved commits live on the resolver's branch (`alt/resolve-{taskID}`).

Then `CleanupResolver` at `resolver.go:306-309` deletes the resolver's branch BEFORE the re-merge happens. Result: re-merge attempts the original worker branch again, producing the same conflicts.

**Fix** — Change the re-merge flow in `checkResolvers`:
1. Before cleanup, record the resolver's branch name: `resolverBranch := "alt/resolve-" + r.CurrentTask`
2. Do NOT delete the resolver branch during cleanup — add a `preserveBranch bool` param to `CleanupResolver`, or split cleanup into two phases
3. In `addToMergeQueue`, pass the resolver branch instead of the worker branch

**Simplest approach**: Add a `addToMergeQueueWithBranch(taskID, branch, agentID string)` method, and have `checkResolvers` call it with the resolver's branch. After merge succeeds, the original worker branch can be cleaned up too.

**Files**:
- `internal/daemon/daemon.go` — `checkResolvers()`, `addToMergeQueue()`, `processMergeQueue()`
- `internal/resolver/resolver.go` — `CleanupResolver()` split into cleanup (tmux, agent status) vs branch cleanup (deferred)

---

## Critical Bug C: markAgentDead doesn't handle resolver role

`daemon.go:394-434` — When a resolver dies, `markAgentDead` calls `reclaimTask` which resets task to open and clears `Branch`/`AssignedTo`. Then `cleanupBranch` expects `worker/{id}` format but resolver branch is `alt/resolve-{taskID}`.

**Fix** — Add resolver-specific handling in `markAgentDead`:
```go
if a.Role == agent.RoleResolver {
    // Don't reclaim task — it should go back to merge queue
    // Use resolverMgr.CleanupResolver for proper cleanup
    d.resolverMgr.CleanupResolver(a)
    // Re-queue the original task for merge (will retry with a new resolver)
    t, _ := d.tasks.Get(a.CurrentTask)
    if t != nil {
        d.addToMergeQueue(t)
    }
    return
}
```

---

## Medium: No retry limit / escalation

After N failed resolution attempts, escalate to liaison instead of spawning another resolver.

**Fix**: Add `ResolveAttempts int` field to `MergeItem` struct. Increment when re-queuing after resolution. In `processMergeQueue`, if conflict detected and `item.ResolveAttempts >= 3`, send help message to liaison instead of spawning resolver.

---

## Minor: buildConflictContext doesn't set BaseBranch

`daemon.go:1007-1022` — `BaseBranch` left empty, falls through to `"main"` default.

**Fix**: Look up rig config's `DefaultBranch` and set it on the context.

---

## Files to modify

| File | Changes |
|------|---------|
| `internal/resolver/resolver.go` | PID capture, split cleanup into two phases |
| `internal/daemon/daemon.go` | `markAgentDead` resolver handling, `checkResolvers` branch fix, `addToMergeQueueWithBranch`, retry counting in `MergeItem`, `buildConflictContext` base branch |

## Verification

- Unit test: spawn resolver, verify PID is set
- E2e test (extend `internal/daemon/e2e_test.go`): full resolve -> re-merge -> success path
- E2e test: resolver dies -> task re-queued (not reclaimed)
- E2e test: retry limit reached -> escalation to liaison
