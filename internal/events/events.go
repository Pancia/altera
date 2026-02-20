// Package events provides an append-only event log for the Altera multi-agent
// orchestration system. Events are written as JSONL to .alt/events.jsonl with
// file-level locking (flock) for concurrent append safety.
package events

import (
	"encoding/json"
	"time"
)

// Type represents the kind of event that occurred.
type Type string

const (
	TaskCreated    Type = "task_created"
	TaskAssigned   Type = "task_assigned"
	TaskStarted    Type = "task_started"
	TaskDone       Type = "task_done"
	TaskFailed     Type = "task_failed"
	AgentSpawned   Type = "agent_spawned"
	AgentDied      Type = "agent_died"
	MergeStarted   Type = "merge_started"
	MergeSuccess   Type = "merge_success"
	MergeConflict  Type = "merge_conflict"
	MergeFailed    Type = "merge_failed"
	BudgetExceeded Type = "budget_exceeded"
	WorkerStalled  Type = "worker_stalled"
)

// Event represents a single event in the system log.
type Event struct {
	Timestamp time.Time      `json:"timestamp"`
	Type      Type           `json:"type"`
	AgentID   string         `json:"agent_id"`
	TaskID    string         `json:"task_id"`
	Data      map[string]any `json:"data,omitempty"`
}

// MarshalJSON produces a compact JSON line for an event.
func (e Event) MarshalJSON() ([]byte, error) {
	type Alias Event
	return json.Marshal(&struct {
		Timestamp string `json:"timestamp"`
		*Alias
	}{
		Timestamp: e.Timestamp.UTC().Format(time.RFC3339Nano),
		Alias:     (*Alias)(&e),
	})
}

// UnmarshalJSON parses an event from JSON.
func (e *Event) UnmarshalJSON(data []byte) error {
	type Alias Event
	aux := &struct {
		Timestamp string `json:"timestamp"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	t, err := time.Parse(time.RFC3339Nano, aux.Timestamp)
	if err != nil {
		return err
	}
	e.Timestamp = t
	return nil
}
