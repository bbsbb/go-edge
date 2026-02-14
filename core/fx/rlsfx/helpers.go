package rlsfx

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/bbsbb/go-edge/core/fx/psqlfx"
)

// Query runs fn inside an RLS transaction and translates pgx errors to domain errors.
func Query[T any](db *DB, ctx context.Context, fn func(context.Context, pgx.Tx) (T, error)) (T, error) {
	var result T
	err := db.Tx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		v, err := fn(ctx, tx)
		if err != nil {
			return err
		}
		result = v
		return nil
	})
	if err != nil {
		return result, psqlfx.TranslateError(err)
	}
	return result, nil
}

// Exec runs fn inside an RLS transaction and translates pgx errors to domain errors.
func Exec(db *DB, ctx context.Context, fn func(context.Context, pgx.Tx) error) error {
	err := db.Tx(ctx, fn)
	if err != nil {
		return psqlfx.TranslateError(err)
	}
	return nil
}
