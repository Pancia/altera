package message

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Type represents the kind of message.
type Type string

const (
	TypeTaskDone    Type = "task_done"
	TypeMergeResult Type = "merge_result"
	TypeHelp        Type = "help"
	TypeCheckpoint  Type = "checkpoint"
	TypeUserMessage Type = "user_message"
)

// Message is the data model for inter-agent communication.
type Message struct {
	ID        string         `json:"id"`
	Type      Type           `json:"type"`
	From      string         `json:"from"`
	To        string         `json:"to"`
	TaskID    string         `json:"task_id,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

var (
	ErrNotFound    = errors.New("message not found")
	ErrInvalidType = errors.New("invalid message type")
)

// validTypes enumerates the accepted message types.
var validTypes = map[Type]bool{
	TypeTaskDone:    true,
	TypeMergeResult: true,
	TypeHelp:        true,
	TypeCheckpoint:  true,
	TypeUserMessage: true,
}

// Store manages message persistence in the filesystem.
type Store struct {
	dir string // e.g. ".alt/messages"
}

// NewStore creates a Store rooted at the given directory.
// The directory and its archive subdirectory are created if they do not exist.
func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create message store dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "archive"), 0o755); err != nil {
		return nil, fmt.Errorf("create message archive dir: %w", err)
	}
	return &Store{dir: dir}, nil
}

// generateID produces an ID in the form m-{random6chars}.
func generateID() (string, error) {
	b := make([]byte, 3) // 3 bytes = 6 hex chars
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random id: %w", err)
	}
	return "m-" + hex.EncodeToString(b), nil
}

// filename builds the canonical filename for a message:
// {timestamp}-{type}-{id}.json where timestamp is Unix nanos.
func filename(m *Message) string {
	return fmt.Sprintf("%d-%s-%s.json", m.CreatedAt.UnixNano(), m.Type, m.ID)
}

// writeAtomic writes data to path via temp-file + rename.
func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-msg-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}

// Create persists a new message. The ID and CreatedAt fields are set
// automatically. Returns the created message.
func (s *Store) Create(msgType Type, from, to, taskID string, payload map[string]any) (*Message, error) {
	if !validTypes[msgType] {
		return nil, ErrInvalidType
	}
	id, err := generateID()
	if err != nil {
		return nil, err
	}
	m := &Message{
		ID:        id,
		Type:      msgType,
		From:      from,
		To:        to,
		TaskID:    taskID,
		Payload:   payload,
		CreatedAt: time.Now(),
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal message: %w", err)
	}
	if err := writeAtomic(filepath.Join(s.dir, filename(m)), data); err != nil {
		return nil, err
	}
	return m, nil
}

// findFile scans the store directory for a file whose name ends with -{id}.json.
func (s *Store) findFile(id string) (string, error) {
	suffix := "-" + id + ".json"
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return "", fmt.Errorf("read message dir: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), suffix) {
			return filepath.Join(s.dir, e.Name()), nil
		}
	}
	return "", ErrNotFound
}

// Get reads a message by ID. Returns ErrNotFound if absent.
func (s *Store) Get(id string) (*Message, error) {
	path, err := s.findFile(id)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read message file: %w", err)
	}
	var m Message
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("unmarshal message: %w", err)
	}
	return &m, nil
}

// Delete removes a message by ID. Returns ErrNotFound if absent.
func (s *Store) Delete(id string) error {
	path, err := s.findFile(id)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete message file: %w", err)
	}
	return nil
}

// ListPending returns all messages addressed to the given recipient,
// ordered by timestamp (oldest first). The ordering is naturally
// provided by the filename prefix (Unix nanos).
func (s *Store) ListPending(to string) ([]*Message, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("read message dir: %w", err)
	}

	// Sort entries by name to guarantee timestamp ordering.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var msgs []*Message
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			continue
		}
		var m Message
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		if m.To == to {
			msgs = append(msgs, &m)
		}
	}
	return msgs, nil
}

// Archive moves a message to the archive subdirectory.
// Returns ErrNotFound if the message does not exist.
func (s *Store) Archive(id string) error {
	path, err := s.findFile(id)
	if err != nil {
		return err
	}
	dest := filepath.Join(s.dir, "archive", filepath.Base(path))
	if err := os.Rename(path, dest); err != nil {
		return fmt.Errorf("archive message: %w", err)
	}
	return nil
}
