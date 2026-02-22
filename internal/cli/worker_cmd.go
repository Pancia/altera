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
	"github.com/anthropics/altera/internal/events"
	"github.com/anthropics/altera/internal/git"
	"github.com/anthropics/altera/internal/tmux"
	"github.com/anthropics/altera/internal/worker"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(workerCmd)
	workerCmd.AddCommand(workerListCmd)
	workerCmd.AddCommand(workerAttachCmd)
	workerCmd.AddCommand(workerPeekCmd)
	workerCmd.AddCommand(workerKillCmd)
	workerCmd.AddCommand(workerInspectCmd)
	workerPeekCmd.Flags().IntVar(&workerPeekLines, "lines", 50, "number of lines to capture")
}

var workerPeekLines int

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
	Use:               "attach <id>",
	Short:             "Attach to a worker's tmux session",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeWorkerIDs,
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
	Use:               "peek <id>",
	Short:             "Capture recent output from a worker's tmux pane",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeWorkerIDs,
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

		output, err := tmux.CapturePane(a.TmuxSession, workerPeekLines)
		if err != nil {
			return fmt.Errorf("capturing pane: %w", err)
		}
		fmt.Print(output)
		return nil
	},
}

var workerKillCmd = &cobra.Command{
	Use:               "kill <id>",
	Short:             "Kill a worker (tmux session, worktree, mark dead)",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeWorkerIDs,
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
	Use:               "inspect <id>",
	Short:             "Show detailed info about a worker",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeWorkerIDs,
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
