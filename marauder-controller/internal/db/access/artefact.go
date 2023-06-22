package access

import (
	"context"
	"fmt"
	"gitea.knockturnmc.com/marauder/controller/internal/db/models"
	"github.com/jmoiron/sqlx"
)

// FindArtefact locates a specific artefact based on its identifier and version in the database.
func FindArtefact(db *sqlx.DB, ctx context.Context, identifier string, version string) (models.ArtefactModel, error) {
	var result models.ArtefactModel
	if err := db.GetContext(ctx, &result, "SELECT * FROM artefact WHERE identifier = $1 AND version = $2", identifier, version); err != nil {
		return models.ArtefactModel{}, fmt.Errorf("failed to find artefact: %w", err)
	}

	return result, nil
}

// QueryArtefactVersions queries the database for all currently hosted versions of the artefact.
func QueryArtefactVersions(db *sqlx.DB, ctx context.Context, identifier string) ([]models.ArtefactModel, error) {
	var result []models.ArtefactModel
	if err := db.SelectContext(ctx, &result, "SELECT * FROM artefact WHERE identifier = $1", identifier); err != nil {
		return result, fmt.Errorf("failed to find artefact: %w", err)
	}

	return result, nil
}

// InsertArtefact inserts an artefact model into the database.
func InsertArtefact(db *sqlx.DB, ctx context.Context, model models.ArtefactModel) error {
	db.PreparexContext(ctx, "INSERT INTO artefact ()")
}
