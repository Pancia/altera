package liaison

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/anthropics/altera/internal/agent"
	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/events"
	"github.com/anthropics/altera/internal/message"
	"github.com/anthropics/altera/internal/task"
	"github.com/anthropics/altera/internal/tmux"
)

// setupProject creates a minimal project directory with .alt/ and stores.
func setupProject(t *testing.T) (projectRoot string, m *Manager) {
	t.Helper()

	projectRoot = t.TempDir()
	altDir := filepath.Join(projectRoot, config.DirName)

	for _, d := range []string{
		altDir,
		filepath.Join(altDir, "agents"),
		filepath.Join(altDir, "tasks"),
		filepath.Join(altDir, "messages"),
		filepath.Join(altDir, "messages", "archive"),
		filepath.Join(altDir, "merge-queue"),
	} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	agents, err := agent.NewStore(filepath.Join(altDir, "agents"))
	if err != nil {
		t.Fatalf("NewStore agents: %v", err)
	}
	tasks, err := task.NewStore(projectRoot)
	if err != nil {
		t.Fatalf("NewStore tasks: %v", err)
	}
	msgs, err := message.NewStore(filepath.Join(altDir, "messages"))
	if err != nil {
		t.Fatalf("NewStore messages: %v", err)
	}
	evPath := filepath.Join(altDir, "events.jsonl")
	evReader := events.NewReader(evPath)

	mgr := NewManager(projectRoot, agents, tasks, msgs, evReader)
	return projectRoot, mgr
}

func TestLiaisonPrompt(t *testing.T) {
	prompt := LiaisonPrompt()

	for _, want := range []string{
		"Liaison Agent",
		"alt task create",
		"alt daemon status",
		"SessionStart",
		"UserPromptSubmit",
		"PreCompact",
		"help",
		"merge",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("LiaisonPrompt() missing %q", want)
		}
	}
}

func TestWriteClaudeMD(t *testing.T) {
	projectRoot, m := setupProject(t)

	if err := m.writeClaudeMD(); err != nil {
		t.Fatalf("writeClaudeMD: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(projectRoot, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}

	if !strings.Contains(string(data), "Liaison Agent") {
		t.Error("CLAUDE.md missing 'Liaison Agent'")
	}
}

func TestWriteClaudeSettings(t *testing.T) {
	projectRoot, m := setupProject(t)

	if err := m.writeClaudeSettings(); err != nil {
		t.Fatalf("writeClaudeSettings: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(projectRoot, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}

	var settings ClaudeSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("unmarshal settings: %v", err)
	}

	// Verify SessionStart hook.
	ss, ok := settings.Hooks["SessionStart"]
	if !ok || len(ss) == 0 {
		t.Fatal("missing SessionStart hook")
	}
	if ss[0].Command != "alt liaison prime" {
		t.Errorf("SessionStart command = %q, want %q", ss[0].Command, "alt liaison prime")
	}

	// Verify UserPromptSubmit hook.
	ups, ok := settings.Hooks["UserPromptSubmit"]
	if !ok || len(ups) == 0 {
		t.Fatal("missing UserPromptSubmit hook")
	}
	if ups[0].Command != "alt liaison check-messages" {
		t.Errorf("UserPromptSubmit command = %q, want %q", ups[0].Command, "alt liaison check-messages")
	}

	// Verify PreCompact hook.
	pc, ok := settings.Hooks["PreCompact"]
	if !ok || len(pc) == 0 {
		t.Fatal("missing PreCompact hook")
	}
	if pc[0].Command != "alt liaison prime" {
		t.Errorf("PreCompact command = %q, want %q", pc[0].Command, "alt liaison prime")
	}
}

func TestPrime_Empty(t *testing.T) {
	_, m := setupProject(t)

	summary, err := m.Prime()
	if err != nil {
		t.Fatalf("Prime: %v", err)
	}

	for _, want := range []string{
		"Altera System State",
		"Tasks",
		"No tasks",
		"Active Agents",
		"No active agents",
		"Merge Queue",
		"Empty",
	} {
		if !strings.Contains(summary, want) {
			t.Errorf("Prime() missing %q", want)
		}
	}
}

func TestPrime_WithState(t *testing.T) {
	projectRoot, m := setupProject(t)
	altDir := filepath.Join(projectRoot, config.DirName)

	// Create some tasks.
	for _, tk := range []*task.Task{
		{Title: "Build the widget", Status: task.StatusOpen, Rig: "my-rig"},
		{Title: "Fix the bug", Status: task.StatusInProgress, AssignedTo: "worker-01"},
	} {
		if err := m.tasks.Create(tk); err != nil {
			t.Fatalf("Create task: %v", err)
		}
	}

	// Create an active agent.
	a := &agent.Agent{
		ID:          "worker-01",
		Role:        agent.RoleWorker,
		Status:      agent.StatusActive,
		CurrentTask: "t-abc123",
		Heartbeat:   time.Now(),
		StartedAt:   time.Now(),
	}
	if err := m.agents.Create(a); err != nil {
		t.Fatalf("Create agent: %v", err)
	}

	// Add a merge queue item.
	mqDir := filepath.Join(altDir, "merge-queue")
	mqItem := map[string]any{
		"task_id":  "t-def456",
		"branch":   "alt/t-def456",
		"agent_id": "worker-02",
	}
	mqData, _ := json.MarshalIndent(mqItem, "", "  ")
	os.WriteFile(filepath.Join(mqDir, "1234567890-t-def456.json"), mqData, 0o644)

	// Write some events.
	evWriter := events.NewWriter(filepath.Join(altDir, "events.jsonl"))
	_ = evWriter.Append(events.Event{
		Timestamp: time.Now(),
		Type:      events.AgentSpawned,
		AgentID:   "worker-01",
		TaskID:    "t-abc123",
	})

	summary, err := m.Prime()
	if err != nil {
		t.Fatalf("Prime: %v", err)
	}

	for _, want := range []string{
		"Build the widget",
		"Fix the bug",
		"worker-01",
		"t-abc123",
		"t-def456",
		"agent_spawned",
	} {
		if !strings.Contains(summary, want) {
			t.Errorf("Prime() missing %q\nsummary:\n%s", want, summary)
		}
	}
}

func TestCheckMessages_Empty(t *testing.T) {
	_, m := setupProject(t)

	result, err := m.CheckMessages()
	if err != nil {
		t.Fatalf("CheckMessages: %v", err)
	}
	if result != "" {
		t.Errorf("CheckMessages() = %q, want empty", result)
	}
}

func TestCheckMessages_WithMessages(t *testing.T) {
	_, m := setupProject(t)

	// Create a help message addressed to the liaison.
	_, err := m.messages.Create(
		message.TypeHelp,
		"worker-01",
		AgentID,
		"t-abc123",
		map[string]any{
			"message": "stuck on implementation",
		},
	)
	if err != nil {
		t.Fatalf("Create message: %v", err)
	}

	result, err := m.CheckMessages()
	if err != nil {
		t.Fatalf("CheckMessages: %v", err)
	}

	for _, want := range []string{
		"Pending Messages (1)",
		"help",
		"worker-01",
		"t-abc123",
		"stuck on implementation",
	} {
		if !strings.Contains(result, want) {
			t.Errorf("CheckMessages() missing %q\nresult:\n%s", want, result)
		}
	}
}

func TestStartLiaison(t *testing.T) {
	if _, err := tmux.ListSessions(); err != nil {
		t.Skip("tmux not available")
	}

	_, m := setupProject(t)

	if err := m.StartLiaison(); err != nil {
		t.Fatalf("StartLiaison: %v", err)
	}
	t.Cleanup(func() {
		_ = tmux.KillSession(SessionName)
	})

	// Verify tmux session exists.
	if !tmux.SessionExists(SessionName) {
		t.Error("tmux session does not exist")
	}

	// Verify agent record was created.
	a, err := m.agents.Get(AgentID)
	if err != nil {
		t.Fatalf("Get agent: %v", err)
	}
	if a.Role != agent.RoleLiaison {
		t.Errorf("agent role = %q, want %q", a.Role, agent.RoleLiaison)
	}
	if a.Status != agent.StatusActive {
		t.Errorf("agent status = %q, want %q", a.Status, agent.StatusActive)
	}
	if a.TmuxSession != SessionName {
		t.Errorf("agent tmux session = %q, want %q", a.TmuxSession, SessionName)
	}

	// Verify CLAUDE.md was written.
	claudeMD := filepath.Join(m.projectRoot, "CLAUDE.md")
	if _, err := os.Stat(claudeMD); err != nil {
		t.Errorf("CLAUDE.md not found: %v", err)
	}

	// Verify .claude/settings.json was written.
	settingsPath := filepath.Join(m.projectRoot, ".claude", "settings.json")
	if _, err := os.Stat(settingsPath); err != nil {
		t.Errorf("settings.json not found: %v", err)
	}
}

func TestStartLiaison_AlreadyExists(t *testing.T) {
	if _, err := tmux.ListSessions(); err != nil {
		t.Skip("tmux not available")
	}

	_, m := setupProject(t)

	if err := m.StartLiaison(); err != nil {
		t.Fatalf("first StartLiaison: %v", err)
	}
	t.Cleanup(func() {
		_ = tmux.KillSession(SessionName)
	})

	// Second start should fail.
	err := m.StartLiaison()
	if err == nil {
		t.Fatal("expected error for double start")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error = %q, want it to mention 'already exists'", err)
	}
}

func TestStartLiaison_ReactivatesDeadAgent(t *testing.T) {
	if _, err := tmux.ListSessions(); err != nil {
		t.Skip("tmux not available")
	}

	_, m := setupProject(t)

	// Create a dead agent record first.
	dead := &agent.Agent{
		ID:        AgentID,
		Role:      agent.RoleLiaison,
		Status:    agent.StatusDead,
		Heartbeat: time.Now().Add(-1 * time.Hour),
		StartedAt: time.Now().Add(-2 * time.Hour),
	}
	if err := m.agents.Create(dead); err != nil {
		t.Fatalf("Create dead agent: %v", err)
	}

	if err := m.StartLiaison(); err != nil {
		t.Fatalf("StartLiaison: %v", err)
	}
	t.Cleanup(func() {
		_ = tmux.KillSession(SessionName)
	})

	// Should have reactivated the existing record.
	a, err := m.agents.Get(AgentID)
	if err != nil {
		t.Fatalf("Get agent: %v", err)
	}
	if a.Status != agent.StatusActive {
		t.Errorf("agent status = %q, want %q", a.Status, agent.StatusActive)
	}
}

func TestAttachLiaison_NoSession(t *testing.T) {
	// Make sure there's no session.
	if tmux.SessionExists(SessionName) {
		_ = tmux.KillSession(SessionName)
	}

	err := AttachLiaison()
	if err == nil {
		t.Fatal("expected error when no session exists")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("error = %q, want it to mention 'does not exist'", err)
	}
}

func TestPrime_TaskStatusGrouping(t *testing.T) {
	_, m := setupProject(t)

	// Create tasks with different statuses.
	statuses := []struct {
		title  string
		status task.Status
	}{
		{"Open task A", task.StatusOpen},
		{"Done task B", task.StatusDone},
		{"Failed task C", task.StatusFailed},
	}
	for _, s := range statuses {
		tk := &task.Task{Title: s.title, Status: s.status}
		if err := m.tasks.Create(tk); err != nil {
			t.Fatalf("Create task %q: %v", s.title, err)
		}
	}

	summary, err := m.Prime()
	if err != nil {
		t.Fatalf("Prime: %v", err)
	}

	// Verify all task titles appear.
	for _, s := range statuses {
		if !strings.Contains(summary, s.title) {
			t.Errorf("Prime() missing task %q", s.title)
		}
	}

	// Verify status headers appear.
	for _, status := range []string{"open", "done", "failed"} {
		if !strings.Contains(summary, status) {
			t.Errorf("Prime() missing status header %q", status)
		}
	}
}
