package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/anthropics/altera/internal/message"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(liaisonCmd)
	liaisonCmd.AddCommand(liaisonPrimeCmd)
	liaisonCmd.AddCommand(liaisonCheckCmd)
}

var liaisonCmd = &cobra.Command{
	Use:   "liaison",
	Short: "Liaison agent communication",
	Long:  `Prime the liaison or check messages for an agent.`,
}

var liaisonPrimeCmd = &cobra.Command{
	Use:   "prime",
	Short: "Prime the liaison agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := resolveAltDir()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}
		fmt.Println("Liaison primed (not yet implemented).")
		return nil
	},
}

var liaisonCheckCmd = &cobra.Command{
	Use:   "check-messages <agent-id>",
	Short: "Check pending messages for an agent",
	Long:  `Lists all pending messages addressed to the given agent.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}

		msgDir := filepath.Join(altDir, "messages")
		store, err := message.NewStore(msgDir)
		if err != nil {
			return fmt.Errorf("opening message store: %w", err)
		}

		msgs, err := store.ListPending(args[0])
		if err != nil {
			return fmt.Errorf("listing messages: %w", err)
		}

		if len(msgs) == 0 {
			fmt.Println("No pending messages.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTYPE\tFROM\tTASK")
		for _, m := range msgs {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				m.ID, m.Type, m.From, m.TaskID)
		}
		w.Flush()

		return nil
	},
}
