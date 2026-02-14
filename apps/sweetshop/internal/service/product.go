// Package service provides application services for the sweetshop.
package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	coredomain "github.com/bbsbb/go-edge/core/domain"
	"github.com/bbsbb/go-edge/sweetshop/internal/domain"
)

type ProductService struct {
	repo   domain.ProductRepository
	logger *slog.Logger
}

func NewProductService(repo domain.ProductRepository, logger *slog.Logger) *ProductService {
	return &ProductService{repo: repo, logger: logger}
}

func (s *ProductService) Create(ctx context.Context, name string, category domain.ProductCategory, priceCents int32) (*domain.Product, error) {
	if !category.IsValid() {
		return nil, coredomain.NewError(coredomain.CodeValidation, "invalid product category")
	}
	if priceCents <= 0 {
		return nil, coredomain.NewError(coredomain.CodeValidation, "price must be positive")
	}

	org, err := coredomain.OrganizationFromContext(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	product := &domain.Product{
		ID:             uuid.Must(uuid.NewV7()),
		OrganizationID: org.ID,
		CreatedAt:      now,
		UpdatedAt:      now,
		Name:           name,
		Category:       category,
		PriceCents:     priceCents,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		s.logger.Error("failed to create product", "error", err, "name", name)
		return nil, err
	}

	s.logger.Info("product created", "product_id", product.ID, "name", name, "category", category)
	return product, nil
}

func (s *ProductService) Get(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.Debug("product lookup failed", "error", err, "product_id", id)
		return nil, err
	}
	return product, nil
}

func (s *ProductService) List(ctx context.Context) ([]*domain.Product, error) {
	products, err := s.repo.List(ctx)
	if err != nil {
		s.logger.Error("failed to list products", "error", err)
		return nil, err
	}
	return products, nil
}

func (s *ProductService) Update(ctx context.Context, id uuid.UUID, name string, category domain.ProductCategory, priceCents int32) (*domain.Product, error) {
	if !category.IsValid() {
		return nil, coredomain.NewError(coredomain.CodeValidation, "invalid product category")
	}
	if priceCents <= 0 {
		return nil, coredomain.NewError(coredomain.CodeValidation, "price must be positive")
	}

	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	product.Name = name
	product.Category = category
	product.PriceCents = priceCents
	product.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, product); err != nil {
		s.logger.Error("failed to update product", "error", err, "product_id", id)
		return nil, err
	}

	s.logger.Info("product updated", "product_id", id, "name", name, "category", category)
	return product, nil
}

func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete product", "error", err, "product_id", id)
		return err
	}
	s.logger.Info("product deleted", "product_id", id)
	return nil
}
