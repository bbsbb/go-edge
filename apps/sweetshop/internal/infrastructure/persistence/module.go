package persistence

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"github.com/bbsbb/go-edge/core/fx/rlsfx"
	coremiddleware "github.com/bbsbb/go-edge/core/transport/http/middleware"
	"github.com/bbsbb/go-edge/sweetshop/internal/domain"
)

func provideOrganizationRepo(pool *pgxpool.Pool) domain.OrganizationRepository {
	return NewOrganizationRepo(pool)
}

func provideOrganizationLoader(pool *pgxpool.Pool) coremiddleware.OrganizationLoader {
	return NewOrganizationRepo(pool)
}

func provideProductRepo(db *rlsfx.DB) domain.ProductRepository {
	return NewProductRepo(db)
}

func provideOrderRepo(db *rlsfx.DB) domain.OrderRepository {
	return NewOrderRepo(db)
}

var Module = fx.Module(
	"sweetshop/persistence",
	fx.Provide(
		provideOrganizationRepo,
		provideOrganizationLoader,
		provideProductRepo,
		provideOrderRepo,
	),
)
