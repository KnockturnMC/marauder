package cronjobworker

import (
	"context"
	"time"

	"github.com/knockturnmc/marauder/marauder-controller/pkg/cronjob"
)

// SimpleCronjobExecutor is an implementation that has a static cooldown and a simple execution function.
type SimpleCronjobExecutor struct {
	cooldown          time.Duration
	executionFunction func(ctx context.Context, worker *CronjobWorker) error
}

func (c SimpleCronjobExecutor) Execute(ctx context.Context, parent *CronjobWorker) error {
	return c.executionFunction(ctx, parent)
}

func (c SimpleCronjobExecutor) Cooldown() time.Duration {
	return c.cooldown
}

// ComputeCronjobMap creates the cronjob map from the yaml cronjob configuration.
func ComputeCronjobMap(configuration cronjob.CronjobsConfiguration) map[cronjob.Type]CronjobExecutor {
	result := make(map[cronjob.Type]CronjobExecutor)

	if configuration.RemoveUnused != nil {
		result[cronjob.RemoveUnusedIdentifier] = RemoveUnused(configuration.RemoveUnused.Every, configuration.RemoveUnused.RemoveAfter)
	}
	if configuration.RemoveHistoric != nil {
		result[cronjob.RemoveHistoricIdentifier] = RemoveHistoric(configuration.RemoveHistoric.Every, configuration.RemoveHistoric.RemoveAfter)
	}
	if configuration.ClearOperatorCaches != nil {
		result[cronjob.ClearOperatorCacheIdentifier] = ClearOperatorCaches(
			configuration.ClearOperatorCaches.Every,
			configuration.ClearOperatorCaches.RemoveAfter,
		)
	}
	if configuration.ExecuteScheduledLifecycleActions != nil {
		result[cronjob.ExecuteScheduledLifecycleActionsIdentifier] = ExecuteScheduledLifecycleActions(
			configuration.ExecuteScheduledLifecycleActions.Every,
		)
	}

	return result
}
