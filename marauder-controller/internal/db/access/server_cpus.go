package access

import (
	"context"
	"fmt"

	"gitea.knockturnmc.com/marauder/controller/sqlm"

	"gitea.knockturnmc.com/marauder/controller/internal/db/models"
	"github.com/google/uuid"
)

// FindAllocatedCPUs find the cpu allocations specified for the passed server.
func FindAllocatedCPUs(ctx context.Context, db *sqlm.DB, serverUUID uuid.UUID) ([]models.ServerCPUAllocation, error) {
	var allocations []models.ServerCPUAllocation
	if err := db.GetContext(ctx, &allocations, `SELECT * FROM server_cpu_allocation WHERE server_uuid = $1`, serverUUID); err != nil {
		return nil, fmt.Errorf("failed to locate server cpu allocations for %s: %w", serverUUID.String(), err)
	}

	return allocations, nil
}

// InsertCPUAllocation adds a new cpu core to the database for the specific server.
func InsertCPUAllocation(ctx context.Context, db *sqlm.DB, allocation models.ServerCPUAllocation) (models.ServerCPUAllocation, error) {
	if err := db.NamedGetContext(ctx, &allocation, `
            INSERT INTO server_cpu_allocation (server_uuid, server_host, cpu_core)
            VALUES (:server_uuid, :server_host, :cpu_core)
            RETURNING *; 
            `, allocation); err != nil {
		return models.ServerCPUAllocation{}, fmt.Errorf("failed to insert cpu allocation: %w", err)
	}

	return allocation, nil
}

// DeleteCPUAllocation deletes a specific CPU allocation from the database.
func DeleteCPUAllocation(ctx context.Context, db *sqlm.DB, uuid uuid.UUID) error {
	if _, err := db.ExecContext(ctx, `DELETE FROM server_cpu_allocation WHERE uuid = $1`, uuid); err != nil {
		return fmt.Errorf("failed to delete cpu allocation %s: %w", uuid.String(), err)
	}

	return nil
}
