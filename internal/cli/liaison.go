package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/anthropics/altera/internal/agent"
	"github.com/anthropics/altera/internal/events"
	"github.com/anthropics/altera/internal/liaison"
	"github.com/anthropics/altera/internal/message"
	"github.com/anthropics/altera/internal/task"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(liaisonCmd)
	liaisonCmd.AddCommand(liaisonStartCmd)
	liaisonCmd.AddCommand(liaisonAttachCmd)
	liaisonCmd.AddCommand(liaisonPrimeCmd)
	liaisonCmd.AddCommand(liaisonCheckCmd)
}

var liaisonCmd = &cobra.Command{
	Use:   "liaison",
	Short: "Liaison agent management",
	Long:  `Start, attach, prime, or check messages for the liaison agent.`,
}

// newLiaisonManager constructs a liaison.Manager from the resolved .alt/ directory.
func newLiaisonManager() (*liaison.Manager, error) {
	altDir, err := resolveAltDir()
	if err != nil {
		return nil, fmt.Errorf("not an altera project: %w", err)
	}
	root := filepath.Dir(altDir)

	agents, err := agent.NewStore(filepath.Join(altDir, "agents"))
	if err != nil {
		return nil, fmt.Errorf("opening agent store: %w", err)
	}
	tasks, err := task.NewStore(root)
	if err != nil {
		return nil, fmt.Errorf("opening task store: %w", err)
	}
	msgs, err := message.NewStore(filepath.Join(altDir, "messages"))
	if err != nil {
		return nil, fmt.Errorf("opening message store: %w", err)
	}
	evReader := events.NewReader(filepath.Join(altDir, "events.jsonl"))

	return liaison.NewManager(root, agents, tasks, msgs, evReader), nil
}

var liaisonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the liaison agent",
	Long:  `Creates a tmux session and starts Claude Code with the liaison system prompt.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := newLiaisonManager()
		if err != nil {
			return err
		}
		if err := m.StartLiaison(); err != nil {
			return fmt.Errorf("starting liaison: %w", err)
		}
		fmt.Println("Liaison started in tmux session: alt-liaison")
		fmt.Println("Use 'alt liaison attach' to connect.")
		return nil
	},
}

var liaisonAttachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach to the liaison tmux session",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := liaison.AttachLiaison(); err != nil {
			return fmt.Errorf("attaching to liaison: %w", err)
		}
		return nil
	},
}

var liaisonPrimeCmd = &cobra.Command{
	Use:   "prime",
	Short: "Prime the liaison agent with system state",
	Long:  `Reads all tasks, agents, merge queue, and recent events, then outputs a summary.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := newLiaisonManager()
		if err != nil {
			return err
		}
		summary, err := m.Prime()
		if err != nil {
			return fmt.Errorf("priming liaison: %w", err)
		}
		fmt.Print(summary)
		return nil
	},
}

var liaisonCheckCmd = &cobra.Command{
	Use:   "check-messages [agent-id]",
	Short: "Check pending messages for the liaison (or a specific agent)",
	Long: `Lists all pending messages addressed to the liaison agent.
If an agent-id argument is provided, lists messages for that agent instead.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// If an explicit agent-id is given, use the old tabular format.
		if len(args) == 1 {
			return checkMessagesTabular(args[0])
		}

		// Default: use the liaison manager's formatted output.
		m, err := newLiaisonManager()
		if err != nil {
			return err
		}
		result, err := m.CheckMessages()
		if err != nil {
			return fmt.Errorf("checking messages: %w", err)
		}
		if result == "" {
			return nil
		}
		fmt.Print(result)
		return nil
	},
}

// checkMessagesTabular lists messages for a specific agent in tabular format.
func checkMessagesTabular(agentID string) error {
	altDir, err := resolveAltDir()
	if err != nil {
		return fmt.Errorf("not an altera project: %w", err)
	}

	msgDir := filepath.Join(altDir, "messages")
	store, err := message.NewStore(msgDir)
	if err != nil {
		return fmt.Errorf("opening message store: %w", err)
	}

	msgs, err := store.ListPending(agentID)
	if err != nil {
		return fmt.Errorf("listing messages: %w", err)
	}

	if len(msgs) == 0 {
		fmt.Println("No pending messages.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tTYPE\tFROM\tTASK")
	for _, m := range msgs {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			m.ID, m.Type, m.From, m.TaskID)
	}
	_ = w.Flush()

	return nil
}
