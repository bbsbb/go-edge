package config

import (
	"context"
	"os"

	"github.com/bbsbb/go-edge/core/configuration"
	"github.com/bbsbb/go-edge/core/fx/loggerfx"
	"github.com/bbsbb/go-edge/core/fx/psqlfx"
)

var (
	_ loggerfx.WithLogging = (*MigrateConfiguration)(nil)
	_ psqlfx.WithPSQL      = (*MigrateConfiguration)(nil)
)

type MigrateConfiguration struct {
	Environment configuration.Environment `yaml:"-" env:"ENVIRONMENT,overwrite"`
	Logging     *loggerfx.Configuration   `yaml:"logging" env:",prefix=LOGGING_,noinit"`
	PSQL        *psqlfx.Configuration     `yaml:"psql" env:",prefix=PSQL_,noinit"`
}

func NewMigrateConfiguration(ctx context.Context, configPath string) (*MigrateConfiguration, error) {
	env := configuration.Environment(os.Getenv("APP_ENVIRONMENT"))
	return configuration.LoadConfiguration[*MigrateConfiguration](
		ctx,
		env,
		configuration.WithPath(configPath),
		configuration.WithEnvironmentPrefix("APP_SWEETSHOP_"),
	)
}

func (c *MigrateConfiguration) LoggingConfiguration() *loggerfx.Configuration {
	return c.Logging
}

func (c *MigrateConfiguration) PSQLConfiguration() *psqlfx.Configuration {
	return c.PSQL
}

func (c *MigrateConfiguration) PSQLConnectionDefaults() *psqlfx.ConnectionDefaults {
	return &psqlfx.ConnectionDefaults{
		ApplicationName: "sweetshop_migrations",
	}
}
