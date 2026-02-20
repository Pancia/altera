package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/anthropics/altera/internal/agent"
	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/task"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show project status overview",
	Long:  `Displays a formatted table of tasks, agents, and rigs.`,
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

	return nil
}
