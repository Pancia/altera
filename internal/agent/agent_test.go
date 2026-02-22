package agent

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := NewStore(filepath.Join(dir, "agents"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

func sampleAgent(id string) *Agent {
	now := time.Now()
	return &Agent{
		ID:        id,
		Role:      RoleWorker,
		Rig:       "rig-1",
		Status:    StatusActive,
		Heartbeat: now,
		StartedAt: now,
		PID:       os.Getpid(),
	}
}

func TestCreateAndGet(t *testing.T) {
	s := newTestStore(t)
	a := sampleAgent("a1")
	a.CurrentTask = "task-1"
	a.Worktree = "/tmp/wt"
	a.TmuxSession = "sess-1"
	a.LastProgress = "building"

	if err := s.Create(a); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.Get("a1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != "a1" || got.Role != RoleWorker || got.Status != StatusActive {
		t.Errorf("fields mismatch: %+v", got)
	}
	if got.CurrentTask != "task-1" {
		t.Errorf("CurrentTask = %q, want %q", got.CurrentTask, "task-1")
	}
	if got.Worktree != "/tmp/wt" {
		t.Errorf("Worktree = %q, want %q", got.Worktree, "/tmp/wt")
	}
	if got.TmuxSession != "sess-1" {
		t.Errorf("TmuxSession = %q, want %q", got.TmuxSession, "sess-1")
	}
	if got.LastProgress != "building" {
		t.Errorf("LastProgress = %q, want %q", got.LastProgress, "building")
	}
	if got.PID != os.Getpid() {
		t.Errorf("PID = %d, want %d", got.PID, os.Getpid())
	}
}

func TestCreateDuplicate(t *testing.T) {
	s := newTestStore(t)
	a := sampleAgent("a1")
	if err := s.Create(a); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := s.Create(a); err != ErrExists {
		t.Fatalf("expected ErrExists, got %v", err)
	}
}

func TestGetNotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Get("nope")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdate(t *testing.T) {
	s := newTestStore(t)
	a := sampleAgent("a1")
	if err := s.Create(a); err != nil {
		t.Fatalf("Create: %v", err)
	}

	a.Status = StatusIdle
	a.CurrentTask = "task-2"
	if err := s.Update(a); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := s.Get("a1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Status != StatusIdle {
		t.Errorf("Status = %q, want %q", got.Status, StatusIdle)
	}
	if got.CurrentTask != "task-2" {
		t.Errorf("CurrentTask = %q, want %q", got.CurrentTask, "task-2")
	}
}

func TestUpdateNotFound(t *testing.T) {
	s := newTestStore(t)
	a := sampleAgent("nope")
	if err := s.Update(a); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	s := newTestStore(t)
	a := sampleAgent("a1")
	if err := s.Create(a); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := s.Delete("a1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := s.Get("a1")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestDeleteNotFound(t *testing.T) {
	s := newTestStore(t)
	if err := s.Delete("nope"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTouchHeartbeat(t *testing.T) {
	s := newTestStore(t)
	a := sampleAgent("a1")
	a.Heartbeat = time.Now().Add(-1 * time.Hour) // stale
	if err := s.Create(a); err != nil {
		t.Fatalf("Create: %v", err)
	}

	before := time.Now()
	if err := s.TouchHeartbeat("a1"); err != nil {
		t.Fatalf("TouchHeartbeat: %v", err)
	}

	got, _ := s.Get("a1")
	if got.Heartbeat.Before(before) {
		t.Errorf("heartbeat not updated: %v < %v", got.Heartbeat, before)
	}
}

func TestTouchHeartbeatNotFound(t *testing.T) {
	s := newTestStore(t)
	if err := s.TouchHeartbeat("nope"); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCheckPID_Alive(t *testing.T) {
	a := sampleAgent("a1")
	a.PID = os.Getpid()
	if !CheckPID(a) {
		t.Error("expected current process to be alive")
	}
}

func TestCheckPID_DeadProcess(t *testing.T) {
	a := sampleAgent("a1")
	a.PID = 999999999
	if CheckPID(a) {
		t.Error("expected dead process to return false")
	}
}

func TestCheckPID_NoPID(t *testing.T) {
	a := sampleAgent("a1")
	a.PID = 0
	if CheckPID(a) {
		t.Error("expected zero PID to return false")
	}
}

func TestHeartbeatStaleness_Fresh(t *testing.T) {
	a := sampleAgent("a1")
	a.Heartbeat = time.Now()
	s := HeartbeatStaleness(a)
	if s > time.Second {
		t.Errorf("expected fresh heartbeat staleness < 1s, got %v", s)
	}
}

func TestHeartbeatStaleness_Stale(t *testing.T) {
	a := sampleAgent("a1")
	a.Heartbeat = time.Now().Add(-5 * time.Minute)
	s := HeartbeatStaleness(a)
	if s < 4*time.Minute {
		t.Errorf("expected staleness >= 4m, got %v", s)
	}
}

func TestHeartbeatTimeoutConstants(t *testing.T) {
	if HeartbeatWarnTimeout >= HeartbeatCriticalTimeout {
		t.Errorf("warn (%v) should be < critical (%v)", HeartbeatWarnTimeout, HeartbeatCriticalTimeout)
	}
	if HeartbeatCriticalTimeout >= HeartbeatDeadTimeout {
		t.Errorf("critical (%v) should be < dead (%v)", HeartbeatCriticalTimeout, HeartbeatDeadTimeout)
	}
}

func TestCheckLiveness_Alive(t *testing.T) {
	a := sampleAgent("a1")
	a.PID = os.Getpid()
	a.Heartbeat = time.Now()
	if !CheckLiveness(a) {
		t.Error("expected alive agent to be live")
	}
}

func TestCheckLiveness_StaleHeartbeat(t *testing.T) {
	a := sampleAgent("a1")
	a.PID = os.Getpid()
	a.Heartbeat = time.Now().Add(-1 * time.Hour)
	if CheckLiveness(a) {
		t.Error("expected stale heartbeat to be not live")
	}
}

func TestCheckLiveness_DeadProcess(t *testing.T) {
	a := sampleAgent("a1")
	a.Heartbeat = time.Now()
	a.PID = 999999999
	if CheckLiveness(a) {
		t.Error("expected dead process to be not live")
	}
}

func TestCheckLiveness_NoPID(t *testing.T) {
	a := sampleAgent("a1")
	a.Heartbeat = time.Now()
	a.PID = 0
	if CheckLiveness(a) {
		t.Error("expected zero PID to be not live")
	}
}

func TestEscalationFields(t *testing.T) {
	s := newTestStore(t)
	a := sampleAgent("a1")
	a.EscalationLevel = "warning"
	a.LastEscalation = time.Now()

	if err := s.Create(a); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.Get("a1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.EscalationLevel != "warning" {
		t.Errorf("EscalationLevel = %q, want %q", got.EscalationLevel, "warning")
	}
	if got.LastEscalation.IsZero() {
		t.Error("LastEscalation should not be zero")
	}
}

func TestListByRole(t *testing.T) {
	s := newTestStore(t)
	for _, id := range []string{"w1", "w2", "l1"} {
		a := sampleAgent(id)
		if id == "l1" {
			a.Role = RoleLiaison
		}
		if err := s.Create(a); err != nil {
			t.Fatalf("Create %s: %v", id, err)
		}
	}

	workers, err := s.ListByRole(RoleWorker)
	if err != nil {
		t.Fatalf("ListByRole: %v", err)
	}
	if len(workers) != 2 {
		t.Errorf("ListByRole(worker) = %d, want 2", len(workers))
	}

	liaisons, err := s.ListByRole(RoleLiaison)
	if err != nil {
		t.Fatalf("ListByRole: %v", err)
	}
	if len(liaisons) != 1 {
		t.Errorf("ListByRole(liaison) = %d, want 1", len(liaisons))
	}

	resolvers, err := s.ListByRole(RoleResolver)
	if err != nil {
		t.Fatalf("ListByRole: %v", err)
	}
	if len(resolvers) != 0 {
		t.Errorf("ListByRole(resolver) = %d, want 0", len(resolvers))
	}
}

func TestListByStatus(t *testing.T) {
	s := newTestStore(t)
	for i, status := range []Status{StatusActive, StatusActive, StatusIdle, StatusDead} {
		a := sampleAgent(string(rune('a' + i)))
		a.Status = status
		if err := s.Create(a); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	active, err := s.ListByStatus(StatusActive)
	if err != nil {
		t.Fatalf("ListByStatus: %v", err)
	}
	if len(active) != 2 {
		t.Errorf("ListByStatus(active) = %d, want 2", len(active))
	}

	idle, err := s.ListByStatus(StatusIdle)
	if err != nil {
		t.Fatalf("ListByStatus: %v", err)
	}
	if len(idle) != 1 {
		t.Errorf("ListByStatus(idle) = %d, want 1", len(idle))
	}

	dead, err := s.ListByStatus(StatusDead)
	if err != nil {
		t.Fatalf("ListByStatus: %v", err)
	}
	if len(dead) != 1 {
		t.Errorf("ListByStatus(dead) = %d, want 1", len(dead))
	}
}

func TestCountByRole(t *testing.T) {
	s := newTestStore(t)

	// 2 active workers, 1 idle worker, 1 active liaison
	agents := []struct {
		id     string
		role   Role
		status Status
	}{
		{"w1", RoleWorker, StatusActive},
		{"w2", RoleWorker, StatusActive},
		{"w3", RoleWorker, StatusIdle},
		{"l1", RoleLiaison, StatusActive},
	}
	for _, tc := range agents {
		a := sampleAgent(tc.id)
		a.Role = tc.role
		a.Status = tc.status
		if err := s.Create(a); err != nil {
			t.Fatalf("Create %s: %v", tc.id, err)
		}
	}

	n, err := s.CountByRole(RoleWorker)
	if err != nil {
		t.Fatalf("CountByRole: %v", err)
	}
	if n != 2 {
		t.Errorf("CountByRole(worker) = %d, want 2 (only active)", n)
	}

	n, err = s.CountByRole(RoleLiaison)
	if err != nil {
		t.Fatalf("CountByRole: %v", err)
	}
	if n != 1 {
		t.Errorf("CountByRole(liaison) = %d, want 1", n)
	}

	n, err = s.CountByRole(RoleResolver)
	if err != nil {
		t.Fatalf("CountByRole: %v", err)
	}
	if n != 0 {
		t.Errorf("CountByRole(resolver) = %d, want 0", n)
	}
}

func TestAtomicWrite(t *testing.T) {
	s := newTestStore(t)
	a := sampleAgent("a1")
	if err := s.Create(a); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Verify the file exists and is valid JSON
	p := s.path("a1")
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) == 0 {
		t.Error("file is empty")
	}

	// Verify no temp files remain
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
			t.Errorf("stale temp file: %s", e.Name())
		}
	}
}

func TestListEmptyStore(t *testing.T) {
	s := newTestStore(t)

	agents, err := s.ListByRole(RoleWorker)
	if err != nil {
		t.Fatalf("ListByRole: %v", err)
	}
	if len(agents) != 0 {
		t.Errorf("expected empty list, got %d", len(agents))
	}

	agents, err = s.ListByStatus(StatusActive)
	if err != nil {
		t.Fatalf("ListByStatus: %v", err)
	}
	if len(agents) != 0 {
		t.Errorf("expected empty list, got %d", len(agents))
	}

	n, err := s.CountByRole(RoleWorker)
	if err != nil {
		t.Fatalf("CountByRole: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 count, got %d", n)
	}
}
