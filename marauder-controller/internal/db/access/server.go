package access

import (
	"context"
	"fmt"

	"gitea.knockturnmc.com/marauder/controller/sqlm"

	"gitea.knockturnmc.com/marauder/controller/internal/db/models"
	"github.com/google/uuid"
)

// FetchServer locates a server based on its uuid.
// sql.ErrNoRows is returned if no server exists with the passed uuid.
func FetchServer(ctx context.Context, db *sqlm.DB, uuid uuid.UUID) (models.ServerModel, error) {
	var result models.ServerModel
	if err := db.GetContext(ctx, &result, `
    SELECT uuid, environment, name, host, memory, cpu, image FROM server WHERE uuid = $1
    `, uuid); err != nil {
		return models.ServerModel{}, fmt.Errorf("failed to find server: %w", err)
	}

	return result, nil
}

// FetchServerByNameAndEnv looks up a single server based on its name and environment.
// sql.ErrNoRows is returned if no server exists with the passed name and environment.
func FetchServerByNameAndEnv(ctx context.Context, db *sqlm.DB, name string, environment string) (models.ServerModel, error) {
	var result models.ServerModel
	if err := db.GetContext(ctx, &result, `
    SELECT uuid, environment, name, host, memory, cpu, image FROM server WHERE name = $1 AND environment = $2
    `, name, environment); err != nil {
		return models.ServerModel{}, fmt.Errorf("failed to find server: %w", err)
	}

	return result, nil
}

// FetchServersByName queries the database for a collection of servers by their name.
func FetchServersByName(ctx context.Context, db *sqlm.DB, name string) ([]models.ServerModel, error) {
	var result []models.ServerModel
	if err := db.SelectContext(ctx, &result, `
    SELECT * FROM server WHERE name = $1
    `, name); err != nil {
		return result, fmt.Errorf("failed to find servers: %w", err)
	}

	return result, nil
}

// FetchServersByEnvironment queries the database for a collection of servers by their environment.
func FetchServersByEnvironment(ctx context.Context, db *sqlm.DB, environment string) ([]models.ServerModel, error) {
	var result []models.ServerModel
	if err := db.SelectContext(ctx, &result, `
    SELECT * FROM server WHERE environment = $1
    `, environment); err != nil {
		return result, fmt.Errorf("failed to find servers: %w", err)
	}

	return result, nil
}

// InsertServer creates a new server instance on the database.
func InsertServer(ctx context.Context, db *sqlm.DB, server models.ServerModel) (models.ServerModel, error) {
	if err := db.NamedGetContext(ctx, &server, `
            INSERT INTO server (environment, name, host, memory, cpu, image)
            VALUES (:environment, :name, :host, :memory, :cpu, :image)
            RETURNING *; 
            `, server); err != nil {
		return models.ServerModel{}, fmt.Errorf("failed to insert server: %w", err)
	}

	return server, nil
}
