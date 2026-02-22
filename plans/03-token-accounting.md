# 3. Token Accounting / Budget Tracking

Status: **Do Now** | Priority: 4 (independent, straightforward)

## Problem

The constraint system already checks `BudgetCeiling` by summing `token_cost` from events, but nothing actually emits those events. Workers don't report their token usage. The budget system is infrastructure without data.

---

## Changes

### `internal/events/events.go`:

Add `TokenUsage Type = "token_usage"` to const block.

### `internal/cli/report_usage.go` (new file):

`alt report-usage` command:
- Flags: `--agent-id` (required, falls back to `ALT_AGENT_ID` env), `--tokens`, `--cost`, `--task-id`
- Writes event with `Data["token_cost"]` (float64) — matches what `constraints.BudgetUsed()` already sums
- At least one of `--tokens` or `--cost` required

### `internal/cli/status.go`:

Add BUDGET section:
- Load config, create constraints checker, call `BudgetUsed()`
- Display `$X.XX / $Y.YY (Z%)`

### `internal/prompts/help/worker/startup.md`:

Add instruction to report usage before `alt task-done`:
```
alt report-usage --agent-id <your-id> --cost <session-cost-usd>
```

### `internal/prompts/help/worker/task-done.md`:

Add reminder to report usage.

---

## Notes

- The constraint system already checks budget via `BudgetUsed()` summing `Data["token_cost"]` — no changes needed there
- Claude Code exposes cost info; workers can read it and report via `alt report-usage`
- Hook integration is prompt-based (instruct worker to call it) rather than automatic — simpler and avoids the problem of not knowing cost until session end

---

## Files to modify/create

| File | Changes |
|------|---------|
| `internal/events/events.go` | Add `TokenUsage` event type |
| `internal/cli/report_usage.go` | **New**: `alt report-usage` command |
| `internal/cli/status.go` | Add BUDGET section |
| `internal/prompts/help/worker/startup.md` | Add usage reporting instruction |
| `internal/prompts/help/worker/task-done.md` | Add usage reporting reminder |

## Verification

- Unit test: `alt report-usage` writes event with correct `token_cost` key
- Integration: Write `token_usage` event, verify `BudgetUsed()` includes it
- Manual: `alt report-usage --agent-id test --cost 5.00` then `alt status`
