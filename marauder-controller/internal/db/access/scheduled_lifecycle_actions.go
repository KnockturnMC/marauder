package access

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/knockturnmc/marauder/marauder-controller/sqlm"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
)

// InsertOrMergeScheduledLifecycleAction inserts the passed scheduled lifecycle action into the database.
// The created lifecycle action model must not contain a full fetched server model.
func InsertOrMergeScheduledLifecycleAction(
	ctx context.Context,
	db *sqlm.DB,
	action networkmodel.ScheduledLifecycleAction,
) (networkmodel.ScheduledLifecycleAction, error) {
	if err := db.NamedGetContext(
		ctx,
		&action,
		`SELECT * FROM func_insert_or_join_existing_scheduled_lifecycle_action(:server,:action,:time_of_execution);`,
		action,
	); err != nil {
		return networkmodel.ScheduledLifecycleAction{}, fmt.Errorf("failed to insert scheduled lifecycle action: %w", err)
	}

	return action, nil
}

// FetchScheduledLifecycleActionsToBeExecutedAfter fetches all lifeccycle actions that exist and are scheduled to be executed
// after the passed time.Time.
func FetchScheduledLifecycleActionsToBeExecutedAfter(
	ctx context.Context,
	db *sqlm.DB,
	timeOfExecution time.Time,
) ([]networkmodel.ScheduledLifecycleAction, error) {
	result := make([]networkmodel.ScheduledLifecycleAction, 0)
	if err := db.SelectContext(ctx, &result, `
		SELECT * FROM scheduled_lifecycle_actions WHERE time_of_execution < $1
		`, timeOfExecution); err != nil {
		return nil, fmt.Errorf("failed to fetch executable scheduled lifecycle actions: %w", err)
	}

	return fillScheduledLifecycleModelRefsSlice(ctx, db, result)
}

// DeleteScheduledLifecycleAction deletes the passed scheduled lifecycle action by its uuid.
func DeleteScheduledLifecycleAction(
	ctx context.Context,
	db *sqlm.DB,
	uuid uuid.UUID,
) error {
	if _, err := db.ExecContext(ctx, "DELETE FROM scheduled_lifecycle_actions WHERE uuid = $1", uuid); err != nil {
		return fmt.Errorf("failed to delete scheduled lifecycle action: %w", err)
	}

	return nil
}
