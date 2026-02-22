package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/anthropics/altera/internal/agent"
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

		// Kill liaison tmux session and mark agent as dead.
		if tmux.SessionExists(liaison.SessionName) {
			if err := tmux.KillSession(liaison.SessionName); err != nil {
				fmt.Printf("Warning: failed to kill liaison session: %v\n", err)
			} else {
				fmt.Println("Liaison session killed.")
			}
		}
		// Mark liaison agent as dead in the store.
		agents, err := agent.NewStore(filepath.Join(altDir, "agents"))
		if err == nil {
			if a, err := agents.Get(liaison.AgentID); err == nil {
				a.Status = agent.StatusDead
				_ = agents.Update(a)
			}
		}

		// Kill worker and resolver agent sessions.
		if agents != nil {
			for _, status := range []agent.Status{agent.StatusActive, agent.StatusIdle} {
				active, err := agents.ListByStatus(status)
				if err != nil {
					continue
				}
				for _, a := range active {
					if a.Role == agent.RoleLiaison {
						continue
					}
					if a.PID > 0 {
						if proc, err := os.FindProcess(a.PID); err == nil {
							_ = proc.Signal(syscall.SIGTERM)
						}
					}
					if a.TmuxSession != "" && tmux.SessionExists(a.TmuxSession) {
						if err := tmux.KillSession(a.TmuxSession); err != nil {
							fmt.Printf("Warning: failed to kill session %s: %v\n", a.TmuxSession, err)
						} else {
							fmt.Printf("Killed %s session: %s\n", a.Role, a.TmuxSession)
						}
					}
					a.Status = agent.StatusDead
					_ = agents.Update(a)
				}
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
