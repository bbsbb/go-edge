package service

type Registry struct {
	Products *ProductService
	Orders   *OrderService
}

func NewRegistry(products *ProductService, orders *OrderService) *Registry {
	return &Registry{Products: products, Orders: orders}
}
