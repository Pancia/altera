package cli

import (
	"fmt"

	"github.com/anthropics/altera/internal/daemon"
	"github.com/anthropics/altera/internal/liaison"
	"github.com/anthropics/altera/internal/tmux"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stopCmd)
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop daemon and liaison",
	Long:  `Sends stop signal to the daemon and kills the liaison and daemon tmux sessions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return err
		}

		// Stop daemon process.
		st := daemon.ReadStatus(altDir)
		if st.Running {
			if err := daemon.SendStop(altDir); err != nil {
				fmt.Printf("Warning: failed to stop daemon: %v\n", err)
			} else {
				fmt.Println("Daemon stop signal sent.")
			}
		} else {
			fmt.Println("Daemon is not running.")
		}

		// Kill liaison tmux session.
		if tmux.SessionExists(liaison.SessionName) {
			if err := tmux.KillSession(liaison.SessionName); err != nil {
				fmt.Printf("Warning: failed to kill liaison session: %v\n", err)
			} else {
				fmt.Println("Liaison session killed.")
			}
		}

		// Kill daemon tmux session.
		if tmux.SessionExists(daemonSessionName) {
			if err := tmux.KillSession(daemonSessionName); err != nil {
				fmt.Printf("Warning: failed to kill daemon session: %v\n", err)
			} else {
				fmt.Println("Daemon session killed.")
			}
		}

		fmt.Println("Altera stopped.")
		return nil
	},
}
