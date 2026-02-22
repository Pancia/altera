package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/anthropics/altera/internal/tmux"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(tmuxCmd)
	tmuxCmd.AddCommand(tmuxListCmd)
	tmuxCmd.AddCommand(tmuxAttachCmd)
	tmuxCmd.AddCommand(tmuxClientCmd)
}

var tmuxCmd = &cobra.Command{
	Use:   "tmux",
	Short: "Manage alt tmux sessions",
	Long:  `List, attach, or open a client for alt-managed tmux sessions running on the alt server socket.`,
}

var tmuxListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sessions on the alt tmux server",
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := tmux.ListSessions()
		if err != nil {
			return err
		}
		if len(sessions) == 0 {
			fmt.Println("No sessions.")
			return nil
		}
		for _, s := range sessions {
			fmt.Println(s)
		}
		return nil
	},
}

var tmuxAttachCmd = &cobra.Command{
	Use:   "attach <session>",
	Short: "Attach to an alt tmux session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return tmux.AttachSession(args[0])
	},
}

var tmuxClientCmd = &cobra.Command{
	Use:   "client",
	Short: "Open a bare tmux client on the alt server",
	Long:  `Drops into a tmux client connected to the alt server socket, giving full tmux navigation (Ctrl-B+s, etc.).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tmuxPath, err := exec.LookPath("tmux")
		if err != nil {
			return fmt.Errorf("tmux not found: %w", err)
		}
		c := exec.Command(tmuxPath, "-L", tmux.SocketName)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}
