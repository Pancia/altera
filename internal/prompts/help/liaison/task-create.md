# Liaison: Creating Tasks

When a human describes work to be done, translate it into structured tasks.

## Principles

- **Atomic** — each task should be a single worker assignment
- **Clear** — a worker can start without asking questions
- **Complete** — acceptance criteria make "done" unambiguous
- **Small** — a few hours of focused work, not days

## Task Format

```json
{
  "id": "t-<shortid>",
  "title": "Short descriptive title",
  "description": "Detailed description with acceptance criteria",
  "status": "open",
  "rig": "<rig-name>",
  "deps": ["t-other-id"],
  "tags": ["feature", "backend"],
  "priority": 1
}
```

## Breaking Down Goals

1. Identify the distinct pieces of work
2. Determine dependencies between them
3. Assign priorities (lower number = higher priority)
4. Write each as a separate task file

## Creating via CLI

```
alt task create --title "Add login endpoint" \
  --description "Create POST /api/login with JWT auth. Accept email+password, return token."
```

## Common Mistakes

- Tasks too large (should be hours, not days)
- Missing acceptance criteria
- Circular dependencies
- Vague descriptions that force workers to guess
