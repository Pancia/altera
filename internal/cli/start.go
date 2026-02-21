package cli

import (
	"fmt"
	"os"
	"os/exec"

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
	Short: "Start daemon and liaison, then attach",
	Long:  `Starts the daemon in a tmux session, starts the liaison, then attaches to the liaison session.`,
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
			fmt.Println("Daemon starting in tmux session: alt-daemon")
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
			fmt.Println("Liaison started in tmux session: alt-liaison")
		}

		// Attach to liaison session with full TTY access.
		fmt.Println("Attaching to liaison...")
		attach := exec.Command("tmux", "attach-session", "-t", liaison.SessionName)
		attach.Stdin = os.Stdin
		attach.Stdout = os.Stdout
		attach.Stderr = os.Stderr
		return attach.Run()
	},
}
