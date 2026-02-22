package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	if cfg.DefaultBranch != "main" {
		t.Fatalf("expected default branch 'main', got %q", cfg.DefaultBranch)
	}
	if cfg.Constraints.BudgetCeiling != 100.0 {
		t.Fatalf("expected budget ceiling 100, got %f", cfg.Constraints.BudgetCeiling)
	}
	if cfg.Constraints.MaxWorkers != 4 {
		t.Fatalf("expected max workers 4, got %d", cfg.Constraints.MaxWorkers)
	}
	if cfg.Constraints.MaxQueueDepth != 10 {
		t.Fatalf("expected max queue depth 10, got %d", cfg.Constraints.MaxQueueDepth)
	}
}

func TestFindRoot(t *testing.T) {
	// Create a temp directory structure: tmp/a/b/c with .alt at tmp/a/
	tmp := t.TempDir()
	altDir := filepath.Join(tmp, "a", DirName)
	deepDir := filepath.Join(tmp, "a", "b", "c")
	if err := os.MkdirAll(altDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(deepDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Should find .alt from deep subdirectory
	found, err := FindRoot(deepDir)
	if err != nil {
		t.Fatalf("FindRoot: %v", err)
	}
	if found != altDir {
		t.Fatalf("expected %s, got %s", altDir, found)
	}

	// Should find .alt from the directory containing it
	found, err = FindRoot(filepath.Join(tmp, "a"))
	if err != nil {
		t.Fatalf("FindRoot from parent: %v", err)
	}
	if found != altDir {
		t.Fatalf("expected %s, got %s", altDir, found)
	}
}

func TestFindRootNotFound(t *testing.T) {
	tmp := t.TempDir()
	_, err := FindRoot(tmp)
	if err == nil {
		t.Fatal("expected error when no .alt directory exists")
	}
}

func TestEnsureDir(t *testing.T) {
	tmp := t.TempDir()
	altDir, err := EnsureDir(tmp)
	if err != nil {
		t.Fatalf("EnsureDir: %v", err)
	}
	expected := filepath.Join(tmp, DirName)
	if altDir != expected {
		t.Fatalf("expected %s, got %s", expected, altDir)
	}

	// Verify subdirectories exist
	info, err := os.Stat(filepath.Join(altDir, "agents"))
	if err != nil {
		t.Fatalf("agents dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("agents should be a directory")
	}

	// Calling again should not fail (idempotent)
	_, err = EnsureDir(tmp)
	if err != nil {
		t.Fatalf("EnsureDir idempotent: %v", err)
	}
}

func TestLoadSaveConfig(t *testing.T) {
	tmp := t.TempDir()
	altDir, err := EnsureDir(tmp)
	if err != nil {
		t.Fatal(err)
	}

	// Load from missing file returns defaults
	cfg, err := Load(altDir)
	if err != nil {
		t.Fatalf("Load default: %v", err)
	}
	if cfg.Constraints.MaxWorkers != 4 {
		t.Fatal("expected default config")
	}

	// Modify and save
	cfg.Constraints.BudgetCeiling = 250.0
	cfg.Constraints.MaxWorkers = 8
	cfg.RepoPath = "/repos/alpha"
	cfg.DefaultBranch = "develop"
	cfg.TestCommand = "go test ./..."
	if err := Save(altDir, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Load back
	loaded, err := Load(altDir)
	if err != nil {
		t.Fatalf("Load after save: %v", err)
	}
	if loaded.Constraints.BudgetCeiling != 250.0 {
		t.Fatalf("expected 250, got %f", loaded.Constraints.BudgetCeiling)
	}
	if loaded.Constraints.MaxWorkers != 8 {
		t.Fatalf("expected 8, got %d", loaded.Constraints.MaxWorkers)
	}
	if loaded.RepoPath != "/repos/alpha" {
		t.Fatalf("expected /repos/alpha, got %s", loaded.RepoPath)
	}
	if loaded.DefaultBranch != "develop" {
		t.Fatalf("expected develop, got %s", loaded.DefaultBranch)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	altDir, _ := EnsureDir(tmp)
	path := filepath.Join(altDir, "config.json")
	if err := os.WriteFile(path, []byte("{invalid"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(altDir)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestAtomicWrite(t *testing.T) {
	// Verify that save doesn't leave partial files on marshal success
	tmp := t.TempDir()
	altDir, _ := EnsureDir(tmp)

	cfg := NewConfig()
	cfg.RepoPath = "/test"
	if err := Save(altDir, cfg); err != nil {
		t.Fatal(err)
	}

	// Check no temp files left behind
	entries, err := os.ReadDir(altDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		// Skip directories created by EnsureDir and the config file.
		if e.IsDir() || e.Name() == "config.json" {
			continue
		}
		// Any non-directory file without an extension that isn't config.json is suspicious.
		if filepath.Ext(e.Name()) == "" {
			t.Fatalf("unexpected file left behind: %s", e.Name())
		}
	}
}

func TestConstraintsValidate_Valid(t *testing.T) {
	c := Constraints{BudgetCeiling: 100, MaxWorkers: 4, MaxQueueDepth: 10}
	if err := c.Validate(); err != nil {
		t.Fatalf("Validate: unexpected error: %v", err)
	}
}

func TestConstraintsValidate_ZeroBudget(t *testing.T) {
	c := Constraints{BudgetCeiling: 0, MaxWorkers: 4, MaxQueueDepth: 10}
	if err := c.Validate(); err != nil {
		t.Fatalf("Validate: zero budget should be valid: %v", err)
	}
}

func TestConstraintsValidate_NegativeBudget(t *testing.T) {
	c := Constraints{BudgetCeiling: -1, MaxWorkers: 4, MaxQueueDepth: 10}
	if err := c.Validate(); err == nil {
		t.Fatal("Validate: expected error for negative budget")
	}
}

func TestConstraintsValidate_ZeroWorkers(t *testing.T) {
	c := Constraints{BudgetCeiling: 100, MaxWorkers: 0, MaxQueueDepth: 10}
	if err := c.Validate(); err == nil {
		t.Fatal("Validate: expected error for zero workers")
	}
}

func TestConstraintsValidate_ZeroQueueDepth(t *testing.T) {
	c := Constraints{BudgetCeiling: 100, MaxWorkers: 4, MaxQueueDepth: 0}
	if err := c.Validate(); err == nil {
		t.Fatal("Validate: expected error for zero queue depth")
	}
}

func TestLoadMinimalJSON(t *testing.T) {
	// Verify that a minimal config can be loaded without error.
	tmp := t.TempDir()
	altDir, _ := EnsureDir(tmp)
	path := filepath.Join(altDir, "config.json")
	if err := os.WriteFile(path, []byte(`{"constraints":{}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(altDir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultBranch != "" {
		t.Fatalf("expected empty default_branch from minimal JSON, got %q", cfg.DefaultBranch)
	}
}
