package cronjob

import "time"

// Type represents the available cronjob types.
type Type string

// The CronjobsConfiguration defines the available configurations for the known cronjobs.
type CronjobsConfiguration struct {
	RemoveUnused   *RemoveUnused   `yaml:"removeUnused,omitempty"`
	RemoveHistoric *RemoveHistoric `yaml:"removeHistoric,omitempty"`
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

// Execution represents a cronjob the controller should execute on a regular basis.
type Execution struct {
	NextExecution time.Time `db:"next_execution"`
	Type          Type      `db:"type"`
}
