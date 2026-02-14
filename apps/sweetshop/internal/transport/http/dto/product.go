// Package dto provides request and response types for the HTTP transport layer.
package dto

import (
	"github.com/go-chi/render"

	transporthttp "github.com/bbsbb/go-edge/core/transport/http"
	"github.com/bbsbb/go-edge/sweetshop/internal/domain"
)

type CreateProductRequest struct {
	transporthttp.NoOpBinder
	Name       string `json:"name"`
	Category   string `json:"category"`
	PriceCents int32  `json:"price_cents"`
}

type UpdateProductRequest struct {
	transporthttp.NoOpBinder
	Name       string `json:"name"`
	Category   string `json:"category"`
	PriceCents int32  `json:"price_cents"`
}

type ProductResponse struct {
	transporthttp.NoOpRenderer
	ID         string `json:"id"`
	Name       string `json:"name"`
	Category   string `json:"category"`
	PriceCents int32  `json:"price_cents"`
}

func ProductToResponse(p *domain.Product) *ProductResponse {
	return &ProductResponse{
		ID:         p.ID.String(),
		Name:       p.Name,
		Category:   string(p.Category),
		PriceCents: p.PriceCents,
	}
}

func ProductListToResponse(products []*domain.Product) []render.Renderer {
	list := make([]render.Renderer, len(products))
	for i, p := range products {
		list[i] = ProductToResponse(p)
	}
	return list
}
