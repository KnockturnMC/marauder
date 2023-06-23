package access

import (
	"context"
	"fmt"

	"gitea.knockturnmc.com/marauder/controller/internal/db/models"
	"gitea.knockturnmc.com/marauder/controller/sqlm"
	"github.com/google/uuid"
)

// FetchServerState fetches a specific server state based on its unique uuid.
func FetchServerState(ctx context.Context, db *sqlm.DB, uuid uuid.UUID) (models.ServerArtefactStateModel, error) {
	var result models.ServerArtefactStateModel
	if err := db.GetContext(ctx, &result, `
            SELECT uuid, server, artefact, definition_date, type FROM server_state WHERE uuid = $1
            `, uuid); err != nil {
		return models.ServerArtefactStateModel{}, fmt.Errorf("failed to fetch server state from db: %w", err)
	}

	return result, nil
}

// FetchServerTargetState fetches the target state for a specific server.
func FetchServerTargetState(ctx context.Context, db *sqlm.DB, serverUUID uuid.UUID) (models.ServerArtefactStateModel, error) {
	return fetchServerStateSingleRow(ctx, db, serverUUID, models.TARGET)
}

// FetchServerIsState fetches the is state for a specific server.
func FetchServerIsState(ctx context.Context, db *sqlm.DB, serverUUID uuid.UUID) (models.ServerArtefactStateModel, error) {
	return fetchServerStateSingleRow(ctx, db, serverUUID, models.IS)
}

// fetchServerStateSingleRow fetches a specific, single row state for a given server.
// HISTORY is not supported as only the most recent history would be returned.
func fetchServerStateSingleRow(
	ctx context.Context,
	db *sqlm.DB,
	serverUUID uuid.UUID,
	state models.ServerStateType,
) (models.ServerArtefactStateModel, error) {
	var result models.ServerArtefactStateModel
	if err := db.GetContext(ctx, &result, `
        SELECT * FROM server_state WHERE type = $1 AND server = $2
        `, state, serverUUID); err != nil {
		return models.ServerArtefactStateModel{}, fmt.Errorf("failed to find %d state for %s: %w", state, serverUUID.String(), err)
	}

	return result, nil
}

// InsertServerState inserts a new server state into the database.
func InsertServerState(ctx context.Context, db *sqlm.DB, state models.ServerArtefactStateModel) (models.ServerArtefactStateModel, error) {
	if err := db.NamedGetContext(ctx, &state, `
            INSERT INTO server_state (server, artefact, definition_date, type) 
            VALUES (:server, :artefact, :definition_date, :type)
            RETURNING *; 
            `, state); err != nil {
		return models.ServerArtefactStateModel{}, fmt.Errorf("failed to insert server state: %w", err)
	}

	return state, nil
}
