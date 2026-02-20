package message

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := NewStore(filepath.Join(dir, "messages"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

func TestNewStoreCreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	msgDir := filepath.Join(dir, "messages")
	s, err := NewStore(msgDir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil store")
	}
	// Verify archive subdir exists.
	info, err := os.Stat(filepath.Join(msgDir, "archive"))
	if err != nil {
		t.Fatalf("archive dir missing: %v", err)
	}
	if !info.IsDir() {
		t.Error("archive is not a directory")
	}
}

func TestCreateAndGet(t *testing.T) {
	s := newTestStore(t)
	payload := map[string]any{"key": "value", "count": float64(42)}
	m, err := s.Create(TypeTaskDone, "agent-a", "agent-b", "task-1", payload)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if !strings.HasPrefix(m.ID, "m-") {
		t.Errorf("ID = %q, want m- prefix", m.ID)
	}
	if len(m.ID) != 8 { // "m-" + 6 hex chars
		t.Errorf("ID length = %d, want 8", len(m.ID))
	}
	if m.Type != TypeTaskDone {
		t.Errorf("Type = %q, want %q", m.Type, TypeTaskDone)
	}
	if m.From != "agent-a" {
		t.Errorf("From = %q, want %q", m.From, "agent-a")
	}
	if m.To != "agent-b" {
		t.Errorf("To = %q, want %q", m.To, "agent-b")
	}
	if m.TaskID != "task-1" {
		t.Errorf("TaskID = %q, want %q", m.TaskID, "task-1")
	}
	if m.Payload["key"] != "value" {
		t.Errorf("Payload[key] = %v, want %q", m.Payload["key"], "value")
	}
	if m.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}

	got, err := s.Get(m.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != m.ID {
		t.Errorf("Get ID = %q, want %q", got.ID, m.ID)
	}
	if got.Type != m.Type {
		t.Errorf("Get Type = %q, want %q", got.Type, m.Type)
	}
	if got.From != m.From {
		t.Errorf("Get From = %q, want %q", got.From, m.From)
	}
	if got.To != m.To {
		t.Errorf("Get To = %q, want %q", got.To, m.To)
	}
	if got.TaskID != m.TaskID {
		t.Errorf("Get TaskID = %q, want %q", got.TaskID, m.TaskID)
	}
	if got.Payload["count"] != float64(42) {
		t.Errorf("Get Payload[count] = %v, want 42", got.Payload["count"])
	}
}

func TestCreateInvalidType(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Create("bogus", "a", "b", "", nil)
	if err != ErrInvalidType {
		t.Fatalf("expected ErrInvalidType, got %v", err)
	}
}

func TestCreateAllTypes(t *testing.T) {
	s := newTestStore(t)
	for _, mt := range []Type{TypeTaskDone, TypeMergeResult, TypeHelp, TypeCheckpoint} {
		m, err := s.Create(mt, "from", "to", "", nil)
		if err != nil {
			t.Fatalf("Create(%s): %v", mt, err)
		}
		if m.Type != mt {
			t.Errorf("Type = %q, want %q", m.Type, mt)
		}
	}
}

func TestGetNotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Get("m-nope00")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	s := newTestStore(t)
	m, err := s.Create(TypeHelp, "a", "b", "", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := s.Delete(m.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err = s.Get(m.ID)
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestDeleteNotFound(t *testing.T) {
	s := newTestStore(t)
	if err := s.Delete("m-nope00"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestListPending(t *testing.T) {
	s := newTestStore(t)
	// Create messages to different recipients.
	s.Create(TypeTaskDone, "x", "alice", "", nil)
	time.Sleep(time.Millisecond) // ensure distinct timestamps
	s.Create(TypeHelp, "y", "bob", "", nil)
	time.Sleep(time.Millisecond)
	s.Create(TypeCheckpoint, "z", "alice", "", nil)

	msgs, err := s.ListPending("alice")
	if err != nil {
		t.Fatalf("ListPending: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("ListPending(alice) = %d, want 2", len(msgs))
	}
	if msgs[0].Type != TypeTaskDone {
		t.Errorf("first message type = %q, want %q", msgs[0].Type, TypeTaskDone)
	}
	if msgs[1].Type != TypeCheckpoint {
		t.Errorf("second message type = %q, want %q", msgs[1].Type, TypeCheckpoint)
	}

	msgs, err = s.ListPending("bob")
	if err != nil {
		t.Fatalf("ListPending: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("ListPending(bob) = %d, want 1", len(msgs))
	}
}

func TestListPendingEmpty(t *testing.T) {
	s := newTestStore(t)
	msgs, err := s.ListPending("nobody")
	if err != nil {
		t.Fatalf("ListPending: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("expected empty list, got %d", len(msgs))
	}
}

func TestListPendingTimestampOrdering(t *testing.T) {
	s := newTestStore(t)
	// Create 5 messages with slight delays to get distinct timestamps.
	for i := 0; i < 5; i++ {
		s.Create(TypeCheckpoint, "sender", "recipient", "", map[string]any{"seq": float64(i)})
		time.Sleep(time.Millisecond)
	}

	msgs, err := s.ListPending("recipient")
	if err != nil {
		t.Fatalf("ListPending: %v", err)
	}
	if len(msgs) != 5 {
		t.Fatalf("ListPending = %d, want 5", len(msgs))
	}
	// Verify ordering by checking CreatedAt is monotonically increasing.
	for i := 1; i < len(msgs); i++ {
		if !msgs[i].CreatedAt.After(msgs[i-1].CreatedAt) {
			t.Errorf("message %d (%v) not after message %d (%v)",
				i, msgs[i].CreatedAt, i-1, msgs[i-1].CreatedAt)
		}
	}
}

func TestArchive(t *testing.T) {
	s := newTestStore(t)
	m, err := s.Create(TypeMergeResult, "a", "b", "task-1", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := s.Archive(m.ID); err != nil {
		t.Fatalf("Archive: %v", err)
	}

	// Message should no longer be found in the main store.
	_, err = s.Get(m.ID)
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after archive, got %v", err)
	}

	// Verify the file exists in the archive directory.
	archiveDir := filepath.Join(s.dir, "archive")
	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		t.Fatalf("ReadDir archive: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("archive has %d files, want 1", len(entries))
	}
	if !strings.Contains(entries[0].Name(), m.ID) {
		t.Errorf("archive file %q does not contain message ID %q", entries[0].Name(), m.ID)
	}
}

func TestArchiveNotFound(t *testing.T) {
	s := newTestStore(t)
	if err := s.Archive("m-nope00"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestFilenameFormat(t *testing.T) {
	s := newTestStore(t)
	m, err := s.Create(TypeHelp, "a", "b", "", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Verify the file on disk matches the expected naming pattern.
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	found := false
	expected := filename(m)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if e.Name() == expected {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected file %q not found in store", expected)
	}
}

func TestAtomicWrite(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Create(TypeCheckpoint, "a", "b", "", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Verify no temp files remain.
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), ".tmp-") {
			t.Errorf("stale temp file: %s", e.Name())
		}
	}
}

func TestIDUniqueness(t *testing.T) {
	s := newTestStore(t)
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		m, err := s.Create(TypeCheckpoint, "a", "b", "", nil)
		if err != nil {
			t.Fatalf("Create %d: %v", i, err)
		}
		if seen[m.ID] {
			t.Fatalf("duplicate ID: %s", m.ID)
		}
		seen[m.ID] = true
	}
}

func TestCreateWithNilPayload(t *testing.T) {
	s := newTestStore(t)
	m, err := s.Create(TypeTaskDone, "a", "b", "", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := s.Get(m.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Payload != nil {
		t.Errorf("Payload = %v, want nil", got.Payload)
	}
}

func TestCreateWithEmptyTaskID(t *testing.T) {
	s := newTestStore(t)
	m, err := s.Create(TypeHelp, "a", "b", "", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := s.Get(m.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.TaskID != "" {
		t.Errorf("TaskID = %q, want empty", got.TaskID)
	}
}

func TestListPendingExcludesArchived(t *testing.T) {
	s := newTestStore(t)
	m1, _ := s.Create(TypeTaskDone, "a", "alice", "", nil)
	s.Create(TypeHelp, "b", "alice", "", nil)

	// Archive the first message.
	if err := s.Archive(m1.ID); err != nil {
		t.Fatalf("Archive: %v", err)
	}

	msgs, err := s.ListPending("alice")
	if err != nil {
		t.Fatalf("ListPending: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("ListPending = %d, want 1 (archived excluded)", len(msgs))
	}
	if msgs[0].Type != TypeHelp {
		t.Errorf("remaining message type = %q, want %q", msgs[0].Type, TypeHelp)
	}
}
