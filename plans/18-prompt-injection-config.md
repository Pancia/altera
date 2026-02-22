# 18. Per-Agent Prompt Injection Config

Status: **Do Now** | Priority: 3 (simple config addition, quick win)

## Problem

Different projects need different context injected into agent prompts — skills to load, conventions to follow, architectural context. Currently there's no way to configure this per-project.

---

## Changes

### `internal/config/config.go` — Add types and field:

```go
type PromptConfig struct {
    ExtraPrompt string   `json:"extra_prompt,omitempty"`
    PromptFiles []string `json:"prompt_files,omitempty"`
}
```

Add `Prompts map[string]PromptConfig` to `Config` struct. Initialize in `NewConfig()`. Nil-check in `Load()`.

### `internal/config/prompts.go` (new file):

`ResolvePrompt(cfg Config, agentType string, projectRoot string) (string, error)`:
- Read each file in `PromptFiles` (relative to `projectRoot`)
- Append `ExtraPrompt`
- Return concatenated string (files first, then inline, separated by `\n\n`)
- Return `""` if no config for agent type

### `internal/cli/prime.go`:

At end of both `primeWorker()` and `primeLiaison()`:
- Load config, call `ResolvePrompt(cfg, "worker"|"liaison", root)`
- Output as `## Additional Instructions` section

### `internal/resolver/resolver.go`:

In `SpawnResolver`, append resolved prompt to `initialPrompt` string.

---

## Files to modify/create

| File | Changes |
|------|---------|
| `internal/config/config.go` | Add `PromptConfig` struct and `Prompts` field to `Config` |
| `internal/config/prompts.go` | **New**: `ResolvePrompt()` function |
| `internal/config/prompts_test.go` | **New**: Tests for `ResolvePrompt()` |
| `internal/cli/prime.go` | Append resolved prompt to worker and liaison prime output |
| `internal/resolver/resolver.go` | Append resolved prompt to resolver initial prompt |

## Verification

- Unit test: Config round-trip with Prompts field
- Unit test: `ResolvePrompt` reads files, concatenates, orders correctly
- Unit test: `ResolvePrompt` returns `""` for unconfigured agent type
- Manual: Add `prompts` to `.alt/config.json`, run `alt prime`, verify output
