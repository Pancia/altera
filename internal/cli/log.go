package cli

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/anthropics/altera/internal/events"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(logCmd)
	logCmd.Flags().IntVar(&logLast, "last", 0, "show only the last N events")
	logCmd.Flags().BoolVar(&logTail, "tail", false, "follow the event log (poll every 2s)")
}

var (
	logLast int
	logTail bool
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show event log",
	Long:  `Display events from the event log. Use --last N to show only the most recent N events. Use --tail to follow new events.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}

		evtPath := filepath.Join(altDir, "events.jsonl")

		if logTail {
			return tailLog(evtPath)
		}

		reader := events.NewReader(evtPath)

		var evts []events.Event
		if logLast > 0 {
			evts, err = reader.Tail(logLast)
		} else {
			evts, err = reader.ReadAll()
		}
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fmt.Println("No events recorded yet.")
				return nil
			}
			return fmt.Errorf("reading events: %w", err)
		}

		if len(evts) == 0 {
			fmt.Println("No events recorded yet.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "TIME\tTYPE\tAGENT\tTASK")
		for _, ev := range evts {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				ev.Timestamp.Format(time.RFC3339), ev.Type, ev.AgentID, ev.TaskID)
		}
		_ = w.Flush()

		return nil
	},
}

// tailLog shows the last 10 events then polls for new events every 2 seconds.
func tailLog(evtPath string) error {
	reader := events.NewReader(evtPath)

	// Show last 10 events as initial context.
	evts, err := reader.Tail(10)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reading events: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "TIME\tTYPE\tAGENT\tTASK")
	for _, ev := range evts {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			ev.Timestamp.Format(time.RFC3339), ev.Type, ev.AgentID, ev.TaskID)
	}
	_ = w.Flush()

	// Track the count of events we've already seen.
	seen := 0
	allEvts, _ := reader.ReadAll()
	seen = len(allEvts)

	// Set up Ctrl-C handler.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigCh:
			fmt.Println("\nStopped.")
			return nil
		case <-ticker.C:
			allEvts, err := reader.ReadAll()
			if err != nil {
				continue
			}
			if len(allEvts) > seen {
				newEvts := allEvts[seen:]
				w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
				for _, ev := range newEvts {
					_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
						ev.Timestamp.Format(time.RFC3339), ev.Type, ev.AgentID, ev.TaskID)
				}
				_ = w.Flush()
				seen = len(allEvts)
			}
		}
	}
}

// logTaskCreated appends a task_created event to the event log.
// It is a best-effort helper used by task create.
func logTaskCreated(evtPath, taskID string) {
	writer := events.NewWriter(evtPath)
	_ = writer.Append(events.Event{
		Timestamp: time.Now(),
		Type:      events.TaskCreated,
		TaskID:    taskID,
	})
}
