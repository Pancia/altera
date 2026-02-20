package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anthropics/altera/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "alt",
	Short: "Altera - multi-agent orchestration system",
	Long:  `Altera is a multi-agent orchestration system with filesystem-based state (.alt/ directory).`,
}

func Execute() error {
	return rootCmd.Execute()
}

// resolveAltDir finds the .alt directory by walking up from the current working
// directory. Commands that need filesystem state should call this.
func resolveAltDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}
	return config.FindRoot(cwd)
}

// projectRoot returns the parent of the .alt directory.
func projectRoot() (string, error) {
	altDir, err := resolveAltDir()
	if err != nil {
		return "", err
	}
	return filepath.Dir(altDir), nil
}
