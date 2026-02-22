// Package liaison manages the liaison agent, which translates between human
// intent and the Altera task system. The liaison runs as a Claude Code session
// in a tmux session (alt-liaison), primed with system state on each interaction.
package liaison

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/anthropics/altera/internal/agent"
	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/events"
	"github.com/anthropics/altera/internal/message"
	"github.com/anthropics/altera/internal/task"
	"github.com/anthropics/altera/internal/tmux"
)

const (
	// SessionName is the tmux session name for the liaison agent.
	SessionName = "alt-liaison"

	// AgentID is the agent ID used for the liaison in the agent store.
	AgentID = "liaison-01"
)

// Manager handles the liaison agent lifecycle and state injection.
type Manager struct {
	projectRoot string
	agents      *agent.Store
	tasks       *task.Store
	messages    *message.Store
	evReader    *events.Reader
}

// NewManager creates a Manager. projectRoot is the directory containing .alt/.
func NewManager(projectRoot string, agents *agent.Store, tasks *task.Store, msgs *message.Store, evReader *events.Reader) *Manager {
	return &Manager{
		projectRoot: projectRoot,
		agents:      agents,
		tasks:       tasks,
		messages:    msgs,
		evReader:    evReader,
	}
}

// StartLiaison creates a tmux session and starts Claude Code with the liaison
// CLAUDE.md. It registers a liaison agent in the agent store and writes
// .claude/settings.json with hooks.
func (m *Manager) StartLiaison() error {
	// Check if session already exists.
	if tmux.SessionExists(SessionName) {
		return fmt.Errorf("liaison session %q already exists (use AttachLiaison to connect)", SessionName)
	}

	// Write CLAUDE.md and .claude/settings.json to project root.
	if err := m.writeClaudeMD(); err != nil {
		return fmt.Errorf("writing CLAUDE.md: %w", err)
	}
	if err := m.writeClaudeSettings(); err != nil {
		return fmt.Errorf("writing settings.json: %w", err)
	}

	// Create tmux session.
	if err := tmux.CreateSession(SessionName); err != nil {
		return fmt.Errorf("creating tmux session: %w", err)
	}

	// Start Claude Code in the session.
	claudeCmd := fmt.Sprintf("cd %s && claude --dangerously-skip-permissions", m.projectRoot)
	if err := tmux.SendKeys(SessionName, claudeCmd); err != nil {
		_ = tmux.KillSession(SessionName)
		return fmt.Errorf("starting claude code: %w", err)
	}

	// Register agent.
	now := time.Now()
	a := &agent.Agent{
		ID:          AgentID,
		Role:        agent.RoleLiaison,
		Status:      agent.StatusActive,
		TmuxSession: SessionName,
		Heartbeat:   now,
		StartedAt:   now,
	}

	// If agent already exists (e.g. from a previous dead session), update it.
	if existing, err := m.agents.Get(AgentID); err == nil {
		existing.Status = agent.StatusActive
		existing.TmuxSession = SessionName
		existing.Heartbeat = now
		if err := m.agents.Update(existing); err != nil {
			_ = tmux.KillSession(SessionName)
			return fmt.Errorf("updating agent record: %w", err)
		}
	} else {
		if err := m.agents.Create(a); err != nil {
			_ = tmux.KillSession(SessionName)
			return fmt.Errorf("creating agent record: %w", err)
		}
	}

	return nil
}

// AttachLiaison attaches to the existing liaison tmux session.
func AttachLiaison() error {
	if !tmux.SessionExists(SessionName) {
		return fmt.Errorf("liaison session %q does not exist (use StartLiaison first)", SessionName)
	}
	return tmux.AttachSession(SessionName)
}

// Prime reads all system state and formats it as a summary string suitable
// for injection into the liaison's context. It reads tasks, agents, merge
// queue entries, and recent events.
func (m *Manager) Prime() (string, error) {
	var b strings.Builder

	b.WriteString("# Altera System State\n\n")
	b.WriteString(fmt.Sprintf("**Timestamp**: %s\n\n", time.Now().UTC().Format(time.RFC3339)))

	// Tasks summary.
	if err := m.writeTasks(&b); err != nil {
		return "", fmt.Errorf("reading tasks: %w", err)
	}

	// Agents summary.
	if err := m.writeAgents(&b); err != nil {
		return "", fmt.Errorf("reading agents: %w", err)
	}

	// Merge queue summary.
	if err := m.writeMergeQueue(&b); err != nil {
		return "", fmt.Errorf("reading merge queue: %w", err)
	}

	// Recent events.
	if err := m.writeRecentEvents(&b); err != nil {
		return "", fmt.Errorf("reading events: %w", err)
	}

	return b.String(), nil
}

// CheckMessages reads pending messages addressed to the liaison agent and
// formats them for display. Returns an empty string if no messages are pending.
func (m *Manager) CheckMessages() (string, error) {
	msgs, err := m.messages.ListPending(AgentID)
	if err != nil {
		return "", fmt.Errorf("listing pending messages: %w", err)
	}

	if len(msgs) == 0 {
		return "", nil
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Pending Messages (%d)\n\n", len(msgs)))

	for _, msg := range msgs {
		b.WriteString(fmt.Sprintf("## %s from %s\n", msg.Type, msg.From))
		b.WriteString(fmt.Sprintf("- **ID**: %s\n", msg.ID))
		if msg.TaskID != "" {
			b.WriteString(fmt.Sprintf("- **Task**: %s\n", msg.TaskID))
		}
		b.WriteString(fmt.Sprintf("- **Time**: %s\n", msg.CreatedAt.UTC().Format(time.RFC3339)))
		if len(msg.Payload) > 0 {
			b.WriteString("- **Payload**:\n")
			for k, v := range msg.Payload {
				b.WriteString(fmt.Sprintf("  - %s: %v\n", k, v))
			}
		}
		b.WriteString("\n")
	}

	return b.String(), nil
}

// --- File generation ---

// ClaudeSettings represents the .claude/settings.json structure.
type ClaudeSettings struct {
	Hooks map[string][]HookGroup `json:"hooks"`
}

// HookGroup represents a matcher + hooks pair in Claude settings.
type HookGroup struct {
	Matcher string    `json:"matcher"`
	Hooks   []HookCmd `json:"hooks"`
}

// HookCmd represents a single hook command.
type HookCmd struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// writeClaudeSettings creates .claude/settings.json with liaison hooks.
func (m *Manager) writeClaudeSettings() error {
	claudeDir := filepath.Join(m.projectRoot, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		return fmt.Errorf("creating .claude dir: %w", err)
	}

	settings := ClaudeSettings{
		Hooks: map[string][]HookGroup{
			"SessionStart": {
				{
					Matcher: "",
					Hooks:   []HookCmd{{Type: "command", Command: "alt prime"}},
				},
			},
			"UserPromptSubmit": {
				{
					Matcher: "",
					Hooks:   []HookCmd{{Type: "command", Command: "alt liaison check-messages"}},
				},
			},
			"PreCompact": {
				{
					Matcher: "",
					Hooks:   []HookCmd{{Type: "command", Command: "alt prime"}},
				},
			},
		},
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling settings: %w", err)
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0o644)
}

// LiaisonPrompt generates a minimal CLAUDE.md bootstrap prompt for the liaison agent.
func LiaisonPrompt() string {
	return `# Liaison Agent

Run ` + "`alt help liaison startup`" + ` for full instructions.
`
}

// writeClaudeMD writes CLAUDE.md with the liaison system prompt to the
// project root.
func (m *Manager) writeClaudeMD() error {
	content := LiaisonPrompt()
	return os.WriteFile(filepath.Join(m.projectRoot, "CLAUDE.md"), []byte(content), 0o644)
}

// --- Prime helpers ---

// writeTasks appends a summary of all tasks grouped by status.
func (m *Manager) writeTasks(b *strings.Builder) error {
	allTasks, err := m.tasks.List(task.Filter{})
	if err != nil {
		return err
	}

	b.WriteString("## Tasks\n\n")
	if len(allTasks) == 0 {
		b.WriteString("No tasks.\n\n")
		return nil
	}

	// Group by status.
	groups := map[task.Status][]*task.Task{}
	for _, t := range allTasks {
		groups[t.Status] = append(groups[t.Status], t)
	}

	statusOrder := []task.Status{
		task.StatusInProgress,
		task.StatusAssigned,
		task.StatusOpen,
		task.StatusDone,
		task.StatusFailed,
	}

	for _, status := range statusOrder {
		tasks := groups[status]
		if len(tasks) == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("### %s (%d)\n\n", status, len(tasks)))
		for _, t := range tasks {
			line := fmt.Sprintf("- **%s** `%s`", t.Title, t.ID)
			if t.AssignedTo != "" {
				line += fmt.Sprintf(" (assigned: %s)", t.AssignedTo)
			}
			if t.Rig != "" {
				line += fmt.Sprintf(" [rig: %s]", t.Rig)
			}
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}

	return nil
}

// writeAgents appends a summary of active agents.
func (m *Manager) writeAgents(b *strings.Builder) error {
	active, err := m.agents.ListByStatus(agent.StatusActive)
	if err != nil {
		return err
	}

	b.WriteString("## Active Agents\n\n")
	if len(active) == 0 {
		b.WriteString("No active agents.\n\n")
		return nil
	}

	for _, a := range active {
		age := time.Since(a.Heartbeat).Round(time.Second)
		b.WriteString(fmt.Sprintf("- **%s** (%s) heartbeat: %s ago", a.ID, a.Role, age))
		if a.CurrentTask != "" {
			b.WriteString(fmt.Sprintf(" task: %s", a.CurrentTask))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")

	return nil
}

// writeMergeQueue appends a summary of items in the merge queue.
func (m *Manager) writeMergeQueue(b *strings.Builder) error {
	queueDir := filepath.Join(m.projectRoot, config.DirName, "merge-queue")
	entries, err := os.ReadDir(queueDir)
	if err != nil {
		if os.IsNotExist(err) {
			b.WriteString("## Merge Queue\n\nEmpty.\n\n")
			return nil
		}
		return err
	}

	// Filter to .json files and sort by name (FIFO order).
	var jsonFiles []os.DirEntry
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			jsonFiles = append(jsonFiles, e)
		}
	}
	sort.Slice(jsonFiles, func(i, j int) bool {
		return jsonFiles[i].Name() < jsonFiles[j].Name()
	})

	b.WriteString(fmt.Sprintf("## Merge Queue (%d)\n\n", len(jsonFiles)))
	if len(jsonFiles) == 0 {
		b.WriteString("Empty.\n\n")
		return nil
	}

	for _, e := range jsonFiles {
		data, err := os.ReadFile(filepath.Join(queueDir, e.Name()))
		if err != nil {
			continue
		}
		var item struct {
			TaskID  string `json:"task_id"`
			Branch  string `json:"branch"`
			AgentID string `json:"agent_id"`
		}
		if err := json.Unmarshal(data, &item); err != nil {
			continue
		}
		b.WriteString(fmt.Sprintf("- task: %s branch: %s agent: %s\n", item.TaskID, item.Branch, item.AgentID))
	}
	b.WriteString("\n")

	return nil
}

// writeRecentEvents appends the last 20 events from the event log.
func (m *Manager) writeRecentEvents(b *strings.Builder) error {
	evts, err := m.evReader.Tail(20)
	if err != nil {
		// File may not exist yet - that's ok. The error may be wrapped,
		// so check with errors.Is as well.
		if os.IsNotExist(err) || errors.Is(err, os.ErrNotExist) || strings.Contains(err.Error(), "no such file") {
			b.WriteString("## Recent Events\n\nNo events yet.\n\n")
			return nil
		}
		return err
	}

	b.WriteString("## Recent Events\n\n")
	if len(evts) == 0 {
		b.WriteString("No events yet.\n\n")
		return nil
	}

	for _, ev := range evts {
		line := fmt.Sprintf("- `%s` **%s**", ev.Timestamp.UTC().Format("15:04:05"), ev.Type)
		if ev.AgentID != "" {
			line += fmt.Sprintf(" agent:%s", ev.AgentID)
		}
		if ev.TaskID != "" {
			line += fmt.Sprintf(" task:%s", ev.TaskID)
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("\n")

	return nil
}
