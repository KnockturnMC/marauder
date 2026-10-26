package cronjobworker

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/knockturnmc/marauder/marauder-controller/internal/db/access"
	"github.com/knockturnmc/marauder/marauder-controller/pkg/cronjob"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
)

// AutoUpdateOutdatedServers constructs the cronjob executor that automatically updates servers with outdated artefacts.
func AutoUpdateOutdatedServers(configuration cronjob.AutoUpdateOutdatedServers) CronjobExecutor {
	return SimpleCronjobExecutor{
		cooldown: configuration.Every,
		executionFunction: func(ctx context.Context, worker *CronjobWorker) error {
			servers, err := access.FetchServers(ctx, worker.DB)
			if err != nil {
				return fmt.Errorf("failed to fetch server list: %w", err)
			}

			// Remove servers not matching the regex.
			servers = slices.DeleteFunc(servers, func(model networkmodel.ServerModel) bool {
				return !configuration.ServerRegex.MatchString(model.FullyQualifiedName())
			})

			for _, server := range servers {
				missMatch, err := access.FindServerTargetStateMissMatch(ctx, worker.DB, server.UUID, true)
				if err != nil {
					return fmt.Errorf("failed to fetch server missmatch for %s: %w", server.FullyQualifiedName(), err)
				}

				if len(missMatch) == 0 {
					continue
				}

				operatorClient := worker.OperatorClientCache.GetOrCreateFromRef(server.OperatorRef)
				response, err := operatorClient.ManagementGetPlayers(
					ctx, server.UUID,
				)
				if err != nil {
					logger.Warn("Failed to query player count of", server.FullyQualifiedName(), "for automated upgrade ", err)
					continue
				}
				if len(response.Players) != 0 {
					continue
				}
				if time.Since(response.LastPlayerLogoutTimestamp) < configuration.MinimumTimeWithoutPlayers {
					continue
				}

				var action networkmodel.LifecycleAction
				if configuration.Force {
					action = networkmodel.ForceUpdateWithRestart
				} else {
					action = networkmodel.UpdateWithRestart
				}

				if err := operatorClient.ExecuteLifecycleAction(ctx, server.UUID, action); err != nil {
					return fmt.Errorf("failed to update server %s: %w", server.FullyQualifiedName(), err)
				}
			}

			return nil
		},
	}
}
