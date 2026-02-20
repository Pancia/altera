package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(workCmd)
}

var workCmd = &cobra.Command{
	Use:   "work",
	Short: "Start working on the next available task",
	Long:  `Assigns the next ready task to a worker agent and begins execution.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := resolveAltDir()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}
		fmt.Println("Work command not yet implemented (requires worker system).")
		return nil
	},
}
