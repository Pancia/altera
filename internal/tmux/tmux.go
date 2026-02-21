// Package tmux provides session management for the Altera multi-agent system.
// All operations shell out to the tmux binary. Sessions follow the naming
// convention alt-{role}-{id} (e.g., alt-worker-03, alt-liaison).
package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// SessionPrefix is prepended to all Altera-managed tmux session names.
const SessionPrefix = "alt-"

// SessionName builds a canonical session name from role and id.
func SessionName(role, id string) string {
	return SessionPrefix + role + "-" + id
}

// CreateSession creates a new detached tmux session with the given name.
func CreateSession(name string) error {
	cmd := exec.Command("tmux", "new-session", "-d", "-s", name)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tmux new-session %q: %s: %w", name, strings.TrimSpace(string(out)), err)
	}
	// Enable mouse mode for scrolling in this session only.
	mouseCmd := exec.Command("tmux", "set-option", "-t", name, "mouse", "on")
	_ = mouseCmd.Run()
	return nil
}

// KillSession destroys the tmux session with the given name.
func KillSession(name string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tmux kill-session %q: %s: %w", name, strings.TrimSpace(string(out)), err)
	}
	return nil
}

// SessionExists returns true if the named tmux session exists.
func SessionExists(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	return cmd.Run() == nil
}

// AttachSession attaches the terminal to the given tmux session.
// Stdin, stdout, and stderr are connected to the current terminal for
// interactive use.
func AttachSession(name string) error {
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("tmux not found: %w", err)
	}
	cmd := exec.Command(tmuxPath, "attach-session", "-t", name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tmux attach-session %q: %w", name, err)
	}
	return nil
}

// SendKeys sends the given keys (typically a shell command + Enter) to a
// tmux session's active pane.
func SendKeys(session, keys string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", session, keys, "Enter")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tmux send-keys %q: %s: %w", session, strings.TrimSpace(string(out)), err)
	}
	return nil
}

// CapturePane captures the last n lines of output from the session's pane.
// If lines is 0, tmux captures the visible pane content.
func CapturePane(session string, lines int) (string, error) {
	args := []string{"capture-pane", "-t", session, "-p"}
	if lines > 0 {
		start := fmt.Sprintf("-%d", lines)
		args = append(args, "-S", start)
	}
	cmd := exec.Command("tmux", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tmux capture-pane %q: %s: %w", session, strings.TrimSpace(string(out)), err)
	}
	return string(out), nil
}

// ListSessions returns the names of all Altera-managed tmux sessions
// (those starting with SessionPrefix).
func ListSessions() ([]string, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// "no server running" or "no sessions" is not an error for listing.
		if strings.Contains(string(out), "no server running") ||
			strings.Contains(string(out), "no sessions") {
			return nil, nil
		}
		return nil, fmt.Errorf("tmux list-sessions: %s: %w", strings.TrimSpace(string(out)), err)
	}
	var sessions []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, SessionPrefix) {
			sessions = append(sessions, line)
		}
	}
	return sessions, nil
}

// WaitForSessionReady polls until the named session exists or the timeout
// elapses. Returns nil if the session is found, or an error on timeout.
func WaitForSessionReady(name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if SessionExists(name) {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for tmux session %q after %s", name, timeout)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
