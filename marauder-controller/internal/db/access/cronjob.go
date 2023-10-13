package access

import (
	"context"
	"fmt"
	"time"

	"gitea.knockturnmc.com/marauder/controller/pkg/cronjob"
	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
)

// FetchLastCronjobExecutions fetches all known jobs from the database.
func FetchLastCronjobExecutions(ctx context.Context, db *sqlm.DB) ([]cronjob.Execution, error) {
	result := make([]cronjob.Execution, 0)
	if err := db.SelectContext(ctx, &result, `
    SELECT * FROM cronjob;
    `); err != nil {
		return result, fmt.Errorf("failed to fetch job executions: %w", err)
	}

	return result, nil
}

// UpsertCronjobExecution upserts a new cronjob execution into the database as the cronjob has been executed/is rescheduled.
func UpsertCronjobExecution(ctx context.Context, db *sqlm.DB, execution cronjob.Execution) error {
	if _, err := db.NamedExecContext(ctx, `
		INSERT INTO cronjob(type, next_execution) VALUES (:type, :next_execution)
		ON CONFLICT (type) DO UPDATE SET next_execution = excluded.next_execution;
	`, execution); err != nil {
		return fmt.Errorf("failed to upsert cronjob execution: %w", err)
	}

	return nil
}

// FindHistoricArtefactsOlderThan yields all artefacts that are older than the passed date and are not
// currently used as a TARGET or IS state.
func FindHistoricArtefactsOlderThan(ctx context.Context, db *sqlm.DB, timestamp time.Time) ([]networkmodel.ArtefactModel, error) {
	result := make([]networkmodel.ArtefactModel, 0)
	if err := db.SelectContext(ctx, &result, `
		SELECT * FROM func_find_historic_artefacts_older_than($1)
		`, timestamp.UTC()); err != nil {
		return nil, fmt.Errorf("failed to query db: %w", err)
	}

	return result, nil
}

// FindHistoricStateOlderThan yields all server state that are older than the passed date and are HISTORIC.
func FindHistoricStateOlderThan(ctx context.Context, db *sqlm.DB, timestamp time.Time) ([]networkmodel.ServerArtefactStateModel, error) {
	result := make([]networkmodel.ServerArtefactStateModel, 0)
	if err := db.SelectContext(ctx, &result, `
		SELECT * FROM func_find_historic_state_older_than($1)
		`, timestamp.UTC()); err != nil {
		return nil, fmt.Errorf("failed to query db: %w", err)
	}

	return result, nil
}
