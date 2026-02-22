package cli

import (
	"path/filepath"
	"strings"

	"github.com/anthropics/altera/internal/agent"
	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/prompts/help"
	"github.com/anthropics/altera/internal/task"
	"github.com/anthropics/altera/internal/tmux"
	"github.com/spf13/cobra"
)

// completeAgentTypes returns agent types for the help command's first argument.
func completeHelpArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return help.AgentTypes(), cobra.ShellCompDirectiveNoFileComp
	case 1:
		topics, err := help.Topics(args[0])
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return topics, cobra.ShellCompDirectiveNoFileComp
	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// completeTaskIDs returns task IDs from the store.
func completeTaskIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	root, err := projectRoot()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	store, err := task.NewStore(root)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	tasks, err := store.List(task.Filter{})
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var ids []string
	for _, t := range tasks {
		ids = append(ids, t.ID+"\t"+t.Title)
	}
	return ids, cobra.ShellCompDirectiveNoFileComp
}

// completeWorkerIDs returns worker agent IDs from the store.
func completeWorkerIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	altDir, err := resolveAltDir()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	store, err := agent.NewStore(filepath.Join(altDir, "agents"))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	workers, err := store.ListByRole(agent.RoleWorker)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var ids []string
	for _, w := range workers {
		ids = append(ids, w.ID+"\t"+string(w.Status))
	}
	return ids, cobra.ShellCompDirectiveNoFileComp
}

// completeAgentIDs returns all agent IDs from the store.
func completeAgentIDs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	altDir, err := resolveAltDir()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	store, err := agent.NewStore(filepath.Join(altDir, "agents"))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	agents, err := store.ListByStatus(agent.StatusActive)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	idle, _ := store.ListByStatus(agent.StatusIdle)
	agents = append(agents, idle...)
	var ids []string
	for _, a := range agents {
		ids = append(ids, a.ID+"\t"+string(a.Role))
	}
	return ids, cobra.ShellCompDirectiveNoFileComp
}

// completeSessionNames returns Altera tmux session names.
func completeSessionNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	sessions, err := tmux.ListSessions()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	// Strip the "alt-" prefix for friendlier completion since the command auto-prepends it.
	var names []string
	for _, s := range sessions {
		name := strings.TrimPrefix(s, tmux.SessionPrefix)
		names = append(names, name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

// completeRigNames returns configured rig names.
func completeRigNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	altDir, err := resolveAltDir()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	names, err := config.ListRigs(altDir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
