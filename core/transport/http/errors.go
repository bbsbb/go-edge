// Package http provides HTTP transport utilities for domain error response writing,
// health check handlers, and secure cookie construction.
package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"github.com/bbsbb/go-edge/core/domain"
)

// ErrorShape is the RFC 9457-inspired problem details response for all error responses.
// It implements render.Renderer so applications can use render.Render(w, r, shape)
// directly for custom error cases where application/problem+json content type
// negotiation is handled separately.
type ErrorShape struct {
	Status    int      `json:"status"`
	Code      string   `json:"code,omitempty"`
	Detail    string   `json:"detail,omitempty"`
	Instance  string   `json:"instance,omitempty"`
	RequestID string   `json:"request_id,omitempty"`
	Errors    []string `json:"errors,omitempty"`
}

func (p *ErrorShape) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, p.Status)
	return nil
}

var statusFromCode = map[domain.Code]int{
	domain.CodeNotFound:   http.StatusNotFound,
	domain.CodeConflict:   http.StatusConflict,
	domain.CodeValidation: http.StatusBadRequest,
	domain.CodeForbidden:  http.StatusForbidden,
	domain.CodeInvariant:  http.StatusUnprocessableEntity,
}

// WriteError translates a domain error (or any error) into an RFC 9457 problem details response.
// The logger parameter is used for logging unhandled (non-domain) errors and encoding failures.
func WriteError(w http.ResponseWriter, r *http.Request, err error, logger *slog.Logger) {
	var domainErr *domain.Error
	if errors.As(err, &domainErr) {
		status, ok := statusFromCode[domainErr.Code]
		if !ok {
			status = http.StatusInternalServerError
		}

		resp := &ErrorShape{
			Status:    status,
			Code:      string(domainErr.Code),
			Detail:    domainErr.Message,
			Instance:  r.URL.Path,
			RequestID: chimw.GetReqID(r.Context()),
		}

		if unwrapper, ok := domainErr.Err.(interface{ Unwrap() []error }); ok {
			for _, e := range unwrapper.Unwrap() {
				resp.Errors = append(resp.Errors, e.Error())
			}
		}

		renderProblemJSON(w, r, resp, logger)
		return
	}

	logger.ErrorContext(r.Context(), "unhandled error", "error", err)
	renderProblemJSON(w, r, &ErrorShape{
		Status:    http.StatusInternalServerError,
		Instance:  r.URL.Path,
		RequestID: chimw.GetReqID(r.Context()),
	}, logger)
}

func renderProblemJSON(w http.ResponseWriter, r *http.Request, resp *ErrorShape, logger *slog.Logger) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(resp.Status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.ErrorContext(r.Context(), "failed to encode error response", "error", err)
	}
}
