// Package psqlfx provides an fx module for PostgreSQL database connections.
package psqlfx

import (
	"context"
	"fmt"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

type WithPSQL interface {
	PSQLConfiguration() *Configuration
	PSQLConnectionDefaults() *ConnectionDefaults
}

type Params struct {
	fx.In
	Lifecycle fx.Lifecycle
	Config    *Configuration
	Defaults  *ConnectionDefaults `optional:"true"`
}

type Result struct {
	fx.Out
	Pool *pgxpool.Pool
}

func NewPool(p Params) (Result, error) {
	dsn := p.Config.DSN()
	if p.Defaults != nil {
		dsnParams, err := p.Defaults.DSN()
		if err != nil {
			return Result{}, fmt.Errorf("psqlfx: encode connection defaults: %w", err)
		}
		if dsnParams != "" {
			dsn = dsn + " " + dsnParams
		}
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return Result{}, fmt.Errorf("psqlfx: parse pool config: %w", err)
	}

	pool := p.Config.Pool
	if pool == nil {
		pool = DefaultPoolConfiguration()
	}

	if pool.MaxOpenConns > 0 {
		poolConfig.MaxConns = pool.MaxOpenConns
	}
	if pool.MaxIdleConns > 0 {
		poolConfig.MinConns = pool.MaxIdleConns
	}
	if pool.ConnMaxLifetime > 0 {
		poolConfig.MaxConnLifetime = pool.ConnMaxLifetime
	}

	poolConfig.ConnConfig.Tracer = otelpgx.NewTracer()

	pgxPool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return Result{}, fmt.Errorf("psqlfx: create pool: %w", err)
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if err := pgxPool.Ping(ctx); err != nil {
				return fmt.Errorf("psqlfx: health check ping: %w", err)
			}
			return nil
		},
		OnStop: func(_ context.Context) error {
			pgxPool.Close()
			return nil
		},
	})

	return Result{Pool: pgxPool}, nil
}

func provideConfiguration(cfg WithPSQL) *Configuration {
	return cfg.PSQLConfiguration()
}

func provideConnectionDefaults(cfg WithPSQL) *ConnectionDefaults {
	return cfg.PSQLConnectionDefaults()
}

var Module = fx.Module(
	"psqlfx",
	fx.Provide(provideConfiguration, provideConnectionDefaults, NewPool),
)
