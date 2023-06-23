package access

import (
	"context"
	"fmt"

	"gitea.knockturnmc.com/marauder/controller/sqlm"

	"gitea.knockturnmc.com/marauder/controller/internal/db/models"
)

// FindArtefact locates a specific artefact based on its identifier and version in the database.
func FindArtefact(ctx context.Context, db *sqlm.DB, identifier string, version string) (models.ArtefactModel, error) {
	var result models.ArtefactModel
	if err := db.GetContext(ctx, &result, "SELECT * FROM artefact WHERE identifier = $1 AND version = $2", identifier, version); err != nil {
		return models.ArtefactModel{}, fmt.Errorf("failed to find artefact: %w", err)
	}

	return result, nil
}

// QueryArtefactVersions queries the database for all currently hosted versions of the artefact.
func QueryArtefactVersions(ctx context.Context, db *sqlm.DB, identifier string) ([]models.ArtefactModel, error) {
	result := make([]models.ArtefactModel, 0)
	if err := db.SelectContext(ctx, &result, "SELECT * FROM artefact WHERE identifier = $1", identifier); err != nil {
		return result, fmt.Errorf("failed to find artefact: %w", err)
	}

	return result, nil
}

// InsertArtefact inserts an artefact model into the database.
func InsertArtefact(ctx context.Context, db *sqlm.DB, model models.ArtefactModel) (models.ArtefactModel, error) {
	if err := db.NamedGetContext(ctx, &model, `
        INSERT INTO artefact (identifier, version, upload_date, storage_path)
        VALUES (:identifier, :version, :upload_date, :storage_path)
        RETURNING *;`,
		model,
	); err != nil {
		return models.ArtefactModel{}, fmt.Errorf("failed to insert artefact: %w", err)
	}

	return model, nil
}
