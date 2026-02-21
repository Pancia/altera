# Test A: Markdown CLI (Happy Path Smoke Test)

Paste everything below the line to the liaison after `alt start`.

---

I need you to set up a project scaffold, create tasks, and then verify the results after the daemon processes them.

## Step 1: Project Scaffold

Create the following files:

**go.mod:**
```
module md2html

go 1.23
```

**main.go:**
```go
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: md2html <file>")
		os.Exit(1)
	}
	fmt.Println("TODO: convert", os.Args[1])
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

Edit `.alt/config.json` and set `max_workers` to `2` (so it processes 2 tasks at a time, 4 tasks = 2 rounds).

## Step 3: Create Tasks

Create these 4 tasks. They are all independent — no shared files, no dependencies between them:

1. **Heading parser** — create `headings.go` with `func ParseHeadings(line string) string` that converts lines starting with `#` to `<h1>`, `##` to `<h2>`, etc up to `<h6>`. Tests in `headings_test.go`.

2. **Bold/italic parser** — create `bold_italic.go` with `func ParseBoldItalic(line string) string` that converts `**text**` to `<strong>` and `*text*` to `<em>`. Tests in `bold_italic_test.go`.

3. **Link parser** — create `links.go` with `func ParseLinks(line string) string` that converts `[text](url)` to `<a href="url">text</a>`. Tests in `links_test.go`.

4. **Code block parser** — create `codeblocks.go` with `func ParseCodeBlocks(input string) string` that converts triple-backtick blocks to `<pre><code>` tags. Tests in `codeblocks_test.go`.

Use `alt task create` for each one. No deps needed — all 4 are independent.

Then wait for the daemon to process them. The daemon ticks every 60 seconds, so be patient. It will spawn 2 workers first, merge them, then spawn 2 more.

## Step 4: Verify

After all 4 tasks show status `done` in `alt task list`, run these checks and report the results:

1. `alt task list` — confirm all 4 tasks have status `done`
2. `go test ./...` — confirm all tests pass on main branch
3. `ls *.go` — confirm `headings.go`, `bold_italic.go`, `links.go`, `codeblocks.go` exist
4. `alt log --last 30` — confirm the event sequence shows `task_assigned` → `agent_spawned` → `task_done` → `merge_success` for each task
5. `git log --oneline` — confirm 4 merge commits exist

Report a summary: PASS if everything checks out, FAIL with details if anything is wrong.
