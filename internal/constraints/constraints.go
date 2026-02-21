// Package constraints checks system resource constraints before spawning
// new worker agents. It integrates with the events log (budget tracking),
// agent store (worker count), and merge queue directory (queue depth).
package constraints

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/anthropics/altera/internal/agent"
	"github.com/anthropics/altera/internal/config"
	"github.com/anthropics/altera/internal/events"
)

// Checker performs constraint checks against live system state.
type Checker struct {
	cfg           config.Constraints
	agents        *agent.Store
	eventsReader  *events.Reader
	mergeQueueDir string // path to .alt/merge-queue/
}

// NewChecker creates a Checker with the given dependencies.
func NewChecker(cfg config.Constraints, agents *agent.Store, evReader *events.Reader, mergeQueueDir string) *Checker {
	return &Checker{
		cfg:           cfg,
		agents:        agents,
		eventsReader:  evReader,
		mergeQueueDir: mergeQueueDir,
	}
}

// UpdateConstraints replaces the constraint configuration. This allows
// the daemon to pick up config changes without restarting.
func (c *Checker) UpdateConstraints(cfg config.Constraints) {
	c.cfg = cfg
}

// BudgetUsed sums the "token_cost" field from all events in the log.
// Events without a token_cost data field are skipped.
func (c *Checker) BudgetUsed() (float64, error) {
	all, err := c.eventsReader.ReadAll()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, fmt.Errorf("constraints: read events: %w", err)
	}
	var total float64
	for _, ev := range all {
		if cost, ok := ev.Data["token_cost"]; ok {
			if v, ok := cost.(float64); ok {
				total += v
			}
		}
	}
	return total, nil
}

// WorkerCount returns the number of active worker agents.
func (c *Checker) WorkerCount() (int, error) {
	n, err := c.agents.CountByRole(agent.RoleWorker)
	if err != nil {
		return 0, fmt.Errorf("constraints: count workers: %w", err)
	}
	return n, nil
}

// QueueDepth returns the number of items in the merge queue directory.
// Each .json file in the merge queue directory represents one queued item.
func (c *Checker) QueueDepth() (int, error) {
	entries, err := os.ReadDir(c.mergeQueueDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("constraints: read merge queue dir: %w", err)
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			n++
		}
	}
	return n, nil
}

// CheckBudget returns true if budget usage is below the ceiling.
func (c *Checker) CheckBudget() (bool, string, error) {
	used, err := c.BudgetUsed()
	if err != nil {
		return false, "", err
	}
	if used >= c.cfg.BudgetCeiling {
		return false, fmt.Sprintf("budget exhausted: %.2f/%.2f", used, c.cfg.BudgetCeiling), nil
	}
	return true, "", nil
}

// CheckMaxWorkers returns true if the worker count is below the maximum.
func (c *Checker) CheckMaxWorkers() (bool, string, error) {
	count, err := c.WorkerCount()
	if err != nil {
		return false, "", err
	}
	if count >= c.cfg.MaxWorkers {
		return false, fmt.Sprintf("max workers reached: %d/%d", count, c.cfg.MaxWorkers), nil
	}
	return true, "", nil
}

// CheckQueueDepth returns true if the merge queue depth is below the maximum.
func (c *Checker) CheckQueueDepth() (bool, string, error) {
	depth, err := c.QueueDepth()
	if err != nil {
		return false, "", err
	}
	if depth >= c.cfg.MaxQueueDepth {
		return false, fmt.Sprintf("merge queue full: %d/%d", depth, c.cfg.MaxQueueDepth), nil
	}
	return true, "", nil
}

// CanSpawnWorker performs all constraint checks and returns whether a new
// worker agent can be spawned. If not, the reason string explains why.
func (c *Checker) CanSpawnWorker() (bool, string) {
	if ok, reason, err := c.CheckBudget(); err != nil {
		return false, fmt.Sprintf("budget check error: %v", err)
	} else if !ok {
		return false, reason
	}

	if ok, reason, err := c.CheckMaxWorkers(); err != nil {
		return false, fmt.Sprintf("worker check error: %v", err)
	} else if !ok {
		return false, reason
	}

	if ok, reason, err := c.CheckQueueDepth(); err != nil {
		return false, fmt.Sprintf("queue check error: %v", err)
	} else if !ok {
		return false, reason
	}

	return true, ""
}
