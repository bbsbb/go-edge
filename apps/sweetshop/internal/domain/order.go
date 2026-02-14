package domain

import (
	"slices"
	"time"

	"github.com/google/uuid"

	coredomain "github.com/bbsbb/go-edge/core/domain"
)

type OrderStatus string

const (
	OrderStatusOpen   OrderStatus = "open"
	OrderStatusClosed OrderStatus = "closed"
)

func (s OrderStatus) IsValid() bool {
	return slices.Contains([]OrderStatus{OrderStatusOpen, OrderStatusClosed}, s)
}

type Order struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Status         OrderStatus
	Items          []OrderItem
}

func (o *Order) CanAddItem() error {
	if o.Status != OrderStatusOpen {
		return coredomain.NewError(coredomain.CodeInvariant, "cannot add items to a closed order")
	}
	return nil
}

func (o *Order) CanClose() error {
	if o.Status == OrderStatusClosed {
		return coredomain.NewError(coredomain.CodeInvariant, "order is already closed")
	}
	return nil
}

func (o *Order) TotalCents() int32 {
	var total int32
	for _, item := range o.Items {
		total += item.LineTotalCents()
	}
	return total
}

type OrderItem struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	OrderID        uuid.UUID
	ProductID      uuid.UUID
	CreatedAt      time.Time
	Quantity       int32
	PriceCents     int32
}

func (i *OrderItem) LineTotalCents() int32 {
	return i.Quantity * i.PriceCents
}
