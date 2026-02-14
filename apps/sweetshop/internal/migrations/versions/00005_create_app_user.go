package versions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"

	"github.com/bbsbb/go-edge/core/fx/psqlfx"
)

func init() {
	goose.AddMigrationContext(upCreateAppUser, downCreateAppUser)
}

func upCreateAppUser(ctx context.Context, tx *sql.Tx) error {
	cfg, err := getAppConfiguration(ctx)
	if err != nil {
		return err
	}

	appUser := cfg.PSQL.Credentials.Username
	quotedUser := psqlfx.QuoteIdentifier([]string{appUser})

	createUserSQL := fmt.Sprintf(
		"CREATE USER %s WITH LOGIN PASSWORD %s",
		quotedUser,
		psqlfx.QuoteLiteral(cfg.PSQL.Credentials.Password),
	)
	if _, err = tx.ExecContext(ctx, createUserSQL); err != nil {
		return fmt.Errorf("create app user: %w", err)
	}

	grantDBSQL := fmt.Sprintf(
		"GRANT CONNECT, TEMP ON DATABASE %s TO %s",
		psqlfx.QuoteIdentifier([]string{cfg.PSQL.Database}),
		quotedUser,
	)
	if _, err = tx.ExecContext(ctx, grantDBSQL); err != nil {
		return fmt.Errorf("grant database privileges: %w", err)
	}

	quotedSchema := psqlfx.QuoteIdentifier([]string{AppSchemaName})

	grantSchemaSQL := fmt.Sprintf("GRANT USAGE ON SCHEMA %s TO %s", quotedSchema, quotedUser)
	if _, err = tx.ExecContext(ctx, grantSchemaSQL); err != nil {
		return fmt.Errorf("grant schema privileges: %w", err)
	}

	grantTableSQL := fmt.Sprintf( //nolint:gosec // identifiers from trusted config, properly quoted
		"GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA %s TO %s",
		quotedSchema, quotedUser,
	)
	if _, err = tx.ExecContext(ctx, grantTableSQL); err != nil {
		return fmt.Errorf("grant table privileges: %w", err)
	}

	grantFutureSQL := fmt.Sprintf( //nolint:gosec // identifiers from trusted config, properly quoted
		"ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO %s",
		quotedSchema, quotedUser,
	)
	if _, err = tx.ExecContext(ctx, grantFutureSQL); err != nil {
		return fmt.Errorf("alter default privileges: %w", err)
	}

	return nil
}

func downCreateAppUser(ctx context.Context, tx *sql.Tx) error {
	appCfg, err := getAppConfiguration(ctx)
	if err != nil {
		return err
	}

	migrateCfg, err := getMigrateConfiguration(ctx)
	if err != nil {
		return err
	}

	quotedUser := psqlfx.QuoteIdentifier([]string{appCfg.PSQL.Credentials.Username})
	quotedRootUser := psqlfx.QuoteIdentifier([]string{migrateCfg.PSQL.Credentials.Username})

	reassignSQL := fmt.Sprintf("REASSIGN OWNED BY %s TO %s", quotedUser, quotedRootUser)
	if _, err := tx.ExecContext(ctx, reassignSQL); err != nil {
		return fmt.Errorf("reassign owned by app user: %w", err)
	}

	dropOwnedSQL := fmt.Sprintf("DROP OWNED BY %s CASCADE", quotedUser)
	if _, err := tx.ExecContext(ctx, dropOwnedSQL); err != nil {
		return fmt.Errorf("drop owned by app user: %w", err)
	}

	dropUserSQL := fmt.Sprintf("DROP USER IF EXISTS %s", quotedUser)
	if _, err := tx.ExecContext(ctx, dropUserSQL); err != nil {
		return fmt.Errorf("drop app user: %w", err)
	}

	return nil
}
