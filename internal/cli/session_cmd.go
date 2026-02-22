package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/anthropics/altera/internal/tmux"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sessionCmd)
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionSwitchCmd)
}

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage tmux sessions",
	Long:  `List or switch between Altera tmux sessions.`,
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Altera tmux sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := tmux.ListSessions()
		if err != nil {
			return fmt.Errorf("listing sessions: %w", err)
		}

		if len(sessions) == 0 {
			fmt.Println("No Altera tmux sessions.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSTATUS")
		for _, s := range sessions {
			status := "dead"
			if tmux.SessionExists(s) {
				status = "alive"
			}
			fmt.Fprintf(w, "%s\t%s\n", s, status)
		}
		w.Flush()
		return nil
	},
}

var sessionSwitchCmd = &cobra.Command{
	Use:               "switch <name>",
	Short:             "Attach to a tmux session (auto-prepends alt- prefix if needed)",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeSessionNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		// Auto-prepend alt- prefix if not present.
		if !strings.HasPrefix(name, tmux.SessionPrefix) {
			name = tmux.SessionPrefix + name
		}
		if !tmux.SessionExists(name) {
			return fmt.Errorf("session %q does not exist", name)
		}
		return tmux.AttachSession(name)
	},
}
