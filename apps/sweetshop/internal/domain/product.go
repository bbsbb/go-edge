// Package domain contains the sweetshop business types and interfaces.
package domain

import (
	"slices"
	"time"

	"github.com/google/uuid"
)

type ProductCategory string

const (
	ProductCategoryIceCream    ProductCategory = "ice_cream"
	ProductCategoryMarshmallow ProductCategory = "marshmallow"
)

func (c ProductCategory) IsValid() bool {
	return slices.Contains([]ProductCategory{ProductCategoryIceCream, ProductCategoryMarshmallow}, c)
}

type Product struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Name           string
	Category       ProductCategory
	PriceCents     int32
}
