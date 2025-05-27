package cronjobworker

import (
	"context"
	"fmt"
	"time"

	"github.com/knockturnmc/marauder/marauder-controller/internal/db/access"
)

// ClearOperatorCaches constructs the cronjob executor that clears the operator caches.
func ClearOperatorCaches(cooldown time.Duration, removeAfter time.Duration) CronjobExecutor {
	return SimpleCronjobExecutor{
		cooldown: cooldown,
		executionFunction: func(ctx context.Context, worker *CronjobWorker) error {
			operators, err := access.FetchOperators(ctx, worker.DB)
			if err != nil {
				return fmt.Errorf("failed to fetch known operators: %w", err)
			}

			for _, operator := range operators {
				operatorClient := worker.OperatorClientCache.GetOrCreate(operator.Identifier, operator.Host, operator.Port)
				if err := operatorClient.ScheduleCacheClear(ctx, removeAfter); err != nil {
					return fmt.Errorf("failed to clean cache on %s: %w", operator.Identifier, err)
				}
			}

			return nil
		},
	}
}
