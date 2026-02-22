package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/daemon"
	"github.com/anthropics/altera/internal/liaison"
	"github.com/anthropics/altera/internal/tmux"
	"github.com/spf13/cobra"
)

const daemonSessionName = "alt-daemon"

var startDebug bool

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolVar(&startDebug, "debug", false, "enable terminal logging (pipe-pane) for all sessions")
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

		// Persist debug flag so spawned agents inherit it.
		if startDebug {
			if err := config.SetDebug(altDir, true); err != nil {
				return fmt.Errorf("setting debug flag: %w", err)
			}
			// Ensure logs directory exists.
			_ = os.MkdirAll(filepath.Join(altDir, "logs"), 0o755)
			fmt.Println("Debug mode enabled (terminal logging active).")
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
			// Start terminal logging for daemon if debug mode is enabled.
			if startDebug {
				logPath := filepath.Join(altDir, "logs", "daemon.terminal.log")
				if err := tmux.StartLogging(daemonSessionName, logPath); err != nil {
					fmt.Printf("Warning: could not start daemon terminal logging: %v\n", err)
				}
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
