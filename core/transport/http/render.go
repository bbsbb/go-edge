package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
)

// NoOpBinder is an embeddable type that satisfies render.Binder with a no-op.
type NoOpBinder struct{}

func (NoOpBinder) Bind(*http.Request) error { return nil }

// NoOpRenderer is an embeddable type that satisfies render.Renderer with a no-op.
type NoOpRenderer struct{}

func (NoOpRenderer) Render(http.ResponseWriter, *http.Request) error { return nil }

// RenderOrLog renders v and logs encoding failures instead of silently discarding them.
func RenderOrLog(w http.ResponseWriter, r *http.Request, v render.Renderer, logger *slog.Logger) {
	if err := render.Render(w, r, v); err != nil {
		logger.ErrorContext(r.Context(), "failed to render response", "error", err)
	}
}

// RenderListOrLog renders a list and logs encoding failures instead of silently discarding them.
func RenderListOrLog(w http.ResponseWriter, r *http.Request, v []render.Renderer, logger *slog.Logger) {
	if err := render.RenderList(w, r, v); err != nil {
		logger.ErrorContext(r.Context(), "failed to render response list", "error", err)
	}
}
