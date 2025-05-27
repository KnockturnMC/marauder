package cronjobworker

import (
	"context"
	"fmt"
	"time"

	"github.com/knockturnmc/marauder/marauder-controller/internal/db/access"
)

// RemoveHistoric constructs the cronjob executor that removes historic state older than the passed duration.
func RemoveHistoric(cooldown time.Duration, removeAfter time.Duration) CronjobExecutor {
	return SimpleCronjobExecutor{
		cooldown: cooldown,
		executionFunction: func(ctx context.Context, worker *CronjobWorker) error {
			toRemove, err := access.FindHistoricStateOlderThan(ctx, worker.DB, time.Now().UTC().Add(-removeAfter))
			if err != nil {
				return fmt.Errorf("failed to fetch artefacts to remove: %w", err)
			}

			for _, model := range toRemove {
				if err := access.DeleteHistoricServerState(ctx, worker.DB, model.UUID); err != nil {
					return fmt.Errorf("failed to delete server state %s: %w", model.UUID, err)
				}
			}

			return nil
		},
	}
}
