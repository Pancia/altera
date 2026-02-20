package cli

import (
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
