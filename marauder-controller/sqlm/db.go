package sqlm

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// DB is a utility helper wrapper for the sqlx.DB instance.
type DB struct {
	*sqlx.DB
}

// NamedGetContext runs a single named query context and scans the first resulting row into the passed destination.
func (f *DB) NamedGetContext(ctx context.Context, dest interface{}, query string, args interface{}) error {
	queryContext, err := f.NamedQueryContext(ctx, query, args)
	if err != nil {
		return fmt.Errorf("failed to run named query: %w", err)
	}

	defer func() { _ = queryContext.Close() }()

	if !queryContext.Next() {
		return fmt.Errorf("could not find row: %w", sql.ErrNoRows)
	}

	if err := queryContext.StructScan(dest); err != nil {
		return fmt.Errorf("failed to scan result: %w", err)
	}

	return nil
}
