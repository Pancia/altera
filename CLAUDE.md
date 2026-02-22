# Altera

Multi-agent orchestration system with filesystem-based state (`.alt/` directory).

## Project Structure

- `cmd/alt/main.go` - CLI entrypoint (Cobra)
- `internal/cli/` - Cobra command definitions
- `internal/config/` - Configuration loading and management
- `internal/task/` - Task representation and lifecycle
- `internal/agent/` - Agent spawning and management
- `internal/message/` - Inter-agent messaging
- `internal/events/` - Event system
- `internal/daemon/` - Background daemon process
- `internal/worker/` - Task execution workers
- `internal/liaison/` - External system integration
- `internal/resolver/` - Conflict resolution
- `internal/merge/` - State merging
- `internal/git/` - Git operations
- `internal/tmux/` - Tmux session management
- `internal/constraints/` - Constraint validation

## Build

```
make          # Build and install to ~/.local/bin/alt (default)
make install  # Same as above
make build    # Build binary to bin/alt only
make test     # Run all tests
make lint     # Run linter
make clean    # Remove build artifacts
```
