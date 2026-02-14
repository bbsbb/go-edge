// Package rlsfx provides an fx module for row-level security enforced database access.
package rlsfx

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"github.com/bbsbb/go-edge/core/domain"
	"github.com/bbsbb/go-edge/core/fx/psqlfx"
)

var ErrMissingSchema = errors.New("rlsfx: schema is required")

// DB wraps a pgxpool.Pool and enforces row-level security on every transaction.
// There is no way to run a query without going through an RLS transaction.
type DB struct {
	pool   *pgxpool.Pool
	schema string
	field  string
	logger *slog.Logger
}

// Tx runs fn inside a transaction with RLS applied.
// The organization is extracted from ctx and used to SET LOCAL the RLS variable.
// Supports nesting: if ctx already carries an active transaction (via psqlfx.TxFromContext),
// creates a savepoint instead of a new top-level transaction.
func (db *DB) Tx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	org, err := domain.OrganizationFromContext(ctx)
	if err != nil {
		return err
	}

	variable := db.schema + "." + db.field

	parent := psqlfx.TxFromContext(ctx)

	var tx pgx.Tx
	if parent != nil {
		tx, err = parent.Begin(ctx) // SAVEPOINT
	} else {
		tx, err = db.pool.Begin(ctx) // BEGIN
	}
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// For nested transactions, save the current RLS value before overwriting.
	// After RELEASE SAVEPOINT, SET LOCAL changes persist to the parent —
	// we restore the previous value after commit to prevent scope leakage.
	var savedValue string
	if parent != nil {
		if err := tx.QueryRow(ctx, "SELECT current_setting($1, true)", variable).Scan(&savedValue); err != nil {
			if db.logger != nil {
				db.logger.WarnContext(ctx, "failed to read current RLS value", "variable", variable, "error", err)
			}
		}
	}

	if db.logger != nil {
		db.logger.DebugContext(ctx, "applying RLS", "variable", variable)
	}

	if _, err := tx.Exec(ctx, "SET LOCAL "+variable+" = "+psqlfx.QuoteLiteral(org.ID.String())); err != nil {
		return err
	}

	if err := fn(psqlfx.ContextWithTx(ctx, tx), tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	// Restore the parent transaction's RLS variable after savepoint release.
	if parent != nil && savedValue != "" {
		if _, err := parent.Exec(ctx, "SET LOCAL "+variable+" = "+psqlfx.QuoteLiteral(savedValue)); err != nil {
			return err
		}
	}

	return nil
}

// NewDB creates a DB without FX dependency injection.
func NewDB(pool *pgxpool.Pool, cfg *Configuration, logger *slog.Logger) (*DB, error) {
	if cfg.Schema == "" {
		return nil, ErrMissingSchema
	}
	return &DB{
		pool:   pool,
		schema: cfg.Schema,
		field:  cfg.Field,
		logger: logger,
	}, nil
}

type params struct {
	fx.In
	Pool   *pgxpool.Pool
	Config *Configuration
	Logger *slog.Logger `optional:"true"`
}

func NewRLS(p params) (*DB, error) {
	return NewDB(p.Pool, p.Config, p.Logger)
}

func provideConfiguration(cfg WithRLS) *Configuration {
	return cfg.RLSConfiguration()
}

var Module = fx.Module(
	"rlsfx",
	fx.Provide(provideConfiguration, NewRLS),
)
