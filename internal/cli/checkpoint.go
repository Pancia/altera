package cli

import (
	"fmt"

	"github.com/anthropics/altera/internal/task"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(checkpointCmd)
	checkpointCmd.Flags().StringVar(&checkpointMsg, "message", "", "checkpoint message")
}

var checkpointMsg string

var checkpointCmd = &cobra.Command{
	Use:   "checkpoint <task-id>",
	Short: "Save a checkpoint for a task",
	Long:  `Records a checkpoint message on a task, capturing progress state.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := projectRoot()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}

		store, err := task.NewStore(root)
		if err != nil {
			return fmt.Errorf("opening task store: %w", err)
		}

		taskID := args[0]
		if err := store.Update(taskID, func(t *task.Task) error {
			t.Checkpoint = checkpointMsg
			return nil
		}); err != nil {
			// Best-effort: the task may already be done and merged by the
			// daemon before the stop hook fires. Silently succeed so hooks
			// don't produce noisy errors.
			fmt.Printf("Checkpoint skipped for task %s (already finished)\n", taskID)
			return nil
		}

		fmt.Printf("Checkpoint saved for task %s\n", taskID)
		return nil
	},
}
