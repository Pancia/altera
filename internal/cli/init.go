package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/git"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize an altera project",
	Long:  `Creates .alt/ with full directory structure and initializes git if not already a repo.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}

		// Initialize git if no .git/ exists.
		gitDir := filepath.Join(cwd, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			if err := git.Init(cwd); err != nil {
				return fmt.Errorf("initializing git: %w", err)
			}
			fmt.Println("Initialized git repository.")
		}

		// Create .alt/ with all subdirectories.
		altDir, err := config.EnsureDir(cwd)
		if err != nil {
			return fmt.Errorf("creating .alt/ directory: %w", err)
		}

		// Write default config.json if it doesn't exist.
		cfgPath := filepath.Join(altDir, "config.json")
		if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
			cfg := config.NewConfig()
			if err := config.Save(altDir, cfg); err != nil {
				return fmt.Errorf("writing default config: %w", err)
			}
		}

		fmt.Printf("Initialized altera project in %s\n", cwd)
		return nil
	},
}
