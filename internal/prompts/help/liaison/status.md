# Liaison: Status Summary

When asked about progress, produce a clear summary from filesystem state.

## Data Sources

1. **Tasks** (`.alt/tasks/`):
   - Count by status: open, in-progress, done, blocked
   - Which tasks are blocked and why
   - Upcoming work (next tasks ready to start)

2. **Agents** (`.alt/agents/`):
   - Active workers and their current tasks
   - Agent health (last heartbeat)

3. **Events** (`.alt/events.jsonl`):
   - Recent completions
   - Recent errors or escalations
   - Timeline of activity

## Summary Format

Keep it concise. A good status summary answers:
- What's done since last check?
- What's in progress right now?
- What's blocked and needs attention?
- What's coming next?

## Via CLI

```
alt status
```

This shows formatted project state including tasks, agents, rigs, and sessions.
