package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	if cfg.Rigs == nil {
		t.Fatal("Rigs map should be initialized")
	}
	if len(cfg.Rigs) != 0 {
		t.Fatalf("expected 0 rigs, got %d", len(cfg.Rigs))
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
	info, err := os.Stat(filepath.Join(altDir, "rigs"))
	if err != nil {
		t.Fatalf("rigs dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("rigs should be a directory")
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
	cfg.Rigs["alpha"] = RigConfig{
		RepoPath:      "/repos/alpha",
		DefaultBranch: "main",
		TestCommand:   "go test ./...",
	}
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
	rig, ok := loaded.Rigs["alpha"]
	if !ok {
		t.Fatal("expected rig 'alpha'")
	}
	if rig.RepoPath != "/repos/alpha" {
		t.Fatalf("expected /repos/alpha, got %s", rig.RepoPath)
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

func TestRigCRUD(t *testing.T) {
	tmp := t.TempDir()
	altDir, err := EnsureDir(tmp)
	if err != nil {
		t.Fatal(err)
	}

	// No rigs initially
	names, err := ListRigs(altDir)
	if err != nil {
		t.Fatalf("ListRigs empty: %v", err)
	}
	if len(names) != 0 {
		t.Fatalf("expected 0 rigs, got %d", len(names))
	}

	// Create a rig
	rc := RigConfig{
		RepoPath:      "/repos/beta",
		DefaultBranch: "develop",
		TestCommand:   "make test",
	}
	if err := SaveRig(altDir, "beta", rc); err != nil {
		t.Fatalf("SaveRig: %v", err)
	}

	// Read it back
	loaded, err := LoadRig(altDir, "beta")
	if err != nil {
		t.Fatalf("LoadRig: %v", err)
	}
	if loaded.RepoPath != "/repos/beta" {
		t.Fatalf("expected /repos/beta, got %s", loaded.RepoPath)
	}
	if loaded.DefaultBranch != "develop" {
		t.Fatalf("expected develop, got %s", loaded.DefaultBranch)
	}
	if loaded.TestCommand != "make test" {
		t.Fatalf("expected make test, got %s", loaded.TestCommand)
	}

	// Update it
	rc.TestCommand = "go test ./..."
	if err := SaveRig(altDir, "beta", rc); err != nil {
		t.Fatalf("SaveRig update: %v", err)
	}
	loaded, err = LoadRig(altDir, "beta")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.TestCommand != "go test ./..." {
		t.Fatalf("expected updated test command, got %s", loaded.TestCommand)
	}

	// Add another rig and list
	if err := SaveRig(altDir, "gamma", RigConfig{RepoPath: "/repos/gamma"}); err != nil {
		t.Fatal(err)
	}
	names, err = ListRigs(altDir)
	if err != nil {
		t.Fatalf("ListRigs: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 rigs, got %d", len(names))
	}

	// Delete a rig
	if err := DeleteRig(altDir, "beta"); err != nil {
		t.Fatalf("DeleteRig: %v", err)
	}
	_, err = LoadRig(altDir, "beta")
	if err == nil {
		t.Fatal("expected error loading deleted rig")
	}
	names, err = ListRigs(altDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 {
		t.Fatalf("expected 1 rig after delete, got %d", len(names))
	}
}

func TestLoadRigNotFound(t *testing.T) {
	tmp := t.TempDir()
	altDir, _ := EnsureDir(tmp)
	_, err := LoadRig(altDir, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent rig")
	}
}

func TestDeleteRigNonexistent(t *testing.T) {
	tmp := t.TempDir()
	altDir, _ := EnsureDir(tmp)
	// Deleting a nonexistent rig should not error (RemoveAll on missing path is fine)
	if err := DeleteRig(altDir, "nonexistent"); err != nil {
		t.Fatalf("DeleteRig nonexistent: %v", err)
	}
}

func TestAtomicWrite(t *testing.T) {
	// Verify that save doesn't leave partial files on marshal success
	tmp := t.TempDir()
	altDir, _ := EnsureDir(tmp)

	cfg := NewConfig()
	cfg.Rigs["test"] = RigConfig{RepoPath: "/test"}
	if err := Save(altDir, cfg); err != nil {
		t.Fatal(err)
	}

	// Check no temp files left behind
	entries, err := os.ReadDir(altDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == "" && e.Name() != "config.json" && e.Name() != "rigs" {
			t.Fatalf("unexpected file left behind: %s", e.Name())
		}
	}
}

func TestLoadNilRigsMap(t *testing.T) {
	// Verify that a config with null rigs field gets initialized
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
	if cfg.Rigs == nil {
		t.Fatal("Rigs should be initialized even if missing from JSON")
	}
}
