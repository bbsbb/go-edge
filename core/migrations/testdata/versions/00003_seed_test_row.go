package versions

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upSeedTestRow, downSeedTestRow)
}

func upSeedTestRow(_ context.Context, tx *sql.Tx) error {
	_, err := tx.Exec("INSERT INTO test_migrations_table (id, name) VALUES ('00000000-0000-0000-0000-000000000001', 'seeded-by-go-migration')")
	return err
}

func downSeedTestRow(_ context.Context, tx *sql.Tx) error {
	_, err := tx.Exec("DELETE FROM test_migrations_table WHERE id = '00000000-0000-0000-0000-000000000001'")
	return err
}
