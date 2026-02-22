package cli

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/daemon"
	"github.com/spf13/cobra"
)

var (
	daemonStatusVerbose bool
	daemonLogsFollow    bool
	daemonLogsLines     int
)

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
	daemonCmd.AddCommand(daemonTickCmd)
	daemonCmd.AddCommand(daemonLogsCmd)
	daemonStatusCmd.Flags().BoolVar(&daemonStatusVerbose, "verbose", false, "show detailed daemon state")
	daemonLogsCmd.Flags().BoolVarP(&daemonLogsFollow, "follow", "f", false, "follow log output")
	daemonLogsCmd.Flags().IntVarP(&daemonLogsLines, "lines", "n", 50, "number of lines to show")
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

var daemonLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show daemon log output",
	Long: `Show the daemon log file (.alt/logs/daemon.log).

Examples:
  alt daemon logs           Show last 50 lines
  alt daemon logs -n 100    Show last 100 lines
  alt daemon logs -f        Follow log output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return err
		}

		logPath := filepath.Join(config.LogsDir(altDir), "daemon.log")

		if daemonLogsFollow {
			return tailDaemonLog(logPath)
		}

		data, err := os.ReadFile(logPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No daemon log file found. Has the daemon been started?")
				return nil
			}
			return fmt.Errorf("reading daemon log: %w", err)
		}

		output := string(data)
		if daemonLogsLines > 0 {
			output = lastNLines(output, daemonLogsLines)
		}
		if output != "" {
			fmt.Print(output)
			// Ensure trailing newline.
			if output[len(output)-1] != '\n' {
				fmt.Println()
			}
		}
		return nil
	},
}

// tailDaemonLog follows the daemon log file, printing new content as it appears.
// It handles file truncation (daemon restart) by resetting the read offset.
func tailDaemonLog(logPath string) error {
	// Show initial context.
	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No daemon log file found. Has the daemon been started?")
			fmt.Println("Waiting for log file...")
		}
	} else {
		initial := lastNLines(string(data), 50)
		if initial != "" {
			fmt.Print(initial)
			if initial[len(initial)-1] != '\n' {
				fmt.Println()
			}
		}
	}

	offset := int64(len(data))

	// Set up signal handling for clean exit.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigCh:
			return nil
		case <-ticker.C:
			f, err := os.Open(logPath)
			if err != nil {
				continue // file may not exist yet
			}

			info, err := f.Stat()
			if err != nil {
				f.Close()
				continue
			}

			// Detect file truncation (daemon restart).
			if info.Size() < offset {
				offset = 0
			}

			if info.Size() > offset {
				if _, err := f.Seek(offset, io.SeekStart); err != nil {
					f.Close()
					continue
				}
				newData, err := io.ReadAll(f)
				if err != nil {
					f.Close()
					continue
				}
				if len(newData) > 0 {
					fmt.Print(string(newData))
					offset += int64(len(newData))
				}
			}
			f.Close()
		}
	}
}
