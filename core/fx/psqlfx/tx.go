package psqlfx

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type txContextKey struct{}

// TxFromContext returns the pgx.Tx stored in ctx, or nil if none is present.
func TxFromContext(ctx context.Context) pgx.Tx {
	tx, _ := ctx.Value(txContextKey{}).(pgx.Tx)
	return tx
}

// ContextWithTx returns a new context carrying the given pgx.Tx.
func ContextWithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}
