package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/anthropics/altera/internal/agent"
	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/events"
	"github.com/anthropics/altera/internal/liaison"
	"github.com/anthropics/altera/internal/message"
	"github.com/anthropics/altera/internal/task"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(primeCmd)
	primeCmd.Flags().StringVar(&primeRole, "role", "", "force role (liaison or worker)")
	primeCmd.Flags().StringVar(&primeAgentID, "agent-id", "", "explicit agent ID")
}

var (
	primeRole    string
	primeAgentID string
)

var primeCmd = &cobra.Command{
	Use:   "prime",
	Short: "Generate a dynamic system prompt for the current agent",
	Long: `Detects the current agent context and outputs a full system prompt
with runtime state. Detection order: ALT_AGENT_ID env var, --agent-id flag,
tmux session name, working directory, default to liaison.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}
		root := filepath.Dir(altDir)

		role, agentID := detectRole(root, altDir)

		switch role {
		case "worker":
			return primeWorker(root, altDir, agentID)
		default:
			return primeLiaison(root, altDir)
		}
	},
}

// detectRole determines the agent role and ID using the detection order:
// 1. --role/--agent-id flags
// 2. ALT_AGENT_ID env var
// 3. tmux session name
// 4. working directory (inside worktrees/{id} = worker)
// 5. default to liaison
func detectRole(root, altDir string) (string, string) {
	// Explicit flags take priority.
	if primeRole != "" {
		return primeRole, primeAgentID
	}

	// Check ALT_AGENT_ID env var.
	if envID := os.Getenv("ALT_AGENT_ID"); envID != "" {
		if strings.HasPrefix(envID, "worker-") {
			return "worker", envID
		}
		return "liaison", envID
	}

	// Explicit --agent-id flag.
	if primeAgentID != "" {
		if strings.HasPrefix(primeAgentID, "worker-") {
			return "worker", primeAgentID
		}
		return "liaison", primeAgentID
	}

	// Try tmux session name.
	tmuxCmd := exec.Command("tmux", "display-message", "-p", "#{session_name}")
	if out, err := tmuxCmd.Output(); err == nil {
		sessionName := strings.TrimSpace(string(out))
		if strings.HasPrefix(sessionName, "alt-worker-") {
			id := strings.TrimPrefix(sessionName, "alt-")
			return "worker", id
		}
		if sessionName == "alt-liaison" {
			return "liaison", "liaison-01"
		}
	}

	// Check working directory - if inside worktrees/{id}, it's a worker.
	if cwd, err := os.Getwd(); err == nil {
		worktreeDir := filepath.Join(root, "worktrees")
		if rel, err := filepath.Rel(worktreeDir, cwd); err == nil && !strings.HasPrefix(rel, "..") {
			parts := strings.SplitN(rel, string(filepath.Separator), 2)
			if len(parts) > 0 && parts[0] != "" {
				return "worker", parts[0]
			}
		}
	}

	return "liaison", "liaison-01"
}

// primeLiaison outputs a slim role header + runtime state.
func primeLiaison(root, altDir string) error {
	fmt.Println("# Liaison Agent")
	fmt.Println()
	fmt.Println("You are the liaison agent in the Altera multi-agent orchestration system.")
	fmt.Println("You translate between human intent and the task/agent system.")
	fmt.Println()
	fmt.Println("Use `alt help liaison startup` for instructions. Use `alt <command> --help` for syntax.")
	fmt.Println()

	// Output runtime state using the liaison manager's Prime() logic.
	agentStore, err := agent.NewStore(filepath.Join(altDir, "agents"))
	if err != nil {
		return fmt.Errorf("opening agent store: %w", err)
	}
	taskStore, err := task.NewStore(root)
	if err != nil {
		return fmt.Errorf("opening task store: %w", err)
	}
	msgStore, err := message.NewStore(filepath.Join(altDir, "messages"))
	if err != nil {
		return fmt.Errorf("opening message store: %w", err)
	}
	evReader := events.NewReader(filepath.Join(altDir, "events.jsonl"))

	m := liaison.NewManager(root, agentStore, taskStore, msgStore, evReader)
	summary, err := m.Prime()
	if err != nil {
		return fmt.Errorf("generating state summary: %w", err)
	}
	fmt.Print(summary)

	return nil
}

// primeWorker outputs the worker system prompt + task.json contents + checkpoint state.
func primeWorker(root, altDir, agentID string) error {
	agentStore, err := agent.NewStore(filepath.Join(altDir, "agents"))
	if err != nil {
		return fmt.Errorf("opening agent store: %w", err)
	}

	a, err := agentStore.Get(agentID)
	if err != nil {
		// If agent not found, output a generic worker prompt.
		fmt.Printf("# Worker Agent: %s\n\nAgent record not found. Operating in standalone mode.\n", agentID)
		return nil
	}

	// Load task details if available.
	var t *task.Task
	if a.CurrentTask != "" {
		taskStore, err := task.NewStore(root)
		if err == nil {
			t, _ = taskStore.Get(a.CurrentTask)
		}
	}

	// Output slim role header.
	fmt.Printf("# Worker Agent: %s\n\n", agentID)
	if t != nil {
		fmt.Printf("- **Task**: %s `%s`\n", t.Title, t.ID)
	} else {
		fmt.Println("No task currently assigned.")
	}
	if a.Rig != "" {
		fmt.Printf("- **Rig**: %s\n", a.Rig)
	}
	fmt.Println()
	fmt.Println("Use `alt help worker startup` for instructions. Use `alt <command> --help` for syntax.")
	fmt.Println()

	// Output task.json contents if available in the worktree.
	if a.Worktree != "" {
		taskJSONPath := filepath.Join(a.Worktree, "task.json")
		if data, err := os.ReadFile(taskJSONPath); err == nil {
			fmt.Println("## task.json")
			fmt.Println("```json")
			fmt.Print(string(data))
			fmt.Println("```")
			fmt.Println()
		}
	}

	// Output checkpoint state if available.
	if t != nil && t.Checkpoint != "" {
		fmt.Println("## Checkpoint (resuming)")
		fmt.Println()
		fmt.Println(t.Checkpoint)
		fmt.Println()
	}

	// Output rig config if available.
	if a.Rig != "" {
		rc, err := config.LoadRig(altDir, a.Rig)
		if err == nil {
			data, _ := json.MarshalIndent(rc, "", "  ")
			fmt.Println("## Rig Configuration")
			fmt.Println("```json")
			fmt.Println(string(data))
			fmt.Println("```")
			fmt.Println()
		}
	}

	return nil
}
