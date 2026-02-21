# Test C: REST API (Dependency / Phasing Test)

Paste everything below the line to the liaison after `alt start`.

---

I need you to set up a project scaffold, create 6 tasks in 3 phases with dependency chains, and then verify the phasing was respected.

## Step 1: Project Scaffold

Create the following files:

**go.mod:**
```
module tinyapi

go 1.23
```

**main.go:**
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

**main_test.go:**
```go
package main

import "testing"

func TestPlaceholder(t *testing.T) {
	// placeholder
}
```

Commit these files as "initial scaffold".

## Step 2: Configure

Edit `.alt/config.json` and set `max_workers` to `2`.

## Step 3: Create Tasks with Dependencies

Create all 6 tasks, then set up the dependency chain by editing the task JSON files in `.alt/tasks/`. The `deps` field is an array of task IDs that must be `done` before the task becomes ready.

**Phase 1** (no dependencies, run in parallel):

1. **User model and store** — create `models/user.go` (`User` struct: `ID int`, `Name string`, `Email string`) and `store/memory.go` (in-memory `UserStore` with `Create`/`Get`/`List`/`Delete`, mutex-protected map, auto-incrementing IDs). Tests in `store/memory_test.go`.

2. **Todo model and store** — create `models/todo.go` (`Todo` struct: `ID int`, `Title string`, `Done bool`, `UserID int`) and `store/todo_store.go` (in-memory `TodoStore` with CRUD + `ListByUser`, mutex-protected). Tests in `store/todo_store_test.go`.

**Phase 2** (depends on Phase 1):

3. **User HTTP handlers** — depends on task 1. Create `handlers/user_handler.go` with `GET /users`, `POST /users`, `GET /users/{id}`, `DELETE /users/{id}`. JSON in/out. Use `UserStore`. Tests with `httptest` in `handlers/user_handler_test.go`.

4. **Todo HTTP handlers** — depends on task 2. Create `handlers/todo_handler.go` with full CRUD for `/todos`. JSON in/out. Use `TodoStore`. Tests with `httptest` in `handlers/todo_handler_test.go`.

**Phase 3** (depends on Phase 2):

5. **Logging middleware** — depends on tasks 3 and 4. Create `middleware/logging.go` with `LoggingMiddleware` that wraps an `http.Handler` and logs method, path, status code, and duration. Wire it into `main.go` around all routes. Tests in `middleware/logging_test.go`.

6. **Integration tests** — depends on tasks 3 and 4. Create `integration_test.go` in the root package that starts `httptest.NewServer`, creates a user via POST, creates a todo via POST, does full CRUD cycle on both, and verifies all responses.

### Setting up dependencies

After creating all 6 tasks with `alt task create`, note the task IDs from `alt task list`. Then edit the task JSON files directly to add the `deps` field:

- Task 3's `.alt/tasks/{id}.json`: add `"deps": ["{task-1-id}"]`
- Task 4's `.alt/tasks/{id}.json`: add `"deps": ["{task-2-id}"]`
- Task 5's `.alt/tasks/{id}.json`: add `"deps": ["{task-3-id}", "{task-4-id}"]`
- Task 6's `.alt/tasks/{id}.json`: add `"deps": ["{task-3-id}", "{task-4-id}"]`

Then wait for the daemon to process them. The daemon ticks every 60 seconds. It should:
- Tick 1: Spawn Phase 1 tasks (1 and 2) — Phase 2 deps are unresolved
- After Phase 1 merges: Spawn Phase 2 tasks (3 and 4)
- After Phase 2 merges: Spawn Phase 3 tasks (5 and 6)

This will take several tick cycles. Be patient.

## Step 4: Verify

After all 6 tasks show status `done` in `alt task list`, run these checks and report the results:

1. `alt task list` — confirm all 6 tasks have status `done`
2. `alt log --last 50` — confirm Phase 1 tasks were assigned BEFORE Phase 2 tasks, and Phase 2 BEFORE Phase 3. Look at timestamps on `task_assigned` events.
3. Directory structure check:
   - `ls models/` — should contain `user.go` and `todo.go`
   - `ls store/` — should contain `memory.go` and `todo_store.go`
   - `ls handlers/` — should contain `user_handler.go` and `todo_handler.go`
   - `ls middleware/` — should contain `logging.go`
4. `go test ./...` — confirm all tests pass
5. `git log --oneline` — confirm ~6 merge commits in correct phase order (Phase 1 merges first, then Phase 2, then Phase 3)

Report a summary: PASS if phasing was respected and everything works, FAIL with details if tasks ran out of order or anything broke.
