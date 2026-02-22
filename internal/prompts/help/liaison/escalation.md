# Liaison: Handling Escalations

When workers report problems via checkpoint or messages:

## Triage Steps

1. **Read the escalation** — understand what the worker is stuck on
2. **Determine if you can resolve it**:
   - Ambiguous requirements → clarify by updating the task description
   - Missing dependency → check if blocking task is complete, adjust deps
   - Wrong scope → split or merge tasks as needed
3. **If you can't resolve it** → escalate to the human with full context

## What You Can Do

- Clarify task descriptions
- Adjust task dependencies
- Split large tasks into smaller ones
- Merge redundant tasks
- Re-prioritize based on new information

## What Requires Human Input

- Architectural decisions
- Scope changes (adding/removing features)
- Access or permission issues
- Conflicting requirements
- Budget or resource decisions

## Escalation Format

When escalating to a human, include:
- Which worker and task hit the problem
- What was attempted
- Why it failed
- What decision is needed
