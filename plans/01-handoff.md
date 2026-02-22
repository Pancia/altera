# 1. Session Continuity / Handoff

Status: **Do Now** | Priority: 6 (depends on Checkpoint #07)

## Problem

Dead workers get tasks reclaimed and fresh workers start from scratch. All context is lost.

---

## Changes

### `internal/git/git.go` — Add helpers:

- `Diff(path string, staged bool) (string, error)` — git diff output
- `DiffStat(path string) (string, error)` — compact diff summary
- `Stash(path, message string) (bool, error)` — stash with message, return whether anything stashed
- `StashPop(path string) error`

### `internal/handoff/handoff.go` (new package):

```go
type Handoff struct {
    PreviousAgent      string                `json:"previous_agent"`
    Timestamp          time.Time             `json:"timestamp"`
    ExitType           string                `json:"exit_type"` // clean|crash|killed
    GitSHA             string                `json:"git_sha,omitempty"`
    UncommittedChanges string                `json:"uncommitted_changes,omitempty"`
    StashRef           string                `json:"stash_ref,omitempty"`
    RecentCommits      []string              `json:"recent_commits,omitempty"`
    Checkpoint         *checkpoint.Checkpoint `json:"checkpoint,omitempty"`
    ProgressSummary    string                `json:"progress_summary,omitempty"`
    OpenQuestions      []string              `json:"open_questions,omitempty"`
}
```

Store with `Save(h, taskID)`, `Load(taskID)`. Storage at `.alt/tasks/{task-id}/handoff.json` (single file, overwritten each time).

### `internal/handoff/gather.go` (new):

`GatherFromWorktree(agent, taskID, exitType, cpStore)`:
- Collect git SHA, uncommitted changes (diff, truncated to 10KB), recent commits (git log -10), latest checkpoint
- `StashUncommitted(worktree, taskID, agentID)` — stash before worktree deletion

### `internal/daemon/daemon.go` — `markAgentDead()`:

- **BEFORE** `reclaimTask` and `cleanupBranch`: call `gatherAndSaveHandoff(a, "crash")`
- `gatherAndSaveHandoff` gathers data from worktree, stashes uncommitted changes, saves handoff.json
- Check for recent clean handoff (from Stop hook) — skip overwriting if < 5 min old

### `internal/cli/handoff.go` (new):

`alt handoff <agent-id> --exit-type clean|crash`:
- Gather handoff data from agent's worktree
- Save to `.alt/tasks/{task-id}/handoff.json`
- Called by Stop hook for clean exits

### Worker hooks — Update Stop hook:

In both `internal/worker/worker.go` and `internal/daemon/daemon.go` (both generate `.claude/settings.json`):
```
Stop hooks: alt checkpoint {agentID} && alt handoff {agentID} --exit-type clean
```

### `internal/cli/prime.go`:

Add "## Handoff (from previous session)" section after checkpoint:
- Show: previous agent, exit type, time, SHA, progress summary, recent commits, uncommitted changes (as diff block), open questions

---

## Files to modify/create

| File | Changes |
|------|---------|
| `internal/git/git.go` | Add Diff, DiffStat, Stash, StashPop helpers |
| `internal/handoff/handoff.go` | **New**: Handoff struct, Save/Load |
| `internal/handoff/gather.go` | **New**: GatherFromWorktree, StashUncommitted |
| `internal/handoff/handoff_test.go` | **New**: Tests for Save, Load, Gather |
| `internal/cli/handoff.go` | **New**: `alt handoff` command |
| `internal/daemon/daemon.go` | Gather handoff before reclaimTask in markAgentDead |
| `internal/worker/worker.go` | Update Stop hook to include handoff |
| `internal/cli/prime.go` | Add Handoff section to worker prime output |

## Verification

- Unit test: `handoff.Save` + `Load` round-trip
- Unit test: `GatherFromWorktree` with temp git repo
- Integration: Agent dies -> handoff.json exists before cleanup
- Manual: Kill a worker, verify handoff data appears in next worker's prime output
