package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// Role represents the function an agent serves.
type Role string

const (
	RoleWorker   Role = "worker"
	RoleLiaison  Role = "liaison"
	RoleResolver Role = "resolver"
)

// Status represents the lifecycle state of an agent.
type Status string

const (
	StatusActive Status = "active"
	StatusIdle   Status = "idle"
	StatusDead   Status = "dead"
)

// Agent is the data model for a running agent instance.
type Agent struct {
	ID           string    `json:"id"`
	Role         Role      `json:"role"`
	Status       Status    `json:"status"`
	CurrentTask  string    `json:"current_task,omitempty"`
	Worktree     string    `json:"worktree,omitempty"`
	SessionDir   string    `json:"session_dir,omitempty"`
	TmuxSession  string    `json:"tmux_session,omitempty"`
	PID          int       `json:"pid,omitempty"`
	Heartbeat    time.Time `json:"heartbeat"`
	LastProgress      string    `json:"last_progress,omitempty"`
	StartedAt         time.Time `json:"started_at"`
	LastStallNotified time.Time `json:"last_stall_notified,omitempty"`
	EscalationLevel   string    `json:"escalation_level,omitempty"`
	LastEscalation    time.Time `json:"last_escalation,omitempty"`
}

var (
	ErrNotFound = errors.New("agent not found")
	ErrExists   = errors.New("agent already exists")
)

// Store manages agent persistence in the filesystem.
type Store struct {
	dir string // e.g. ".alt/agents"
}

// NewStore creates a Store rooted at the given directory.
// The directory is created if it does not exist.
func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create agent store dir: %w", err)
	}
	return &Store{dir: dir}, nil
}

func (s *Store) path(id string) string {
	return filepath.Join(s.dir, id+".json")
}

// writeAtomic writes data to path via temp-file + rename.
func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-agent-*")
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

// Create persists a new agent. Returns ErrExists if the ID is taken.
func (s *Store) Create(a *Agent) error {
	p := s.path(a.ID)
	if _, err := os.Stat(p); err == nil {
		return ErrExists
	}
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal agent: %w", err)
	}
	return writeAtomic(p, data)
}

// Get reads an agent by ID. Returns ErrNotFound if absent.
func (s *Store) Get(id string) (*Agent, error) {
	data, err := os.ReadFile(s.path(id))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("read agent file: %w", err)
	}
	var a Agent
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("unmarshal agent: %w", err)
	}
	return &a, nil
}

// Update overwrites an existing agent. Returns ErrNotFound if absent.
func (s *Store) Update(a *Agent) error {
	if _, err := os.Stat(s.path(a.ID)); os.IsNotExist(err) {
		return ErrNotFound
	}
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal agent: %w", err)
	}
	return writeAtomic(s.path(a.ID), data)
}

// Delete removes an agent by ID. Returns ErrNotFound if absent.
func (s *Store) Delete(id string) error {
	err := os.Remove(s.path(id))
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return fmt.Errorf("delete agent file: %w", err)
	}
	return nil
}

// TouchHeartbeat updates the heartbeat timestamp to now.
func (s *Store) TouchHeartbeat(id string) error {
	a, err := s.Get(id)
	if err != nil {
		return err
	}
	a.Heartbeat = time.Now()
	return s.Update(a)
}

// Escalation level values for the EscalationLevel field.
const (
	EscalationWarning  = "warning"
	EscalationCritical = "critical"
)

// Heartbeat escalation thresholds.
var (
	HeartbeatWarnTimeout     = 3 * time.Minute
	HeartbeatCriticalTimeout = 6 * time.Minute
	HeartbeatDeadTimeout     = 10 * time.Minute
)

// HeartbeatTimeout is kept for backward compatibility with callers that
// reference the original threshold. It equals HeartbeatWarnTimeout.
var HeartbeatTimeout = HeartbeatWarnTimeout

// CheckPID returns true if the agent's OS process is still running
// (verified via signal 0).
func CheckPID(a *Agent) bool {
	if a.PID <= 0 {
		return false
	}
	proc, err := os.FindProcess(a.PID)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

// HeartbeatStaleness returns how long it has been since the agent's last
// heartbeat. A zero or negative duration means the heartbeat is fresh.
func HeartbeatStaleness(a *Agent) time.Duration {
	return time.Since(a.Heartbeat)
}

// CheckLiveness returns true if the agent's heartbeat is fresh and its
// OS process still exists. It is preserved for backward compatibility;
// new code should use CheckPID and HeartbeatStaleness directly.
func CheckLiveness(a *Agent) bool {
	if HeartbeatStaleness(a) > HeartbeatTimeout {
		return false
	}
	return CheckPID(a)
}

// listAll reads every agent file in the store directory.
func (s *Store) listAll() ([]*Agent, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("read agent dir: %w", err)
	}
	var agents []*Agent
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		id := e.Name()[:len(e.Name())-len(".json")]
		a, err := s.Get(id)
		if err != nil {
			continue // skip corrupt files
		}
		agents = append(agents, a)
	}
	return agents, nil
}

// ListByRole returns all agents with the given role.
func (s *Store) ListByRole(role Role) ([]*Agent, error) {
	all, err := s.listAll()
	if err != nil {
		return nil, err
	}
	var out []*Agent
	for _, a := range all {
		if a.Role == role {
			out = append(out, a)
		}
	}
	return out, nil
}

// ListByStatus returns all agents with the given status.
func (s *Store) ListByStatus(status Status) ([]*Agent, error) {
	all, err := s.listAll()
	if err != nil {
		return nil, err
	}
	var out []*Agent
	for _, a := range all {
		if a.Status == status {
			out = append(out, a)
		}
	}
	return out, nil
}

// CountByRole returns the count of active agents with the given role.
func (s *Store) CountByRole(role Role) (int, error) {
	all, err := s.listAll()
	if err != nil {
		return 0, err
	}
	n := 0
	for _, a := range all {
		if a.Role == role && a.Status == StatusActive {
			n++
		}
	}
	return n, nil
}
