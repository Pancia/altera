package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/anthropics/altera/internal/task"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskShowCmd)
	taskCmd.AddCommand(taskCreateCmd)

	taskListCmd.Flags().StringVar(&taskListStatus, "status", "", "filter by status (open, assigned, in_progress, done, failed)")
	taskListCmd.Flags().StringVar(&taskListRig, "rig", "", "filter by rig")
	taskListCmd.Flags().StringVar(&taskListAssignee, "assignee", "", "filter by assignee")
	taskListCmd.Flags().StringVar(&taskListTag, "tag", "", "filter by tag")

	taskListCmd.RegisterFlagCompletionFunc("status", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"open", "assigned", "in_progress", "done", "failed"}, cobra.ShellCompDirectiveNoFileComp
	})
	taskListCmd.RegisterFlagCompletionFunc("rig", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return completeRigNames(cmd, nil, toComplete)
	})

	taskCreateCmd.Flags().StringVar(&taskCreateTitle, "title", "", "task title (required)")
	taskCreateCmd.Flags().StringVar(&taskCreateDesc, "description", "", "task description")
}

var (
	taskListStatus   string
	taskListRig      string
	taskListAssignee string
	taskListTag      string

	taskCreateTitle string
	taskCreateDesc  string
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
	Long:  `Create, list, and show tasks.`,
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long:  `List all tasks, optionally filtered by status, rig, assignee, or tag.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := projectRoot()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}

		store, err := task.NewStore(root)
		if err != nil {
			return fmt.Errorf("opening task store: %w", err)
		}

		var f task.Filter
		if taskListStatus != "" {
			s, err := task.ParseStatus(taskListStatus)
			if err != nil {
				return err
			}
			f.Status = s
		}
		f.Rig = taskListRig
		f.AssignedTo = taskListAssignee
		f.Tag = taskListTag

		tasks, err := store.List(f)
		if err != nil {
			return fmt.Errorf("listing tasks: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tSTATUS\tASSIGNED\tRIG\tTITLE")
		for _, t := range tasks {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				t.ID, t.Status, t.AssignedTo, t.Rig, t.Title)
		}
		w.Flush()

		if len(tasks) == 0 {
			fmt.Println("No tasks found.")
		}
		return nil
	},
}

var taskShowCmd = &cobra.Command{
	Use:               "show <id>",
	Short:             "Show task details",
	Long:              `Display full details for a task.`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeTaskIDs,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := projectRoot()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}

		store, err := task.NewStore(root)
		if err != nil {
			return fmt.Errorf("opening task store: %w", err)
		}

		t, err := store.Get(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("ID:          %s\n", t.ID)
		fmt.Printf("Title:       %s\n", t.Title)
		fmt.Printf("Status:      %s\n", t.Status)
		if t.Description != "" {
			fmt.Printf("Description: %s\n", t.Description)
		}
		if t.AssignedTo != "" {
			fmt.Printf("Assigned To: %s\n", t.AssignedTo)
		}
		if t.Branch != "" {
			fmt.Printf("Branch:      %s\n", t.Branch)
		}
		if t.Rig != "" {
			fmt.Printf("Rig:         %s\n", t.Rig)
		}
		if t.CreatedBy != "" {
			fmt.Printf("Created By:  %s\n", t.CreatedBy)
		}
		fmt.Printf("Created At:  %s\n", t.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated At:  %s\n", t.UpdatedAt.Format("2006-01-02 15:04:05"))
		if t.ParentID != "" {
			fmt.Printf("Parent:      %s\n", t.ParentID)
		}
		if len(t.Deps) > 0 {
			fmt.Printf("Deps:        %s\n", strings.Join(t.Deps, ", "))
		}
		if len(t.Tags) > 0 {
			fmt.Printf("Tags:        %s\n", strings.Join(t.Tags, ", "))
		}
		if t.Priority != 0 {
			fmt.Printf("Priority:    %d\n", t.Priority)
		}
		if t.Result != "" {
			fmt.Printf("Result:      %s\n", t.Result)
		}
		if t.Checkpoint != "" {
			fmt.Printf("Checkpoint:  %s\n", t.Checkpoint)
		}

		return nil
	},
}

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	Long:  `Create a new task with --title and optional --description.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if taskCreateTitle == "" {
			return fmt.Errorf("--title is required")
		}

		root, err := projectRoot()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}

		store, err := task.NewStore(root)
		if err != nil {
			return fmt.Errorf("opening task store: %w", err)
		}

		t := &task.Task{
			Title:       taskCreateTitle,
			Description: taskCreateDesc,
		}
		if err := store.Create(t); err != nil {
			return fmt.Errorf("creating task: %w", err)
		}

		altDir := filepath.Join(root, ".alt")
		evtPath := filepath.Join(altDir, "events.jsonl")
		logTaskCreated(evtPath, t.ID)

		fmt.Printf("Created task %s: %s\n", t.ID, t.Title)
		return nil
	},
}
