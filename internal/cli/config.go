package cli

import (
	"fmt"
	"strconv"

	"github.com/anthropics/altera/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
}

var validKeys = []string{
	"repo_path", "default_branch", "test_command",
	"budget_ceiling", "max_workers", "max_queue_depth",
}

func getField(cfg config.Config, key string) (string, error) {
	switch key {
	case "repo_path":
		return cfg.RepoPath, nil
	case "default_branch":
		return cfg.DefaultBranch, nil
	case "test_command":
		return cfg.TestCommand, nil
	case "budget_ceiling":
		return strconv.FormatFloat(cfg.Constraints.BudgetCeiling, 'f', -1, 64), nil
	case "max_workers":
		return strconv.Itoa(cfg.Constraints.MaxWorkers), nil
	case "max_queue_depth":
		return strconv.Itoa(cfg.Constraints.MaxQueueDepth), nil
	default:
		return "", fmt.Errorf("unknown config key %q (valid keys: %v)", key, validKeys)
	}
}

func setField(cfg *config.Config, key, value string) error {
	switch key {
	case "repo_path":
		cfg.RepoPath = value
	case "default_branch":
		cfg.DefaultBranch = value
	case "test_command":
		cfg.TestCommand = value
	case "budget_ceiling":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float for budget_ceiling: %w", err)
		}
		cfg.Constraints.BudgetCeiling = v
	case "max_workers":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer for max_workers: %w", err)
		}
		cfg.Constraints.MaxWorkers = v
	case "max_queue_depth":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer for max_queue_depth: %w", err)
		}
		cfg.Constraints.MaxQueueDepth = v
	default:
		return fmt.Errorf("unknown config key %q (valid keys: %v)", key, validKeys)
	}
	return nil
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View and modify configuration",
	Long:  `Read and write fields in .alt/config.json.`,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Print a config value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return err
		}
		cfg, err := config.Load(altDir)
		if err != nil {
			return err
		}
		val, err := getField(cfg, args[0])
		if err != nil {
			return err
		}
		fmt.Println(val)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Update a config value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return err
		}
		cfg, err := config.Load(altDir)
		if err != nil {
			return err
		}
		if err := setField(&cfg, args[0], args[1]); err != nil {
			return err
		}
		if err := config.Save(altDir, cfg); err != nil {
			return err
		}
		fmt.Printf("%s = %s\n", args[0], args[1])
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "Print all config values",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return err
		}
		cfg, err := config.Load(altDir)
		if err != nil {
			return err
		}
		for _, key := range validKeys {
			val, _ := getField(cfg, key)
			fmt.Printf("%s = %s\n", key, val)
		}
		return nil
	},
}
