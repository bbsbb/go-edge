//go:build testing

// Package testing provides test utilities including transaction-isolated database
// helpers, mock RLS setup, and log capture.
package testing

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bbsbb/go-edge/core/fx/psqlfx"
)

// DB provides pgx-based test database isolation.
// Each test runs inside a transaction that rolls back on cleanup.
type DB struct {
	Pool *pgxpool.Pool
}

// NewDB creates a connection pool from the given psqlfx configuration.
func NewDB(t *testing.T, dbConfig *psqlfx.Configuration) *DB {
	t.Helper()

	pool, err := pgxpool.New(context.Background(), dbConfig.DSN())
	if err != nil {
		t.Fatalf("Unable to create pgxpool: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("Unable to ping database: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return &DB{Pool: pool}
}

// WithTx begins a transaction, stores it in context via psqlfx.ContextWithTx,
// and registers a rollback on t.Cleanup. All repositories — RLS and non-RLS —
// see the transaction in context and use it automatically.
func (d *DB) WithTx(t *testing.T, fn func(ctx context.Context)) {
	t.Helper()

	tx, err := d.Pool.Begin(context.Background())
	if err != nil {
		t.Fatalf("Unable to begin transaction: %v", err)
	}

	ctx := psqlfx.ContextWithTx(context.Background(), tx)

	t.Cleanup(func() {
		_ = tx.Rollback(context.Background())
	})

	fn(ctx)
}

// MockRLS sets a PostgreSQL session variable on the transaction from context.
// This simulates RLS context for tests without going through rlsfx.
func MockRLS(ctx context.Context, t *testing.T, schema, field, value string) {
	t.Helper()

	tx := psqlfx.TxFromContext(ctx)
	if tx == nil {
		t.Fatal("MockRLS requires a transaction in context — call WithTx first")
	}

	variable := field
	if schema != "" {
		variable = schema + "." + field
	}

	setSQL := fmt.Sprintf("SET LOCAL %s TO %s", variable, psqlfx.QuoteLiteral(value))
	if _, err := tx.Exec(ctx, setSQL); err != nil {
		t.Fatalf("Failed to set RLS variable: %v", err)
	}
}
