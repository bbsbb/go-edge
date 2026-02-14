package migrations

import "errors"

var (
	ErrResetNotAllowed = errors.New("migrations: reset is only allowed in development and testing environments")
	ErrNoMigrations    = errors.New("migrations: no migration files found")
	ErrVersionMismatch = errors.New("migrations: database version does not match latest migration file")
)
