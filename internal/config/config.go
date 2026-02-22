// Package config handles loading and saving Altera configuration from the
// .alt/ directory. It provides path resolution (walking up from cwd to find
// .alt/), CRUD operations on the root config, and atomic file writes to
// prevent corruption.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// DirName is the name of the Altera project directory.
const DirName = ".alt"

// Constraints holds resource limits for the orchestration system.
type Constraints struct {
	BudgetCeiling float64 `json:"budget_ceiling"`
	MaxWorkers    int     `json:"max_workers"`
	MaxQueueDepth int     `json:"max_queue_depth"`
}

// Validate checks that constraint values are within acceptable ranges.
func (c Constraints) Validate() error {
	if c.BudgetCeiling < 0 {
		return fmt.Errorf("budget_ceiling must be >= 0, got %v", c.BudgetCeiling)
	}
	if c.MaxWorkers < 1 {
		return fmt.Errorf("max_workers must be >= 1, got %d", c.MaxWorkers)
	}
	if c.MaxQueueDepth < 1 {
		return fmt.Errorf("max_queue_depth must be >= 1, got %d", c.MaxQueueDepth)
	}
	return nil
}

// Config is the root configuration stored in .alt/config.json.
type Config struct {
	RepoPath      string      `json:"repo_path"`
	DefaultBranch string      `json:"default_branch"`
	TestCommand   string      `json:"test_command"`
	Constraints   Constraints `json:"constraints"`
}

// NewConfig returns a Config with sensible defaults.
func NewConfig() Config {
	return Config{
		DefaultBranch: "main",
		Constraints: Constraints{
			BudgetCeiling: 100.0,
			MaxWorkers:    4,
			MaxQueueDepth: 10,
		},
	}
}

// FindRoot walks up from startDir looking for a DirName directory.
// Returns the path to the DirName directory (e.g. /path/to/.alt), or an error
// if none is found.
func FindRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolving absolute path: %w", err)
	}
	for {
		candidate := filepath.Join(dir, DirName)
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no %s directory found (searched up from %s)", DirName, startDir)
		}
		dir = parent
	}
}

// EnsureDir creates the .alt/ directory and standard subdirectories under
// parentDir if they don't already exist.
func EnsureDir(parentDir string) (string, error) {
	altDir := filepath.Join(parentDir, DirName)
	dirs := []string{
		altDir,
		filepath.Join(altDir, "agents"),
		filepath.Join(altDir, "tasks"),
		filepath.Join(altDir, "messages"),
		filepath.Join(altDir, "messages", "archive"),
		filepath.Join(altDir, "merge-queue"),
		filepath.Join(altDir, "worktrees"),
		filepath.Join(altDir, "logs"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return "", fmt.Errorf("creating directory %s: %w", d, err)
		}
	}
	return altDir, nil
}

// LogsDir returns the path to the logs directory within the given .alt/ dir.
func LogsDir(altDir string) string {
	return filepath.Join(altDir, "logs")
}

// DebugEnabled returns true if the debug marker file exists in the .alt/ dir.
func DebugEnabled(altDir string) bool {
	_, err := os.Stat(filepath.Join(altDir, "debug"))
	return err == nil
}

// SetDebug creates or removes the debug marker file.
func SetDebug(altDir string, enabled bool) error {
	path := filepath.Join(altDir, "debug")
	if enabled {
		return os.WriteFile(path, []byte("1\n"), 0o644)
	}
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// Load reads and parses the root config.json from the given .alt/ directory.
func Load(altDir string) (Config, error) {
	path := filepath.Join(altDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NewConfig(), nil
		}
		return Config{}, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}

// Save writes the root config.json to the given .alt/ directory using an
// atomic temp+rename pattern.
func Save(altDir string, cfg Config) error {
	path := filepath.Join(altDir, "config.json")
	return atomicWriteJSON(path, cfg)
}

// atomicWriteJSON marshals v as indented JSON and writes it atomically using
// a temp file + rename in the same directory.
func atomicWriteJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}
