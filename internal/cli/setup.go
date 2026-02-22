package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.AddCommand(setupFishCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup shell integrations",
	Long:  `Configure shell completions and other integrations.`,
}

var setupFishCmd = &cobra.Command{
	Use:   "fish",
	Short: "Install fish shell completions",
	Long:  `Writes Cobra-generated fish completions to ~/.config/fish/completions/alt.fish.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting home directory: %w", err)
		}

		completionDir := filepath.Join(home, ".config", "fish", "completions")
		if err := os.MkdirAll(completionDir, 0o755); err != nil {
			return fmt.Errorf("creating completions directory: %w", err)
		}

		completionFile := filepath.Join(completionDir, "alt.fish")
		f, err := os.Create(completionFile)
		if err != nil {
			return fmt.Errorf("creating completion file: %w", err)
		}
		defer func() { _ = f.Close() }()

		if err := rootCmd.GenFishCompletion(f, true); err != nil {
			return fmt.Errorf("generating fish completions: %w", err)
		}

		fmt.Printf("Fish completions written to %s\n", completionFile)
		fmt.Println("Restart your shell or run 'source' on the file to activate.")
		return nil
	},
}
