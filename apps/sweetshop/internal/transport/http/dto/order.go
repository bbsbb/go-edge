package dto

import (
	transporthttp "github.com/bbsbb/go-edge/core/transport/http"
	"github.com/bbsbb/go-edge/sweetshop/internal/domain"
)

type AddOrderItemRequest struct {
	transporthttp.NoOpBinder
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
}

type OrderItemResponse struct {
	ID             string `json:"id"`
	ProductID      string `json:"product_id"`
	Quantity       int32  `json:"quantity"`
	PriceCents     int32  `json:"price_cents"`
	LineTotalCents int32  `json:"line_total_cents"`
}

type OrderResponse struct {
	transporthttp.NoOpRenderer
	ID         string              `json:"id"`
	Status     string              `json:"status"`
	Items      []OrderItemResponse `json:"items"`
	TotalCents int32               `json:"total_cents"`
}

func OrderToResponse(o *domain.Order) *OrderResponse {
	items := make([]OrderItemResponse, len(o.Items))
	for i, item := range o.Items {
		items[i] = OrderItemResponse{
			ID:             item.ID.String(),
			ProductID:      item.ProductID.String(),
			Quantity:       item.Quantity,
			PriceCents:     item.PriceCents,
			LineTotalCents: item.LineTotalCents(),
		}
	}
	return &OrderResponse{
		ID:         o.ID.String(),
		Status:     string(o.Status),
		Items:      items,
		TotalCents: o.TotalCents(),
	}
}

type OrderItemCreatedResponse struct {
	transporthttp.NoOpRenderer
	ID             string `json:"id"`
	ProductID      string `json:"product_id"`
	Quantity       int32  `json:"quantity"`
	PriceCents     int32  `json:"price_cents"`
	LineTotalCents int32  `json:"line_total_cents"`
}

func OrderItemToCreatedResponse(item *domain.OrderItem) *OrderItemCreatedResponse {
	return &OrderItemCreatedResponse{
		ID:             item.ID.String(),
		ProductID:      item.ProductID.String(),
		Quantity:       item.Quantity,
		PriceCents:     item.PriceCents,
		LineTotalCents: item.LineTotalCents(),
	}
}
