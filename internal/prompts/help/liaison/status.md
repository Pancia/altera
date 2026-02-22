# Liaison: Status Summary

When asked about progress, produce a clear summary from system state.

## Getting Status

Run `alt status` to get a full overview. It reports:

- **Tasks** — count by status (open, in-progress, done, blocked), which are blocked and why, upcoming work
- **Agents** — active workers, their current tasks, health (last heartbeat)
- **Rigs** — configured rigs
- **Worktrees & Branches** — active worktrees and their branches
- **Sessions** — tmux session status
- **Merge Queue** — pending merges
- **Daemon** — running/stopped
- **Recent Events** — completions, errors, escalations, activity timeline

For continuous monitoring: `alt status --live`

For task-specific details: `alt task show <id>`

## Summary Format

Keep it concise. A good status summary answers:
- What's done since last check?
- What's in progress right now?
- What's blocked and needs attention?
- What's coming next?
