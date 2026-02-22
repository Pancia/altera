package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/anthropics/altera/internal/daemon"
	"github.com/spf13/cobra"
)

var daemonStatusVerbose bool

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
	daemonCmd.AddCommand(daemonTickCmd)
	daemonStatusCmd.Flags().BoolVar(&daemonStatusVerbose, "verbose", false, "show detailed daemon state")
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage the altera daemon",
	Long:  `Start, stop, or check the status of the altera daemon process.`,
}

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := projectRoot()
		if err != nil {
			return err
		}
		d, err := daemon.New(root)
		if err != nil {
			return err
		}
		fmt.Println("Daemon starting...")
		if err := d.Run(); err != nil {
			if strings.Contains(err.Error(), "flock") {
				return fmt.Errorf("daemon is already running")
			}
			return err
		}
		return nil
	},
}

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return err
		}
		if err := daemon.SendStop(altDir); err != nil {
			return err
		}
		fmt.Println("Daemon stop signal sent.")
		return nil
	},
}

var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show daemon status",
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return err
		}
		st := daemon.ReadStatus(altDir)
		if st.Running {
			fmt.Printf("Daemon is running (PID %d)\n", st.PID)
		} else {
			fmt.Println("Daemon is not running.")
		}

		if daemonStatusVerbose {
			state, err := daemon.ReadState(altDir)
			if err != nil {
				fmt.Println("\nNo daemon state available.")
				return nil
			}
			fmt.Printf("\nLast tick:       %s (%s ago)\n", state.LastTick.Format(time.RFC3339), time.Since(state.LastTick).Round(time.Second))
			fmt.Printf("Tick number:     %d\n", state.TickNum)
			fmt.Printf("Active workers:  %d\n", state.ActiveWorkers)
			fmt.Printf("Dead workers:    %d\n", state.DeadWorkers)
			if state.LastSpawnTask != "" {
				fmt.Printf("Last spawn task: %s\n", state.LastSpawnTask)
				if state.LastSpawnError != "" {
					fmt.Printf("Last spawn error: %s\n", state.LastSpawnError)
				}
			}
			if len(state.RecentErrors) > 0 {
				fmt.Println("\nRecent errors:")
				for _, e := range state.RecentErrors {
					fmt.Printf("  %s\n", e)
				}
			}
		}

		return nil
	},
}

var daemonTickCmd = &cobra.Command{
	Use:   "tick",
	Short: "Force an immediate daemon tick",
	Long:  `Sends SIGUSR1 to the daemon to trigger an immediate tick cycle.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return err
		}
		if err := daemon.SendTickNow(altDir); err != nil {
			return err
		}
		fmt.Println("Tick signal sent.")
		return nil
	},
}
