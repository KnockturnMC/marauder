package access

import (
	"context"
	"fmt"

	"github.com/knockturnmc/marauder/marauder-controller/sqlm"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
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
