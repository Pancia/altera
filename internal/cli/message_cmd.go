package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/altera/internal/agent"
	"github.com/anthropics/altera/internal/message"
	"github.com/anthropics/altera/internal/tmux"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(messageCmd)
	messageCmd.AddCommand(messageSendCmd)
	messageCmd.AddCommand(messageReadCmd)
}

var messageCmd = &cobra.Command{
	Use:     "message",
	Aliases: []string{"msg"},
	Short:   "Send and read messages to agents",
}

var messageSendCmd = &cobra.Command{
	Use:   "send <agent-id> <text...>",
	Short: "Send a message to a running agent",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}

		agentID := args[0]
		body := strings.Join(args[1:], " ")

		// Store the message.
		msgStore, err := message.NewStore(filepath.Join(altDir, "messages"))
		if err != nil {
			return fmt.Errorf("opening message store: %w", err)
		}
		payload := map[string]any{"body": body}
		if _, err := msgStore.Create(message.TypeUserMessage, "user", agentID, "", payload); err != nil {
			return fmt.Errorf("creating message: %w", err)
		}

		// Look up agent to get tmux session.
		agentStore, err := agent.NewStore(filepath.Join(altDir, "agents"))
		if err != nil {
			return fmt.Errorf("opening agent store: %w", err)
		}
		a, err := agentStore.Get(agentID)
		if err != nil {
			// Message stored but can't notify — still useful.
			fmt.Printf("Message stored for %s (could not look up agent for notification: %v)\n", agentID, err)
			return nil
		}
		if a.TmuxSession == "" {
			fmt.Printf("Message stored for %s (no tmux session to notify)\n", agentID)
			return nil
		}

		notification := "You have a new message. Read it with: alt message read"
		if err := tmux.SendText(a.TmuxSession, notification); err != nil {
			fmt.Printf("Message stored for %s (tmux notification failed: %v)\n", agentID, err)
			return nil
		}
		if err := tmux.SendEnter(a.TmuxSession); err != nil {
			fmt.Printf("Message stored for %s (tmux Enter failed: %v)\n", agentID, err)
			return nil
		}

		fmt.Printf("Message sent to %s\n", agentID)
		return nil
	},
}

var messageReadCmd = &cobra.Command{
	Use:   "read [agent-id]",
	Short: "Read and archive pending messages for an agent",
	Long:  `Reads pending user messages for the given agent. If agent-id is omitted, uses the ALT_AGENT_ID environment variable.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		altDir, err := resolveAltDir()
		if err != nil {
			return fmt.Errorf("not an altera project: %w", err)
		}

		var agentID string
		if len(args) > 0 {
			agentID = args[0]
		} else {
			agentID = os.Getenv("ALT_AGENT_ID")
		}
		if agentID == "" {
			return fmt.Errorf("agent-id required (pass as argument or set ALT_AGENT_ID)")
		}

		store, err := message.NewStore(filepath.Join(altDir, "messages"))
		if err != nil {
			return fmt.Errorf("opening message store: %w", err)
		}

		msgs, err := store.ListPending(agentID)
		if err != nil {
			return fmt.Errorf("listing messages: %w", err)
		}

		// Filter to user_message type only.
		var userMsgs []*message.Message
		for _, m := range msgs {
			if m.Type == message.TypeUserMessage {
				userMsgs = append(userMsgs, m)
			}
		}

		if len(userMsgs) == 0 {
			fmt.Println("No messages.")
			return nil
		}

		for _, m := range userMsgs {
			body, _ := m.Payload["body"].(string)
			fmt.Printf("[%s] %s\n", m.CreatedAt.Format("15:04:05"), body)
			if err := store.Archive(m.ID); err != nil {
				return fmt.Errorf("archiving message %s: %w", m.ID, err)
			}
		}
		return nil
	},
}
