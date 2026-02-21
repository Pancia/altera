package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/anthropics/altera/internal/agent"
	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/daemon"
	"github.com/anthropics/altera/internal/events"
	"github.com/anthropics/altera/internal/git"
	"github.com/anthropics/altera/internal/task"
	"github.com/anthropics/altera/internal/tmux"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show project status overview",
	Long:  `Displays a formatted table of tasks, agents, rigs, worktrees, branches, tmux sessions, merge queue, daemon status, and recent events.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}

		root := filepath.Dir(altDir)
		return runStatus(root, altDir)
	},
}

func runStatus(root, altDir string) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)

	// Tasks section
	taskStore, err := task.NewStore(root)
	if err != nil {
		return fmt.Errorf("opening task store: %w", err)
	}
	tasks, err := taskStore.List(task.Filter{})
	if err != nil {
		return fmt.Errorf("listing tasks: %w", err)
	}

	fmt.Fprintln(w, "TASKS")
	fmt.Fprintln(w, "ID\tSTATUS\tASSIGNED\tRIG\tTITLE")
	for _, t := range tasks {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			t.ID, t.Status, t.AssignedTo, t.Rig, t.Title)
	}
	if len(tasks) == 0 {
		fmt.Fprintln(w, "(none)")
	}
	w.Flush()

	fmt.Println()

	// Agents section
	w = tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	agentStore, err := agent.NewStore(filepath.Join(altDir, "agents"))
	if err != nil {
		return fmt.Errorf("opening agent store: %w", err)
	}
	agents, err := agentStore.ListByStatus(agent.StatusActive)
	if err != nil {
		return fmt.Errorf("listing agents: %w", err)
	}

	fmt.Fprintln(w, "AGENTS")
	fmt.Fprintln(w, "ID\tROLE\tSTATUS\tRIG\tTASK")
	for _, a := range agents {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			a.ID, a.Role, a.Status, a.Rig, a.CurrentTask)
	}
	if len(agents) == 0 {
		fmt.Fprintln(w, "(none)")
	}
	w.Flush()

	fmt.Println()

	// Rigs section
	rigs, err := config.ListRigs(altDir)
	if err != nil {
		return fmt.Errorf("listing rigs: %w", err)
	}

	fmt.Println("RIGS")
	if len(rigs) == 0 {
		fmt.Println("(none)")
	} else {
		for _, name := range rigs {
			fmt.Printf("  %s\n", name)
		}
	}

	fmt.Println()

	// Worktrees section
	w = tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "WORKTREES")
	fmt.Fprintln(w, "ID\tBRANCH")
	worktreeDir := filepath.Join(altDir, "worktrees")
	if entries, err := os.ReadDir(worktreeDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			wtPath := filepath.Join(worktreeDir, e.Name())
			branch, brErr := git.CurrentBranch(wtPath)
			if brErr != nil {
				branch = "(unknown)"
			}
			fmt.Fprintf(w, "%s\t%s\n", e.Name(), branch)
		}
	}
	// Also check the project-level worktrees directory
	projectWorktreeDir := filepath.Join(root, "worktrees")
	if projectWorktreeDir != worktreeDir {
		if entries, err := os.ReadDir(projectWorktreeDir); err == nil {
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				wtPath := filepath.Join(projectWorktreeDir, e.Name())
				branch, brErr := git.CurrentBranch(wtPath)
				if brErr != nil {
					branch = "(unknown)"
				}
				fmt.Fprintf(w, "%s\t%s\n", e.Name(), branch)
			}
		}
	}
	w.Flush()

	fmt.Println()

	// Branches section
	fmt.Println("BRANCHES")
	altBranches, _ := git.ListBranches(root, "alt/")
	workerBranches, _ := git.ListBranches(root, "worker/")
	allBranches := append(altBranches, workerBranches...)
	if len(allBranches) == 0 {
		fmt.Println("(none)")
	} else {
		for _, b := range allBranches {
			fmt.Printf("  %s\n", b)
		}
	}

	fmt.Println()

	// Tmux Sessions section
	w = tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "TMUX SESSIONS")
	fmt.Fprintln(w, "NAME\tSTATUS")
	sessions, err := tmux.ListSessions()
	if err != nil {
		fmt.Fprintf(w, "(error: %v)\n", err)
	} else if len(sessions) == 0 {
		fmt.Fprintln(w, "(none)")
	} else {
		for _, s := range sessions {
			status := "dead"
			if tmux.SessionExists(s) {
				status = "alive"
			}
			fmt.Fprintf(w, "%s\t%s\n", s, status)
		}
	}
	w.Flush()

	fmt.Println()

	// Merge Queue section
	mergeQueueDir := filepath.Join(altDir, "merge-queue")
	count := 0
	if entries, err := os.ReadDir(mergeQueueDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
				count++
			}
		}
	}
	fmt.Printf("MERGE QUEUE: %d items\n", count)

	fmt.Println()

	// Daemon section
	st := daemon.ReadStatus(altDir)
	fmt.Print("DAEMON: ")
	if st.Running {
		fmt.Printf("running (PID %d)\n", st.PID)
	} else {
		fmt.Println("stopped")
	}

	fmt.Println()

	// Recent Events section
	w = tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "RECENT EVENTS")
	fmt.Fprintln(w, "TIME\tTYPE\tAGENT\tTASK")
	evtPath := filepath.Join(altDir, "events.jsonl")
	reader := events.NewReader(evtPath)
	evts, err := reader.Tail(5)
	if err == nil && len(evts) > 0 {
		for _, ev := range evts {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				ev.Timestamp.Format(time.RFC3339), ev.Type, ev.AgentID, ev.TaskID)
		}
	} else {
		fmt.Fprintln(w, "(none)")
	}
	w.Flush()

	return nil
}
