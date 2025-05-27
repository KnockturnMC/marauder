package access

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/knockturnmc/marauder/marauder-controller/sqlm"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
)

// FetchArtefact locates a specific artefact based on its identifier and version in the database.
func FetchArtefact(ctx context.Context, db *sqlm.DB, identifier string, version string) (networkmodel.ArtefactModel, error) {
	var result networkmodel.ArtefactModel
	if err := db.GetContext(ctx, &result, `
	SELECT * FROM artefact WHERE identifier = $1 AND version = $2
	`, identifier, version); err != nil {
		return networkmodel.ArtefactModel{}, fmt.Errorf("failed to find artefact: %w", err)
	}

	return result, nil
}

// FetchArtefactByUUID locates a specific artefact based on its uuid.
func FetchArtefactByUUID(ctx context.Context, db *sqlm.DB, uuid uuid.UUID) (networkmodel.ArtefactModel, error) {
	var result networkmodel.ArtefactModel
	if err := db.GetContext(ctx, &result, `
	SELECT * FROM artefact WHERE uuid = $1
	`, uuid); err != nil {
		return networkmodel.ArtefactModel{}, fmt.Errorf("failed to find artefact: %w", err)
	}

	return result, nil
}

// FetchArtefactVersions queries the database for all currently hosted versions of the artefact.
func FetchArtefactVersions(ctx context.Context, db *sqlm.DB, identifier string) ([]networkmodel.ArtefactModel, error) {
	result := make([]networkmodel.ArtefactModel, 0)
	if err := db.SelectContext(ctx, &result, `
    SELECT * FROM artefact WHERE identifier = $1
    `, identifier); err != nil {
		return result, fmt.Errorf("failed to find artefact: %w", err)
	}

	return result, nil
}

// InsertArtefact inserts an artefact model into the database.
func InsertArtefact(ctx context.Context, db *sqlm.DB, model networkmodel.ArtefactModelWithBinary) (networkmodel.ArtefactModel, error) {
	transaction, err := db.Beginx()
	if err != nil {
		return networkmodel.ArtefactModel{}, fmt.Errorf("failed to begin insertion transaction: %w", err)
	}

	defer func() { _ = transaction.Rollback() }() // Rollback in case, this explodes. If Commit is called prior, this is a noop.

	var result networkmodel.ArtefactModel
	if err := transaction.NamedGetContext(ctx, &result, `
        INSERT INTO artefact (identifier, version, upload_date, requires_restart)
        VALUES (:identifier, :version, :upload_date, :requires_restart)
        RETURNING *;`,
		&model,
	); err != nil {
		return networkmodel.ArtefactModel{}, fmt.Errorf("failed to insert artefact: %w", err)
	}

	if _, err := transaction.ExecContext(ctx, `
        INSERT INTO artefact_file (artefact, tarball, hash) VALUES ($1, $2, $3);`,
		result.UUID, model.TarballBlob, model.Hash,
	); err != nil {
		return networkmodel.ArtefactModel{}, fmt.Errorf("failed to insert tarball into database for %s: %w", result.UUID, err)
	}

	if err := transaction.Commit(); err != nil {
		return networkmodel.ArtefactModel{}, fmt.Errorf("failed to commit insertion transaction: %w", err)
	}

	return result, nil
}

// FetchArtefactTarball fetches a full ArtefactModelWithBinary from the database including the full tarball binary.
func FetchArtefactTarball(ctx context.Context, db *sqlm.DB, uuid uuid.UUID) (networkmodel.ArtefactModelWithBinary, error) {
	var result networkmodel.ArtefactModelWithBinary
	if err := db.GetContext(ctx, &result, `
        SELECT uuid, identifier, version, upload_date, hash, tarball FROM artefact
            JOIN artefact_file af on artefact.uuid = af.artefact WHERE af.artefact = $1
        `, uuid); err != nil {
		return networkmodel.ArtefactModelWithBinary{}, fmt.Errorf("failed to fetch artefact model with binary from database: %w", err)
	}

	return result, nil
}

// DeleteArtefact deletes an artefact from the database.
func DeleteArtefact(ctx context.Context, db *sqlm.DB, uuid uuid.UUID) error {
	if _, err := db.ExecContext(ctx, "DELETE FROM artefact WHERE uuid = $1;", uuid); err != nil {
		return fmt.Errorf("failed to delete artefact %s: %w", uuid, err)
	}

	return nil
}
