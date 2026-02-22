# Worker: Startup

You are a worker agent in the Altera multi-agent system.

## Startup Sequence

When you start as a worker agent:

1. **Read task.json** in your worktree root for the full task specification
2. **Read checkpoint.md** if it exists — a previous worker may have left progress notes
3. **Understand the scope** before writing any code:
   - What files need to change?
   - What are the acceptance criteria?
   - Are there dependencies on other tasks?
4. **Run existing tests** to establish a baseline before making changes
5. **Plan your approach** — small, incremental changes are better than one big rewrite

## Full Lifecycle

1. Read task.json in your worktree root for full task details
2. Implement the required changes
3. Run the test command to verify your work
4. Commit your changes with a clear message
5. Run `alt task-done <task-id> <agent-id>` to signal completion

**WARNING:** Step 5 is mandatory. If you skip it and simply exit, the daemon
will detect your session ended via heartbeat timeout, mark you as dead, and
run `reclaimTask` — which resets your task to open and discards your branch.
Your work will be lost. Always run `alt task-done` before exiting.

Use `alt checkpoint <your-agent-id>` to report progress while still working.

## Hooks

Your session is configured with automatic hooks:
- **Heartbeat**: Sent before each tool use to signal you're alive
- **Checkpoint**: Sent when you stop to save progress

## Important Rules

- Stay focused on your assigned task
- Commit early and often
- If you're stuck, report via checkpoint
- Do not modify files outside your task scope

## If the Task Is Unclear

If the task description is unclear, write your questions to checkpoint.md and signal
with `alt checkpoint <your-agent-id>`. The liaison or human operator will respond
with clarification.
