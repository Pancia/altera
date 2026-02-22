// Package tmux provides session management for the Altera multi-agent system.
// All operations shell out to the tmux binary. Sessions follow the naming
// convention alt-{role}-{id} (e.g., alt-worker-03, alt-liaison).
package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// SocketName is the tmux server socket name used to isolate alt sessions.
const SocketName = "alt"

// SessionPrefix is prepended to all Altera-managed tmux session names.
const SessionPrefix = "alt-"

// socketArgs returns the args to select the alt tmux server socket.
func socketArgs() []string {
	return []string{"-L", SocketName}
}

// SessionName builds a canonical session name from role and id.
func SessionName(role, id string) string {
	return SessionPrefix + role + "-" + id
}

// CreateSession creates a new detached tmux session with the given name.
func CreateSession(name string) error {
	args := append(socketArgs(), "new-session", "-d", "-s", name)
	cmd := exec.Command("tmux", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tmux new-session %q: %s: %w", name, strings.TrimSpace(string(out)), err)
	}
	// Enable mouse mode for scrolling in this session only.
	mouseArgs := append(socketArgs(), "set-option", "-t", name, "mouse", "on")
	mouseCmd := exec.Command("tmux", mouseArgs...)
	_ = mouseCmd.Run()
	return nil
}

// KillSession destroys the tmux session with the given name.
func KillSession(name string) error {
	args := append(socketArgs(), "kill-session", "-t", name)
	cmd := exec.Command("tmux", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tmux kill-session %q: %s: %w", name, strings.TrimSpace(string(out)), err)
	}
	return nil
}

// SessionExists returns true if the named tmux session exists.
func SessionExists(name string) bool {
	args := append(socketArgs(), "has-session", "-t", name)
	cmd := exec.Command("tmux", args...)
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
	args := append(socketArgs(), "attach-session", "-t", name)
	cmd := exec.Command(tmuxPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tmux attach-session %q: %w", name, err)
	}
	return nil
}

// PanePID returns the PID of the foreground process in the session's active pane.
func PanePID(session string) (int, error) {
	args := append(socketArgs(), "list-panes", "-t", session, "-F", "#{pane_pid}")
	cmd := exec.Command("tmux", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("tmux list-panes %q: %s: %w", session, strings.TrimSpace(string(out)), err)
	}
	line := strings.TrimSpace(strings.Split(strings.TrimSpace(string(out)), "\n")[0])
	pid, err := strconv.Atoi(line)
	if err != nil {
		return 0, fmt.Errorf("parse pane pid %q: %w", line, err)
	}
	return pid, nil
}

// SendKeys sends the given keys (typically a shell command + Enter) to a
// tmux session's active pane.
func SendKeys(session, keys string) error {
	args := append(socketArgs(), "send-keys", "-t", session, keys, "Enter")
	cmd := exec.Command("tmux", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tmux send-keys %q: %s: %w", session, strings.TrimSpace(string(out)), err)
	}
	return nil
}

// SendText sends text to a session WITHOUT pressing Enter.
func SendText(session, text string) error {
	args := append(socketArgs(), "send-keys", "-t", session, text)
	cmd := exec.Command("tmux", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tmux send-text %q: %s: %w", session, strings.TrimSpace(string(out)), err)
	}
	return nil
}

// SendEnter sends the Enter key to a session.
func SendEnter(session string) error {
	args := append(socketArgs(), "send-keys", "-t", session, "Enter")
	cmd := exec.Command("tmux", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tmux send-enter %q: %s: %w", session, strings.TrimSpace(string(out)), err)
	}
	return nil
}

// CapturePane captures the last n lines of output from the session's pane.
// If lines is 0, tmux captures the visible pane content.
func CapturePane(session string, lines int) (string, error) {
	args := append(socketArgs(), "capture-pane", "-t", session, "-p")
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
	args := append(socketArgs(), "list-sessions", "-F", "#{session_name}")
	cmd := exec.Command("tmux", args...)
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
