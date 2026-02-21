# Altera Test Plans

Three integration tests to validate the full system end-to-end via the liaison.

---

## Test A: Markdown CLI (Happy Path Smoke Test)

**Goal:** 4 independent tasks, no deps, no conflicts. Verify daemon spawns workers, workers code, merge pipeline merges cleanly.

**Setup:**
```bash
mkdir ~/projects/altera-tests/test-a-md2html && cd $_
alt init
```

Create `go.mod` (`module md2html`, `go 1.23`), a stub `main.go` that prints usage, and a placeholder `main_test.go`. Commit as "initial scaffold".

```bash
alt start   # daemon + liaison + attach
```

**Tell the liaison:**
> I need 4 features for this markdown-to-HTML converter. Each is independent — no shared files, no dependencies between them:
> 1. Heading parser — create headings.go with func ParseHeadings(line string) string that converts lines starting with # to h1, ## to h2, etc up to h6. Tests in headings_test.go.
> 2. Bold/italic parser — create bold_italic.go with func ParseBoldItalic(line string) string that converts \*\*text\*\* to strong and \*text\* to em. Tests in bold_italic_test.go.
> 3. Link parser — create links.go with func ParseLinks(line string) string that converts \[text\](url) to anchor tags. Tests in links_test.go.
> 4. Code block parser — create codeblocks.go with func ParseCodeBlocks(input string) string that converts triple-backtick blocks to pre/code tags. Tests in codeblocks_test.go.
>
> Create all 4 tasks and let the daemon handle them.

**Config note:** Set `max_workers: 2` in `.alt/config.json` so it processes 2 at a time (4 tasks = 2 rounds).

**Expected behavior:**
1. Daemon spawns 2 workers on first tick (heading + bold/italic)
2. Workers code in isolated worktrees, commit, send task_done
3. Daemon merges both (no conflicts — different files)
4. Daemon spawns 2 more workers (link + codeblock)
5. Same cycle, all 4 merge cleanly

**Verify:**
- `alt task list` — all 4 tasks status `done`
- `go test ./...` passes on main
- 4 new .go files exist
- `alt log` shows task_assigned → agent_spawned → task_done → merge_success for each

**Cleanup:** `alt stop`

---

## Test B: Data Structures (Resolver / Conflict Test)

**Goal:** 2 tasks that both modify the same file (`types.go`), forcing a merge conflict. Verify the resolver spawns, resolves the conflict, and the re-merge succeeds.

**Setup:**
```bash
mkdir ~/projects/altera-tests/test-b-conflicts && cd $_
alt init
```

Create `go.mod` (`module datastructs`, `go 1.23`).

Create `types.go` — the shared file both tasks will modify:
```go
package datastructs

// Element is the type stored in all data structures.
type Element = int

// Collection is the interface all data structures implement.
type Collection interface {
    Size() int
    IsEmpty() bool
}
```

Placeholder test file. Commit as "initial scaffold".

```bash
alt start
```

**Tell the liaison:**
> I need two data structures implemented. IMPORTANT: both tasks must add their interface to types.go right after the Collection interface — this is intentional, I want to test merge conflict resolution.
> 1. Stack — create stack.go implementing a Stack type. Add a Stacker interface (Push, Pop, Peek methods) to types.go after Collection. Stack implements both Collection and Stacker. Use []Element internally. Tests in stack_test.go.
> 2. Queue — create queue.go implementing a Queue type. Add a Queuer interface (Enqueue, Dequeue, Front methods) to types.go after Collection. Queue implements both Collection and Queuer. Use []Element internally. Tests in queue_test.go.
>
> Create both tasks. They should run simultaneously.

**Config note:** `max_workers: 2` (both start at once, guaranteeing conflict on types.go).

**Expected behavior:**
1. Both workers spawn simultaneously, both modify types.go in their worktrees
2. First to finish gets merged cleanly
3. Second hits merge conflict on types.go
4. Daemon detects conflict, emits merge_conflict event
5. Daemon spawns resolver agent (alt-resolver-* tmux session)
6. Resolver reads conflict-context.json, resolves markers in types.go, commits
7. Daemon detects resolution, re-queues merge
8. Re-merge succeeds

**Verify:**
- Both tasks `done`
- `types.go` has BOTH Stacker and Queuer interfaces
- `stack.go` and `queue.go` exist
- `go test ./...` passes
- `alt log` shows: merge_conflict event, resolver agent_spawned, merge_success

**Cleanup:** `alt stop`

---

## Test C: REST API (Dependency / Phasing Test)

**Goal:** 6 tasks in 3 phases with dependency chains. Verify Phase 2 doesn't start until Phase 1 is done, Phase 3 doesn't start until Phase 2 is done.

**Setup:**
```bash
mkdir ~/projects/altera-tests/test-c-rest-api && cd $_
alt init
```

Create `go.mod` (`module tinyapi`, `go 1.23`), stub `main.go`:
```go
package main

import (
    "fmt"
    "net/http"
)

func main() {
    fmt.Println("Starting tinyapi on :8080")
    http.ListenAndServe(":8080", nil)
}
```

Placeholder test. Commit as "initial scaffold".

```bash
alt start
```

**Tell the liaison:**
> I need a tiny REST API built in 3 phases with proper dependencies:
>
> **Phase 1** (no dependencies, run in parallel):
> 1. User model and in-memory store — create models/user.go (User struct: ID int, Name string, Email string) and store/memory.go (in-memory UserStore with Create/Get/List/Delete, mutex-protected map). Tests in store/memory_test.go.
> 2. Todo model and in-memory store — create models/todo.go (Todo struct: ID int, Title string, Done bool, UserID int) and store/todo_store.go (in-memory TodoStore with CRUD + ListByUser, mutex-protected). Tests in store/todo_store_test.go.
>
> **Phase 2** (depends on Phase 1):
> 3. User HTTP handlers — depends on task 1. Create handlers/user_handler.go with GET/POST /users, GET/DELETE /users/{id}. JSON in/out. Use UserStore. Tests with httptest.
> 4. Todo HTTP handlers — depends on task 2. Create handlers/todo_handler.go with full CRUD for /todos. JSON in/out. Use TodoStore. Tests with httptest.
>
> **Phase 3** (depends on Phase 2):
> 5. Logging middleware — depends on tasks 3+4. Create middleware/logging.go with LoggingMiddleware that logs method/path/status/duration. Wire into main.go. Tests.
> 6. Integration tests — depends on tasks 3+4. Create integration_test.go that starts httptest.NewServer, creates user, creates todo, does full CRUD cycle, verifies everything.
>
> Create all 6 tasks with the dependency chain. The daemon should respect the phasing.

**Config note:** `max_workers: 2`.

**Note on deps:** The liaison will need to create tasks and set the `deps` field. Since `alt task create` doesn't support `--deps`, the liaison should either edit the task JSON files directly or create tasks and then manually add deps to `.alt/tasks/{id}.json`.

**Expected behavior:**
1. Tick 1: Only Phase 1 tasks are ready (Phase 2 deps unresolved). 2 workers spawn.
2. Phase 1 completes and merges (different files, no conflicts)
3. Next tick: Phase 2 tasks become ready (deps now done). 2 workers spawn.
4. Phase 2 completes and merges
5. Next tick: Phase 3 tasks become ready. 2 workers spawn.
6. Phase 3 completes and merges

**Verify:**
- All 6 tasks `done`
- `alt log` shows Phase 1 assigned BEFORE Phase 2, Phase 2 BEFORE Phase 3
- Directory structure: `models/`, `store/`, `handlers/`, `middleware/`
- `go test ./...` passes
- Git log shows ~6 merge commits in correct phase order

**Cleanup:** `alt stop`

---

## General Notes for All Tests

- Daemon tick interval is 60 seconds — be patient between ticks
- `tmux list-sessions` to see active workers/resolvers
- `tmux attach -t alt-worker-XXX` to watch Claude work (Ctrl-B D to detach)
- `alt` binary must be on PATH globally (worker hooks call `alt heartbeat`)
- `alt liaison attach` to re-enter liaison after detaching
- If worker dies immediately: check `claude` is on PATH and authenticated
- Cleanup all tests: `alt stop` in each project dir, then `tmux kill-server` if needed
