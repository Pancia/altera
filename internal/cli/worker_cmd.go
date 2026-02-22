package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/anthropics/altera/internal/agent"
	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/events"
	"github.com/anthropics/altera/internal/git"
	"github.com/anthropics/altera/internal/session"
	"github.com/anthropics/altera/internal/tmux"
	"github.com/anthropics/altera/internal/worker"
	"github.com/spf13/cobra"
)

var (
	workerPeekLines   int
	workerPeekAll     bool
	workerPeekSession bool
)

func init() {
	rootCmd.AddCommand(workerCmd)
	workerCmd.AddCommand(workerListCmd)
	workerCmd.AddCommand(workerAttachCmd)
	workerCmd.AddCommand(workerPeekCmd)
	workerCmd.AddCommand(workerKillCmd)
	workerCmd.AddCommand(workerInspectCmd)
	workerPeekCmd.Flags().IntVar(&workerPeekLines, "lines", 200, "number of lines to capture")
	workerPeekCmd.Flags().BoolVar(&workerPeekAll, "all", false, "show full scrollback history")
	workerPeekCmd.Flags().BoolVar(&workerPeekSession, "session", false, "show JSONL transcript instead of terminal output")
}

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Manage worker agents",
	Long:  `List, attach, peek, kill, or inspect worker agents.`,
}

var workerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workers",
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}

		agentStore, err := agent.NewStore(filepath.Join(altDir, "agents"))
		if err != nil {
			return fmt.Errorf("opening agent store: %w", err)
		}

		workers, err := agentStore.ListByRole(agent.RoleWorker)
		if err != nil {
			return fmt.Errorf("listing workers: %w", err)
		}

		if len(workers) == 0 {
			fmt.Println("No workers.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tSTATUS\tTASK\tHEARTBEAT\tWORKTREE\tTMUX")
		for _, a := range workers {
			hbAge := time.Since(a.Heartbeat).Round(time.Second)
			wtBase := filepath.Base(a.Worktree)
			if a.Worktree == "" {
				wtBase = "-"
			}
			tmuxSession := a.TmuxSession
			if tmuxSession == "" {
				tmuxSession = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s ago\t%s\t%s\n",
				a.ID, a.Status, a.CurrentTask, hbAge, wtBase, tmuxSession)
		}
		w.Flush()
		return nil
	},
}

var workerAttachCmd = &cobra.Command{
	Use:   "attach <id>",
	Short: "Attach to a worker's tmux session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}

		agentStore, err := agent.NewStore(filepath.Join(altDir, "agents"))
		if err != nil {
			return fmt.Errorf("opening agent store: %w", err)
		}

		a, err := agentStore.Get(args[0])
		if err != nil {
			return fmt.Errorf("agent %q: %w", args[0], err)
		}
		if a.TmuxSession == "" {
			return fmt.Errorf("agent %q has no tmux session", args[0])
		}
		return tmux.AttachSession(a.TmuxSession)
	},
}

var workerPeekCmd = &cobra.Command{
	Use:   "peek <id>",
	Short: "Capture recent output from a worker",
	Long: `Capture terminal output from a worker agent.

For live sessions, captures from the tmux pane scrollback.
For dead sessions, falls back to reading the terminal log file from .alt/logs/.

Flags:
  --lines N    Number of lines to capture (default 200)
  --all        Show full scrollback history
  --session    Show the JSONL transcript instead of terminal output`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}

		agentStore, err := agent.NewStore(filepath.Join(altDir, "agents"))
		if err != nil {
			return fmt.Errorf("opening agent store: %w", err)
		}

		a, err := agentStore.Get(args[0])
		if err != nil {
			return fmt.Errorf("agent %q: %w", args[0], err)
		}

		// --session: show JSONL transcript instead of terminal output.
		if workerPeekSession {
			return peekSession(altDir, a)
		}

		// Check if tmux session is alive.
		sessionAlive := a.TmuxSession != "" && tmux.SessionExists(a.TmuxSession)

		if sessionAlive {
			lines := workerPeekLines
			if workerPeekAll {
				lines = -1 // full history
			}
			output, err := tmux.CapturePane(a.TmuxSession, lines)
			if err != nil {
				return fmt.Errorf("capturing pane: %w", err)
			}
			fmt.Print(output)
			return nil
		}

		// Session is dead — fall back to terminal log file.
		logPath := filepath.Join(config.LogsDir(altDir), a.ID+".terminal.log")
		data, err := os.ReadFile(logPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("no live session and no terminal log found for %s", a.ID)
			}
			return fmt.Errorf("reading terminal log: %w", err)
		}

		output := string(data)
		if !workerPeekAll && workerPeekLines > 0 {
			output = lastNLines(output, workerPeekLines)
		}
		fmt.Print(output)
		return nil
	},
}

// peekSession finds and renders the JSONL transcript for an agent.
func peekSession(altDir string, a *agent.Agent) error {
	// Try the agent's session directory first.
	if a.SessionDir != "" {
		transcripts, err := session.FindTranscripts(a.SessionDir)
		if err == nil && len(transcripts) > 0 {
			return renderTranscript(transcripts[0])
		}
	}

	// Fall back to .alt/logs/{id}.jsonl (copied on cleanup).
	logPath := filepath.Join(config.LogsDir(altDir), a.ID+".jsonl")
	if _, err := os.Stat(logPath); err == nil {
		return renderTranscript(logPath)
	}

	// Try computing session dir from worktree path.
	if a.Worktree != "" {
		dir := session.TranscriptDir(a.Worktree)
		transcripts, err := session.FindTranscripts(dir)
		if err == nil && len(transcripts) > 0 {
			return renderTranscript(transcripts[0])
		}
	}

	return fmt.Errorf("no JSONL transcript found for %s", a.ID)
}

// renderTranscript renders a JSONL file to stdout.
func renderTranscript(path string) error {
	output, err := session.RenderTranscript(path)
	if err != nil {
		return fmt.Errorf("rendering transcript: %w", err)
	}
	fmt.Print(output)
	return nil
}

// lastNLines returns the last n lines of a string.
func lastNLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[len(lines)-n:], "\n")
}

var workerKillCmd = &cobra.Command{
	Use:   "kill <id>",
	Short: "Kill a worker (tmux session, worktree, mark dead)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}
		root := filepath.Dir(altDir)

		agentStore, err := agent.NewStore(filepath.Join(altDir, "agents"))
		if err != nil {
			return fmt.Errorf("opening agent store: %w", err)
		}

		a, err := agentStore.Get(args[0])
		if err != nil {
			return fmt.Errorf("agent %q: %w", args[0], err)
		}

		evWriter := events.NewWriter(filepath.Join(altDir, "events.jsonl"))
		wm := worker.NewManager(root, agentStore, evWriter)
		if err := wm.CleanupWorker(a); err != nil {
			return fmt.Errorf("killing worker: %w", err)
		}

		fmt.Printf("Worker %s killed.\n", args[0])
		return nil
	},
}

var workerInspectCmd = &cobra.Command{
	Use:   "inspect <id>",
	Short: "Show detailed info about a worker",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}

		agentStore, err := agent.NewStore(filepath.Join(altDir, "agents"))
		if err != nil {
			return fmt.Errorf("opening agent store: %w", err)
		}

		a, err := agentStore.Get(args[0])
		if err != nil {
			return fmt.Errorf("agent %q: %w", args[0], err)
		}

		// Agent JSON
		data, _ := json.MarshalIndent(a, "", "  ")
		fmt.Println("AGENT")
		fmt.Println(string(data))
		fmt.Println()

		// Tmux status
		tmuxStatus := "no session"
		if a.TmuxSession != "" {
			if tmux.SessionExists(a.TmuxSession) {
				tmuxStatus = "alive"
			} else {
				tmuxStatus = "dead"
			}
		}
		fmt.Printf("TMUX: %s (%s)\n", a.TmuxSession, tmuxStatus)
		fmt.Println()

		// Git info from worktree
		if a.Worktree != "" {
			fmt.Println("WORKTREE")
			fmt.Printf("  Path: %s\n", a.Worktree)

			branch, err := git.CurrentBranch(a.Worktree)
			if err == nil {
				fmt.Printf("  Branch: %s\n", branch)
			}

			clean, err := git.IsClean(a.Worktree)
			if err == nil {
				if clean {
					fmt.Println("  Status: clean")
				} else {
					fmt.Println("  Status: dirty")
				}
			}

			log, err := git.Log(a.Worktree, 5)
			if err == nil && log != "" {
				fmt.Println("  Recent commits:")
				for _, line := range strings.Split(log, "\n") {
					if line != "" {
						fmt.Printf("    %s\n", line)
					}
				}
			}
		}

		return nil
	},
}
