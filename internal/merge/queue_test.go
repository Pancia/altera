package merge

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestQueue(t *testing.T) *Queue {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "merge-queue")
	q, err := NewQueue(dir)
	if err != nil {
		t.Fatalf("NewQueue: %v", err)
	}
	return q
}

// --- NewQueue ---

func TestNewQueue_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "merge-queue")
	q, err := NewQueue(dir)
	if err != nil {
		t.Fatalf("NewQueue: %v", err)
	}
	info, err := os.Stat(q.Dir())
	if err != nil {
		t.Fatalf("Stat queue dir: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected queue dir to be a directory")
	}
}

// --- Enqueue ---

func TestEnqueue(t *testing.T) {
	q := newTestQueue(t)

	if err := q.Enqueue("t-abc123"); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	n, err := q.Len()
	if err != nil {
		t.Fatalf("Len: %v", err)
	}
	if n != 1 {
		t.Errorf("Len = %d, want 1", n)
	}
}

func TestEnqueue_Multiple(t *testing.T) {
	q := newTestQueue(t)

	ids := []string{"t-aaa111", "t-bbb222", "t-ccc333"}
	for _, id := range ids {
		if err := q.Enqueue(id); err != nil {
			t.Fatalf("Enqueue %s: %v", id, err)
		}
		// Small delay to ensure distinct timestamps.
		time.Sleep(time.Millisecond)
	}

	n, err := q.Len()
	if err != nil {
		t.Fatalf("Len: %v", err)
	}
	if n != 3 {
		t.Errorf("Len = %d, want 3", n)
	}
}

func TestEnqueue_Duplicate(t *testing.T) {
	q := newTestQueue(t)

	if err := q.Enqueue("t-dup111"); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	err := q.Enqueue("t-dup111")
	if err != ErrAlreadyQueued {
		t.Errorf("expected ErrAlreadyQueued, got %v", err)
	}
}

func TestEnqueue_AtomicWrite(t *testing.T) {
	q := newTestQueue(t)

	if err := q.Enqueue("t-atom11"); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	// Verify no temp files left behind.
	entries, err := os.ReadDir(q.Dir())
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
			t.Errorf("unexpected non-json file: %s", e.Name())
		}
	}
}

// --- Dequeue ---

func TestDequeue(t *testing.T) {
	q := newTestQueue(t)

	_ = q.Enqueue("t-first1")
	time.Sleep(time.Millisecond)
	_ = q.Enqueue("t-second")

	id, err := q.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue: %v", err)
	}
	if id != "t-first1" {
		t.Errorf("Dequeue = %q, want %q", id, "t-first1")
	}

	n, _ := q.Len()
	if n != 1 {
		t.Errorf("Len after dequeue = %d, want 1", n)
	}
}

func TestDequeue_FIFO_Order(t *testing.T) {
	q := newTestQueue(t)

	ids := []string{"t-one111", "t-two222", "t-three3"}
	for _, id := range ids {
		_ = q.Enqueue(id)
		time.Sleep(time.Millisecond)
	}

	for _, want := range ids {
		got, err := q.Dequeue()
		if err != nil {
			t.Fatalf("Dequeue: %v", err)
		}
		if got != want {
			t.Errorf("Dequeue = %q, want %q", got, want)
		}
	}
}

func TestDequeue_Empty(t *testing.T) {
	q := newTestQueue(t)

	_, err := q.Dequeue()
	if err != ErrQueueEmpty {
		t.Errorf("expected ErrQueueEmpty, got %v", err)
	}
}

func TestDequeue_RemovesFile(t *testing.T) {
	q := newTestQueue(t)

	_ = q.Enqueue("t-remove")
	_, _ = q.Dequeue()

	entries, err := os.ReadDir(q.Dir())
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	jsonCount := 0
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".json" {
			jsonCount++
		}
	}
	if jsonCount != 0 {
		t.Errorf("expected 0 json files after dequeue, got %d", jsonCount)
	}
}

// --- Peek ---

func TestPeek(t *testing.T) {
	q := newTestQueue(t)

	_ = q.Enqueue("t-peek11")
	time.Sleep(time.Millisecond)
	_ = q.Enqueue("t-peek22")

	id, err := q.Peek()
	if err != nil {
		t.Fatalf("Peek: %v", err)
	}
	if id != "t-peek11" {
		t.Errorf("Peek = %q, want %q", id, "t-peek11")
	}

	// Peek should not remove the item.
	n, _ := q.Len()
	if n != 2 {
		t.Errorf("Len after peek = %d, want 2", n)
	}
}

func TestPeek_Empty(t *testing.T) {
	q := newTestQueue(t)

	_, err := q.Peek()
	if err != ErrQueueEmpty {
		t.Errorf("expected ErrQueueEmpty, got %v", err)
	}
}

func TestPeek_Idempotent(t *testing.T) {
	q := newTestQueue(t)

	_ = q.Enqueue("t-idem11")

	for i := 0; i < 3; i++ {
		id, err := q.Peek()
		if err != nil {
			t.Fatalf("Peek %d: %v", i, err)
		}
		if id != "t-idem11" {
			t.Errorf("Peek %d = %q, want %q", i, id, "t-idem11")
		}
	}
}

// --- Len ---

func TestLen_Empty(t *testing.T) {
	q := newTestQueue(t)

	n, err := q.Len()
	if err != nil {
		t.Fatalf("Len: %v", err)
	}
	if n != 0 {
		t.Errorf("Len = %d, want 0", n)
	}
}

func TestLen_AfterEnqueueDequeue(t *testing.T) {
	q := newTestQueue(t)

	_ = q.Enqueue("t-len001")
	time.Sleep(time.Millisecond)
	_ = q.Enqueue("t-len002")
	time.Sleep(time.Millisecond)
	_ = q.Enqueue("t-len003")

	n, _ := q.Len()
	if n != 3 {
		t.Errorf("Len = %d, want 3", n)
	}

	_, _ = q.Dequeue()
	n, _ = q.Len()
	if n != 2 {
		t.Errorf("Len = %d, want 2", n)
	}

	_, _ = q.Dequeue()
	_, _ = q.Dequeue()
	n, _ = q.Len()
	if n != 0 {
		t.Errorf("Len = %d, want 0", n)
	}
}

// --- Dir ---

func TestDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "my-queue")
	q, err := NewQueue(dir)
	if err != nil {
		t.Fatalf("NewQueue: %v", err)
	}
	if q.Dir() != dir {
		t.Errorf("Dir = %q, want %q", q.Dir(), dir)
	}
}

// --- Compatibility with constraints package ---

func TestQueue_ConstraintsCompatible(t *testing.T) {
	// The constraints package counts .json files in the merge queue directory
	// to determine queue depth. Verify our queue stores one .json file per entry.
	q := newTestQueue(t)

	_ = q.Enqueue("t-compat1")
	time.Sleep(time.Millisecond)
	_ = q.Enqueue("t-compat2")
	time.Sleep(time.Millisecond)
	_ = q.Enqueue("t-compat3")

	entries, err := os.ReadDir(q.Dir())
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	jsonCount := 0
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			jsonCount++
		}
	}
	if jsonCount != 3 {
		t.Errorf("expected 3 .json files (for constraints compatibility), got %d", jsonCount)
	}
}
