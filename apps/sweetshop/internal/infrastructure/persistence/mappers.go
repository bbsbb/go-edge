package persistence

import (
	"time"

	"github.com/google/uuid"

	coredomain "github.com/bbsbb/go-edge/core/domain"
	"github.com/bbsbb/go-edge/sweetshop/internal/domain"
	"github.com/bbsbb/go-edge/sweetshop/internal/infrastructure/persistence/sqlcgen"
)

func organizationToDomain(m sqlcgen.Organization) *coredomain.Organization {
	return &coredomain.Organization{
		ID:   m.ID,
		Slug: m.Slug,
	}
}

func organizationCreateParams(org *coredomain.Organization, now time.Time) sqlcgen.CreateOrganizationParams {
	return sqlcgen.CreateOrganizationParams{
		ID:              org.ID,
		SystemCreatedAt: now,
		SystemUpdatedAt: now,
		Name:            org.Slug,
		Slug:            org.Slug,
	}
}

func productToDomain(m sqlcgen.Product) *domain.Product {
	return &domain.Product{
		ID:             m.ID,
		OrganizationID: m.OrganizationID,
		CreatedAt:      m.SystemCreatedAt,
		UpdatedAt:      m.SystemUpdatedAt,
		Name:           m.Name,
		Category:       domain.ProductCategory(m.Category),
		PriceCents:     m.PriceCents,
	}
}

func productCreateParams(p *domain.Product) sqlcgen.CreateProductParams {
	return sqlcgen.CreateProductParams{
		ID:              p.ID,
		OrganizationID:  p.OrganizationID,
		SystemCreatedAt: p.CreatedAt,
		SystemUpdatedAt: p.UpdatedAt,
		Name:            p.Name,
		Category:        string(p.Category),
		PriceCents:      p.PriceCents,
	}
}

func productUpdateParams(p *domain.Product) sqlcgen.UpdateProductParams {
	return sqlcgen.UpdateProductParams{
		ID:              p.ID,
		SystemUpdatedAt: p.UpdatedAt,
		Name:            p.Name,
		Category:        string(p.Category),
		PriceCents:      p.PriceCents,
	}
}

func orderToDomain(m sqlcgen.Order) *domain.Order {
	return &domain.Order{
		ID:             m.ID,
		OrganizationID: m.OrganizationID,
		CreatedAt:      m.SystemCreatedAt,
		UpdatedAt:      m.SystemUpdatedAt,
		Status:         domain.OrderStatus(m.Status),
	}
}

func orderCreateParams(o *domain.Order) sqlcgen.CreateOrderParams {
	return sqlcgen.CreateOrderParams{
		ID:              o.ID,
		OrganizationID:  o.OrganizationID,
		SystemCreatedAt: o.CreatedAt,
		SystemUpdatedAt: o.UpdatedAt,
		Status:          string(o.Status),
	}
}

func orderCloseParams(id uuid.UUID, updatedAt time.Time) sqlcgen.CloseOrderParams {
	return sqlcgen.CloseOrderParams{
		ID:              id,
		SystemUpdatedAt: updatedAt,
	}
}

func orderItemToDomain(m sqlcgen.OrderItem) domain.OrderItem {
	return domain.OrderItem{
		ID:             m.ID,
		OrganizationID: m.OrganizationID,
		OrderID:        m.OrderID,
		ProductID:      m.ProductID,
		CreatedAt:      m.SystemCreatedAt,
		Quantity:       m.Quantity,
		PriceCents:     m.PriceCents,
	}
}

func orderItemCreateParams(i *domain.OrderItem) sqlcgen.CreateOrderItemParams {
	return sqlcgen.CreateOrderItemParams{
		ID:              i.ID,
		OrganizationID:  i.OrganizationID,
		OrderID:         i.OrderID,
		ProductID:       i.ProductID,
		SystemCreatedAt: i.CreatedAt,
		Quantity:        i.Quantity,
		PriceCents:      i.PriceCents,
	}
}
