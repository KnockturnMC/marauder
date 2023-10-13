package cronjobworker

import (
	"context"
	"fmt"
	"time"

	"gitea.knockturnmc.com/marauder/controller/internal/db/access"
)

// ExecuteScheduledLifecycleActions is responsible for executing lifecycle actions scheduled against an operator and a respective server.
// This cron job is a lot more dynamic in its nature as cooldown might be dynamically adjusted for e.g. batched updates by the cicd to the
// integration environment.
func ExecuteScheduledLifecycleActions(cooldown time.Duration) CronjobExecutor {
	return SimpleCronjobExecutor{
		cooldown: cooldown,
		executionFunction: func(ctx context.Context, worker *CronjobWorker) error {
			actions, err := access.FetchScheduledLifecycleActionsToBeExecutedAfter(ctx, worker.DB, time.Now())
			if err != nil {
				return fmt.Errorf("failed to fetch scheduled lifecycle actions: %w", err)
			}

			for _, action := range actions {
				operatorClient := worker.OperatorClientCache.GetOrCreateFromRef(action.Server.OperatorRef)
				if err := operatorClient.ExecuteLifecycleAction(
					ctx,
					action.ServerUUID,
					action.LifecycleAction,
				); err != nil {
					return fmt.Errorf("failed to execute scheduled lifecycle action %s on %s: %w", action.LifecycleAction, action.ServerUUID, err)
				}

				if err := access.DeleteScheduledLifecycleAction(ctx, worker.DB, action.UUID); err != nil {
					return fmt.Errorf("failed to delete executed scheduled action %s from db: %w", action.UUID, err)
				}
			}

			return nil
		},
	}
}
