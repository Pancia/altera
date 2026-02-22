package daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/events"
	"github.com/anthropics/altera/internal/task"
	"github.com/anthropics/altera/internal/tmux"
)

// --- Shared test infrastructure ---

var (
	altBinOnce sync.Once
	altBinPath string
	altBinErr  error
)

// buildAlt compiles cmd/alt to a temp directory, cached across all tests.
func buildAlt(t *testing.T) string {
	t.Helper()
	altBinOnce.Do(func() {
		dir, err := os.MkdirTemp("", "alt-integ-bin-*")
		if err != nil {
			altBinErr = fmt.Errorf("create temp dir: %w", err)
			return
		}
		binPath := filepath.Join(dir, "alt")
		_, thisFile, _, ok := runtime.Caller(0)
		if !ok {
			altBinErr = fmt.Errorf("runtime.Caller failed")
			return
		}
		projectRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
		cmd := exec.Command("go", "build", "-o", binPath, "./cmd/alt")
		cmd.Dir = projectRoot
		if out, err := cmd.CombinedOutput(); err != nil {
			altBinErr = fmt.Errorf("go build: %s: %w", out, err)
			return
		}
		altBinPath = binPath
	})
	if altBinErr != nil {
		t.Fatalf("buildAlt: %v", altBinErr)
	}
	return altBinPath
}

// writeMockWorker writes a bash script that acts as a deterministic worker.
// It reads task.json, creates a file (or modifies main.go for CONFLICT:
// descriptions), commits, and calls alt task-done.
func writeMockWorker(t *testing.T, altBin string) string {
	t.Helper()

	script := `#!/usr/bin/env bash
set -euo pipefail

# Parse task ID from task.json.
TASK_ID=$(sed -n 's/.*"id": *"\([^"]*\)".*/\1/p' task.json | head -1)
AGENT_ID=${ALT_AGENT_ID}

# Worktree is at {root}/.alt/worktrees/{id}, so .alt/ is at ../..
ALT_DIR=$(cd ../.. && pwd)

# Check for sleep marker (used by crash recovery test).
if [ -f "${ALT_DIR}/sleep-marker" ]; then
    sleep 3600
fi

# Parse description for special behavior markers.
DESC=$(sed -n 's/.*"description": *"\([^"]*\)".*/\1/p' task.json | head -1 || true)

if echo "$DESC" | grep -q 'CONFLICT:'; then
    CONFLICT_VAL=$(echo "$DESC" | sed 's/.*CONFLICT://')
    printf 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println("%s")\n}\n' "$CONFLICT_VAL" > main.go
    git add main.go
else
    echo "completed by $AGENT_ID" > "${TASK_ID}.txt"
    git add "${TASK_ID}.txt"
fi

git commit -m "complete task $TASK_ID"

# Signal task done, then sleep briefly so the daemon has time to process
# the message before the shell exits.
__ALT_BIN__ task-done "$TASK_ID" "$AGENT_ID"
sleep 1
`
	script = strings.ReplaceAll(script, "__ALT_BIN__", altBin)

	dir := t.TempDir()
	path := filepath.Join(dir, "mock-worker.sh")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write mock worker: %v", err)
	}
	return path
}

// startDaemon creates a daemon with fast tick interval and mock worker
// command, runs it in a goroutine, and registers cleanup.
func startDaemon(t *testing.T, root, mockScript string) *Daemon {
	t.Helper()
	d, err := New(root,
		WithTickInterval(200*time.Millisecond),
		WithWorkerCommand(mockScript),
	)
	if err != nil {
		t.Fatalf("startDaemon: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- d.Run()
	}()

	t.Cleanup(func() {
		d.Stop()
		select {
		case err := <-errCh:
			if err != nil {
				t.Errorf("daemon exited with error: %v", err)
			}
		case <-time.After(10 * time.Second):
			t.Error("daemon did not shut down within 10s")
		}
	})

	// Let the daemon start and run its first tick.
	time.Sleep(100 * time.Millisecond)

	return d
}

// waitForTaskStatus polls the task store until the task reaches the target
// status or the timeout expires.
func waitForTaskStatus(t *testing.T, store *task.Store, taskID string, status task.Status, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		tk, err := store.Get(taskID)
		if err == nil && tk.Status == status {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	tk, _ := store.Get(taskID)
	if tk != nil {
		t.Fatalf("task %s: status=%s, want %s (timeout %s)", taskID, tk.Status, status, timeout)
	} else {
		t.Fatalf("task %s: not found (timeout %s)", taskID, timeout)
	}
}

// waitForFile polls until a file exists at root/path.
func waitForFile(t *testing.T, root, path string, timeout time.Duration) {
	t.Helper()
	fullPath := filepath.Join(root, path)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(fullPath); err == nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("file %s not found (timeout %s)", path, timeout)
}

// waitForMergeQueueEmpty polls until the merge queue directory has no .json files.
func waitForMergeQueueEmpty(t *testing.T, altDir string, timeout time.Duration) {
	t.Helper()
	queueDir := filepath.Join(altDir, "merge-queue")
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		entries, err := os.ReadDir(queueDir)
		if err == nil {
			count := 0
			for _, e := range entries {
				if filepath.Ext(e.Name()) == ".json" && !strings.HasPrefix(e.Name(), ".tmp-") {
					count++
				}
			}
			if count == 0 {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("merge queue not empty (timeout %s)", timeout)
}

// --- Test 1: Single task happy path ---

func TestIntegration_SingleTask(t *testing.T) {
	root := setupE2EProject(t)
	altBin := buildAlt(t)
	mockScript := writeMockWorker(t, altBin)
	d := startDaemon(t, root, mockScript)

	tk := &task.Task{
		ID:    "t-single01",
		Title: "Single task test",
	}
	if err := d.tasks.Create(tk); err != nil {
		t.Fatalf("create task: %v", err)
	}

	waitForTaskStatus(t, d.tasks, "t-single01", task.StatusDone, 30*time.Second)
	waitForFile(t, root, "t-single01.txt", 30*time.Second)
	waitForMergeQueueEmpty(t, d.altDir, 30*time.Second)

	// Verify the task result.
	done, err := d.tasks.Get("t-single01")
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if done.Status != task.StatusDone {
		t.Errorf("task status = %s, want done", done.Status)
	}

	// Verify events were emitted.
	evts, _ := d.evReader.ReadAll()
	hasAssigned := false
	hasDone := false
	hasMerge := false
	for _, e := range evts {
		switch e.Type {
		case events.TaskAssigned:
			if e.TaskID == "t-single01" {
				hasAssigned = true
			}
		case events.TaskDone:
			if e.TaskID == "t-single01" {
				hasDone = true
			}
		case events.MergeSuccess:
			if e.TaskID == "t-single01" {
				hasMerge = true
			}
		}
	}
	if !hasAssigned {
		t.Error("missing TaskAssigned event")
	}
	if !hasDone {
		t.Error("missing TaskDone event")
	}
	if !hasMerge {
		t.Error("missing MergeSuccess event")
	}
}

// --- Test 2: Multiple tasks in parallel ---

func TestIntegration_MultiTaskParallel(t *testing.T) {
	root := setupE2EProject(t)
	altBin := buildAlt(t)
	mockScript := writeMockWorker(t, altBin)
	d := startDaemon(t, root, mockScript)

	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("t-par%02d", i)
		tk := &task.Task{
			ID:    id,
			Title: fmt.Sprintf("Parallel task %d", i),
		}
		if err := d.tasks.Create(tk); err != nil {
			t.Fatalf("create task %s: %v", id, err)
		}
	}

	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("t-par%02d", i)
		waitForTaskStatus(t, d.tasks, id, task.StatusDone, 30*time.Second)
	}

	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("t-par%02d", i)
		waitForFile(t, root, id+".txt", 30*time.Second)
	}

	waitForMergeQueueEmpty(t, d.altDir, 30*time.Second)
}

// --- Test 3: Task dependencies ---

func TestIntegration_TaskDependencies(t *testing.T) {
	root := setupE2EProject(t)
	altBin := buildAlt(t)
	mockScript := writeMockWorker(t, altBin)
	d := startDaemon(t, root, mockScript)

	taskA := &task.Task{
		ID:    "t-depA",
		Title: "Base task",
	}
	if err := d.tasks.Create(taskA); err != nil {
		t.Fatalf("create task A: %v", err)
	}

	taskB := &task.Task{
		ID:    "t-depB",
		Title: "Dependent task",
		Deps:  []string{"t-depA"},
	}
	if err := d.tasks.Create(taskB); err != nil {
		t.Fatalf("create task B: %v", err)
	}

	// A must complete first.
	waitForTaskStatus(t, d.tasks, "t-depA", task.StatusDone, 30*time.Second)
	waitForFile(t, root, "t-depA.txt", 30*time.Second)

	// B should now be assigned and eventually complete.
	waitForTaskStatus(t, d.tasks, "t-depB", task.StatusDone, 30*time.Second)
	waitForFile(t, root, "t-depB.txt", 30*time.Second)

	waitForMergeQueueEmpty(t, d.altDir, 30*time.Second)
}

// --- Test 4: Max workers limit ---

func TestIntegration_MaxWorkersLimit(t *testing.T) {
	root := setupE2EProject(t)
	altBin := buildAlt(t)
	mockScript := writeMockWorker(t, altBin)

	// Set max_workers=2 before starting daemon.
	altDir := filepath.Join(root, ".alt")
	cfg := config.Config{
		Constraints: config.Constraints{
			BudgetCeiling: 100.0,
			MaxWorkers:    2,
			MaxQueueDepth: 10,
		},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(filepath.Join(altDir, "config.json"), data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	d := startDaemon(t, root, mockScript)

	for i := 1; i <= 4; i++ {
		id := fmt.Sprintf("t-max%02d", i)
		tk := &task.Task{
			ID:    id,
			Title: fmt.Sprintf("Max workers task %d", i),
		}
		if err := d.tasks.Create(tk); err != nil {
			t.Fatalf("create task %s: %v", id, err)
		}
	}

	// All 4 tasks should eventually complete despite max_workers=2.
	for i := 1; i <= 4; i++ {
		id := fmt.Sprintf("t-max%02d", i)
		waitForTaskStatus(t, d.tasks, id, task.StatusDone, 60*time.Second)
	}

	for i := 1; i <= 4; i++ {
		id := fmt.Sprintf("t-max%02d", i)
		waitForFile(t, root, id+".txt", 30*time.Second)
	}

	waitForMergeQueueEmpty(t, d.altDir, 30*time.Second)
}

// --- Test 5: Merge conflict ---

func TestIntegration_MergeConflict(t *testing.T) {
	root := setupE2EProject(t)
	altBin := buildAlt(t)
	mockScript := writeMockWorker(t, altBin)
	d := startDaemon(t, root, mockScript)

	tk1 := &task.Task{
		ID:          "t-conf01",
		Title:       "Conflict task 1",
		Description: "CONFLICT:version_A",
	}
	if err := d.tasks.Create(tk1); err != nil {
		t.Fatalf("create task 1: %v", err)
	}

	tk2 := &task.Task{
		ID:          "t-conf02",
		Title:       "Conflict task 2",
		Description: "CONFLICT:version_B",
	}
	if err := d.tasks.Create(tk2); err != nil {
		t.Fatalf("create task 2: %v", err)
	}

	// Both workers complete their work.
	waitForTaskStatus(t, d.tasks, "t-conf01", task.StatusDone, 30*time.Second)
	waitForTaskStatus(t, d.tasks, "t-conf02", task.StatusDone, 30*time.Second)

	// Wait for merge queue to drain.
	waitForMergeQueueEmpty(t, d.altDir, 30*time.Second)

	// First merge succeeds, second conflicts.
	evts, _ := d.evReader.ReadAll()
	successCount := 0
	conflictCount := 0
	for _, e := range evts {
		switch e.Type {
		case events.MergeSuccess:
			successCount++
		case events.MergeConflict:
			conflictCount++
		}
	}
	if successCount != 1 {
		t.Errorf("MergeSuccess events = %d, want 1", successCount)
	}
	if conflictCount != 1 {
		t.Errorf("MergeConflict events = %d, want 1", conflictCount)
	}

	// Repo should be clean (conflict was aborted).
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git status: %v", err)
	}
	var dirty []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, ".alt/") || strings.HasSuffix(line, "worktrees/") {
			continue
		}
		dirty = append(dirty, line)
	}
	if len(dirty) != 0 {
		t.Errorf("repo dirty after conflict: %v", dirty)
	}
}

// --- Test 6: Worker crash recovery ---

func TestIntegration_WorkerCrashRecovery(t *testing.T) {
	root := setupE2EProject(t)
	altBin := buildAlt(t)
	mockScript := writeMockWorker(t, altBin)

	altDir := filepath.Join(root, ".alt")

	// Create a sleep marker so the first worker blocks.
	sleepMarker := filepath.Join(altDir, "sleep-marker")
	if err := os.WriteFile(sleepMarker, []byte("1\n"), 0o644); err != nil {
		t.Fatalf("write sleep marker: %v", err)
	}

	d := startDaemon(t, root, mockScript)

	tk := &task.Task{
		ID:    "t-crash01",
		Title: "Crash recovery task",
	}
	if err := d.tasks.Create(tk); err != nil {
		t.Fatalf("create task: %v", err)
	}

	// Wait for the task to be assigned (worker spawned, sleeping on marker).
	waitForTaskStatus(t, d.tasks, "t-crash01", task.StatusAssigned, 30*time.Second)

	// Find the worker's tmux session and kill it.
	assigned, err := d.tasks.Get("t-crash01")
	if err != nil {
		t.Fatalf("get assigned task: %v", err)
	}
	worker, err := d.agents.Get(assigned.AssignedTo)
	if err != nil {
		t.Fatalf("get worker agent: %v", err)
	}
	if err := tmux.KillSession(worker.TmuxSession); err != nil {
		t.Fatalf("kill worker session: %v", err)
	}

	// Daemon detects dead PID, reclaims the task.
	waitForTaskStatus(t, d.tasks, "t-crash01", task.StatusOpen, 30*time.Second)

	// Remove the sleep marker so the respawned worker completes normally.
	_ = os.Remove(sleepMarker)

	// Daemon respawns a new worker which completes the task.
	waitForTaskStatus(t, d.tasks, "t-crash01", task.StatusDone, 30*time.Second)
	waitForFile(t, root, "t-crash01.txt", 30*time.Second)
	waitForMergeQueueEmpty(t, d.altDir, 30*time.Second)

	// Verify the original worker was marked dead.
	original, _ := d.agents.Get(worker.ID)
	if original.Status != "dead" {
		t.Errorf("original worker status = %s, want dead", original.Status)
	}
}
