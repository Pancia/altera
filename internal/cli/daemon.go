package cli

import (
	"fmt"
	"strings"

	"github.com/anthropics/altera/internal/daemon"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
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
		return nil
	},
}
