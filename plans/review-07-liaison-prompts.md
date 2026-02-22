# Code Review Area 7: Liaison & Prompt System

## Why This Matters

The liaison is the human-facing agent that translates intent into tasks. Prompt quality directly affects agent behavior - a misleading prompt can cause workers to take wrong actions. The state summary given to the liaison determines its situational awareness.

## Files to Review

- `internal/liaison/liaison.go` - System state summarization (`Prime()`), message checking
- `internal/prompts/help/help.go` - Prompt template loading
- `internal/prompts/help/liaison/` - 5 liaison prompt templates
- `internal/prompts/help/worker/` - 5 worker prompt templates

## What to Check

- `Prime()` state summary: is it complete and accurate?
  - Does it include all relevant system state?
  - Could it mislead the liaison about system health?
  - Formatting: is the markdown well-structured for LLM consumption?
- `CheckMessages()`: edge cases with empty messages, malformed payloads
- Prompt template quality:
  - Are instructions clear and unambiguous?
  - Could any prompt lead to dangerous behavior (e.g., force-pushing, deleting files)?
  - Are the `alt` CLI commands referenced in prompts correct and up-to-date?
- Template loading: what happens if a template file is missing?
- Prompt injection risks: could task descriptions or message payloads inject instructions?

## Severity Estimate

**LOW-MEDIUM** - Prompt issues cause behavioral problems rather than data corruption, but bad agent behavior could still damage the codebase.
