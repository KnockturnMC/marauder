package cronjob

import (
	"time"
)

const (
	RemoveUnusedIdentifier                     Type = "removeUnused"
	RemoveHistoricIdentifier                   Type = "removeHistoric"
	ClearOperatorCacheIdentifier               Type = "clearOperatorCaches"
	ExecuteScheduledLifecycleActionsIdentifier Type = "executeScheduledLifecycleActions"
)

// Type is a specific cronjob type runnable by marauder.
type Type string

// Execution creates a new execution record for the specific cronjob type that expects to be re-run next
// at the given time.
func (c Type) Execution(nextExecution time.Time) Execution {
	return Execution{
		NextExecution: nextExecution,
		Type:          c,
	}
}

// The CronjobsConfiguration defines the available configurations for the known cronjobs.
type CronjobsConfiguration struct {
	RemoveUnused                     *RemoveUnused                     `yaml:"removeUnused,omitempty"`
	RemoveHistoric                   *RemoveHistoric                   `yaml:"removeHistoric,omitempty"`
	ClearOperatorCaches              *ClearOperatorCaches              `yaml:"clearOperatorCaches,omitempty"`
	ExecuteScheduledLifecycleActions *ExecuteScheduledLifecycleActions `yaml:"executeScheduledLifecycleActions"`
}

// BaseCronjobConfiguration defines a base struct for all cronjobs configurations.
type BaseCronjobConfiguration struct {
	Every time.Duration `yaml:"every"`
}

// RemoveUnused defines the configuration for the cronjob remove unused.
type RemoveUnused struct {
	BaseCronjobConfiguration `yaml:",inline"`
	RemoveAfter              time.Duration `yaml:"removeAfter"`
}

// RemoveHistoric defines the configuration for the cronjob to remove historic server state.
type RemoveHistoric struct {
	BaseCronjobConfiguration `yaml:",inline"`
	RemoveAfter              time.Duration `yaml:"removeAfter"`
}

// ClearOperatorCaches defines the configuration for the cronjob to clean operator caches.
type ClearOperatorCaches struct {
	BaseCronjobConfiguration `yaml:",inline"`
	RemoveAfter              time.Duration `yaml:"removeAfter"`
}

// ExecuteScheduledLifecycleActions holds the configuration for the cronjob that is responsible for executing scheduled lifecycle actions.
type ExecuteScheduledLifecycleActions struct {
	BaseCronjobConfiguration `yaml:",inline"`
}

// Execution represents a cronjob the controller should execute on a regular basis.
type Execution struct {
	NextExecution time.Time `db:"next_execution"`
	Type          Type      `db:"type"`
}
