package cronjobworker

import (
	"time"

	"gitea.knockturnmc.com/marauder/controller/pkg/cronjob"
)

// SimpleCronjobExecutor is an implementation that has a static cooldown and a simple execution function.
type SimpleCronjobExecutor struct {
	cooldown          time.Duration
	executionFunction func(worker *CronjobWorker) error
}

func (c SimpleCronjobExecutor) Execute(parent *CronjobWorker) error {
	return c.executionFunction(parent)
}

func (c SimpleCronjobExecutor) Cooldown() time.Duration {
	return c.cooldown
}

// ComputeCronjobMap creates the cronjob map from the yaml cronjob configuration.
func ComputeCronjobMap(configuration cronjob.CronjobsConfiguration) map[cronjob.Type]CronjobExecutor {
	result := make(map[cronjob.Type]CronjobExecutor)

	if configuration.RemoveUnused != nil {
		result["removeUnused"] = RemoveUnused(configuration.RemoveUnused.Every, configuration.RemoveUnused.RemoveAfter)
	}

	return result
}
