# Liaison: Debugging & Inspecting Workers

When a worker is stalled, stuck, or behaving unexpectedly, you have several tools to investigate.

## Peeking at Worker Output

The `alt worker peek` command shows recent terminal output from a worker:

```
alt worker peek <id>                # Last 200 lines of terminal output
alt worker peek <id> --lines 500    # Last 500 lines
alt worker peek <id> --all          # Full scrollback history
alt worker peek <id> --session      # Show the JSONL conversation transcript
```

For **live workers**, peek captures from the tmux pane scrollback buffer (50,000 lines).
For **dead workers**, peek falls back to reading terminal log files from `.alt/logs/`.

## Viewing Conversation Transcripts

Every worker runs a Claude Code session that writes a JSONL transcript. Use `--session` to see a human-readable summary of the conversation — messages, tool calls, and progress:

```
alt worker peek <id> --session
```

After a worker dies or is cleaned up, the transcript is automatically copied to `.alt/logs/<id>.jsonl` so it remains accessible even after the worktree is deleted.

## Inspecting Worker State

For a detailed overview of a worker's state:

```
alt worker inspect <id>
```

This shows the agent record (JSON), tmux session status, worktree path, git branch, and recent commits.

## Debug Mode

Start Altera with `--debug` to enable continuous terminal logging via tmux pipe-pane:

```
alt start --debug
```

This writes raw terminal output to `.alt/logs/<id>.terminal.log` for every agent session. These files persist after `alt stop`, making them useful for post-mortem analysis of crashes or unexpected behavior.

## Viewing Daemon Logs

The daemon writes logs to `.alt/logs/daemon.log`. View them without attaching to tmux:

```
alt daemon logs              # Last 50 lines
alt daemon logs -n 100       # Last 100 lines
alt daemon logs -f           # Follow log output (Ctrl-C to stop)
```

The log file is truncated each time the daemon starts, so it always reflects the current session. Logs from previous sessions are not preserved.

## Useful Debugging Workflow

1. Check worker status: `alt worker list`
2. Check daemon logs: `alt daemon logs`
3. Peek at recent output: `alt worker peek <id>`
4. If stuck, read the conversation: `alt worker peek <id> --session`
5. For detailed state: `alt worker inspect <id>`
6. If needed, attach interactively: `alt worker attach <id>`
7. Send guidance: `alt message send <id> "<advice>"`

## Log Files

When debug mode is enabled or workers die, logs are stored in `.alt/logs/`:

```
.alt/logs/
├── daemon.log               # Daemon log (current session, always written)
├── w-abc123.terminal.log   # Raw terminal output (debug mode only)
├── w-abc123.jsonl           # Claude Code conversation transcript
├── worker-01.terminal.log
├── worker-01.jsonl
├── resolver-01.terminal.log
├── resolver-01.jsonl
└── liaison-01.terminal.log
```
