package merge

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/anthropics/altera/internal/events"
	"github.com/anthropics/altera/internal/git"
	"github.com/anthropics/altera/internal/message"
	"github.com/anthropics/altera/internal/task"
)

// Outcome represents the result of a merge attempt.
type Outcome string

const (
	OutcomeSuccess     Outcome = "success"
	OutcomeConflict    Outcome = "conflict"
	OutcomeTestFailure Outcome = "test_failure"
)

// Result holds the details of a merge attempt.
type Result struct {
	TaskID     string
	Outcome    Outcome
	Conflicts  []ConflictInfo // populated for OutcomeConflict
	TestOutput string         // populated for OutcomeTestFailure
}

// ConflictInfo describes a single conflicting file with its parsed markers.
type ConflictInfo struct {
	Path    string
	Markers []ConflictMarker
}

// ConflictMarker represents a single conflict region in a file.
type ConflictMarker struct {
	OursStart   int // line number of <<<<<<< marker
	OursEnd     int // line number of ======= marker
	TheirsStart int // line number of ======= marker (same as OursEnd)
	TheirsEnd   int // line number of >>>>>>> marker
}

// Pipeline coordinates merge attempts for completed tasks.
type Pipeline struct {
	tasks    *task.Store
	events   *events.Writer
	messages *message.Store
	queue    *Queue
}

// NewPipeline creates a Pipeline with the given dependencies.
func NewPipeline(tasks *task.Store, ev *events.Writer, msgs *message.Store, q *Queue) *Pipeline {
	return &Pipeline{
		tasks:    tasks,
		events:   ev,
		messages: msgs,
		queue:    q,
	}
}

// AttemptMerge tries to merge a task's branch into the default branch in the
// given worktree. It handles three outcomes:
//   - success: push, emit merge_success, send merge_result message
//   - conflict: extract conflicts, abort merge, emit merge_conflict
//   - test failure: revert merge, emit merge_failed, send merge_result message
func (p *Pipeline) AttemptMerge(taskID string, defaultBranch string, testCommand string, worktree string) (*Result, error) {
	t, err := p.tasks.Get(taskID)
	if err != nil {
		return nil, fmt.Errorf("get task %q: %w", taskID, err)
	}
	if t.Branch == "" {
		return nil, fmt.Errorf("task %q has no branch", taskID)
	}

	// Record pre-merge HEAD so we can revert if tests fail.
	preHead, err := git.Rev(worktree, "HEAD")
	if err != nil {
		return nil, fmt.Errorf("get pre-merge HEAD: %w", err)
	}

	// Emit merge_started event.
	p.events.Append(events.Event{
		Timestamp: time.Now().UTC(),
		Type:      events.MergeStarted,
		TaskID:    taskID,
		Data:      map[string]any{"branch": t.Branch},
	})

	// Attempt the git merge.
	mr, err := git.Merge(worktree, t.Branch)
	if err != nil {
		return nil, fmt.Errorf("merge branch %q: %w", t.Branch, err)
	}

	// Handle conflict outcome.
	if !mr.Clean {
		conflicts := make([]ConflictInfo, 0, len(mr.Conflicts))
		for _, path := range mr.Conflicts {
			fullPath := filepath.Join(worktree, path)
			info := ExtractConflicts(fullPath)
			info.Path = path
			conflicts = append(conflicts, info)
		}

		git.AbortMerge(worktree)

		conflictPaths := make([]any, len(mr.Conflicts))
		for i, c := range mr.Conflicts {
			conflictPaths[i] = c
		}

		p.events.Append(events.Event{
			Timestamp: time.Now().UTC(),
			Type:      events.MergeConflict,
			TaskID:    taskID,
			Data: map[string]any{
				"branch":    t.Branch,
				"conflicts": conflictPaths,
			},
		})

		return &Result{
			TaskID:    taskID,
			Outcome:   OutcomeConflict,
			Conflicts: conflicts,
		}, nil
	}

	// Merge was clean — run tests if a test command is configured.
	if testCommand != "" {
		testOutput, testErr := runTests(worktree, testCommand)
		if testErr != nil {
			// Tests failed — revert the merge commit.
			resetHard(worktree, preHead)

			p.events.Append(events.Event{
				Timestamp: time.Now().UTC(),
				Type:      events.MergeFailed,
				TaskID:    taskID,
				Data: map[string]any{
					"branch": t.Branch,
					"reason": "tests_failed",
					"output": testOutput,
				},
			})

			p.messages.Create(message.TypeMergeResult, "merge-pipeline", t.AssignedTo, taskID, map[string]any{
				"outcome": string(OutcomeTestFailure),
				"output":  testOutput,
			})

			return &Result{
				TaskID:     taskID,
				Outcome:    OutcomeTestFailure,
				TestOutput: testOutput,
			}, nil
		}
	}

	// Tests passed (or no test command) — push.
	if err := git.Push(worktree, "origin", defaultBranch); err != nil {
		return nil, fmt.Errorf("pushing merge: %w", err)
	}

	p.events.Append(events.Event{
		Timestamp: time.Now().UTC(),
		Type:      events.MergeSuccess,
		TaskID:    taskID,
		Data:      map[string]any{"branch": t.Branch},
	})

	p.messages.Create(message.TypeMergeResult, "merge-pipeline", t.AssignedTo, taskID, map[string]any{
		"outcome": string(OutcomeSuccess),
	})

	return &Result{
		TaskID:  taskID,
		Outcome: OutcomeSuccess,
	}, nil
}

// ExtractConflicts parses git conflict markers from a file at the given path,
// returning structured conflict information. If the file cannot be read, an
// empty ConflictInfo is returned.
func ExtractConflicts(path string) ConflictInfo {
	f, err := os.Open(path)
	if err != nil {
		return ConflictInfo{}
	}
	defer f.Close()

	var markers []ConflictMarker
	var current *ConflictMarker
	scanner := bufio.NewScanner(f)
	line := 0
	for scanner.Scan() {
		line++
		text := scanner.Text()
		switch {
		case strings.HasPrefix(text, "<<<<<<<"):
			current = &ConflictMarker{OursStart: line}
		case strings.HasPrefix(text, "=======") && current != nil:
			current.OursEnd = line
			current.TheirsStart = line
		case strings.HasPrefix(text, ">>>>>>>") && current != nil:
			current.TheirsEnd = line
			markers = append(markers, *current)
			current = nil
		}
	}

	return ConflictInfo{Markers: markers}
}

// runTests executes the rig's test command in the given directory and returns
// the combined output. A non-nil error means the tests failed.
func runTests(dir, command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

// resetHard resets the worktree to the given commit, discarding any changes.
func resetHard(dir, commit string) error {
	cmd := exec.Command("git", "reset", "--hard", commit)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git reset --hard %s: %s: %w", commit, strings.TrimSpace(stderr.String()), err)
	}
	return nil
}
