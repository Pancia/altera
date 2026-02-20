# Altera Development Guide

You are working on the Altera codebase — a multi-agent orchestration system that
coordinates Claude Code workers to parallelize software development tasks.

## Architecture Overview

Altera is a Go CLI (`alt`) that manages:
- **Workers**: Claude Code instances running in git worktrees, each on a focused task
- **Liaisons**: Translators between human intent and structured task files
- **Resolvers**: Merge conflict resolution agents
- **Daemon**: Background process that monitors agents, enforces constraints, and
  reclaims resources from dead workers

## Project Structure

```
cmd/alt/              # CLI entrypoint
internal/
├── agent/            # Agent data model and persistence
├── cli/              # Cobra command definitions
├── config/           # .alt/ directory management and rig configs
├── constraints/      # Resource limits (budget, max workers, queue depth)
├── daemon/           # Background monitoring and lifecycle management
├── events/           # Append-only event log (JSONL)
├── git/              # Git operations (branch, worktree, merge)
├── liaison/          # Liaison agent logic
├── merge/            # Merge pipeline and queue
├── message/          # Inter-agent messaging
├── prompts/          # System prompt templates (Go embed)
├── resolver/         # Merge conflict resolution
├── task/             # Task CRUD with status machine
├── tmux/             # Tmux session management
└── worker/           # Worker lifecycle (spawn, cleanup, list)
```

## Key Design Decisions

1. **File-based state**: All state lives in `.alt/` as JSON files with atomic writes
   (temp file + rename). No database required.
2. **Agent isolation**: Each worker gets its own git worktree and tmux session.
   Workers cannot interfere with each other.
3. **Event sourcing**: All significant actions emit events to `events.jsonl` for
   debugging and auditing.
4. **Constraint enforcement**: The daemon enforces budget ceilings, max worker counts,
   and queue depth limits.

## Development Workflow

```bash
make build    # Build the alt binary
make test     # Run all tests
make lint     # Run golangci-lint
```

## Testing Conventions

- Tests use `t.TempDir()` for isolation — no shared state between tests
- Integration tests that need tmux check for availability and skip if missing
- Git tests create real repos with initial commits
- Use table-driven tests for pure functions

## Code Style

- Packages are small and focused — one concept per package
- Error messages include context: `fmt.Errorf("doing X: %w", err)`
- Atomic file writes everywhere state is persisted
- Comments explain "why", not "what"
