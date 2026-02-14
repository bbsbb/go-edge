package psqlfx

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/bbsbb/go-edge/core/domain"
)

const pgUniqueViolation = "23505"

// TranslateError converts pgx errors that have clear domain meaning into domain errors.
// Errors without domain meaning (connection failures, unexpected pgx errors) are returned
// unwrapped so the transport layer treats them as internal errors (HTTP 500).
func TranslateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.NewError(domain.CodeNotFound, "not found")
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return domain.NewError(domain.CodeConflict, "already exists")
	}
	return err
}
