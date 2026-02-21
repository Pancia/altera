package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Set via ldflags at build time.
var (
	Version = "dev"
	Commit  = "unknown"
)

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate(versionOutput())
}

func versionOutput() string {
	return fmt.Sprintf("alt %s (commit %s)\n", Version, Commit)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version and build info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(versionOutput())
	},
}
