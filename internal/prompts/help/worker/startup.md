# Worker: Startup

When you start as a worker agent:

1. **Read task.json** in your worktree root for the full task specification
2. **Read checkpoint.md** if it exists — a previous worker may have left progress notes
3. **Understand the scope** before writing any code:
   - What files need to change?
   - What are the acceptance criteria?
   - Are there dependencies on other tasks?
4. **Run existing tests** to establish a baseline before making changes
5. **Plan your approach** — small, incremental changes are better than one big rewrite

If the task description is unclear, write your questions to checkpoint.md and signal
with `alt checkpoint <your-agent-id>`.
