# In-Depth Code Review Plan: Altera

## Project Summary

Altera is a ~17.5k LOC Go multi-agent orchestration system with filesystem-based state (`.alt/`), a single daemon tick loop, and Claude Code workers running in git worktrees via tmux. Only direct dependency is Cobra.

## Review Areas

| # | Area | File | Severity |
|---|------|------|----------|
| 1 | Concurrency & Data Integrity | `review-01-concurrency.md` | HIGH |
| 2 | Error Handling & Recovery | `review-02-error-handling.md` | MEDIUM |
| 3 | Process Management & Subprocess Safety | `review-03-process-management.md` | MEDIUM |
| 4 | Daemon Tick Loop & Orchestration | `review-04-daemon-orchestration.md` | HIGH |
| 5 | Configuration & Constraint Validation | `review-05-config-constraints.md` | LOW-MED |
| 6 | Git & Merge Operations | `review-06-git-merge.md` | MEDIUM |
| 7 | Liaison & Prompt System | `review-07-liaison-prompts.md` | LOW-MED |
| 8 | Test Quality & Coverage Gaps | `review-08-test-quality.md` | MEDIUM |
| 9 | API Design & Code Organization | `review-09-api-organization.md` | LOW |

## Review Order

Areas 1 and 4 (HIGH) first, then 2/3/6/8 (MEDIUM), then 5/7/9 (LOW).

## Deliverable Per Area

Each area will produce findings with:
- Specific issues with `file:line` references
- Severity: Critical / High / Medium / Low / Nit
- Concrete recommendations
