package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/fx"

	"github.com/bbsbb/go-edge/core/configuration"
	"github.com/bbsbb/go-edge/core/fx/loggerfx"
	"github.com/bbsbb/go-edge/core/fx/psqlfx"
	"github.com/bbsbb/go-edge/sweetshop/internal/config"
	"github.com/bbsbb/go-edge/sweetshop/internal/migrations"
)

type migrationFunc func(ctx context.Context, db *sql.DB, logger *slog.Logger, env configuration.Environment) error

func RunMigrateUp(configPath string) error {
	return runMigration(configPath, func(_ context.Context, db *sql.DB, logger *slog.Logger, _ configuration.Environment) error {
		return migrations.MigrateUp(db, logger)
	})
}

func RunMigrateReset(configPath string) error {
	return runMigration(configPath, func(_ context.Context, db *sql.DB, logger *slog.Logger, env configuration.Environment) error {
		return migrations.MigrateReset(db, logger, env)
	})
}

func RunMigrateVerify(configPath string) error {
	return runMigration(configPath, func(ctx context.Context, db *sql.DB, logger *slog.Logger, _ configuration.Environment) error {
		return migrations.VerifyVersion(ctx, db, logger)
	})
}

// RunMigrateCreate creates a new migration file. The migrationType must be "sql" or "go".
func RunMigrateCreate(name, migrationType, dir string) error {
	if !migrations.IsValidMigrationType(migrationType) {
		return fmt.Errorf("invalid migration type %q: must be \"sql\" or \"go\"", migrationType)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	return migrations.CreateMigration(context.Background(), logger, name, migrationType, dir)
}

func runMigration(configPath string, fn migrationFunc) error {
	ctx := context.Background()
	env := configuration.Environment(os.Getenv("APP_ENVIRONMENT"))

	migrateCfg, err := config.NewMigrateConfiguration(ctx, path.Join(configPath, "migrate"))
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	var migrationErr error
	app := fx.New(
		fx.NopLogger,
		fx.Supply(
			migrateCfg.LoggingConfiguration(),
			migrateCfg.PSQLConfiguration(),
			migrateCfg.PSQLConnectionDefaults(),
		),
		fx.Provide(loggerfx.NewLogger, psqlfx.NewPool),
		fx.Invoke(func(pool *pgxpool.Pool, logger *slog.Logger, shutdowner fx.Shutdowner) error {
			db := stdlib.OpenDBFromPool(pool)
			migrationErr = fn(ctx, db, logger, env)
			return shutdowner.Shutdown()
		}),
	)

	app.Run()

	if err := app.Stop(ctx); err != nil {
		return err
	}

	return migrationErr
}
