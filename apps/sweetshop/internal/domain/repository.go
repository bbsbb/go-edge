package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	coredomain "github.com/bbsbb/go-edge/core/domain"
)

type OrganizationRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*coredomain.Organization, error)
	FindBySlug(ctx context.Context, slug string) (*coredomain.Organization, error)
	Create(ctx context.Context, org *coredomain.Organization) error
}

type ProductRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Product, error)
	List(ctx context.Context) ([]*Product, error)
	Create(ctx context.Context, product *Product) error
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type OrderRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Order, error)
	Create(ctx context.Context, order *Order) error
	Close(ctx context.Context, id uuid.UUID, updatedAt time.Time) (*Order, error)
	CreateItem(ctx context.Context, item *OrderItem) error
	ListItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]OrderItem, error)
}
