// Package merge implements a FIFO merge queue and merge pipeline for the
// Altera orchestration system. The queue stores entries as individual JSON
// files in a directory (compatible with the constraints package's QueueDepth
// counter), and the pipeline coordinates merge attempts with three outcomes:
// success, conflict, and test failure.
package merge

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Sentinel errors for queue operations.
var (
	ErrQueueEmpty   = errors.New("merge queue is empty")
	ErrAlreadyQueued = errors.New("task already in merge queue")
)

// QueueEntry represents a single item in the merge queue.
type QueueEntry struct {
	TaskID     string    `json:"task_id"`
	EnqueuedAt time.Time `json:"enqueued_at"`
}

// Queue is a FIFO merge queue backed by individual JSON files in a directory.
// Each entry is stored as {unix_nanos}-{taskID}.json to provide natural FIFO
// ordering via filename sort. This layout is also compatible with the
// constraints package, which counts .json files in the queue directory.
type Queue struct {
	dir string
}

// NewQueue creates a Queue backed by the given directory, creating it if needed.
func NewQueue(dir string) (*Queue, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create merge queue dir: %w", err)
	}
	return &Queue{dir: dir}, nil
}

// Enqueue appends a task to the end of the queue. Returns ErrAlreadyQueued
// if the task is already present.
func (q *Queue) Enqueue(taskID string) error {
	files, err := q.list()
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.entry.TaskID == taskID {
			return ErrAlreadyQueued
		}
	}

	entry := QueueEntry{
		TaskID:     taskID,
		EnqueuedAt: time.Now().UTC(),
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal queue entry: %w", err)
	}
	data = append(data, '\n')

	name := fmt.Sprintf("%d-%s.json", entry.EnqueuedAt.UnixNano(), taskID)
	return atomicWrite(filepath.Join(q.dir, name), data)
}

// Dequeue removes and returns the first task ID in the queue.
// Returns ErrQueueEmpty if the queue has no items.
func (q *Queue) Dequeue() (string, error) {
	files, err := q.list()
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", ErrQueueEmpty
	}
	first := files[0]
	if err := os.Remove(first.path); err != nil {
		return "", fmt.Errorf("remove queue entry: %w", err)
	}
	return first.entry.TaskID, nil
}

// Peek returns the first task ID without removing it.
// Returns ErrQueueEmpty if the queue has no items.
func (q *Queue) Peek() (string, error) {
	files, err := q.list()
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", ErrQueueEmpty
	}
	return files[0].entry.TaskID, nil
}

// Len returns the number of items in the queue.
func (q *Queue) Len() (int, error) {
	files, err := q.list()
	if err != nil {
		return 0, err
	}
	return len(files), nil
}

// Dir returns the queue's backing directory path.
func (q *Queue) Dir() string {
	return q.dir
}

// queueFile pairs a parsed entry with its on-disk path for internal use.
type queueFile struct {
	path  string
	entry QueueEntry
}

// list reads all queue entries sorted by filename (timestamp order = FIFO).
func (q *Queue) list() ([]queueFile, error) {
	entries, err := os.ReadDir(q.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read merge queue dir: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var files []queueFile
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		p := filepath.Join(q.dir, e.Name())
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var qe QueueEntry
		if err := json.Unmarshal(data, &qe); err != nil {
			continue
		}
		files = append(files, queueFile{path: p, entry: qe})
	}
	return files, nil
}

// atomicWrite writes data to path via temp-file + rename.
func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-mq-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}
