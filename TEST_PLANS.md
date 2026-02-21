# Altera Test Plans

Three integration tests to validate the full system end-to-end. Each test has a self-contained prompt file in `test-plans/` — paste it to the liaison and it handles scaffold, tasks, and verification.

## Running a Test

For each test:

```bash
mkdir -p ~/projects/altera-tests/<test-dir> && cd $_
alt init
alt start    # launches daemon + liaison, attaches you to liaison
# paste the contents of the test file (below the --- line) to the liaison
# wait for it to finish (several daemon tick cycles, 60s each)
# when done:
alt stop
```

## Tests

| Test | File | What it validates |
|------|------|-------------------|
| **A: Markdown CLI** | `test-plans/test-a-md2html.md` | Happy path — 4 independent tasks, no conflicts, clean merges |
| **B: Data Structures** | `test-plans/test-b-conflicts.md` | Conflict resolution — 2 tasks modify same file, resolver fixes it |
| **C: REST API** | `test-plans/test-c-rest-api.md` | Dependency phasing — 6 tasks in 3 phases, deps respected |

Run them sequentially: A → B → C.

## General Notes

- Daemon tick interval is 60 seconds — be patient between ticks
- `tmux list-sessions` to see active workers/resolvers
- `tmux attach -t alt-worker-XXX` to watch Claude work (Ctrl-B D to detach)
- `alt` binary must be on PATH globally (worker hooks call `alt heartbeat`)
- `alt liaison attach` to re-enter liaison after detaching
- If a worker dies immediately: check `claude` is on PATH and authenticated
- Cleanup: `alt stop` in each project dir, then `tmux kill-server` if needed
