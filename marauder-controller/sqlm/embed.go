package sqlm

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Migrations hold the embedded migrations for the marauder controller.
//
//go:embed migrations/*
var Migrations embed.FS

// ApplyMigrations runs all known migrations on the passed database driver.
func ApplyMigrations(databaseDriver database.Driver, database string) error {
	fsDriver, err := iofs.New(Migrations, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create new iofs from embedded fs: %w", err)
	}

	instance, err := migrate.NewWithInstance("iofs", fsDriver, database, databaseDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := instance.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil // We are fine with no changes
		}

		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
