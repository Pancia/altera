package tmux

import (
	"crypto/rand"
	"encoding/hex"
	"os/exec"
	"testing"
)

// UseTestSocket sets a unique tmux server socket for the duration of the
// test, isolating it from the production "alt" socket and other test runs.
// It registers a cleanup that kills the entire tmux server on that socket.
func UseTestSocket(t *testing.T) {
	t.Helper()

	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		t.Fatalf("generate random socket name: %v", err)
	}
	name := "alt-test-" + hex.EncodeToString(buf[:])

	old := socketName
	socketName = name

	t.Cleanup(func() {
		_ = exec.Command("tmux", "-L", name, "kill-server").Run()
		socketName = old
	})
}
