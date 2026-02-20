package events

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Filter controls which events are returned by Read.
type Filter struct {
	Type    Type      // If non-empty, match only this event type.
	AgentID string    // If non-empty, match only this agent.
	TaskID  string    // If non-empty, match only this task.
	After   time.Time // If non-zero, match events after this time.
	Before  time.Time // If non-zero, match events before this time.
}

func (f Filter) matches(e *Event) bool {
	if f.Type != "" && e.Type != f.Type {
		return false
	}
	if f.AgentID != "" && e.AgentID != f.AgentID {
		return false
	}
	if f.TaskID != "" && e.TaskID != f.TaskID {
		return false
	}
	if !f.After.IsZero() && !e.Timestamp.After(f.After) {
		return false
	}
	if !f.Before.IsZero() && !e.Timestamp.Before(f.Before) {
		return false
	}
	return true
}

// Reader reads events from a JSONL file.
type Reader struct {
	path string
}

// NewReader creates a Reader for the given file path.
// If path is empty, it defaults to .alt/events.jsonl.
func NewReader(path string) *Reader {
	if path == "" {
		path = defaultPath
	}
	return &Reader{path: path}
}

// Read returns all events matching the given filter.
// A zero-value Filter matches all events.
func (r *Reader) Read(filter Filter) ([]Event, error) {
	f, err := os.Open(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("events: open: %w", err)
	}
	defer f.Close()

	var result []Event
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var ev Event
		if err := json.Unmarshal(line, &ev); err != nil {
			continue // skip corrupt lines
		}
		if filter.matches(&ev) {
			result = append(result, ev)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("events: scan: %w", err)
	}
	return result, nil
}

// ReadAll returns every event in the log.
func (r *Reader) ReadAll() ([]Event, error) {
	return r.Read(Filter{})
}

// Tail returns the last n events from the log.
func (r *Reader) Tail(n int) ([]Event, error) {
	if n <= 0 {
		return nil, nil
	}

	all, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(all) <= n {
		return all, nil
	}
	return all[len(all)-n:], nil
}

// Path returns the file path this reader reads from.
func (r *Reader) Path() string {
	return r.path
}
