// Package task provides file-based task management with status tracking.
//
// Tasks are stored as JSON files in .alt/tasks/{id}.json with atomic writes
// (temp file + rename) to prevent corruption.
package task

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Status represents the lifecycle state of a task.
type Status string

const (
	StatusOpen       Status = "open"
	StatusAssigned   Status = "assigned"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
	StatusFailed     Status = "failed"
)

// validTransitions defines which status transitions are allowed.
// open -> assigned -> in_progress -> done|failed
var validTransitions = map[Status][]Status{
	StatusOpen:       {StatusAssigned},
	StatusAssigned:   {StatusInProgress},
	StatusInProgress: {StatusDone, StatusFailed},
}

// ValidateTransition checks if a status transition is allowed.
func ValidateTransition(from, to Status) error {
	allowed, ok := validTransitions[from]
	if !ok {
		return fmt.Errorf("no transitions from status %q", from)
	}
	for _, s := range allowed {
		if s == to {
			return nil
		}
	}
	return fmt.Errorf("invalid transition from %q to %q", from, to)
}

// ParseStatus converts a string to a Status, returning an error for unknown values.
func ParseStatus(s string) (Status, error) {
	switch Status(s) {
	case StatusOpen, StatusAssigned, StatusInProgress, StatusDone, StatusFailed:
		return Status(s), nil
	default:
		return "", fmt.Errorf("unknown status %q", s)
	}
}

// Task represents a unit of work in the system.
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Status      Status    `json:"status"`
	AssignedTo  string    `json:"assigned_to,omitempty"`
	Branch      string    `json:"branch,omitempty"`
	Rig         string    `json:"rig,omitempty"`
	CreatedBy   string    `json:"created_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Result      string    `json:"result,omitempty"`
	ParentID    string    `json:"parent_id,omitempty"`
	Deps        []string  `json:"deps,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Priority    int       `json:"priority,omitempty"`
	Checkpoint  string    `json:"checkpoint,omitempty"`
}

// GenerateID creates a new task ID in the format t-{6 random hex chars}.
func GenerateID() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating task ID: %w", err)
	}
	return "t-" + hex.EncodeToString(b), nil
}

// Store provides file-based CRUD operations for tasks.
// Tasks are stored in {root}/.alt/tasks/{id}.json.
type Store struct {
	root string // project root directory
}

// NewStore creates a Store rooted at the given directory.
// It ensures the .alt/tasks/ directory exists.
func NewStore(root string) (*Store, error) {
	dir := filepath.Join(root, ".alt", "tasks")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating tasks directory: %w", err)
	}
	return &Store{root: root}, nil
}

func (s *Store) tasksDir() string {
	return filepath.Join(s.root, ".alt", "tasks")
}

func (s *Store) taskPath(id string) string {
	return filepath.Join(s.tasksDir(), id+".json")
}

// Create persists a new task. It sets CreatedAt, UpdatedAt, and generates an
// ID if one is not already set. The task must have a title.
func (s *Store) Create(t *Task) error {
	if t.Title == "" {
		return errors.New("task title is required")
	}
	if t.ID == "" {
		id, err := GenerateID()
		if err != nil {
			return err
		}
		t.ID = id
	}
	if t.Status == "" {
		t.Status = StatusOpen
	}
	now := time.Now().UTC()
	t.CreatedAt = now
	t.UpdatedAt = now

	if _, err := os.Stat(s.taskPath(t.ID)); err == nil {
		return fmt.Errorf("task %q already exists", t.ID)
	}
	return s.writeTask(t)
}

// Get reads a task by ID.
func (s *Store) Get(id string) (*Task, error) {
	data, err := os.ReadFile(s.taskPath(id))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("task %q not found", id)
		}
		return nil, fmt.Errorf("reading task %q: %w", id, err)
	}
	var t Task
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("parsing task %q: %w", id, err)
	}
	return &t, nil
}

// Update applies a mutation function to an existing task, validating any
// status change against the allowed transitions.
func (s *Store) Update(id string, fn func(*Task) error) error {
	t, err := s.Get(id)
	if err != nil {
		return err
	}
	oldStatus := t.Status
	if err := fn(t); err != nil {
		return err
	}
	if t.Status != oldStatus {
		if err := ValidateTransition(oldStatus, t.Status); err != nil {
			return err
		}
	}
	t.UpdatedAt = time.Now().UTC()
	return s.writeTask(t)
}

// Delete removes a task file.
func (s *Store) Delete(id string) error {
	path := s.taskPath(id)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("task %q not found", id)
		}
		return fmt.Errorf("deleting task %q: %w", id, err)
	}
	return nil
}

// Filter specifies criteria for listing tasks.
type Filter struct {
	Status     Status
	Rig        string
	AssignedTo string
	Tag        string
}

// List returns all tasks matching the given filter. Zero-value filter fields
// are ignored (match all).
func (s *Store) List(f Filter) ([]*Task, error) {
	entries, err := os.ReadDir(s.tasksDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("listing tasks: %w", err)
	}

	var tasks []*Task
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		t, err := s.Get(id)
		if err != nil {
			continue // skip corrupt files
		}
		if matchesFilter(t, f) {
			tasks = append(tasks, t)
		}
	}
	return tasks, nil
}

// FindReady returns open tasks whose dependencies are all done.
func (s *Store) FindReady() ([]*Task, error) {
	all, err := s.List(Filter{Status: StatusOpen})
	if err != nil {
		return nil, err
	}

	var ready []*Task
	for _, t := range all {
		if len(t.Deps) == 0 {
			ready = append(ready, t)
			continue
		}
		allDone := true
		for _, depID := range t.Deps {
			dep, err := s.Get(depID)
			if err != nil {
				// Missing dep means not done.
				allDone = false
				break
			}
			if dep.Status != StatusDone {
				allDone = false
				break
			}
		}
		if allDone {
			ready = append(ready, t)
		}
	}
	return ready, nil
}

// ForceWrite writes a task to disk without validating status transitions.
// This is intended for privileged operations like the daemon reclaiming
// tasks from dead agents.
func (s *Store) ForceWrite(t *Task) error {
	return s.writeTask(t)
}

// writeTask atomically writes a task to disk (temp file + rename).
func (s *Store) writeTask(t *Task) error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling task: %w", err)
	}
	data = append(data, '\n')

	dir := s.tasksDir()
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", err)
	}

	dest := s.taskPath(t.ID)
	if err := os.Rename(tmpName, dest); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming temp to %s: %w", dest, err)
	}
	return nil
}

func matchesFilter(t *Task, f Filter) bool {
	if f.Status != "" && t.Status != f.Status {
		return false
	}
	if f.Rig != "" && t.Rig != f.Rig {
		return false
	}
	if f.AssignedTo != "" && t.AssignedTo != f.AssignedTo {
		return false
	}
	if f.Tag != "" && !containsTag(t.Tags, f.Tag) {
		return false
	}
	return true
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
