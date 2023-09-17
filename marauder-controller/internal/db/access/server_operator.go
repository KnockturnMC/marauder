package access

import (
	"context"
	"fmt"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"

	"gitea.knockturnmc.com/marauder/controller/sqlm"
)

// FetchOperators fetches all operators known to the controller.
func FetchOperators(ctx context.Context, db *sqlm.DB) ([]networkmodel.ServerOperator, error) {
	result := make([]networkmodel.ServerOperator, 0)

	if err := db.SelectContext(ctx, &result, `
    SELECT * FROM server_operator
    `); err != nil {
		return result, fmt.Errorf("failed to list operators: %w", err)
	}

	return result, nil
}
