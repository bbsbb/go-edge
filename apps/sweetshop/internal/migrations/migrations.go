// Package migrations contains database schema migrations for the sweetshop application.
package migrations

import (
	"context"
	"database/sql"
	"embed"
	"log/slog"

	"github.com/bbsbb/go-edge/core/configuration"
	coremigrations "github.com/bbsbb/go-edge/core/migrations"

	_ "github.com/bbsbb/go-edge/sweetshop/internal/migrations/versions"
)

const (
	relativeDir  = "versions"
	versionTable = "public.sweetshop_goose_db_version"
)

//go:embed versions/*.sql
//go:embed versions/*.go
var schemaFS embed.FS

func IsValidMigrationType(t string) bool {
	return coremigrations.IsValidMigrationType(t)
}

func CreateMigration(ctx context.Context, logger *slog.Logger, name, migrationType, dir string) error {
	return coremigrations.CreateMigration(ctx, logger, schemaFS, versionTable, name, migrationType, dir)
}

func MigrateUp(db *sql.DB, logger *slog.Logger) error {
	return coremigrations.MigrateUp(db, logger, schemaFS, versionTable, relativeDir)
}

func MigrateReset(db *sql.DB, logger *slog.Logger, environment configuration.Environment) error {
	return coremigrations.MigrateReset(db, logger, schemaFS, versionTable, relativeDir, environment)
}

func VerifyVersion(ctx context.Context, db *sql.DB, logger *slog.Logger) error {
	return coremigrations.VerifyVersion(ctx, db, logger, schemaFS, versionTable, relativeDir)
}
