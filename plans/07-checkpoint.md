# 7. Checkpoint / Resume

Status: **Do Now** | Priority: 5 (foundation for handoff)

## Problem

`alt checkpoint` saves a single string to `task.Checkpoint`. Need structured checkpoints with completed/remaining/decisions/blockers, stored as individual timestamped files.

---

## Changes

### `internal/checkpoint/checkpoint.go` (new package):

```go
type Checkpoint struct {
    TaskID    string    `json:"task_id"`
    AgentID   string    `json:"agent_id"`
    Timestamp time.Time `json:"timestamp"`
    GitSHA    string    `json:"git_sha"`
    Summary   string    `json:"summary"`
    Completed []string  `json:"completed,omitempty"`
    Remaining []string  `json:"remaining,omitempty"`
    Blockers  []string  `json:"blockers,omitempty"`
    Decisions []string  `json:"decisions,omitempty"`
    Notes     string    `json:"notes,omitempty"`
}
```

Store with `Save(cp)`, `Latest(taskID)`, `List(taskID)`. Storage at `.alt/tasks/{task-id}/checkpoints/{timestamp}.json`. The existing `task.Store.List()` already skips directories, so the new `{task-id}/` subdirectory won't interfere.

### `internal/cli/checkpoint.go` — Rewrite:

- Change positional arg from `<task-id>` to `<agent-id>` (Stop hook passes agent-id)
- Look up agent's CurrentTask and Worktree
- Add flags: `--summary`, `--completed` (StringSlice), `--remaining`, `--blockers`, `--decisions`, `--notes`, `--json` (path or `-` for stdin)
- Auto-populate: timestamp, git SHA from worktree
- Save via `checkpoint.Store.Save()`
- Also update legacy `task.Checkpoint` string for backward compat

### `internal/cli/prime.go`:

Replace simple checkpoint output with structured display:
- Load `checkpoint.Store.Latest(taskID)`
- Show: agent, time, SHA, summary, completed, remaining, blockers, decisions, notes
- Fall back to `task.Checkpoint` string if no structured checkpoint exists

### `internal/prompts/help/worker/checkpoint.md`:

Update instructions for structured format.

---

## Key detail: Stop hook compatibility

Current Stop hook: `alt checkpoint {agentID}`. Currently this FAILS because the command expects a task-id. The rewrite fixes this by accepting agent-id and looking up the task. Even with no flags, the checkpoint records timestamp + git SHA, which is useful.

---

## Files to modify/create

| File | Changes |
|------|---------|
| `internal/checkpoint/checkpoint.go` | **New**: Checkpoint struct, Store with Save/Latest/List |
| `internal/checkpoint/checkpoint_test.go` | **New**: Tests for Save, Latest, List, empty dir, ordering |
| `internal/cli/checkpoint.go` | Rewrite: agent-id arg, structured flags, auto-populate |
| `internal/cli/prime.go` | Structured checkpoint display, fallback to legacy |
| `internal/prompts/help/worker/checkpoint.md` | Update instructions for structured format |

## Verification

- Unit test: `Save`, `Latest`, `List`, empty dir, multiple checkpoints ordering
- Unit test: Checkpoint CLI with agent-id resolves to correct task
- Manual: `alt checkpoint <agent-id> --summary "halfway done" --completed "auth,routing" --remaining "tests"`
