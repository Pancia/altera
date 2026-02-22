# Code Review Area 5: Configuration & Constraint Validation

## Why This Matters

Config is reloaded every tick. Constraints gate critical operations like worker spawning. Malformed config or incorrect budget calculation could cause the system to over-provision or refuse work.

## Files to Review

- `internal/config/config.go` - Config loading, validation, defaults, rig configs
- `internal/constraints/constraints.go` - Budget ceiling, max workers, queue depth checks

## What to Check

- What happens with malformed config JSON (parse errors, missing fields, wrong types)?
- Are there sensible defaults for all config values?
- What happens with negative or zero constraint values?
- Budget calculation correctness: summing `token_cost` from event log
  - What if events have no `token_cost` field?
  - Is the budget cumulative (all-time) or windowed?
  - Precision issues with float summation?
- Whether constraint violations are actually enforced or just logged
- Config hot-reload implications: can a mid-tick config change cause inconsistency?
- Rig config validation: what if repo path doesn't exist?
- `FindRoot()` directory walk: behavior at filesystem root, symlinks, mount points

## Severity Estimate

**LOW-MEDIUM** - Config issues are usually caught early, but budget miscalculation could silently waste resources.
