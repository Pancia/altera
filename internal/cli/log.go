package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/anthropics/altera/internal/events"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(logCmd)
	logCmd.Flags().IntVar(&logLast, "last", 0, "show only the last N events")
}

var logLast int

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show event log",
	Long:  `Display events from the event log. Use --last N to show only the most recent N events.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}

		evtPath := filepath.Join(altDir, "events.jsonl")
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
		fmt.Fprintln(w, "TIME\tTYPE\tAGENT\tTASK")
		for _, ev := range evts {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				ev.Timestamp.Format(time.RFC3339), ev.Type, ev.AgentID, ev.TaskID)
		}
		w.Flush()

		return nil
	},
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
