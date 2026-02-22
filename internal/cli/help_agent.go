package cli

import (
	"fmt"
	"strings"

	"github.com/anthropics/altera/internal/prompts/help"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(helpAgentCmd)
}

var helpAgentCmd = &cobra.Command{
	Use:   "help <agent-type> <topic> [subtopic...]",
	Short: "Look up agent help topics",
	Long: `Look up embedded help content for Altera agent types.

Agent types: liaison, worker

Examples:
  alt help liaison startup
  alt help worker startup
  alt help worker task-done`,
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Available agent types:")
			for _, t := range help.AgentTypes() {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", t)
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "\nUsage: alt help <agent-type> <topic> [subtopic...]")
			return nil
		}

		agentType := args[0]

		if len(args) == 1 {
			topics, err := help.Topics(agentType)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Available topics for %s:\n", agentType)
			for _, t := range topics {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", t)
			}
			return nil
		}

		content, err := help.Lookup(agentType, args[1:]...)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprint(cmd.OutOrStdout(), content)

		// Ensure trailing newline.
		if !strings.HasSuffix(content, "\n") {
			_, _ = fmt.Fprintln(cmd.OutOrStdout())
		}
		return nil
	},
}
