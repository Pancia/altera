package prompts

import (
	"strings"
	"testing"
)

func TestWorker(t *testing.T) {
	v := Vars{
		AgentID:         "worker-05",
		TaskID:          "t-abc123",
		TaskTitle:       "Implement widgets",
		TaskDescription: "Add the widget feature to the dashboard",
		RigName:         "my-rig",
	}

	got, err := Worker(v)
	if err != nil {
		t.Fatalf("Worker: %v", err)
	}

	for _, want := range []string{
		"worker-05",
		"t-abc123",
		"Implement widgets",
		"my-rig",
		"Add the widget feature to the dashboard",
		"alt checkpoint worker-05",
		"task.json",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Worker() missing %q", want)
		}
	}
}

func TestLiaison(t *testing.T) {
	v := Vars{
		AgentID: "liaison-01",
		RigName: "test-rig",
	}

	got, err := Liaison(v)
	if err != nil {
		t.Fatalf("Liaison: %v", err)
	}

	for _, want := range []string{
		"liaison-01",
		"test-rig",
		"translator",
		".alt/tasks/",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Liaison() missing %q", want)
		}
	}
}

func TestResolver(t *testing.T) {
	v := Vars{
		AgentID:    "resolver-01",
		RigName:    "test-rig",
		BranchName: "alt/t-abc123",
	}

	got, err := Resolver(v)
	if err != nil {
		t.Fatalf("Resolver: %v", err)
	}

	for _, want := range []string{
		"resolver-01",
		"test-rig",
		"alt/t-abc123",
		"conflict",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Resolver() missing %q", want)
		}
	}
}

func TestDeveloper(t *testing.T) {
	got, err := Developer()
	if err != nil {
		t.Fatalf("Developer: %v", err)
	}

	for _, want := range []string{
		"Altera",
		"internal/",
		"make test",
		"worker",
		"liaison",
		"resolver",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Developer() missing %q", want)
		}
	}
}

func TestWorker_AllVarsRendered(t *testing.T) {
	v := Vars{
		AgentID:         "AGENT_SENTINEL",
		TaskID:          "TASK_SENTINEL",
		TaskTitle:       "TITLE_SENTINEL",
		TaskDescription: "DESC_SENTINEL",
		RigName:         "RIG_SENTINEL",
	}

	got, err := Worker(v)
	if err != nil {
		t.Fatalf("Worker: %v", err)
	}

	// Verify no unrendered template directives remain.
	if strings.Contains(got, "{{") {
		t.Error("Worker() output contains unrendered template directives")
	}

	// Verify all sentinel values appear.
	for _, sentinel := range []string{
		"AGENT_SENTINEL",
		"TASK_SENTINEL",
		"TITLE_SENTINEL",
		"DESC_SENTINEL",
		"RIG_SENTINEL",
	} {
		if !strings.Contains(got, sentinel) {
			t.Errorf("Worker() missing sentinel %q", sentinel)
		}
	}
}

func TestResolver_AllVarsRendered(t *testing.T) {
	v := Vars{
		AgentID:    "AGENT_SENTINEL",
		RigName:    "RIG_SENTINEL",
		BranchName: "BRANCH_SENTINEL",
	}

	got, err := Resolver(v)
	if err != nil {
		t.Fatalf("Resolver: %v", err)
	}

	if strings.Contains(got, "{{") {
		t.Error("Resolver() output contains unrendered template directives")
	}

	for _, sentinel := range []string{
		"AGENT_SENTINEL",
		"RIG_SENTINEL",
		"BRANCH_SENTINEL",
	} {
		if !strings.Contains(got, sentinel) {
			t.Errorf("Resolver() missing sentinel %q", sentinel)
		}
	}
}
