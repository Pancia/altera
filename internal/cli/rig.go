package cli

import (
	"fmt"

	"github.com/anthropics/altera/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(rigCmd)
	rigCmd.AddCommand(rigAddCmd)
	rigCmd.AddCommand(rigListCmd)

	rigAddCmd.Flags().StringVar(&rigAddRepo, "repo", "", "path to the repository")
	rigAddCmd.Flags().StringVar(&rigAddBranch, "branch", "main", "default branch")
	rigAddCmd.Flags().StringVar(&rigAddTest, "test", "", "test command")
}

var (
	rigAddRepo   string
	rigAddBranch string
	rigAddTest   string
)

var rigCmd = &cobra.Command{
	Use:   "rig",
	Short: "Manage rigs",
	Long:  `Add and list rigs (repository configurations).`,
}

var rigAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a rig",
	Long:  `Register a new rig with its repository path and configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}

		name := args[0]
		rc := config.RigConfig{
			RepoPath:      rigAddRepo,
			DefaultBranch: rigAddBranch,
			TestCommand:   rigAddTest,
		}

		cfg, err := config.Load(altDir)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		cfg.Rigs[name] = rc
		if err := config.Save(altDir, cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		if err := config.SaveRig(altDir, name, rc); err != nil {
			return fmt.Errorf("saving rig config: %w", err)
		}

		fmt.Printf("Added rig %q\n", name)
		return nil
	},
}

var rigListCmd = &cobra.Command{
	Use:   "list",
	Short: "List rigs",
	Long:  `List all registered rigs.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}

		names, err := config.ListRigs(altDir)
		if err != nil {
			return fmt.Errorf("listing rigs: %w", err)
		}

		if len(names) == 0 {
			fmt.Println("No rigs configured.")
			return nil
		}

		for _, name := range names {
			rc, err := config.LoadRig(altDir, name)
			if err != nil {
				fmt.Printf("  %s  (error loading config)\n", name)
				continue
			}
			fmt.Printf("  %s  repo=%s  branch=%s\n", name, rc.RepoPath, rc.DefaultBranch)
		}
		return nil
	},
}
