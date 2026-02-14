// Package handler provides HTTP handlers for the sweetshop service.
package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	coredomain "github.com/bbsbb/go-edge/core/domain"
	transporthttp "github.com/bbsbb/go-edge/core/transport/http"
	"github.com/bbsbb/go-edge/sweetshop/internal/domain"
	"github.com/bbsbb/go-edge/sweetshop/internal/service"
	"github.com/bbsbb/go-edge/sweetshop/internal/transport/http/dto"
)

type ProductHandler struct {
	services *service.Registry
	logger   *slog.Logger
}

func NewProductHandler(services *service.Registry, logger *slog.Logger) *ProductHandler {
	return &ProductHandler{services: services, logger: logger}
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	products, err := h.services.Products.List(r.Context())
	if err != nil {
		transporthttp.WriteError(w, r, err, h.logger)
		return
	}
	transporthttp.RenderListOrLog(w, r, dto.ProductListToResponse(products), h.logger)
}

func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := coredomain.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		transporthttp.WriteError(w, r, err, h.logger)
		return
	}

	product, err := h.services.Products.Get(r.Context(), id.UUID())
	if err != nil {
		transporthttp.WriteError(w, r, err, h.logger)
		return
	}
	render.Status(r, http.StatusOK)
	transporthttp.RenderOrLog(w, r, dto.ProductToResponse(product), h.logger)
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProductRequest
	if err := render.Bind(r, &req); err != nil {
		transporthttp.WriteError(w, r, coredomain.NewError(coredomain.CodeValidation, "invalid request body"), h.logger)
		return
	}

	product, err := h.services.Products.Create(r.Context(), req.Name, domain.ProductCategory(req.Category), req.PriceCents)
	if err != nil {
		transporthttp.WriteError(w, r, err, h.logger)
		return
	}
	render.Status(r, http.StatusCreated)
	transporthttp.RenderOrLog(w, r, dto.ProductToResponse(product), h.logger)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := coredomain.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		transporthttp.WriteError(w, r, err, h.logger)
		return
	}

	var req dto.UpdateProductRequest
	if err := render.Bind(r, &req); err != nil {
		transporthttp.WriteError(w, r, coredomain.NewError(coredomain.CodeValidation, "invalid request body"), h.logger)
		return
	}

	product, err := h.services.Products.Update(r.Context(), id.UUID(), req.Name, domain.ProductCategory(req.Category), req.PriceCents)
	if err != nil {
		transporthttp.WriteError(w, r, err, h.logger)
		return
	}
	render.Status(r, http.StatusOK)
	transporthttp.RenderOrLog(w, r, dto.ProductToResponse(product), h.logger)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := coredomain.ParseID(chi.URLParam(r, "id"))
	if err != nil {
		transporthttp.WriteError(w, r, err, h.logger)
		return
	}

	if err := h.services.Products.Delete(r.Context(), id.UUID()); err != nil {
		transporthttp.WriteError(w, r, err, h.logger)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
