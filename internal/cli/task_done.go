package cli

import (
	"fmt"
	"path/filepath"

	"github.com/anthropics/altera/internal/message"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(taskDoneCmd)
	taskDoneCmd.Flags().StringVar(&taskDoneResult, "result", "", "summary of what was accomplished")
}

var taskDoneResult string

var taskDoneCmd = &cobra.Command{
	Use:   "task-done <task-id> <agent-id>",
	Short: "Signal that a task is complete",
	Long:  `Sends a task_done message to the daemon so it can mark the task done and queue it for merge.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}

		store, err := message.NewStore(filepath.Join(altDir, "messages"))
		if err != nil {
			return fmt.Errorf("opening message store: %w", err)
		}

		taskID := args[0]
		agentID := args[1]

		var payload map[string]any
		if taskDoneResult != "" {
			payload = map[string]any{"result": taskDoneResult}
		}

		if _, err := store.Create(message.TypeTaskDone, agentID, "daemon", taskID, payload); err != nil {
			return fmt.Errorf("creating task_done message: %w", err)
		}

		fmt.Printf("Task %s marked done by agent %s\n", taskID, agentID)
		return nil
	},
}
