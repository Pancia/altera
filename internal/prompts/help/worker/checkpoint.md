# Worker: Checkpoint

Checkpoints save your progress so a fresh worker can continue where you left off.

## When to Checkpoint

- **Context filling up** — before you run out of context window
- **Stuck on a problem** — need human or liaison input
- **Partial progress** — good stopping point between phases

## How to Checkpoint

1. Write `checkpoint.md` in your worktree root with:
   - Current task state (what's done, what's not)
   - Progress percentage estimate
   - Key decisions you made and why
   - Current blockers (if any)
   - Concrete next steps for your replacement

2. Commit your work so far:
   ```
   git add <files>
   git commit -m "wip: <progress summary>"
   ```

3. Signal the daemon:
   ```
   alt checkpoint <your-agent-id>
   ```

## Checkpoint Format

```markdown
## Status: ~60% complete

## Done
- Implemented X
- Added tests for Y

## Remaining
- Need to implement Z
- Tests for edge case W

## Decisions
- Chose approach A over B because ...

## Next Steps
1. Start with Z implementation
2. Then add W edge case tests
```
