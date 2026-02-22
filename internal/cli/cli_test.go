package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/anthropics/altera/internal/events"
	"github.com/anthropics/altera/internal/task"
)

// executeCmd runs rootCmd with the given args and returns stdout output.
func executeCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

// setupProject creates a temporary .alt project directory and returns
// the project root. It changes the working directory to root and
// restores it on cleanup.
func setupProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	altDir := filepath.Join(root, ".alt")
	for _, sub := range []string{"tasks", "agents", "messages", "messages/archive", "rigs"} {
		if err := os.MkdirAll(filepath.Join(altDir, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	return root
}

func TestRootCommand(t *testing.T) {
	// Root command with no args should not error (prints help).
	_, err := executeCmd(t)
	if err != nil {
		t.Fatalf("root command failed: %v", err)
	}
}

func TestSubcommandRegistration(t *testing.T) {
	expected := []string{
		"status", "task", "log", "rig", "work",
		"daemon", "heartbeat", "checkpoint", "liaison",
		"worker", "session", "prime", "setup", "help",
		"task-done",
	}
	cmds := rootCmd.Commands()
	names := make(map[string]bool)
	for _, c := range cmds {
		names[c.Name()] = true
	}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}

func TestTaskSubcommands(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "task" {
			expected := []string{"list", "show", "create"}
			subs := c.Commands()
			names := make(map[string]bool)
			for _, s := range subs {
				names[s.Name()] = true
			}
			for _, name := range expected {
				if !names[name] {
					t.Errorf("expected task subcommand %q not found", name)
				}
			}
			return
		}
	}
	t.Fatal("task command not found")
}

func TestDaemonSubcommands(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "daemon" {
			expected := []string{"start", "stop", "status"}
			subs := c.Commands()
			names := make(map[string]bool)
			for _, s := range subs {
				names[s.Name()] = true
			}
			for _, name := range expected {
				if !names[name] {
					t.Errorf("expected daemon subcommand %q not found", name)
				}
			}
			return
		}
	}
	t.Fatal("daemon command not found")
}

func TestLiaisonSubcommands(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "liaison" {
			expected := []string{"prime", "check-messages"}
			subs := c.Commands()
			names := make(map[string]bool)
			for _, s := range subs {
				names[s.Name()] = true
			}
			for _, name := range expected {
				if !names[name] {
					t.Errorf("expected liaison subcommand %q not found", name)
				}
			}
			return
		}
	}
	t.Fatal("liaison command not found")
}

func TestRigSubcommands(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "rig" {
			expected := []string{"add", "list"}
			subs := c.Commands()
			names := make(map[string]bool)
			for _, s := range subs {
				names[s.Name()] = true
			}
			for _, name := range expected {
				if !names[name] {
					t.Errorf("expected rig subcommand %q not found", name)
				}
			}
			return
		}
	}
	t.Fatal("rig command not found")
}

func TestTaskListFlags(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "task" {
			for _, s := range c.Commands() {
				if s.Name() == "list" {
					for _, flag := range []string{"status", "rig", "assignee", "tag"} {
						f := s.Flags().Lookup(flag)
						if f == nil {
							t.Errorf("task list missing --%s flag", flag)
						}
					}
					return
				}
			}
		}
	}
	t.Fatal("task list command not found")
}

func TestTaskShowRequiresArg(t *testing.T) {
	root := setupProject(t)
	_ = root
	_, err := executeCmd(t, "task", "show")
	if err == nil {
		t.Error("expected error when no task ID provided")
	}
}

func TestTaskCreateRequiresTitle(t *testing.T) {
	setupProject(t)
	_, err := executeCmd(t, "task", "create")
	if err == nil {
		t.Error("expected error when --title not provided")
	}
}

func TestTaskCreateAndShow(t *testing.T) {
	root := setupProject(t)

	// Create a task directly via the store so we can verify show works.
	store, err := task.NewStore(root)
	if err != nil {
		t.Fatal(err)
	}
	tk := &task.Task{
		ID:    "t-test01",
		Title: "Test Task",
	}
	if err := store.Create(tk); err != nil {
		t.Fatal(err)
	}

	_, err = executeCmd(t, "task", "show", "t-test01")
	if err != nil {
		t.Fatalf("task show failed: %v", err)
	}
}

func TestTaskListEmpty(t *testing.T) {
	setupProject(t)
	_, err := executeCmd(t, "task", "list")
	if err != nil {
		t.Fatalf("task list failed: %v", err)
	}
}

func TestLogEmpty(t *testing.T) {
	setupProject(t)
	_, err := executeCmd(t, "log")
	if err != nil {
		t.Fatalf("log command failed: %v", err)
	}
}

func TestLogWithEvents(t *testing.T) {
	root := setupProject(t)

	evtPath := filepath.Join(root, ".alt", "events.jsonl")
	writer := events.NewWriter(evtPath)
	if err := writer.Append(events.Event{
		Timestamp: time.Now(),
		Type:      events.TaskCreated,
		TaskID:    "t-abc123",
	}); err != nil {
		t.Fatal(err)
	}

	_, err := executeCmd(t, "log")
	if err != nil {
		t.Fatalf("log command failed: %v", err)
	}
}

func TestLogLastFlag(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "log" {
			f := c.Flags().Lookup("last")
			if f == nil {
				t.Error("log missing --last flag")
			}
			return
		}
	}
	t.Fatal("log command not found")
}

func TestHeartbeatRequiresArg(t *testing.T) {
	setupProject(t)
	_, err := executeCmd(t, "heartbeat")
	if err == nil {
		t.Error("expected error when no agent ID provided")
	}
}

func TestCheckpointRequiresArg(t *testing.T) {
	setupProject(t)
	_, err := executeCmd(t, "checkpoint")
	if err == nil {
		t.Error("expected error when no task ID provided")
	}
}

func TestLiaisonCheckMessagesNoArg(t *testing.T) {
	setupProject(t)
	// check-messages without args defaults to checking liaison messages.
	_, err := executeCmd(t, "liaison", "check-messages")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRigAddRequiresName(t *testing.T) {
	setupProject(t)
	_, err := executeCmd(t, "rig", "add")
	if err == nil {
		t.Error("expected error when no rig name provided")
	}
}

func TestRigAddAndList(t *testing.T) {
	root := setupProject(t)

	// Write a minimal config so rig add can load it.
	cfgPath := filepath.Join(root, ".alt", "config.json")
	cfgData, _ := json.Marshal(map[string]any{
		"rigs":        map[string]any{},
		"constraints": map[string]any{},
	})
	if err := os.WriteFile(cfgPath, cfgData, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := executeCmd(t, "rig", "add", "myrig", "--repo", "/tmp/repo", "--branch", "develop")
	if err != nil {
		t.Fatalf("rig add failed: %v", err)
	}

	_, err = executeCmd(t, "rig", "list")
	if err != nil {
		t.Fatalf("rig list failed: %v", err)
	}
}

func TestStatusCommand(t *testing.T) {
	setupProject(t)
	_, err := executeCmd(t, "status")
	if err != nil {
		t.Fatalf("status command failed: %v", err)
	}
}

func TestRigAddFlags(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "rig" {
			for _, s := range c.Commands() {
				if s.Name() == "add" {
					for _, flag := range []string{"repo", "branch", "test"} {
						f := s.Flags().Lookup(flag)
						if f == nil {
							t.Errorf("rig add missing --%s flag", flag)
						}
					}
					return
				}
			}
		}
	}
	t.Fatal("rig add command not found")
}

func TestCheckpointMessageFlag(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "checkpoint" {
			f := c.Flags().Lookup("message")
			if f == nil {
				t.Error("checkpoint missing --message flag")
			}
			return
		}
	}
	t.Fatal("checkpoint command not found")
}

func TestTaskCreateFlags(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "task" {
			for _, s := range c.Commands() {
				if s.Name() == "create" {
					for _, flag := range []string{"title", "description"} {
						f := s.Flags().Lookup(flag)
						if f == nil {
							t.Errorf("task create missing --%s flag", flag)
						}
					}
					return
				}
			}
		}
	}
	t.Fatal("task create command not found")
}

func TestWorkerSubcommands(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "worker" {
			expected := []string{"list", "attach", "peek", "kill", "inspect"}
			subs := c.Commands()
			names := make(map[string]bool)
			for _, s := range subs {
				names[s.Name()] = true
			}
			for _, name := range expected {
				if !names[name] {
					t.Errorf("expected worker subcommand %q not found", name)
				}
			}
			return
		}
	}
	t.Fatal("worker command not found")
}

func TestSessionSubcommands(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "session" {
			expected := []string{"list", "switch"}
			subs := c.Commands()
			names := make(map[string]bool)
			for _, s := range subs {
				names[s.Name()] = true
			}
			for _, name := range expected {
				if !names[name] {
					t.Errorf("expected session subcommand %q not found", name)
				}
			}
			return
		}
	}
	t.Fatal("session command not found")
}

func TestPrimeCommand(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "prime" {
			found = true
			// Verify flags exist.
			for _, flag := range []string{"role", "agent-id"} {
				f := c.Flags().Lookup(flag)
				if f == nil {
					t.Errorf("prime missing --%s flag", flag)
				}
			}
			break
		}
	}
	if !found {
		t.Fatal("prime command not found")
	}
}

func TestSetupSubcommands(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "setup" {
			expected := []string{"fish"}
			subs := c.Commands()
			names := make(map[string]bool)
			for _, s := range subs {
				names[s.Name()] = true
			}
			for _, name := range expected {
				if !names[name] {
					t.Errorf("expected setup subcommand %q not found", name)
				}
			}
			return
		}
	}
	t.Fatal("setup command not found")
}

func TestHelpAgentNoArgs(t *testing.T) {
	out, err := executeCmd(t, "help")
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}
	if !strings.Contains(out, "liaison") || !strings.Contains(out, "worker") {
		t.Error("expected output to list agent types")
	}
}

func TestHelpAgentListTopics(t *testing.T) {
	out, err := executeCmd(t, "help", "worker")
	if err != nil {
		t.Fatalf("help worker failed: %v", err)
	}
	if !strings.Contains(out, "startup") || !strings.Contains(out, "task-done") {
		t.Error("expected output to list worker topics")
	}
}

func TestHelpAgentShowTopic(t *testing.T) {
	out, err := executeCmd(t, "help", "worker", "startup")
	if err != nil {
		t.Fatalf("help worker startup failed: %v", err)
	}
	if !strings.Contains(out, "Worker: Startup") {
		t.Error("expected output to contain topic content")
	}
}

func TestHelpAgentUnknownType(t *testing.T) {
	_, err := executeCmd(t, "help", "badtype")
	if err == nil {
		t.Error("expected error for unknown agent type")
	}
}

func TestHelpAgentUnknownTopic(t *testing.T) {
	_, err := executeCmd(t, "help", "worker", "nonexistent")
	if err == nil {
		t.Error("expected error for unknown topic")
	}
}

func TestTaskDoneRequiresArgs(t *testing.T) {
	setupProject(t)
	_, err := executeCmd(t, "task-done")
	if err == nil {
		t.Error("expected error when no args provided")
	}
	_, err = executeCmd(t, "task-done", "t-123")
	if err == nil {
		t.Error("expected error when only one arg provided")
	}
}

func TestTaskDoneCreatesMessage(t *testing.T) {
	root := setupProject(t)

	_, err := executeCmd(t, "task-done", "t-abc", "agent-1", "--result", "implemented feature X")
	if err != nil {
		t.Fatalf("task-done failed: %v", err)
	}

	// Verify a message file was created in .alt/messages.
	msgDir := filepath.Join(root, ".alt", "messages")
	entries, err := os.ReadDir(msgDir)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.Contains(e.Name(), "task_done") {
			data, err := os.ReadFile(filepath.Join(msgDir, e.Name()))
			if err != nil {
				t.Fatal(err)
			}
			var msg map[string]any
			if err := json.Unmarshal(data, &msg); err != nil {
				t.Fatal(err)
			}
			if msg["from"] != "agent-1" || msg["to"] != "daemon" || msg["task_id"] != "t-abc" {
				t.Errorf("unexpected message fields: %v", msg)
			}
			payload, ok := msg["payload"].(map[string]any)
			if !ok || payload["result"] != "implemented feature X" {
				t.Errorf("expected result in payload, got: %v", msg["payload"])
			}
			found = true
			break
		}
	}
	if !found {
		t.Error("no task_done message file found in messages directory")
	}
}

func TestTaskDoneResultFlag(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "task-done" {
			f := c.Flags().Lookup("result")
			if f == nil {
				t.Error("task-done missing --result flag")
			}
			return
		}
	}
	t.Fatal("task-done command not found")
}

func TestLogTailFlag(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "log" {
			f := c.Flags().Lookup("tail")
			if f == nil {
				t.Error("log missing --tail flag")
			}
			return
		}
	}
	t.Fatal("log command not found")
}
