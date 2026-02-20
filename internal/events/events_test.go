package events

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func tmpPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "events.jsonl")
}

func mustWrite(t *testing.T, w *Writer, events ...Event) {
	t.Helper()
	if err := w.Append(events...); err != nil {
		t.Fatalf("Append failed: %v", err)
	}
}

func testEvent(typ Type, agentID, taskID string) Event {
	return Event{
		Timestamp: time.Now().UTC().Truncate(time.Millisecond),
		Type:      typ,
		AgentID:   agentID,
		TaskID:    taskID,
	}
}

// --- Event struct tests ---

func TestEventMarshalRoundtrip(t *testing.T) {
	ev := Event{
		Timestamp: time.Date(2026, 2, 19, 12, 0, 0, 0, time.UTC),
		Type:      TaskCreated,
		AgentID:   "polecat-1",
		TaskID:    "al-92a",
		Data:      map[string]any{"key": "value", "count": float64(42)},
	}

	data, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got Event
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !got.Timestamp.Equal(ev.Timestamp) {
		t.Errorf("Timestamp: got %v, want %v", got.Timestamp, ev.Timestamp)
	}
	if got.Type != ev.Type {
		t.Errorf("Type: got %v, want %v", got.Type, ev.Type)
	}
	if got.AgentID != ev.AgentID {
		t.Errorf("AgentID: got %v, want %v", got.AgentID, ev.AgentID)
	}
	if got.TaskID != ev.TaskID {
		t.Errorf("TaskID: got %v, want %v", got.TaskID, ev.TaskID)
	}
	if got.Data["key"] != ev.Data["key"] {
		t.Errorf("Data[key]: got %v, want %v", got.Data["key"], ev.Data["key"])
	}
	if got.Data["count"] != ev.Data["count"] {
		t.Errorf("Data[count]: got %v, want %v", got.Data["count"], ev.Data["count"])
	}
}

func TestEventMarshalOmitsEmptyData(t *testing.T) {
	ev := Event{
		Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Type:      AgentSpawned,
		AgentID:   "a1",
		TaskID:    "t1",
	}

	data, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var raw map[string]any
	json.Unmarshal(data, &raw)
	if _, ok := raw["data"]; ok {
		t.Errorf("Expected data field to be omitted when nil, got: %s", data)
	}
}

func TestEventTimestampUTC(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	ev := Event{
		Timestamp: time.Date(2026, 6, 15, 12, 0, 0, 0, loc),
		Type:      TaskDone,
		AgentID:   "a1",
		TaskID:    "t1",
	}

	data, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got Event
	json.Unmarshal(data, &got)
	if got.Timestamp.Location() != time.UTC {
		t.Errorf("Expected UTC timestamp, got %v", got.Timestamp.Location())
	}
}

// --- Writer tests ---

func TestWriterAppendSingle(t *testing.T) {
	path := tmpPath(t)
	w := NewWriter(path)

	ev := testEvent(TaskCreated, "agent-1", "task-1")
	mustWrite(t, w, ev)

	r := NewReader(path)
	events, err := r.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	if events[0].Type != TaskCreated {
		t.Errorf("Type: got %v, want %v", events[0].Type, TaskCreated)
	}
}

func TestWriterAppendMultiple(t *testing.T) {
	path := tmpPath(t)
	w := NewWriter(path)

	mustWrite(t, w,
		testEvent(TaskCreated, "a1", "t1"),
		testEvent(TaskAssigned, "a1", "t1"),
		testEvent(TaskStarted, "a1", "t1"),
	)

	r := NewReader(path)
	events, err := r.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("Expected 3 events, got %d", len(events))
	}
}

func TestWriterAppendEmpty(t *testing.T) {
	path := tmpPath(t)
	w := NewWriter(path)

	if err := w.Append(); err != nil {
		t.Fatalf("Append with no events should not error: %v", err)
	}

	// File should not be created for empty append.
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("Expected file to not exist after empty append")
	}
}

func TestWriterCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "events.jsonl")
	w := NewWriter(path)

	mustWrite(t, w, testEvent(AgentSpawned, "a1", ""))

	r := NewReader(path)
	events, err := r.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
}

func TestWriterMultipleAppendCalls(t *testing.T) {
	path := tmpPath(t)
	w := NewWriter(path)

	mustWrite(t, w, testEvent(TaskCreated, "a1", "t1"))
	mustWrite(t, w, testEvent(TaskStarted, "a1", "t1"))
	mustWrite(t, w, testEvent(TaskDone, "a1", "t1"))

	r := NewReader(path)
	events, err := r.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("Expected 3 events, got %d", len(events))
	}
	if events[0].Type != TaskCreated {
		t.Errorf("events[0].Type: got %v, want %v", events[0].Type, TaskCreated)
	}
	if events[2].Type != TaskDone {
		t.Errorf("events[2].Type: got %v, want %v", events[2].Type, TaskDone)
	}
}

func TestWriterDefaultPath(t *testing.T) {
	w := NewWriter("")
	if w.Path() != defaultPath {
		t.Errorf("Path: got %v, want %v", w.Path(), defaultPath)
	}
}

// --- Reader filter tests ---

func TestReadFilterByType(t *testing.T) {
	path := tmpPath(t)
	w := NewWriter(path)

	mustWrite(t, w,
		testEvent(TaskCreated, "a1", "t1"),
		testEvent(AgentSpawned, "a2", ""),
		testEvent(TaskDone, "a1", "t1"),
		testEvent(AgentSpawned, "a3", ""),
	)

	r := NewReader(path)
	events, err := r.Read(Filter{Type: AgentSpawned})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("Expected 2 AgentSpawned events, got %d", len(events))
	}
	for _, ev := range events {
		if ev.Type != AgentSpawned {
			t.Errorf("Expected AgentSpawned, got %v", ev.Type)
		}
	}
}

func TestReadFilterByAgent(t *testing.T) {
	path := tmpPath(t)
	w := NewWriter(path)

	mustWrite(t, w,
		testEvent(TaskCreated, "a1", "t1"),
		testEvent(TaskCreated, "a2", "t2"),
		testEvent(TaskDone, "a1", "t1"),
	)

	r := NewReader(path)
	events, err := r.Read(Filter{AgentID: "a1"})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("Expected 2 events for a1, got %d", len(events))
	}
}

func TestReadFilterByTask(t *testing.T) {
	path := tmpPath(t)
	w := NewWriter(path)

	mustWrite(t, w,
		testEvent(TaskCreated, "a1", "t1"),
		testEvent(TaskCreated, "a2", "t2"),
		testEvent(TaskDone, "a1", "t1"),
	)

	r := NewReader(path)
	events, err := r.Read(Filter{TaskID: "t2"})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event for t2, got %d", len(events))
	}
}

func TestReadFilterByTimeRange(t *testing.T) {
	path := tmpPath(t)
	w := NewWriter(path)

	t1 := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 1, 1, 14, 0, 0, 0, time.UTC)

	mustWrite(t, w,
		Event{Timestamp: t1, Type: TaskCreated, AgentID: "a1", TaskID: "t1"},
		Event{Timestamp: t2, Type: TaskStarted, AgentID: "a1", TaskID: "t1"},
		Event{Timestamp: t3, Type: TaskDone, AgentID: "a1", TaskID: "t1"},
	)

	r := NewReader(path)

	// After t1 (exclusive), before t3 (exclusive) → should get t2 only
	events, err := r.Read(Filter{
		After:  t1,
		Before: t3,
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event in range, got %d", len(events))
	}
	if events[0].Type != TaskStarted {
		t.Errorf("Expected TaskStarted, got %v", events[0].Type)
	}
}

func TestReadFilterCombined(t *testing.T) {
	path := tmpPath(t)
	w := NewWriter(path)

	mustWrite(t, w,
		testEvent(TaskCreated, "a1", "t1"),
		testEvent(TaskCreated, "a2", "t2"),
		testEvent(AgentSpawned, "a1", ""),
		testEvent(TaskDone, "a1", "t1"),
	)

	r := NewReader(path)
	events, err := r.Read(Filter{Type: TaskCreated, AgentID: "a1"})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	if events[0].TaskID != "t1" {
		t.Errorf("TaskID: got %v, want t1", events[0].TaskID)
	}
}

func TestReadEmptyFilter(t *testing.T) {
	path := tmpPath(t)
	w := NewWriter(path)

	mustWrite(t, w,
		testEvent(TaskCreated, "a1", "t1"),
		testEvent(TaskDone, "a1", "t1"),
	)

	r := NewReader(path)
	events, err := r.Read(Filter{})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(events))
	}
}

func TestReadNonexistentFile(t *testing.T) {
	r := NewReader(filepath.Join(t.TempDir(), "nonexistent.jsonl"))
	_, err := r.ReadAll()
	if err == nil {
		t.Fatal("Expected error reading nonexistent file")
	}
}

func TestReaderDefaultPath(t *testing.T) {
	r := NewReader("")
	if r.Path() != defaultPath {
		t.Errorf("Path: got %v, want %v", r.Path(), defaultPath)
	}
}

// --- Tail tests ---

func TestTailLastN(t *testing.T) {
	path := tmpPath(t)
	w := NewWriter(path)

	for i := 0; i < 10; i++ {
		mustWrite(t, w, testEvent(TaskCreated, "a1", "t1"))
	}

	r := NewReader(path)
	events, err := r.Tail(3)
	if err != nil {
		t.Fatalf("Tail: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("Expected 3 events, got %d", len(events))
	}
}

func TestTailMoreThanExists(t *testing.T) {
	path := tmpPath(t)
	w := NewWriter(path)

	mustWrite(t, w,
		testEvent(TaskCreated, "a1", "t1"),
		testEvent(TaskDone, "a1", "t1"),
	)

	r := NewReader(path)
	events, err := r.Tail(100)
	if err != nil {
		t.Fatalf("Tail: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(events))
	}
}

func TestTailZero(t *testing.T) {
	path := tmpPath(t)
	w := NewWriter(path)
	mustWrite(t, w, testEvent(TaskCreated, "a1", "t1"))

	r := NewReader(path)
	events, err := r.Tail(0)
	if err != nil {
		t.Fatalf("Tail: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("Expected 0 events, got %d", len(events))
	}
}

func TestTailNegative(t *testing.T) {
	r := NewReader(tmpPath(t))
	events, err := r.Tail(-1)
	if err != nil {
		t.Fatalf("Tail: %v", err)
	}
	if events != nil {
		t.Fatalf("Expected nil, got %v", events)
	}
}

func TestTailPreservesOrder(t *testing.T) {
	path := tmpPath(t)
	w := NewWriter(path)

	types := []Type{TaskCreated, TaskAssigned, TaskStarted, TaskDone, TaskFailed}
	for _, typ := range types {
		mustWrite(t, w, testEvent(typ, "a1", "t1"))
	}

	r := NewReader(path)
	events, err := r.Tail(3)
	if err != nil {
		t.Fatalf("Tail: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("Expected 3, got %d", len(events))
	}
	if events[0].Type != TaskStarted {
		t.Errorf("events[0].Type: got %v, want TaskStarted", events[0].Type)
	}
	if events[1].Type != TaskDone {
		t.Errorf("events[1].Type: got %v, want TaskDone", events[1].Type)
	}
	if events[2].Type != TaskFailed {
		t.Errorf("events[2].Type: got %v, want TaskFailed", events[2].Type)
	}
}

// --- Concurrent append safety test ---

func TestConcurrentAppendSafety(t *testing.T) {
	path := tmpPath(t)

	const numGoroutines = 10
	const eventsPerGoroutine = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(id int) {
			defer wg.Done()
			w := NewWriter(path)
			for i := 0; i < eventsPerGoroutine; i++ {
				ev := Event{
					Timestamp: time.Now().UTC(),
					Type:      TaskCreated,
					AgentID:   "agent-" + string(rune('a'+id)),
					TaskID:    "task-1",
					Data:      map[string]any{"goroutine": id, "index": i},
				}
				if err := w.Append(ev); err != nil {
					t.Errorf("goroutine %d: Append failed: %v", id, err)
					return
				}
			}
		}(g)
	}

	wg.Wait()

	r := NewReader(path)
	events, err := r.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll after concurrent writes: %v", err)
	}

	expected := numGoroutines * eventsPerGoroutine
	if len(events) != expected {
		t.Errorf("Expected %d events, got %d", expected, len(events))
	}

	// Verify each line is valid JSON (no interleaving).
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	count := 0
	for dec.More() {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			t.Fatalf("Line %d: invalid JSON: %v", count+1, err)
		}
		count++
	}
	if count != expected {
		t.Errorf("JSON line count: got %d, want %d", count, expected)
	}
}

// --- Event type constants test ---

func TestAllEventTypes(t *testing.T) {
	types := []Type{
		TaskCreated, TaskAssigned, TaskStarted, TaskDone, TaskFailed,
		AgentSpawned, AgentDied,
		MergeStarted, MergeSuccess, MergeConflict, MergeFailed,
		BudgetExceeded, WorkerStalled,
	}

	if len(types) != 13 {
		t.Errorf("Expected 13 event types, got %d", len(types))
	}

	seen := make(map[Type]bool)
	for _, typ := range types {
		if typ == "" {
			t.Error("Event type should not be empty string")
		}
		if seen[typ] {
			t.Errorf("Duplicate event type: %v", typ)
		}
		seen[typ] = true
	}
}

func TestEventWithData(t *testing.T) {
	path := tmpPath(t)
	w := NewWriter(path)

	ev := Event{
		Timestamp: time.Now().UTC().Truncate(time.Millisecond),
		Type:      MergeFailed,
		AgentID:   "refinery-1",
		TaskID:    "al-123",
		Data: map[string]any{
			"branch":  "polecat/capable/al-92a",
			"error":   "conflict in events.go",
			"retries": float64(3),
		},
	}

	mustWrite(t, w, ev)

	r := NewReader(path)
	events, err := r.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	got := events[0]
	if got.Data["branch"] != "polecat/capable/al-92a" {
		t.Errorf("Data[branch]: got %v", got.Data["branch"])
	}
	if got.Data["error"] != "conflict in events.go" {
		t.Errorf("Data[error]: got %v", got.Data["error"])
	}
	if got.Data["retries"] != float64(3) {
		t.Errorf("Data[retries]: got %v", got.Data["retries"])
	}
}
