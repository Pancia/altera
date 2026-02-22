# Worker: Task Done

When you've completed your assigned task:

1. **Run tests** — make sure all tests pass before declaring done
2. **Review your changes** — `git diff` to verify nothing unintended was modified
3. **Commit all work** — no uncommitted changes should remain:
   ```
   git add <specific-files>
   git commit -m "feat: <what you did>"
   ```
4. **Verify scope** — did you only modify files relevant to your task?
5. **Exit cleanly** — the daemon will detect your session ended and handle merge

Do NOT:
- Leave uncommitted changes
- Batch everything into one giant commit
- Modify files outside your task scope
- Skip running tests
