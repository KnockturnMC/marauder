package access

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"gitea.knockturnmc.com/marauder/controller/sqlm"

	"gitea.knockturnmc.com/marauder/controller/internal/db/models"
)

// FetchArtefact locates a specific artefact based on its identifier and version in the database.
func FetchArtefact(ctx context.Context, db *sqlm.DB, identifier string, version string) (models.ArtefactModel, error) {
	var result models.ArtefactModel
	if err := db.GetContext(ctx, &result, `
    SELECT * FROM artefact WHERE identifier = $1 AND version = $2
    `, identifier, version); err != nil {
		return models.ArtefactModel{}, fmt.Errorf("failed to find artefact: %w", err)
	}

	return result, nil
}

// FetchArtefactVersions queries the database for all currently hosted versions of the artefact.
func FetchArtefactVersions(ctx context.Context, db *sqlm.DB, identifier string) ([]models.ArtefactModel, error) {
	result := make([]models.ArtefactModel, 0)
	if err := db.SelectContext(ctx, &result, `
    SELECT uuid, identifier, version, upload_date FROM artefact WHERE identifier = $1
    `, identifier); err != nil {
		return result, fmt.Errorf("failed to find artefact: %w", err)
	}

	return result, nil
}

// InsertArtefact inserts an artefact model into the database.
func InsertArtefact(ctx context.Context, db *sqlm.DB, model models.ArtefactModelWithBinary) (models.ArtefactModel, error) {
	transaction, err := db.Beginx()
	if err != nil {
		return models.ArtefactModel{}, fmt.Errorf("failed to begin insertion transaction: %w", err)
	}

	var result models.ArtefactModel
	if err := transaction.NamedGetContext(ctx, &result, `
        INSERT INTO artefact (identifier, version, upload_date)
        VALUES (:identifier, :version, :upload_date)
        RETURNING *;`,
		&model,
	); err != nil {
		return models.ArtefactModel{}, fmt.Errorf("failed to insert artefact: %w", err)
	}

	if _, err := transaction.ExecContext(ctx, `
        INSERT INTO artefact_file (artefact, tarball) VALUES ($1, $2);`,
		result.UUID, model.TarballBlob,
	); err != nil {
		return models.ArtefactModel{}, fmt.Errorf("failed to insert tarball into database for %s: %w", result.UUID, err)
	}

	if err := transaction.Commit(); err != nil {
		return models.ArtefactModel{}, fmt.Errorf("failed to commit insertion transaction: %w", err)
	}

	return result, nil
}

// FetchArtefactTarball fetches a full ArtefactModelWithBinary from the database including the full tarball binary.
func FetchArtefactTarball(ctx context.Context, db *sqlm.DB, uuid uuid.UUID) (models.ArtefactModelWithBinary, error) {
	var result models.ArtefactModelWithBinary
	if err := db.GetContext(ctx, &result, `
        SELECT uuid, identifier, version, upload_date, tarball FROM artefact JOIN artefact_file af on artefact.uuid = af.artefact WHERE af.artefact = $1
        `, uuid); err != nil {
		return models.ArtefactModelWithBinary{}, fmt.Errorf("failed to fetch artefact model with binary from database: %w", err)
	}

	return result, nil
}
