package tmux

import (
	"os/exec"
	"strings"
	"testing"
	"time"
)

// requireTmux skips the test if tmux is not available.
func requireTmux(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux not available, skipping integration test")
	}
}

// testSession creates a uniquely named session for the test and registers
// cleanup to kill it afterward.
func testSession(t *testing.T) string {
	t.Helper()
	requireTmux(t)
	UseTestSocket(t)
	name := SessionPrefix + "test-" + t.Name()
	// Sanitize: tmux session names can't contain dots or colons.
	name = strings.NewReplacer(".", "_", ":", "_", "/", "_").Replace(name)
	t.Cleanup(func() {
		// Best-effort cleanup on the test socket.
		_ = exec.Command("tmux", "-L", socketName, "kill-session", "-t", name).Run()
	})
	return name
}

func TestSessionName(t *testing.T) {
	got := SessionName("worker", "03")
	if got != "alt-worker-03" {
		t.Fatalf("expected alt-worker-03, got %s", got)
	}
	got = SessionName("liaison", "abc")
	if got != "alt-liaison-abc" {
		t.Fatalf("expected alt-liaison-abc, got %s", got)
	}
}

func TestCreateAndKillSession(t *testing.T) {
	name := testSession(t)

	if err := CreateSession(name); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if !SessionExists(name) {
		t.Fatal("session should exist after create")
	}

	if err := KillSession(name); err != nil {
		t.Fatalf("KillSession: %v", err)
	}
	if SessionExists(name) {
		t.Fatal("session should not exist after kill")
	}
}

func TestCreateDuplicateSession(t *testing.T) {
	name := testSession(t)

	if err := CreateSession(name); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	// Creating a session with the same name should fail.
	if err := CreateSession(name); err == nil {
		t.Fatal("expected error creating duplicate session")
	}
}

func TestKillNonexistentSession(t *testing.T) {
	requireTmux(t)
	UseTestSocket(t)
	err := KillSession("alt-test-nonexistent-xyz")
	if err == nil {
		t.Fatal("expected error killing nonexistent session")
	}
}

func TestSessionExistsNonexistent(t *testing.T) {
	requireTmux(t)
	UseTestSocket(t)
	if SessionExists("alt-test-nonexistent-xyz") {
		t.Fatal("nonexistent session should return false")
	}
}

func TestSendKeysAndCapturePane(t *testing.T) {
	name := testSession(t)

	if err := CreateSession(name); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// Send a command that produces predictable output.
	if err := SendKeys(name, "echo ALTERA_TMUX_TEST_MARKER"); err != nil {
		t.Fatalf("SendKeys: %v", err)
	}

	// Give the shell a moment to process the command.
	time.Sleep(500 * time.Millisecond)

	output, err := CapturePane(name, 20)
	if err != nil {
		t.Fatalf("CapturePane: %v", err)
	}

	if !strings.Contains(output, "ALTERA_TMUX_TEST_MARKER") {
		t.Fatalf("expected output to contain ALTERA_TMUX_TEST_MARKER, got:\n%s", output)
	}
}

func TestCapturePaneNoLines(t *testing.T) {
	name := testSession(t)

	if err := CreateSession(name); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// Capture with lines=0 (visible pane content) should not error.
	_, err := CapturePane(name, 0)
	if err != nil {
		t.Fatalf("CapturePane(0): %v", err)
	}
}

func TestListSessions(t *testing.T) {
	name := testSession(t)

	if err := CreateSession(name); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	sessions, err := ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}

	found := false
	for _, s := range sessions {
		if s == name {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected to find %q in sessions: %v", name, sessions)
	}
}

func TestListSessionsFiltersNonAlt(t *testing.T) {
	requireTmux(t)
	UseTestSocket(t)

	// Create a non-alt session on the test socket.
	nonAlt := "noalt-test-" + t.Name()
	nonAlt = strings.NewReplacer(".", "_", ":", "_", "/", "_").Replace(nonAlt)
	_ = exec.Command("tmux", "-L", socketName, "new-session", "-d", "-s", nonAlt).Run()
	t.Cleanup(func() {
		_ = exec.Command("tmux", "-L", socketName, "kill-session", "-t", nonAlt).Run()
	})

	sessions, err := ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}

	for _, s := range sessions {
		if s == nonAlt {
			t.Fatalf("ListSessions should not include non-alt session %q", nonAlt)
		}
	}
}

func TestWaitForSessionReady(t *testing.T) {
	name := testSession(t)

	if err := CreateSession(name); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// Session already exists, should return immediately.
	if err := WaitForSessionReady(name, 2*time.Second); err != nil {
		t.Fatalf("WaitForSessionReady: %v", err)
	}
}

func TestWaitForSessionReadyTimeout(t *testing.T) {
	requireTmux(t)
	UseTestSocket(t)

	err := WaitForSessionReady("alt-test-never-exists-xyz", 300*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("expected timeout error, got: %v", err)
	}
}

func TestSendKeysNonexistent(t *testing.T) {
	requireTmux(t)
	UseTestSocket(t)
	err := SendKeys("alt-test-nonexistent-xyz", "echo hi")
	if err == nil {
		t.Fatal("expected error sending keys to nonexistent session")
	}
}

func TestCapturePaneNonexistent(t *testing.T) {
	requireTmux(t)
	UseTestSocket(t)
	_, err := CapturePane("alt-test-nonexistent-xyz", 10)
	if err == nil {
		t.Fatal("expected error capturing from nonexistent session")
	}
}
