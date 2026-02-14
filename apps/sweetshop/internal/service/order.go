package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	coredomain "github.com/bbsbb/go-edge/core/domain"
	"github.com/bbsbb/go-edge/sweetshop/internal/domain"
)

type OrderService struct {
	orders   domain.OrderRepository
	products domain.ProductRepository
	logger   *slog.Logger
}

func NewOrderService(orders domain.OrderRepository, products domain.ProductRepository, logger *slog.Logger) *OrderService {
	return &OrderService{orders: orders, products: products, logger: logger}
}

func (s *OrderService) OpenOrder(ctx context.Context) (*domain.Order, error) {
	org, err := coredomain.OrganizationFromContext(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	order := &domain.Order{
		ID:             uuid.Must(uuid.NewV7()),
		OrganizationID: org.ID,
		CreatedAt:      now,
		UpdatedAt:      now,
		Status:         domain.OrderStatusOpen,
	}

	if err := s.orders.Create(ctx, order); err != nil {
		s.logger.Error("failed to create order", "error", err)
		return nil, err
	}

	s.logger.Info("order opened", "order_id", order.ID)
	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	order, err := s.orders.FindByID(ctx, id)
	if err != nil {
		s.logger.Debug("order lookup failed", "error", err, "order_id", id)
		return nil, err
	}
	return order, nil
}

func (s *OrderService) AddItem(ctx context.Context, orderID uuid.UUID, productID uuid.UUID, quantity int32) (*domain.OrderItem, error) {
	if quantity <= 0 {
		return nil, coredomain.NewError(coredomain.CodeValidation, "quantity must be positive")
	}

	order, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if err := order.CanAddItem(); err != nil {
		s.logger.Warn("attempted to add item to closed order", "order_id", orderID)
		return nil, err
	}

	product, err := s.products.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	item := &domain.OrderItem{
		ID:             uuid.Must(uuid.NewV7()),
		OrganizationID: order.OrganizationID,
		OrderID:        orderID,
		ProductID:      productID,
		CreatedAt:      time.Now(),
		Quantity:       quantity,
		PriceCents:     product.PriceCents,
	}

	if err := s.orders.CreateItem(ctx, item); err != nil {
		s.logger.Error("failed to add order item", "error", err, "order_id", orderID, "product_id", productID)
		return nil, err
	}

	s.logger.Info("item added to order",
		"order_id", orderID, "product_id", productID,
		"quantity", quantity, "price_cents", product.PriceCents)
	return item, nil
}

func (s *OrderService) CloseOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	order, err := s.orders.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := order.CanClose(); err != nil {
		s.logger.Warn("attempted to close already-closed order", "order_id", id)
		return nil, err
	}

	if _, err := s.orders.Close(ctx, id, time.Now()); err != nil {
		s.logger.Error("failed to close order", "error", err, "order_id", id)
		return nil, err
	}

	s.logger.Info("order closed", "order_id", id)
	return s.orders.FindByID(ctx, id)
}
