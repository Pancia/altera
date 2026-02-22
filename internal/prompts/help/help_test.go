package help

import (
	"strings"
	"testing"
)

func TestAgentTypes(t *testing.T) {
	types := AgentTypes()
	if len(types) != 2 {
		t.Fatalf("expected 2 agent types, got %d", len(types))
	}
	want := map[string]bool{"liaison": true, "worker": true}
	for _, at := range types {
		if !want[at] {
			t.Errorf("unexpected agent type: %s", at)
		}
	}
}

func TestTopics_Worker(t *testing.T) {
	topics, err := Topics("worker")
	if err != nil {
		t.Fatalf("Topics(worker) failed: %v", err)
	}
	expected := []string{"checkpoint", "commit", "startup", "stuck", "task-done"}
	if len(topics) != len(expected) {
		t.Fatalf("expected %d worker topics, got %d: %v", len(expected), len(topics), topics)
	}
	for i, want := range expected {
		if topics[i] != want {
			t.Errorf("topic[%d] = %q, want %q", i, topics[i], want)
		}
	}
}

func TestTopics_Liaison(t *testing.T) {
	topics, err := Topics("liaison")
	if err != nil {
		t.Fatalf("Topics(liaison) failed: %v", err)
	}
	expected := []string{"debugging", "escalation", "startup", "status", "task-create"}
	if len(topics) != len(expected) {
		t.Fatalf("expected %d liaison topics, got %d: %v", len(expected), len(topics), topics)
	}
	for i, want := range expected {
		if topics[i] != want {
			t.Errorf("topic[%d] = %q, want %q", i, topics[i], want)
		}
	}
}

func TestTopics_UnknownType(t *testing.T) {
	_, err := Topics("unknown")
	if err == nil {
		t.Fatal("expected error for unknown agent type")
	}
}

func TestLookup(t *testing.T) {
	content, err := Lookup("worker", "startup")
	if err != nil {
		t.Fatalf("Lookup(worker, startup) failed: %v", err)
	}
	if !strings.Contains(content, "Worker: Startup") {
		t.Error("expected content to contain 'Worker: Startup'")
	}
}

func TestLookup_LiaisonTopic(t *testing.T) {
	content, err := Lookup("liaison", "task-create")
	if err != nil {
		t.Fatalf("Lookup(liaison, task-create) failed: %v", err)
	}
	if !strings.Contains(content, "Creating Tasks") {
		t.Error("expected content to contain 'Creating Tasks'")
	}
}

func TestLookup_NotFound(t *testing.T) {
	_, err := Lookup("worker", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent topic")
	}
	if !strings.Contains(err.Error(), "topic not found") {
		t.Errorf("expected 'topic not found' in error, got: %v", err)
	}
}

func TestLookup_NoTopic(t *testing.T) {
	_, err := Lookup("worker")
	if err == nil {
		t.Fatal("expected error when no topic specified")
	}
}

func TestLookup_AllTopicsLoadable(t *testing.T) {
	for _, agentType := range AgentTypes() {
		topics, err := Topics(agentType)
		if err != nil {
			t.Fatalf("Topics(%s) failed: %v", agentType, err)
		}
		for _, topic := range topics {
			content, err := Lookup(agentType, topic)
			if err != nil {
				t.Errorf("Lookup(%s, %s) failed: %v", agentType, topic, err)
				continue
			}
			if len(content) == 0 {
				t.Errorf("Lookup(%s, %s) returned empty content", agentType, topic)
			}
		}
	}
}
