package task

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func tempStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

// --- ID Generation ---

func TestGenerateID(t *testing.T) {
	id, err := GenerateID()
	if err != nil {
		t.Fatalf("GenerateID: %v", err)
	}
	if !strings.HasPrefix(id, "t-") {
		t.Errorf("ID should start with 't-', got %q", id)
	}
	// t- + 6 hex chars = 8 total
	if len(id) != 8 {
		t.Errorf("expected ID length 8, got %d (%q)", len(id), id)
	}
}

func TestGenerateID_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for range 100 {
		id, err := GenerateID()
		if err != nil {
			t.Fatalf("GenerateID: %v", err)
		}
		if seen[id] {
			t.Fatalf("duplicate ID: %s", id)
		}
		seen[id] = true
	}
}

// --- Status Parsing ---

func TestParseStatus(t *testing.T) {
	cases := []struct {
		in   string
		want Status
		ok   bool
	}{
		{"open", StatusOpen, true},
		{"assigned", StatusAssigned, true},
		{"in_progress", StatusInProgress, true},
		{"done", StatusDone, true},
		{"failed", StatusFailed, true},
		{"bogus", "", false},
		{"", "", false},
	}
	for _, tc := range cases {
		got, err := ParseStatus(tc.in)
		if tc.ok {
			if err != nil {
				t.Errorf("ParseStatus(%q): unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("ParseStatus(%q) = %q, want %q", tc.in, got, tc.want)
			}
		} else {
			if err == nil {
				t.Errorf("ParseStatus(%q): expected error, got %q", tc.in, got)
			}
		}
	}
}

// --- Status Transitions ---

func TestValidateTransition(t *testing.T) {
	valid := []struct{ from, to Status }{
		{StatusOpen, StatusAssigned},
		{StatusAssigned, StatusInProgress},
		{StatusInProgress, StatusDone},
		{StatusInProgress, StatusFailed},
	}
	for _, tc := range valid {
		if err := ValidateTransition(tc.from, tc.to); err != nil {
			t.Errorf("expected valid transition %s->%s, got error: %v", tc.from, tc.to, err)
		}
	}

	invalid := []struct{ from, to Status }{
		{StatusOpen, StatusInProgress},
		{StatusOpen, StatusDone},
		{StatusOpen, StatusFailed},
		{StatusAssigned, StatusOpen},
		{StatusAssigned, StatusDone},
		{StatusInProgress, StatusOpen},
		{StatusInProgress, StatusAssigned},
		{StatusDone, StatusOpen},
		{StatusFailed, StatusOpen},
	}
	for _, tc := range invalid {
		if err := ValidateTransition(tc.from, tc.to); err == nil {
			t.Errorf("expected invalid transition %s->%s, got nil", tc.from, tc.to)
		}
	}
}

// --- CRUD ---

func TestCreateAndGet(t *testing.T) {
	s := tempStore(t)

	task := &Task{Title: "Test task"}
	if err := s.Create(task); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if task.ID == "" {
		t.Fatal("expected ID to be set")
	}
	if task.Status != StatusOpen {
		t.Errorf("expected status %q, got %q", StatusOpen, task.Status)
	}
	if task.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	got, err := s.Get(task.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Title != "Test task" {
		t.Errorf("title = %q, want %q", got.Title, "Test task")
	}
}

func TestCreate_RequiresTitle(t *testing.T) {
	s := tempStore(t)
	err := s.Create(&Task{})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestCreate_DuplicateID(t *testing.T) {
	s := tempStore(t)
	task := &Task{ID: "t-aaaaaa", Title: "First"}
	if err := s.Create(task); err != nil {
		t.Fatalf("Create: %v", err)
	}
	err := s.Create(&Task{ID: "t-aaaaaa", Title: "Second"})
	if err == nil {
		t.Fatal("expected error for duplicate ID")
	}
}

func TestGet_NotFound(t *testing.T) {
	s := tempStore(t)
	_, err := s.Get("t-nonexist")
	if err == nil {
		t.Fatal("expected error for missing task")
	}
}

func TestUpdate(t *testing.T) {
	s := tempStore(t)
	task := &Task{Title: "Original"}
	if err := s.Create(task); err != nil {
		t.Fatalf("Create: %v", err)
	}

	err := s.Update(task.ID, func(t *Task) error {
		t.Title = "Updated"
		t.Status = StatusAssigned
		t.AssignedTo = "polecat/slit"
		return nil
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := s.Get(task.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Title != "Updated" {
		t.Errorf("title = %q, want %q", got.Title, "Updated")
	}
	if got.Status != StatusAssigned {
		t.Errorf("status = %q, want %q", got.Status, StatusAssigned)
	}
	if got.AssignedTo != "polecat/slit" {
		t.Errorf("assigned_to = %q, want %q", got.AssignedTo, "polecat/slit")
	}
	if !got.UpdatedAt.After(got.CreatedAt) {
		t.Error("expected UpdatedAt > CreatedAt after update")
	}
}

func TestUpdate_InvalidTransition(t *testing.T) {
	s := tempStore(t)
	task := &Task{Title: "Test"}
	if err := s.Create(task); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// open -> in_progress is not allowed (must go through assigned)
	err := s.Update(task.ID, func(t *Task) error {
		t.Status = StatusInProgress
		return nil
	})
	if err == nil {
		t.Fatal("expected error for invalid transition open->in_progress")
	}
}

func TestUpdate_NotFound(t *testing.T) {
	s := tempStore(t)
	err := s.Update("t-nonexist", func(t *Task) error { return nil })
	if err == nil {
		t.Fatal("expected error for missing task")
	}
}

func TestDelete(t *testing.T) {
	s := tempStore(t)
	task := &Task{Title: "To delete"}
	if err := s.Create(task); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := s.Delete(task.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := s.Get(task.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestDelete_NotFound(t *testing.T) {
	s := tempStore(t)
	err := s.Delete("t-nonexist")
	if err == nil {
		t.Fatal("expected error for missing task")
	}
}

// --- List with Filters ---

func TestList_ByStatus(t *testing.T) {
	s := tempStore(t)
	_ = s.Create(&Task{ID: "t-aaaaaa", Title: "Open1"})
	_ = s.Create(&Task{ID: "t-bbbbbb", Title: "Open2"})

	// Transition one to assigned
	_ = s.Update("t-aaaaaa", func(t *Task) error {
		t.Status = StatusAssigned
		return nil
	})

	open, err := s.List(Filter{Status: StatusOpen})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(open) != 1 {
		t.Errorf("expected 1 open task, got %d", len(open))
	}

	assigned, err := s.List(Filter{Status: StatusAssigned})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(assigned) != 1 {
		t.Errorf("expected 1 assigned task, got %d", len(assigned))
	}
}

func TestList_ByAssignee(t *testing.T) {
	s := tempStore(t)
	_ = s.Create(&Task{ID: "t-aaaaaa", Title: "Assigned", AssignedTo: "slit"})
	_ = s.Create(&Task{ID: "t-bbbbbb", Title: "Unassigned"})

	tasks, err := s.List(Filter{AssignedTo: "slit"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tasks) != 1 || tasks[0].AssignedTo != "slit" {
		t.Errorf("expected 1 task assigned to slit, got %d", len(tasks))
	}
}

func TestList_ByTag(t *testing.T) {
	s := tempStore(t)
	_ = s.Create(&Task{ID: "t-aaaaaa", Title: "Tagged", Tags: []string{"urgent", "bug"}})
	_ = s.Create(&Task{ID: "t-bbbbbb", Title: "Untagged"})

	tasks, err := s.List(Filter{Tag: "urgent"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 tagged task, got %d", len(tasks))
	}
}

func TestList_NoFilter(t *testing.T) {
	s := tempStore(t)
	_ = s.Create(&Task{ID: "t-aaaaaa", Title: "One"})
	_ = s.Create(&Task{ID: "t-bbbbbb", Title: "Two"})

	tasks, err := s.List(Filter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestList_EmptyStore(t *testing.T) {
	s := tempStore(t)
	tasks, err := s.List(Filter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
}

// --- FindReady ---

func TestFindReady_NoDeps(t *testing.T) {
	s := tempStore(t)
	_ = s.Create(&Task{ID: "t-aaaaaa", Title: "No deps"})

	ready, err := s.FindReady()
	if err != nil {
		t.Fatalf("FindReady: %v", err)
	}
	if len(ready) != 1 {
		t.Errorf("expected 1 ready task, got %d", len(ready))
	}
}

func TestFindReady_DepsNotDone(t *testing.T) {
	s := tempStore(t)
	_ = s.Create(&Task{ID: "t-dep001", Title: "Dependency"})
	_ = s.Create(&Task{ID: "t-main01", Title: "Main", Deps: []string{"t-dep001"}})

	ready, err := s.FindReady()
	if err != nil {
		t.Fatalf("FindReady: %v", err)
	}

	// t-dep001 has no deps -> ready. t-main01 depends on t-dep001 (open) -> not ready.
	if len(ready) != 1 || ready[0].ID != "t-dep001" {
		t.Errorf("expected only t-dep001 ready, got %v", readyIDs(ready))
	}
}

func TestFindReady_DepsDone(t *testing.T) {
	s := tempStore(t)
	_ = s.Create(&Task{ID: "t-dep001", Title: "Dependency"})
	_ = s.Create(&Task{ID: "t-main01", Title: "Main", Deps: []string{"t-dep001"}})

	// Walk t-dep001 through to done.
	_ = s.Update("t-dep001", func(t *Task) error { t.Status = StatusAssigned; return nil })
	_ = s.Update("t-dep001", func(t *Task) error { t.Status = StatusInProgress; return nil })
	_ = s.Update("t-dep001", func(t *Task) error { t.Status = StatusDone; return nil })

	ready, err := s.FindReady()
	if err != nil {
		t.Fatalf("FindReady: %v", err)
	}

	// t-dep001 is done (not open), t-main01 is open with deps done -> ready.
	if len(ready) != 1 || ready[0].ID != "t-main01" {
		t.Errorf("expected only t-main01 ready, got %v", readyIDs(ready))
	}
}

func TestFindReady_MissingDep(t *testing.T) {
	s := tempStore(t)
	_ = s.Create(&Task{ID: "t-main01", Title: "Main", Deps: []string{"t-nonexist"}})

	ready, err := s.FindReady()
	if err != nil {
		t.Fatalf("FindReady: %v", err)
	}
	if len(ready) != 0 {
		t.Errorf("expected 0 ready (missing dep), got %d", len(ready))
	}
}

func TestFindReady_OnlyOpenTasks(t *testing.T) {
	s := tempStore(t)
	_ = s.Create(&Task{ID: "t-aaaaaa", Title: "Done"})
	_ = s.Update("t-aaaaaa", func(t *Task) error { t.Status = StatusAssigned; return nil })
	_ = s.Update("t-aaaaaa", func(t *Task) error { t.Status = StatusInProgress; return nil })
	_ = s.Update("t-aaaaaa", func(t *Task) error { t.Status = StatusDone; return nil })

	ready, err := s.FindReady()
	if err != nil {
		t.Fatalf("FindReady: %v", err)
	}
	if len(ready) != 0 {
		t.Errorf("expected 0 ready (no open tasks), got %d", len(ready))
	}
}

// --- Atomic Write ---

func TestAtomicWrite_NoTempFiles(t *testing.T) {
	s := tempStore(t)
	_ = s.Create(&Task{ID: "t-aaaaaa", Title: "Test"})

	entries, err := os.ReadDir(s.tasksDir())
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp-") {
			t.Errorf("found leftover temp file: %s", e.Name())
		}
	}
}

func TestNewStore_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "sub", "project")
	s, err := NewStore(target)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	info, err := os.Stat(s.tasksDir())
	if err != nil {
		t.Fatalf("tasks dir does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("tasks path is not a directory")
	}
}

// --- Full Lifecycle ---

func TestFullLifecycle(t *testing.T) {
	s := tempStore(t)

	// Create
	task := &Task{
		Title:       "Implement feature",
		Description: "Build the thing",
		CreatedBy:   "witness",
		Tags:        []string{"feature"},
		Priority:    1,
	}
	if err := s.Create(task); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Assign
	_ = s.Update(task.ID, func(t *Task) error {
		t.Status = StatusAssigned
		t.AssignedTo = "polecat/slit"
		t.Branch = "polecat/slit/feature"
		return nil
	})

	// Start
	_ = s.Update(task.ID, func(t *Task) error {
		t.Status = StatusInProgress
		return nil
	})

	// Complete
	_ = s.Update(task.ID, func(t *Task) error {
		t.Status = StatusDone
		t.Result = "Implemented successfully"
		t.Checkpoint = "abc123"
		return nil
	})

	got, _ := s.Get(task.ID)
	if got.Status != StatusDone {
		t.Errorf("status = %q, want %q", got.Status, StatusDone)
	}
	if got.Result != "Implemented successfully" {
		t.Errorf("result = %q, want %q", got.Result, "Implemented successfully")
	}
	if got.Checkpoint != "abc123" {
		t.Errorf("checkpoint = %q, want %q", got.Checkpoint, "abc123")
	}
}

func TestList_CorruptFileSkipped(t *testing.T) {
	s := tempStore(t)

	// Create a valid task.
	_ = s.Create(&Task{ID: "t-aaaaaa", Title: "Valid"})

	// Write a corrupt task file.
	corruptPath := filepath.Join(s.tasksDir(), "t-corrupt.json")
	if err := os.WriteFile(corruptPath, []byte("{invalid json"), 0o644); err != nil {
		t.Fatalf("write corrupt file: %v", err)
	}

	tasks, err := s.List(Filter{})
	if err != nil {
		t.Fatalf("List: unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task (skipping corrupt), got %d", len(tasks))
	}
	if tasks[0].ID != "t-aaaaaa" {
		t.Errorf("expected task t-aaaaaa, got %s", tasks[0].ID)
	}
}

func readyIDs(tasks []*Task) []string {
	ids := make([]string, len(tasks))
	for i, t := range tasks {
		ids[i] = t.ID
	}
	return ids
}
