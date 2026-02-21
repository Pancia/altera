// Package resolver manages the lifecycle of Claude Code resolver agents.
// Each resolver runs in its own git worktree and tmux session, tasked with
// resolving merge conflicts detected by the merge pipeline. Resolvers are
// given a conflict-context.json describing the conflicting files and both
// sides' intent, then use Claude Code to produce a clean resolution.
package resolver

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
	"github.com/anthropics/altera/internal/merge"
	"github.com/anthropics/altera/internal/tmux"
)

// ConflictContext holds all the information a resolver agent needs to
// understand and resolve merge conflicts.
type ConflictContext struct {
	TaskID          string               `json:"task_id"`
	Branch          string               `json:"branch"`
	BaseBranch      string               `json:"base_branch"`
	RigName         string               `json:"rig_name"`
	Conflicts       []merge.ConflictInfo `json:"conflicts"`
	TaskDescription string               `json:"task_description"`
}

// Manager handles spawning, detecting resolution, and cleaning up resolvers.
type Manager struct {
	agents      *agent.Store
	eventWriter *events.Writer
	projectRoot string
}

// NewManager creates a Manager. projectRoot is the directory containing .alt/.
func NewManager(projectRoot string, agents *agent.Store, ew *events.Writer) *Manager {
	return &Manager{
		agents:      agents,
		eventWriter: ew,
		projectRoot: projectRoot,
	}
}

// nextResolverNum determines the next sequential resolver number by scanning
// existing agents. Returns 1 if none exist.
func (m *Manager) nextResolverNum() (int, error) {
	resolvers, err := m.agents.ListByRole(agent.RoleResolver)
	if err != nil {
		return 0, fmt.Errorf("listing resolvers: %w", err)
	}
	if len(resolvers) == 0 {
		return 1, nil
	}

	maxNum := 0
	for _, r := range resolvers {
		num := parseResolverNum(r.ID)
		if num > maxNum {
			maxNum = num
		}
	}
	return maxNum + 1, nil
}

// parseResolverNum extracts the numeric suffix from a resolver ID like "resolver-03".
// Returns 0 if the format doesn't match.
func parseResolverNum(id string) int {
	if !strings.HasPrefix(id, "resolver-") {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimPrefix(id, "resolver-"))
	if err != nil {
		return 0
	}
	return n
}

// resolverID formats a sequential resolver ID like "resolver-01".
func resolverID(num int) string {
	return fmt.Sprintf("resolver-%02d", num)
}

// SpawnResolver creates a new resolver agent for the given conflict context.
// It performs the following steps:
//  1. Create git worktree from rig's repo with conflict state
//  2. Set git author
//  3. Place conflict-context.json with conflict details
//  4. Generate .claude/settings.json with heartbeat hook
//  5. Write CLAUDE.md with resolver system prompt
//  6. Start Claude Code in tmux session (alt-resolver-{id})
//  7. Create agent record in .alt/agents/
func (m *Manager) SpawnResolver(ctx ConflictContext) (*agent.Agent, error) {
	altDir := filepath.Join(m.projectRoot, config.DirName)
	rc, err := config.LoadRig(altDir, ctx.RigName)
	if err != nil {
		return nil, fmt.Errorf("loading rig config: %w", err)
	}

	num, err := m.nextResolverNum()
	if err != nil {
		return nil, err
	}
	id := resolverID(num)

	// 1. Create git branch and worktree with conflict state.
	branchName := "alt/resolve-" + ctx.TaskID
	worktreePath := filepath.Join(m.projectRoot, "worktrees", id)

	baseBranch := ctx.BaseBranch
	if baseBranch == "" {
		baseBranch = rc.DefaultBranch
	}
	if baseBranch == "" {
		baseBranch = "main"
	}

	// Create a branch from the base branch (e.g. main) for the resolver to work on.
	if err := git.CreateBranch(rc.RepoPath, branchName, baseBranch); err != nil {
		return nil, fmt.Errorf("creating branch: %w", err)
	}
	if err := git.CreateWorktree(rc.RepoPath, branchName, worktreePath); err != nil {
		_ = git.DeleteBranch(rc.RepoPath, branchName)
		return nil, fmt.Errorf("creating worktree: %w", err)
	}

	cleanup := func() {
		_ = git.DeleteWorktree(rc.RepoPath, worktreePath)
		_ = git.DeleteBranch(rc.RepoPath, branchName)
	}

	// Merge the conflicting branch into the worktree so the resolver sees
	// actual conflict markers in the files.
	mr, err := git.Merge(worktreePath, ctx.Branch)
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("merging branch for conflict state: %w", err)
	}
	// If the merge was clean, there's nothing to resolve.
	if mr.Clean {
		cleanup()
		return nil, fmt.Errorf("no conflicts found merging %s into %s", ctx.Branch, baseBranch)
	}

	// 2. Set git author.
	authorName := "alt-" + id
	authorEmail := id + "@altera.local"
	if err := git.SetAuthor(worktreePath, authorName, authorEmail); err != nil {
		cleanup()
		return nil, fmt.Errorf("setting author: %w", err)
	}

	// 3. Place conflict-context.json in worktree root.
	if err := writeConflictContext(worktreePath, ctx); err != nil {
		cleanup()
		return nil, fmt.Errorf("writing conflict-context.json: %w", err)
	}

	// 4. Generate .claude/settings.json with hooks.
	if err := writeClaudeSettings(worktreePath, id); err != nil {
		cleanup()
		return nil, fmt.Errorf("writing claude settings: %w", err)
	}

	// 5. Write CLAUDE.md with resolver system prompt.
	if err := writeClaudeMD(worktreePath, ctx, id); err != nil {
		cleanup()
		return nil, fmt.Errorf("writing CLAUDE.md: %w", err)
	}

	// 6. Start Claude Code in tmux session.
	sessionName := tmux.SessionName("resolver", id)
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
		Role:        agent.RoleResolver,
		Rig:         ctx.RigName,
		Status:      agent.StatusActive,
		CurrentTask: ctx.TaskID,
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

	_ = m.eventWriter.Append(events.Event{
		Timestamp: now,
		Type:      events.AgentSpawned,
		AgentID:   id,
		TaskID:    ctx.TaskID,
		Data: map[string]any{
			"rig":      ctx.RigName,
			"role":     "resolver",
			"worktree": worktreePath,
			"branch":   branchName,
		},
	})

	return a, nil
}

// DetectResolution checks whether a resolver agent has completed its work.
// Resolution is detected when:
//  1. No conflict markers remain in the conflicting files
//  2. A new commit exists beyond the merge commit
func DetectResolution(a *agent.Agent, conflicts []merge.ConflictInfo) (bool, error) {
	if a.Worktree == "" {
		return false, fmt.Errorf("agent %s has no worktree", a.ID)
	}

	// Check for remaining conflict markers in the conflicting files.
	for _, c := range conflicts {
		fullPath := filepath.Join(a.Worktree, c.Path)
		if hasConflictMarkers(fullPath) {
			return false, nil
		}
	}

	// Check that the working tree is clean (changes committed).
	clean, err := git.IsClean(a.Worktree)
	if err != nil {
		return false, fmt.Errorf("checking clean status: %w", err)
	}
	if !clean {
		return false, nil
	}

	return true, nil
}

// hasConflictMarkers reads a file and checks for git conflict markers.
func hasConflictMarkers(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	content := string(data)
	return strings.Contains(content, "<<<<<<<") ||
		strings.Contains(content, ">>>>>>>")
}

// CleanupResolver tears down a resolver agent:
//  1. Kill tmux session
//  2. Delete git worktree
//  3. Archive agent record (set status=dead)
func (m *Manager) CleanupResolver(a *agent.Agent) error {
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
			branchName := "alt/resolve-" + a.CurrentTask
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

// ListResolvers returns all resolver agents sorted by ID.
func (m *Manager) ListResolvers() ([]*agent.Agent, error) {
	resolvers, err := m.agents.ListByRole(agent.RoleResolver)
	if err != nil {
		return nil, err
	}
	sort.Slice(resolvers, func(i, j int) bool {
		return resolvers[i].ID < resolvers[j].ID
	})
	return resolvers, nil
}

// writeConflictContext writes the conflict context as JSON to
// {worktree}/conflict-context.json.
func writeConflictContext(worktreePath string, ctx ConflictContext) error {
	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling conflict context: %w", err)
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(worktreePath, "conflict-context.json"), data, 0o644)
}

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

// writeClaudeSettings creates .claude/settings.json with a heartbeat hook
// for the given agent ID.
func writeClaudeSettings(worktreePath, agentID string) error {
	claudeDir := filepath.Join(worktreePath, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		return fmt.Errorf("creating .claude dir: %w", err)
	}

	settings := ClaudeSettings{
		Hooks: map[string][]HookGroup{
			"PreToolUse": {
				{
					Matcher: "",
					Hooks:   []HookCmd{{Type: "command", Command: fmt.Sprintf("alt heartbeat %s", agentID)}},
				},
			},
			"Stop": {
				{
					Matcher: "",
					Hooks:   []HookCmd{{Type: "command", Command: fmt.Sprintf("alt checkpoint %s", agentID)}},
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

// ResolverPrompt generates the CLAUDE.md system prompt for a resolver agent.
func ResolverPrompt(ctx ConflictContext, agentID string) string {
	// Build the list of conflicting files.
	var fileList strings.Builder
	for _, c := range ctx.Conflicts {
		fmt.Fprintf(&fileList, "- `%s` (%d conflict regions)\n", c.Path, len(c.Markers))
	}

	return fmt.Sprintf(`# Resolver Agent: %s

You are a resolver agent in the Altera multi-agent system. Your sole job is to
resolve merge conflicts in this worktree, then commit the resolution and exit.

## Conflict Context

- **Task ID**: %s
- **Branch being merged**: %s
- **Base branch**: %s
- **Rig**: %s

## Task Description

%s

## Conflicting Files

%s
## Instructions

1. Read conflict-context.json in your worktree root for full conflict details
2. Examine each conflicting file — look for <<<<<<< / ======= / >>>>>>> markers
3. Understand both sides of each conflict:
   - **Ours** (between <<<<<<< and =======): changes from the base branch
   - **Theirs** (between ======= and >>>>>>>): changes from the task branch
4. Resolve each conflict by choosing the correct combination of changes
5. Remove ALL conflict markers — no <<<<<<< / ======= / >>>>>>> may remain
6. Stage your resolved files: git add <resolved-files>
7. Commit with message: "resolve: merge conflicts for %s"
8. Verify no conflict markers remain in any file
9. Exit when done

## Rules

- Resolve conflicts to preserve the intent of BOTH sides where possible
- If the changes are incompatible, prefer the task branch changes (theirs)
- Do NOT modify files that are not in conflict
- Do NOT add new features or refactor — only resolve conflicts
- Commit exactly once with all resolutions
- After committing, verify with 'git diff --check' that no markers remain

## Hooks

Your session is configured with automatic hooks:
- **Heartbeat**: Sent before each tool use to signal you're alive
- **Checkpoint**: Sent when you stop to save progress
`, agentID, ctx.TaskID, ctx.Branch, ctx.BaseBranch, ctx.RigName,
		ctx.TaskDescription, fileList.String(), ctx.TaskID)
}

// writeClaudeMD writes CLAUDE.md with the resolver system prompt.
func writeClaudeMD(worktreePath string, ctx ConflictContext, agentID string) error {
	content := ResolverPrompt(ctx, agentID)
	return os.WriteFile(filepath.Join(worktreePath, "CLAUDE.md"), []byte(content), 0o644)
}
