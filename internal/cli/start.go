package cli

import (
	"fmt"

	"github.com/anthropics/altera/internal/daemon"
	"github.com/anthropics/altera/internal/liaison"
	"github.com/anthropics/altera/internal/tmux"
	"github.com/spf13/cobra"
)

const daemonSessionName = "alt-daemon"

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start daemon and liaison",
	Long:  `Starts the daemon in a tmux session and starts the liaison.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return err
		}

		root, err := projectRoot()
		if err != nil {
			return err
		}

		// Start daemon if not already running.
		st := daemon.ReadStatus(altDir)
		if st.Running {
			fmt.Println("Daemon already running.")
		} else {
			// Kill stale daemon session if it exists.
			if tmux.SessionExists(daemonSessionName) {
				_ = tmux.KillSession(daemonSessionName)
			}
			if err := tmux.CreateSession(daemonSessionName); err != nil {
				return fmt.Errorf("creating daemon tmux session: %w", err)
			}
			daemonCmd := fmt.Sprintf("cd %s && alt daemon start", root)
			if err := tmux.SendKeys(daemonSessionName, daemonCmd); err != nil {
				return fmt.Errorf("starting daemon: %w", err)
			}
			fmt.Println("Daemon started.")
		}

		// Start liaison if not already running.
		if tmux.SessionExists(liaison.SessionName) {
			fmt.Println("Liaison already running.")
		} else {
			m, err := newLiaisonManager()
			if err != nil {
				return err
			}
			if err := m.StartLiaison(); err != nil {
				return fmt.Errorf("starting liaison: %w", err)
			}
			fmt.Println("Liaison started.")
		}

		fmt.Println("\nAltera running. Attach to liaison with:")
		fmt.Println("  alt liaison attach")
		return nil
	},
}
