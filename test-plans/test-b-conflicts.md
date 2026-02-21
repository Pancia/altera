# Test B: Data Structures (Resolver / Conflict Test)

Paste everything below the line to the liaison after `alt start`.

---

I need you to set up a project scaffold, create tasks that will intentionally conflict, and then verify the conflict resolution worked.

## Step 1: Project Scaffold

Create the following files:

**go.mod:**
```
module datastructs

go 1.23
```

**types.go:**
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

**types_test.go:**
```go
package datastructs

import "testing"

func TestPlaceholder(t *testing.T) {
	// placeholder
}
```

Commit these files as "initial scaffold".

## Step 2: Configure

Edit `.alt/config.json` and set `max_workers` to `2` (both tasks must start simultaneously to guarantee a conflict on types.go).

## Step 3: Create Tasks

Create exactly 2 tasks. IMPORTANT: both tasks must add their interface to `types.go` right after the `Collection` interface — this is intentional, I want to test merge conflict resolution.

1. **Stack implementation** — create `stack.go` implementing a `Stack` type. Add a `Stacker` interface (`Push(Element)`, `Pop() (Element, bool)`, `Peek() (Element, bool)` methods) to `types.go` right after the `Collection` interface. `Stack` implements both `Collection` and `Stacker`. Use `[]Element` internally. Tests in `stack_test.go`.

2. **Queue implementation** — create `queue.go` implementing a `Queue` type. Add a `Queuer` interface (`Enqueue(Element)`, `Dequeue() (Element, bool)`, `Front() (Element, bool)` methods) to `types.go` right after the `Collection` interface. `Queue` implements both `Collection` and `Queuer`. Use `[]Element` internally. Tests in `queue_test.go`.

Use `alt task create` for each. No deps — they should run simultaneously.

Then wait for the daemon to process them. Both will spawn at once, and one will hit a merge conflict on `types.go`. The daemon should automatically spawn a resolver agent to fix it.

## Step 4: Verify

After both tasks show status `done` in `alt task list`, run these checks and report the results:

1. `alt task list` — confirm both tasks have status `done`
2. `cat types.go` — confirm it contains BOTH the `Stacker` and `Queuer` interfaces (plus the original `Collection`)
3. `ls stack.go queue.go` — confirm both files exist
4. `go test ./...` — confirm all tests pass
5. `alt log --last 30` — confirm the event log shows:
   - Two `task_assigned` + `agent_spawned` events
   - At least one `merge_conflict` event
   - A resolver `agent_spawned` event
   - A `merge_success` event after the resolution
6. `git log --oneline` — confirm merge commits exist

Report a summary: PASS if everything checks out (especially the conflict → resolve → re-merge cycle), FAIL with details if anything is wrong.
