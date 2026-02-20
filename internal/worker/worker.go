// Package worker manages the lifecycle of Claude Code worker agents.
// Each worker runs in its own git worktree and tmux session, with
// sequential naming (worker-01, worker-02, etc.).
package worker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/anthropics/altera/internal/agent"
	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/events"
	"github.com/anthropics/altera/internal/git"
	"github.com/anthropics/altera/internal/task"
	"github.com/anthropics/altera/internal/tmux"
)

// Manager handles spawning and cleaning up workers. It coordinates the
// agent store, event log, and rig configuration needed for worker lifecycle.
type Manager struct {
	agents      *agent.Store
	eventWriter *events.Writer
	projectRoot string // root dir containing .alt/
}

// NewManager creates a Manager. projectRoot is the directory containing .alt/.
func NewManager(projectRoot string, agents *agent.Store, ew *events.Writer) *Manager {
	return &Manager{
		agents:      agents,
		eventWriter: ew,
		projectRoot: projectRoot,
	}
}

// nextWorkerNum determines the next sequential worker number by scanning
// existing agents. Returns 1 if none exist.
func (m *Manager) nextWorkerNum() (int, error) {
	workers, err := m.agents.ListByRole(agent.RoleWorker)
	if err != nil {
		return 0, fmt.Errorf("listing workers: %w", err)
	}
	if len(workers) == 0 {
		return 1, nil
	}

	// Find the highest existing worker number.
	maxNum := 0
	for _, w := range workers {
		num := parseWorkerNum(w.ID)
		if num > maxNum {
			maxNum = num
		}
	}
	return maxNum + 1, nil
}

// parseWorkerNum extracts the numeric suffix from a worker ID like "worker-03".
// Returns 0 if the format doesn't match.
func parseWorkerNum(id string) int {
	if !strings.HasPrefix(id, "worker-") {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimPrefix(id, "worker-"))
	if err != nil {
		return 0
	}
	return n
}

// workerID formats a sequential worker ID like "worker-01".
func workerID(num int) string {
	return fmt.Sprintf("worker-%02d", num)
}

// SpawnWorker creates a new worker for the given task and rig.
// It performs the following steps:
//  1. Create git worktree from rig's repo (branch: alt/t-{task.ID})
//  2. Set git author: alt-worker-{num} <worker-{num}@altera.local>
//  3. Place task.json in worktree root with task details
//  4. Generate .claude/settings.json with hooks
//  5. Write CLAUDE.md with worker system prompt
//  6. Start Claude Code in tmux session (alt-worker-{id})
//  7. Create agent record in .alt/agents/
func (m *Manager) SpawnWorker(t *task.Task, rigName string) (*agent.Agent, error) {
	altDir := filepath.Join(m.projectRoot, config.DirName)
	rc, err := config.LoadRig(altDir, rigName)
	if err != nil {
		return nil, fmt.Errorf("loading rig config: %w", err)
	}

	num, err := m.nextWorkerNum()
	if err != nil {
		return nil, err
	}
	id := workerID(num)

	// 1. Create git branch and worktree.
	branchName := "alt/t-" + t.ID
	worktreePath := filepath.Join(m.projectRoot, "worktrees", id)

	baseBranch := rc.DefaultBranch
	if baseBranch == "" {
		baseBranch = "main"
	}
	if err := git.CreateBranch(rc.RepoPath, branchName, baseBranch); err != nil {
		return nil, fmt.Errorf("creating branch: %w", err)
	}
	if err := git.CreateWorktree(rc.RepoPath, branchName, worktreePath); err != nil {
		// Clean up the branch on failure.
		_ = git.DeleteBranch(rc.RepoPath, branchName)
		return nil, fmt.Errorf("creating worktree: %w", err)
	}

	// From here on, if we fail, we must clean up the worktree and branch.
	cleanup := func() {
		_ = git.DeleteWorktree(rc.RepoPath, worktreePath)
		_ = git.DeleteBranch(rc.RepoPath, branchName)
	}

	// 2. Set git author.
	authorName := "alt-" + id
	authorEmail := id + "@altera.local"
	if err := git.SetAuthor(worktreePath, authorName, authorEmail); err != nil {
		cleanup()
		return nil, fmt.Errorf("setting author: %w", err)
	}

	// 3. Place task.json in worktree root.
	if err := writeTaskJSON(worktreePath, t); err != nil {
		cleanup()
		return nil, fmt.Errorf("writing task.json: %w", err)
	}

	// 4. Generate .claude/settings.json with hooks.
	if err := writeClaudeSettings(worktreePath, id); err != nil {
		cleanup()
		return nil, fmt.Errorf("writing claude settings: %w", err)
	}

	// 5. Write CLAUDE.md with worker system prompt.
	if err := writeClaudeMD(worktreePath, t, id, rigName); err != nil {
		cleanup()
		return nil, fmt.Errorf("writing CLAUDE.md: %w", err)
	}

	// 6. Start Claude Code in tmux session.
	sessionName := tmux.SessionName("worker", id)
	if err := tmux.CreateSession(sessionName); err != nil {
		cleanup()
		return nil, fmt.Errorf("creating tmux session: %w", err)
	}

	claudeCmd := fmt.Sprintf("cd %s && claude --dangerously-skip-permissions", worktreePath)
	if err := tmux.SendKeys(sessionName, claudeCmd); err != nil {
		_ = tmux.KillSession(sessionName)
		cleanup()
		return nil, fmt.Errorf("starting claude code: %w", err)
	}

	// 7. Create agent record.
	now := time.Now()
	a := &agent.Agent{
		ID:          id,
		Role:        agent.RoleWorker,
		Rig:         rigName,
		Status:      agent.StatusActive,
		CurrentTask: t.ID,
		Worktree:    worktreePath,
		TmuxSession: sessionName,
		Heartbeat:   now,
		StartedAt:   now,
	}
	if err := m.agents.Create(a); err != nil {
		_ = tmux.KillSession(sessionName)
		cleanup()
		return nil, fmt.Errorf("creating agent record: %w", err)
	}

	// Emit event.
	_ = m.eventWriter.Append(events.Event{
		Timestamp: now,
		Type:      events.AgentSpawned,
		AgentID:   id,
		TaskID:    t.ID,
		Data: map[string]any{
			"rig":      rigName,
			"worktree": worktreePath,
			"branch":   branchName,
		},
	})

	return a, nil
}

// CleanupWorker tears down a worker agent:
//  1. Kill tmux session
//  2. Delete git worktree
//  3. Archive agent record (set status=dead)
func (m *Manager) CleanupWorker(a *agent.Agent) error {
	var errs []string

	// 1. Kill tmux session.
	if a.TmuxSession != "" {
		if err := tmux.KillSession(a.TmuxSession); err != nil {
			errs = append(errs, fmt.Sprintf("kill tmux session: %v", err))
		}
	}

	// 2. Delete git worktree.
	if a.Worktree != "" && a.Rig != "" {
		altDir := filepath.Join(m.projectRoot, config.DirName)
		rc, err := config.LoadRig(altDir, a.Rig)
		if err != nil {
			errs = append(errs, fmt.Sprintf("load rig config: %v", err))
		} else {
			if err := git.DeleteWorktree(rc.RepoPath, a.Worktree); err != nil {
				errs = append(errs, fmt.Sprintf("delete worktree: %v", err))
			}
			// Also delete the branch.
			branchName := "alt/t-" + a.CurrentTask
			if a.CurrentTask != "" {
				if err := git.DeleteBranch(rc.RepoPath, branchName); err != nil {
					errs = append(errs, fmt.Sprintf("delete branch: %v", err))
				}
			}
		}
	}

	// 3. Archive agent record (set status=dead).
	a.Status = agent.StatusDead
	if err := m.agents.Update(a); err != nil {
		errs = append(errs, fmt.Sprintf("update agent status: %v", err))
	}

	// Emit event.
	_ = m.eventWriter.Append(events.Event{
		Timestamp: time.Now(),
		Type:      events.AgentDied,
		AgentID:   a.ID,
		TaskID:    a.CurrentTask,
	})

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// ListWorkers returns all worker agents sorted by ID.
func (m *Manager) ListWorkers() ([]*agent.Agent, error) {
	workers, err := m.agents.ListByRole(agent.RoleWorker)
	if err != nil {
		return nil, err
	}
	sort.Slice(workers, func(i, j int) bool {
		return workers[i].ID < workers[j].ID
	})
	return workers, nil
}

// writeTaskJSON writes a task as JSON to {worktree}/task.json.
func writeTaskJSON(worktreePath string, t *task.Task) error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling task: %w", err)
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(worktreePath, "task.json"), data, 0o644)
}

// ClaudeSettings represents the .claude/settings.json structure.
type ClaudeSettings struct {
	Hooks map[string][]HookEntry `json:"hooks"`
}

// HookEntry represents a single hook in Claude settings.
type HookEntry struct {
	Matcher string `json:"matcher"`
	Command string `json:"command"`
}

// writeClaudeSettings creates .claude/settings.json with heartbeat and
// checkpoint hooks for the given agent ID.
func writeClaudeSettings(worktreePath, agentID string) error {
	claudeDir := filepath.Join(worktreePath, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		return fmt.Errorf("creating .claude dir: %w", err)
	}

	settings := ClaudeSettings{
		Hooks: map[string][]HookEntry{
			"PreToolUse": {
				{
					Matcher: "",
					Command: fmt.Sprintf("alt heartbeat %s", agentID),
				},
			},
			"Stop": {
				{
					Matcher: "",
					Command: fmt.Sprintf("alt checkpoint %s", agentID),
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

// WorkerPrompt generates the CLAUDE.md system prompt for a worker agent.
func WorkerPrompt(t *task.Task, agentID, rigName string) string {
	return fmt.Sprintf(`# Worker Agent: %s

You are a worker agent in the Altera multi-agent system.

## Your Assignment

- **Task ID**: %s
- **Title**: %s
- **Rig**: %s

## Task Description

%s

## Instructions

1. Read task.json in your worktree root for full task details
2. Implement the required changes
3. Run the test command to verify your work
4. Commit your changes with a clear message
5. Use 'alt checkpoint %s' to report progress

## Hooks

Your session is configured with automatic hooks:
- **Heartbeat**: Sent before each tool use to signal you're alive
- **Checkpoint**: Sent when you stop to save progress

## Important

- Stay focused on your assigned task
- Commit early and often
- If you're stuck, report via checkpoint
- Do not modify files outside your task scope
`, agentID, t.ID, t.Title, rigName, t.Description, agentID)
}

// writeClaudeMD writes CLAUDE.md with the worker system prompt.
func writeClaudeMD(worktreePath string, t *task.Task, agentID, rigName string) error {
	content := WorkerPrompt(t, agentID, rigName)
	return os.WriteFile(filepath.Join(worktreePath, "CLAUDE.md"), []byte(content), 0o644)
}
