//go:build testing

package migrations_test

import (
	"context"
	"embed"
	"testing"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/suite"

	"github.com/bbsbb/go-edge/core/configuration"
	"github.com/bbsbb/go-edge/core/fx/psqlfx"
	"github.com/bbsbb/go-edge/core/migrations"
	_ "github.com/bbsbb/go-edge/core/migrations/testdata/versions"
	coretesting "github.com/bbsbb/go-edge/core/testing"
)

//go:embed testdata/versions/*.sql
//go:embed testdata/versions/*.go
var testSchemaFS embed.FS

const (
	testVersionTable = "public.core_migrations_test_version"
	testRelativeDir  = "testdata/versions"
)

type MigrationSuite struct {
	suite.Suite
	db *coretesting.DB
}

func (s *MigrationSuite) SetupSuite() {
	cfg := &psqlfx.Configuration{
		Host:       "localhost",
		Port:       5432,
		Database:   "test_core",
		DisableSSL: true,
		Credentials: &psqlfx.Credentials{
			Username: "root",
			Password: "root",
		},
	}
	s.db = coretesting.NewDB(s.T(), cfg)
}

func (s *MigrationSuite) SetupTest() {
	sqlDB := stdlib.OpenDBFromPool(s.db.Pool)
	s.T().Cleanup(func() { sqlDB.Close() })

	_, err := sqlDB.Exec("DROP TABLE IF EXISTS test_migrations_table")
	s.Require().NoError(err)
	_, err = sqlDB.Exec("DROP TABLE IF EXISTS " + testVersionTable)
	s.Require().NoError(err)
}

func (s *MigrationSuite) TestMigrateUpAndVerify() {
	logger := coretesting.NewNoopLogger()
	sqlDB := stdlib.OpenDBFromPool(s.db.Pool)
	s.T().Cleanup(func() { sqlDB.Close() })

	err := migrations.MigrateUp(sqlDB, logger, testSchemaFS, testVersionTable, testRelativeDir)
	s.Require().NoError(err)

	err = migrations.VerifyVersion(context.Background(), sqlDB, logger, testSchemaFS, testVersionTable, testRelativeDir)
	s.Assert().NoError(err)
}

func (s *MigrationSuite) TestGoMigrationExecutes() {
	logger := coretesting.NewNoopLogger()
	sqlDB := stdlib.OpenDBFromPool(s.db.Pool)
	s.T().Cleanup(func() { sqlDB.Close() })

	err := migrations.MigrateUp(sqlDB, logger, testSchemaFS, testVersionTable, testRelativeDir)
	s.Require().NoError(err)

	var name string
	err = sqlDB.QueryRow("SELECT name FROM test_migrations_table WHERE id = '00000000-0000-0000-0000-000000000001'").Scan(&name)
	s.Require().NoError(err)
	s.Assert().Equal("seeded-by-go-migration", name)
}

func (s *MigrationSuite) TestVerifyVersionDetectsMismatch() {
	logger := coretesting.NewNoopLogger()
	sqlDB := stdlib.OpenDBFromPool(s.db.Pool)
	s.T().Cleanup(func() { sqlDB.Close() })

	// Create the version table with version 0 (no migrations applied) by running goose
	// against an empty set, then verify should detect the mismatch.
	// Simpler: just call VerifyVersion without running MigrateUp first.
	err := migrations.VerifyVersion(context.Background(), sqlDB, logger, testSchemaFS, testVersionTable, testRelativeDir)
	s.Assert().ErrorIs(err, migrations.ErrVersionMismatch)
}

func (s *MigrationSuite) TestMigrateResetInDevelopment() {
	logger := coretesting.NewNoopLogger()
	sqlDB := stdlib.OpenDBFromPool(s.db.Pool)
	s.T().Cleanup(func() { sqlDB.Close() })

	err := migrations.MigrateUp(sqlDB, logger, testSchemaFS, testVersionTable, testRelativeDir)
	s.Require().NoError(err)

	err = migrations.MigrateReset(sqlDB, logger, testSchemaFS, testVersionTable, testRelativeDir, configuration.Development)
	s.Assert().NoError(err)

	err = migrations.VerifyVersion(context.Background(), sqlDB, logger, testSchemaFS, testVersionTable, testRelativeDir)
	s.Assert().NoError(err)
}

func (s *MigrationSuite) TestMigrateResetInTesting() {
	logger := coretesting.NewNoopLogger()
	sqlDB := stdlib.OpenDBFromPool(s.db.Pool)
	s.T().Cleanup(func() { sqlDB.Close() })

	err := migrations.MigrateReset(sqlDB, logger, testSchemaFS, testVersionTable, testRelativeDir, configuration.Testing)
	s.Assert().NoError(err)
}

func (s *MigrationSuite) TestMigrateResetRejectsProduction() {
	logger := coretesting.NewNoopLogger()
	sqlDB := stdlib.OpenDBFromPool(s.db.Pool)
	s.T().Cleanup(func() { sqlDB.Close() })

	err := migrations.MigrateReset(sqlDB, logger, testSchemaFS, testVersionTable, testRelativeDir, configuration.Production)
	s.Assert().ErrorIs(err, migrations.ErrResetNotAllowed)
}

func (s *MigrationSuite) TestIsValidMigrationType() {
	s.Assert().True(migrations.IsValidMigrationType("sql"))
	s.Assert().True(migrations.IsValidMigrationType("go"))
	s.Assert().False(migrations.IsValidMigrationType("yaml"))
	s.Assert().False(migrations.IsValidMigrationType(""))
}

func TestMigrationSuite(t *testing.T) {
	suite.Run(t, new(MigrationSuite))
}
