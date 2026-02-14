// Package migrations provides a thin wrapper around Goose for managing database
// schema migrations. Applications supply their embedded SQL files and version
// table name; this package handles the Goose configuration and execution.
package migrations

import (
	"context"
	"database/sql"
	"io/fs"
	"log/slog"
	"slices"

	"github.com/pressly/goose/v3"

	"github.com/bbsbb/go-edge/core/configuration"
)

const (
	minVersion = int64(0)
	maxVersion = int64((1 << 63) - 1)
)

// MigrationType represents the type of migration file.
type MigrationType string

const (
	MigrationTypeSQL MigrationType = "sql"
	MigrationTypeGo  MigrationType = "go"
)

// IsValidMigrationType determines if the migration type is supported.
func IsValidMigrationType(t string) bool {
	return t == string(MigrationTypeSQL) || t == string(MigrationTypeGo)
}

func setGooseOptions(schemaFS fs.FS, versionTable string, logger *slog.Logger) {
	goose.SetBaseFS(schemaFS)
	goose.SetLogger(newGooseLogger(logger))
	goose.SetSequential(true)
	goose.SetTableName(versionTable)
}

// CreateMigration creates a new migration file in the specified directory.
// The migrationType must be "sql" or "go".
func CreateMigration(ctx context.Context, logger *slog.Logger, schemaFS fs.FS, versionTable, name, migrationType, dir string) error {
	logger.Info("creating migration", "name", name, "type", migrationType)
	setGooseOptions(schemaFS, versionTable, logger)

	if err := goose.RunContext(ctx, "create", nil, dir, name, migrationType); err != nil {
		logger.Error("failed to create migration", "name", name, "error", err)
		return err
	}

	return nil
}

// MigrateUp migrates the database to the latest version.
func MigrateUp(db *sql.DB, logger *slog.Logger, schemaFS fs.FS, versionTable, relativeDir string) error {
	logger.Info("migrating database to latest version")
	setGooseOptions(schemaFS, versionTable, logger)

	if err := goose.Up(db, relativeDir); err != nil {
		logger.Error("migration failed", "error", err)
		return err
	}

	logger.Info("database migrated to latest version")
	return nil
}

// MigrateReset resets the database by rolling back all migrations and re-applying them.
// Only allowed in development and testing environments.
func MigrateReset(db *sql.DB, logger *slog.Logger, schemaFS fs.FS, versionTable, relativeDir string, environment configuration.Environment) error {
	allowed := []configuration.Environment{
		configuration.Development,
		configuration.Testing,
	}
	if !slices.Contains(allowed, environment) {
		logger.Error("reset is only allowed in development and testing environments", "environment", environment)
		return ErrResetNotAllowed
	}

	logger.Info("resetting database")
	setGooseOptions(schemaFS, versionTable, logger)

	version, err := goose.GetDBVersion(db)
	if err != nil || version == 0 {
		logger.Info("no migrations to reset, running up")
	} else {
		if err := goose.Reset(db, relativeDir); err != nil {
			logger.Error("reset failed", "error", err)
			return err
		}
	}

	if err := goose.Up(db, relativeDir); err != nil {
		logger.Error("migration after reset failed", "error", err)
		return err
	}

	logger.Info("database reset complete")
	return nil
}

// VerifyVersion verifies that the database is on the latest migration version.
func VerifyVersion(ctx context.Context, db *sql.DB, logger *slog.Logger, schemaFS fs.FS, versionTable, relativeDir string) error {
	logger.Info("verifying database is on latest migration version")
	setGooseOptions(schemaFS, versionTable, logger)

	migrations, err := goose.CollectMigrations(relativeDir, minVersion, maxVersion)
	if err != nil {
		logger.Error("failed to collect migrations", "error", err)
		return err
	}

	if len(migrations) == 0 {
		logger.Error("no migration files found")
		return ErrNoMigrations
	}

	fileVersion := migrations[len(migrations)-1].Version
	logger.Debug("latest migration file version", "version", fileVersion)

	dbVersion, err := goose.GetDBVersionContext(ctx, db)
	if err != nil {
		logger.Error("failed to get database version", "error", err)
		return err
	}
	logger.Debug("database migration version", "version", dbVersion)

	if fileVersion != dbVersion {
		logger.Error("version mismatch", "file_version", fileVersion, "db_version", dbVersion)
		return ErrVersionMismatch
	}

	logger.Info("database is on latest migration version", "version", fileVersion)
	return nil
}
