// Package config provides application configuration for the sweetshop service.
package config

import (
	"context"
	"os"

	"go.uber.org/fx"

	"github.com/bbsbb/go-edge/core/configuration"
	"github.com/bbsbb/go-edge/core/fx/httpserverfx"
	"github.com/bbsbb/go-edge/core/fx/loggerfx"
	"github.com/bbsbb/go-edge/core/fx/middlewarefx"
	"github.com/bbsbb/go-edge/core/fx/otelfx"
	"github.com/bbsbb/go-edge/core/fx/psqlfx"
	"github.com/bbsbb/go-edge/core/fx/rlsfx"
)

var (
	_ loggerfx.WithLogging        = (*AppConfiguration)(nil)
	_ httpserverfx.WithHTTPServer = (*AppConfiguration)(nil)
	_ psqlfx.WithPSQL             = (*AppConfiguration)(nil)
	_ rlsfx.WithRLS               = (*AppConfiguration)(nil)
	_ otelfx.WithOTel             = (*AppConfiguration)(nil)
	_ middlewarefx.WithMiddleware = (*AppConfiguration)(nil)
)

type AppConfiguration struct {
	Environment configuration.Environment   `yaml:"-" env:"ENVIRONMENT,overwrite"`
	Logging     *loggerfx.Configuration     `yaml:"logging" env:",prefix=LOGGING_,noinit"`
	HTTPServer  *httpserverfx.Configuration `yaml:"http_server" env:",prefix=HTTP_,noinit"`
	PSQL        *psqlfx.Configuration       `yaml:"psql" env:",prefix=PSQL_,noinit"`
	OTel        *otelfx.Configuration       `yaml:"otel" env:",prefix=OTEL_,noinit"`
	RLS         *rlsfx.Configuration        `yaml:"rls" env:",prefix=RLS_,noinit"`
	Middleware  *middlewarefx.Configuration `yaml:"middleware" env:",prefix=MW_,noinit"`
}

func NewAppConfiguration(ctx context.Context, configPath string) (*AppConfiguration, error) {
	env := configuration.Environment(os.Getenv("APP_ENVIRONMENT"))
	return configuration.LoadConfiguration[*AppConfiguration](
		ctx,
		env,
		configuration.WithPath(configPath),
		configuration.WithEnvironmentPrefix("APP_SWEETSHOP_"),
	)
}

func (c *AppConfiguration) LoggingConfiguration() *loggerfx.Configuration {
	return c.Logging
}

func (c *AppConfiguration) HTTPServerConfiguration() *httpserverfx.Configuration {
	return c.HTTPServer
}

func (c *AppConfiguration) PSQLConfiguration() *psqlfx.Configuration {
	return c.PSQL
}

func (c *AppConfiguration) PSQLConnectionDefaults() *psqlfx.ConnectionDefaults {
	return &psqlfx.ConnectionDefaults{
		ApplicationName: "sweetshop",
	}
}

func (c *AppConfiguration) OTelConfiguration() *otelfx.Configuration {
	return c.OTel
}

func (c *AppConfiguration) RLSConfiguration() *rlsfx.Configuration {
	return c.RLS
}

func (c *AppConfiguration) MiddlewareConfiguration() *middlewarefx.Configuration {
	return c.Middleware
}

func (c *AppConfiguration) AsFx() fx.Option {
	return fx.Supply(
		c,
		fx.Annotate(
			c,
			fx.As(new(loggerfx.WithLogging)),
			fx.As(new(httpserverfx.WithHTTPServer)),
			fx.As(new(psqlfx.WithPSQL)),
			fx.As(new(otelfx.WithOTel)),
			fx.As(new(rlsfx.WithRLS)),
			fx.As(new(middlewarefx.WithMiddleware)),
		),
	)
}
