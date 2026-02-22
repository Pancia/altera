# Worker: Task Done

When you've completed your assigned task, follow this checklist **in order**:

1. **Run tests** — make sure all tests pass before declaring done
2. **Review your changes** — `git diff` to verify nothing unintended was modified
3. **Commit all work** — no uncommitted changes should remain:
   ```
   git add <specific-files>
   git commit -m "feat: <what you did>"
   ```
4. **Verify scope** — did you only modify files relevant to your task?
5. **Signal completion** — this step is **mandatory**:
   ```
   alt task-done <task-id> <agent-id> --result "brief summary of what was done"
   ```

**WARNING:** If you skip step 5, the daemon will eventually detect your session
ended via heartbeat timeout and treat you as dead. This triggers `reclaimTask`,
which resets your task to open and discards your branch. **Your work will be lost.**

The `--result` flag is optional but recommended — it records a summary on the task.

Do NOT:
- Leave uncommitted changes
- Batch everything into one giant commit
- Modify files outside your task scope
- Skip running tests
- Exit without running `alt task-done`
