package events

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

const defaultPath = ".alt/events.jsonl"

// Writer appends events to a JSONL file with flock-based concurrency safety.
type Writer struct {
	path string
}

// NewWriter creates a Writer that appends to the given file path.
// If path is empty, it defaults to .alt/events.jsonl.
func NewWriter(path string) *Writer {
	if path == "" {
		path = defaultPath
	}
	return &Writer{path: path}
}

// Append writes one or more events to the log file atomically.
// It acquires an exclusive flock, appends the events as JSONL lines,
// then releases the lock.
func (w *Writer) Append(events ...Event) error {
	if len(events) == 0 {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(w.path), 0o755); err != nil {
		return fmt.Errorf("events: create dir: %w", err)
	}

	f, err := os.OpenFile(w.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("events: open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Acquire exclusive lock for concurrent append safety.
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("events: flock: %w", err)
	}
	defer func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }()

	enc := json.NewEncoder(f)
	for _, ev := range events {
		if err := enc.Encode(ev); err != nil {
			return fmt.Errorf("events: encode: %w", err)
		}
	}

	return nil
}

// Path returns the file path this writer appends to.
func (w *Writer) Path() string {
	return w.path
}
