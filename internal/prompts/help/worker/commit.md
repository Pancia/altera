# Worker: Commit Protocol

Commit early and often. Each commit should be a logical unit of work.

## Format

```
git add <specific-files>
git commit -m "<type>: <what you did>"
```

## Commit Types

- `feat:` — new functionality
- `fix:` — bug fix
- `refactor:` — code restructuring without behavior change
- `test:` — adding or updating tests
- `docs:` — documentation changes
- `wip:` — work in progress (checkpoints only)

## Rules

- **Small, focused commits** — one logical change per commit
- **Add specific files** — never use `git add .` or `git add -A`
- **Clear messages** — describe what changed and why
- **No giant batches** — do not save all changes for one commit at the end
- **Commit working code** — each commit should leave the codebase in a buildable state

## Examples

```
git add internal/api/handler.go
git commit -m "feat: add validation to create endpoint"

git add internal/api/handler_test.go
git commit -m "test: add validation test cases"
```
