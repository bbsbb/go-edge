package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	coredomain "github.com/bbsbb/go-edge/core/domain"
	transporthttp "github.com/bbsbb/go-edge/core/transport/http"
	"github.com/bbsbb/go-edge/sweetshop/internal/service"
	"github.com/bbsbb/go-edge/sweetshop/internal/transport/http/dto"
)

type OrderHandler struct {
	services *service.Registry
	logger   *slog.Logger
}

func NewOrderHandler(services *service.Registry, logger *slog.Logger) *OrderHandler {
	return &OrderHandler{services: services, logger: logger}
}

func (h *OrderHandler) Open(w http.ResponseWriter, r *http.Request) {
	order, err := h.services.Orders.OpenOrder(r.Context())
	if err != nil {
		transporthttp.WriteError(w, r, err, h.logger)
		return
	}
	render.Status(r, http.StatusCreated)
	transporthttp.RenderOrLog(w, r, dto.OrderToResponse(order), h.logger)
}

func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := coredomain.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		transporthttp.WriteError(w, r, err, h.logger)
		return
	}

	order, err := h.services.Orders.GetOrder(r.Context(), id.UUID())
	if err != nil {
		transporthttp.WriteError(w, r, err, h.logger)
		return
	}
	render.Status(r, http.StatusOK)
	transporthttp.RenderOrLog(w, r, dto.OrderToResponse(order), h.logger)
}

func (h *OrderHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	orderID, err := coredomain.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		transporthttp.WriteError(w, r, err, h.logger)
		return
	}

	var req dto.AddOrderItemRequest
	if err := render.Bind(r, &req); err != nil {
		transporthttp.WriteError(w, r, coredomain.NewError(coredomain.CodeValidation, "invalid request body"), h.logger)
		return
	}

	productID, err := coredomain.ParseID(req.ProductID)
	if err != nil {
		transporthttp.WriteError(w, r, err, h.logger)
		return
	}

	item, err := h.services.Orders.AddItem(r.Context(), orderID.UUID(), productID.UUID(), req.Quantity)
	if err != nil {
		transporthttp.WriteError(w, r, err, h.logger)
		return
	}
	render.Status(r, http.StatusCreated)
	transporthttp.RenderOrLog(w, r, dto.OrderItemToCreatedResponse(item), h.logger)
}

func (h *OrderHandler) Close(w http.ResponseWriter, r *http.Request) {
	id, err := coredomain.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		transporthttp.WriteError(w, r, err, h.logger)
		return
	}

	order, err := h.services.Orders.CloseOrder(r.Context(), id.UUID())
	if err != nil {
		transporthttp.WriteError(w, r, err, h.logger)
		return
	}
	render.Status(r, http.StatusOK)
	transporthttp.RenderOrLog(w, r, dto.OrderToResponse(order), h.logger)
}
