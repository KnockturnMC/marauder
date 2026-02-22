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
func (db *DB) NamedGetContext(ctx context.Context, dest any, query string, args any) error {
	return NamedGetContext(ctx, db, dest, query, args)
}

// Beginx opens a new transaction on the db.
func (db *DB) Beginx() (*Tx, error) {
	begin, err := db.DB.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &Tx{Tx: begin}, nil
}

type Tx struct {
	*sqlx.Tx
}

// NamedGetContext runs a single named query context and scans the first resulting row into the passed destination.
func (db *Tx) NamedGetContext(ctx context.Context, dest any, query string, args any) error {
	return NamedGetContext(ctx, db, dest, query, args)
}

func NamedGetContext(ctx context.Context, e sqlx.ExtContext, dest any, query string, args any) error {
	rows, err := sqlx.NamedQueryContext(ctx, e, query, args)
	if err != nil {
		return fmt.Errorf("failed to run named query: %w", err)
	}

	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		return fmt.Errorf("could not find row: %w", sql.ErrNoRows)
	}

	if err := rows.StructScan(dest); err != nil {
		return fmt.Errorf("failed to scan result: %w", err)
	}

	return nil
}
