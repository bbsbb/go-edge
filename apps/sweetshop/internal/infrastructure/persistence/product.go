package persistence

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/bbsbb/go-edge/core/fx/rlsfx"
	"github.com/bbsbb/go-edge/sweetshop/internal/domain"
	"github.com/bbsbb/go-edge/sweetshop/internal/infrastructure/persistence/sqlcgen"
)

type ProductRepo struct {
	db *rlsfx.DB
}

func NewProductRepo(db *rlsfx.DB) *ProductRepo {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	return rlsfx.Query(r.db, ctx, func(ctx context.Context, tx pgx.Tx) (*domain.Product, error) {
		m, err := sqlcgen.New(tx).FindProductByID(ctx, id)
		if err != nil {
			return nil, err
		}
		return productToDomain(m), nil
	})
}

func (r *ProductRepo) List(ctx context.Context) ([]*domain.Product, error) {
	return rlsfx.Query(r.db, ctx, func(ctx context.Context, tx pgx.Tx) ([]*domain.Product, error) {
		rows, err := sqlcgen.New(tx).ListProducts(ctx)
		if err != nil {
			return nil, err
		}
		products := make([]*domain.Product, len(rows))
		for i, m := range rows {
			products[i] = productToDomain(m)
		}
		return products, nil
	})
}

func (r *ProductRepo) Create(ctx context.Context, product *domain.Product) error {
	return rlsfx.Exec(r.db, ctx, func(ctx context.Context, tx pgx.Tx) error {
		return sqlcgen.New(tx).CreateProduct(ctx, productCreateParams(product))
	})
}

func (r *ProductRepo) Update(ctx context.Context, product *domain.Product) error {
	return rlsfx.Exec(r.db, ctx, func(ctx context.Context, tx pgx.Tx) error {
		n, err := sqlcgen.New(tx).UpdateProduct(ctx, productUpdateParams(product))
		if err != nil {
			return err
		}
		if n == 0 {
			return pgx.ErrNoRows
		}
		return nil
	})
}

func (r *ProductRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return rlsfx.Exec(r.db, ctx, func(ctx context.Context, tx pgx.Tx) error {
		n, err := sqlcgen.New(tx).DeleteProduct(ctx, id)
		if err != nil {
			return err
		}
		if n == 0 {
			return pgx.ErrNoRows
		}
		return nil
	})
}
