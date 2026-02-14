package http

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/bbsbb/go-edge/core/fx/middlewarefx"
	coremiddleware "github.com/bbsbb/go-edge/core/transport/http/middleware"
	"github.com/bbsbb/go-edge/sweetshop/internal/service"
	"github.com/bbsbb/go-edge/sweetshop/internal/transport/http/handler"
)

type routeParams struct {
	fx.In
	Mux            *chi.Mux
	ProductHandler *handler.ProductHandler
	OrderHandler   *handler.OrderHandler
}

func registerRoutes(p routeParams) {
	p.Mux.Route("/products", func(r chi.Router) {
		r.Get("/", p.ProductHandler.List)
		r.Post("/", p.ProductHandler.Create)
		r.Get("/{id}", p.ProductHandler.Get)
		r.Put("/{id}", p.ProductHandler.Update)
		r.Delete("/{id}", p.ProductHandler.Delete)
	})

	p.Mux.Route("/orders", func(r chi.Router) {
		r.Post("/", p.OrderHandler.Open)
		r.Get("/{id}", p.OrderHandler.Get)
		r.Post("/{id}/items", p.OrderHandler.AddItem)
		r.Post("/{id}/close", p.OrderHandler.Close)
	})
}

type orgMiddlewareResult struct {
	fx.Out
	Middleware middlewarefx.Middleware `group:"middleware"`
}

func provideOrganizationMiddleware(loader coremiddleware.OrganizationLoader, logger *slog.Logger) orgMiddlewareResult {
	return orgMiddlewareResult{
		Middleware: middlewarefx.Middleware{
			Name: "organization",
			Handler: coremiddleware.WithOrganization(coremiddleware.WithOrganizationConfig{
				SkipPaths: []string{"/healthz", "/readyz"},
				Logger:    logger,
				Loader:    loader,
			}),
		},
	}
}

var RouteModule = fx.Module(
	"sweetshop/routes",
	fx.Provide(
		service.NewProductService,
		service.NewOrderService,
		service.NewRegistry,
		handler.NewProductHandler,
		handler.NewOrderHandler,
		provideOrganizationMiddleware,
	),
	fx.Invoke(registerRoutes),
)
