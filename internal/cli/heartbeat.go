package cli

import (
	"fmt"
	"path/filepath"

	"github.com/anthropics/altera/internal/agent"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(heartbeatCmd)
}

var heartbeatCmd = &cobra.Command{
	Use:   "heartbeat <agent-id>",
	Short: "Update an agent's heartbeat timestamp",
	Long:  `Touches the heartbeat for the given agent, signaling it is still alive.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}

		store, err := agent.NewStore(filepath.Join(altDir, "agents"))
		if err != nil {
			return fmt.Errorf("opening agent store: %w", err)
		}

		if err := store.TouchHeartbeat(args[0]); err != nil {
			return fmt.Errorf("updating heartbeat: %w", err)
		}

		fmt.Printf("Heartbeat updated for agent %s\n", args[0])
		return nil
	},
}
