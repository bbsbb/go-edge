package persistence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/bbsbb/go-edge/core/fx/rlsfx"
	"github.com/bbsbb/go-edge/sweetshop/internal/domain"
	"github.com/bbsbb/go-edge/sweetshop/internal/infrastructure/persistence/sqlcgen"
)

type OrderRepo struct {
	db *rlsfx.DB
}

func NewOrderRepo(db *rlsfx.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return rlsfx.Query(r.db, ctx, func(ctx context.Context, tx pgx.Tx) (*domain.Order, error) {
		m, err := sqlcgen.New(tx).FindOrderByID(ctx, id)
		if err != nil {
			return nil, err
		}
		order := orderToDomain(m)

		items, err := sqlcgen.New(tx).ListOrderItemsByOrderID(ctx, id)
		if err != nil {
			return nil, err
		}
		order.Items = make([]domain.OrderItem, len(items))
		for i, item := range items {
			order.Items[i] = orderItemToDomain(item)
		}
		return order, nil
	})
}

func (r *OrderRepo) Create(ctx context.Context, order *domain.Order) error {
	return rlsfx.Exec(r.db, ctx, func(ctx context.Context, tx pgx.Tx) error {
		return sqlcgen.New(tx).CreateOrder(ctx, orderCreateParams(order))
	})
}

func (r *OrderRepo) Close(ctx context.Context, id uuid.UUID, updatedAt time.Time) (*domain.Order, error) {
	return rlsfx.Query(r.db, ctx, func(ctx context.Context, tx pgx.Tx) (*domain.Order, error) {
		m, err := sqlcgen.New(tx).CloseOrder(ctx, orderCloseParams(id, updatedAt))
		if err != nil {
			return nil, err
		}
		return orderToDomain(m), nil
	})
}

func (r *OrderRepo) CreateItem(ctx context.Context, item *domain.OrderItem) error {
	return rlsfx.Exec(r.db, ctx, func(ctx context.Context, tx pgx.Tx) error {
		return sqlcgen.New(tx).CreateOrderItem(ctx, orderItemCreateParams(item))
	})
}

func (r *OrderRepo) ListItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error) {
	return rlsfx.Query(r.db, ctx, func(ctx context.Context, tx pgx.Tx) ([]domain.OrderItem, error) {
		rows, err := sqlcgen.New(tx).ListOrderItemsByOrderID(ctx, orderID)
		if err != nil {
			return nil, err
		}
		items := make([]domain.OrderItem, len(rows))
		for i, row := range rows {
			items[i] = orderItemToDomain(row)
		}
		return items, nil
	})
}
