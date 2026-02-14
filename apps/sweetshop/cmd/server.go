// Package cmd provides the application entry points.
package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"github.com/bbsbb/go-edge/core/fx/bootfx"
	"github.com/bbsbb/go-edge/core/fx/httpserverfx"
	"github.com/bbsbb/go-edge/core/fx/middlewarefx"
	"github.com/bbsbb/go-edge/core/fx/otelfx"
	"github.com/bbsbb/go-edge/core/fx/psqlfx"
	"github.com/bbsbb/go-edge/core/fx/rlsfx"
	transporthttp "github.com/bbsbb/go-edge/core/transport/http"
	"github.com/bbsbb/go-edge/sweetshop/internal/config"
	"github.com/bbsbb/go-edge/sweetshop/internal/infrastructure/persistence"
	transportroutes "github.com/bbsbb/go-edge/sweetshop/internal/transport/http"
)

func RunServer(configPath string) error {
	ctx := context.Background()

	cfg, err := config.NewAppConfiguration(ctx, configPath)
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	app := fx.New(
		bootfx.BootFx(cfg,
			httpserverfx.Module,
			psqlfx.Module,
			rlsfx.Module,
			otelfx.Module,
			middlewarefx.Module,
			persistence.Module,
			transportroutes.RouteModule,
			fx.Invoke(registerHealthRoutes),
		),
	)

	app.Run()
	return nil
}

type healthParams struct {
	fx.In
	Mux    *chi.Mux
	Pool   *pgxpool.Pool
	Logger *slog.Logger
}

func registerHealthRoutes(p healthParams) {
	p.Mux.Get("/healthz", transporthttp.LivenessHandler())
	p.Mux.Get("/readyz", transporthttp.ReadinessHandler(p.Pool, 0, p.Logger))
}
