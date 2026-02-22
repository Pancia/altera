# 10. Mission-Type Orders (Auftragstaktik)

Status: **Do Now** | Priority: 2 (simple struct extension, quick win)

## Problem

Tasks only have `Title` and `Description`. Workers don't know why a task matters, what success looks like, or how much freedom they have.

---

## Changes

### `internal/task/task.go` — Add fields after `Checkpoint` (line 79):

```go
Intent          string   `json:"intent,omitempty"`
SuccessCriteria []string `json:"success_criteria,omitempty"`
TaskConstraints []string `json:"task_constraints,omitempty"`
Context         string   `json:"context,omitempty"`
Freedom         string   `json:"freedom,omitempty"`
```

Using `TaskConstraints` to avoid collision with Go's `constraints` package if needed as a field name. All `omitempty` — backward compatible with existing tasks.

### `internal/cli/task.go` — Add flags to `taskCreateCmd`:

- `--intent`, `--success-criteria` (StringSlice), `--constraints` (StringSlice), `--context`, `--freedom`
- Wire into Task struct in `taskCreateCmd.RunE`
- Add display in `taskShowCmd.RunE`

### `internal/cli/prime.go` — In `primeWorker()`:

Add a "## Mission Intent" section before the task.json dump. Show Intent, Success Criteria, Constraints, Context, Freedom when present.

### `internal/prompts/help/worker/startup.md`:

Add section on mission-type orders: intent > instructions, adapt when needed, success criteria define done.

---

## Files to modify

| File | Changes |
|------|---------|
| `internal/task/task.go` | Add 5 new fields to Task struct |
| `internal/cli/task.go` | Add flags to create, display in show |
| `internal/cli/prime.go` | Add Mission Intent section to worker prime |
| `internal/prompts/help/worker/startup.md` | Add mission-type orders guidance |

## Verification

- Unit test: Create task with mission fields, persist, read back, verify all fields
- Unit test: Old task JSON without new fields loads correctly (zero values)
- Manual: `alt task create --title "Test" --intent "Make it work" --success-criteria "tests pass"` then `alt task show`
