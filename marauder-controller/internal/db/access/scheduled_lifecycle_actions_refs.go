package access

import (
	"context"
	"fmt"

	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
)

// fillServerModelNetwork fetches the network configuration of a given server model from the database.
func fillScheduledLifecycleActionModel(
	ctx context.Context,
	db *sqlm.DB,
	model networkmodel.ScheduledLifecycleAction,
) (networkmodel.ScheduledLifecycleAction, error) {
	var err error

	if model, err = fillScheduledLifecycleActionServer(ctx, db, model); err != nil {
		return networkmodel.ScheduledLifecycleAction{}, err
	}

	return model, nil
}

// fillServerModelNetworkSlice executes fillServerModelNetwork for each server in the slice.
func fillScheduledLifecycleModelRefsSlice(
	ctx context.Context,
	db *sqlm.DB,
	scheduledLifecycleActions []networkmodel.ScheduledLifecycleAction,
) ([]networkmodel.ScheduledLifecycleAction, error) {
	for i, el := range scheduledLifecycleActions {
		fullyFetched, err := fillScheduledLifecycleActionModel(ctx, db, el)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch refs for %s: %w", el.UUID.String(), err)
		}
		scheduledLifecycleActions[i] = fullyFetched
	}

	return scheduledLifecycleActions, nil
}

// fillServerModelNetwork fetches the server model the scheduled lifecycle action references.
func fillScheduledLifecycleActionServer(
	ctx context.Context,
	db *sqlm.DB,
	model networkmodel.ScheduledLifecycleAction,
) (networkmodel.ScheduledLifecycleAction, error) {
	fetchServer, err := FetchServer(ctx, db, model.ServerUUID)
	if err != nil {
		return networkmodel.ScheduledLifecycleAction{}, fmt.Errorf("failed to fetch server: %w", err)
	}

	model.Server = fetchServer

	return model, nil
}
