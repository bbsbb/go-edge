// Package middlewarefx provides a default HTTP middleware stack for all applications.
package middlewarefx

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const CorrelationIDHeader = "X-Correlation-ID"

type correlationIDKey struct{}

// CorrelationID returns a middleware that reads the correlation ID from the
// configured request header. If absent, it generates a new UUID. The value is
// stored in context and set on the response header. When cfg.Header is empty,
// it falls back to CorrelationIDHeader.
func CorrelationID(cfg CorrelationIDConfig) func(http.Handler) http.Handler {
	header := cfg.Header
	if header == "" {
		header = CorrelationIDHeader
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(header)
			if id == "" {
				id = uuid.Must(uuid.NewV7()).String()
			}

			ctx := context.WithValue(r.Context(), correlationIDKey{}, id)
			if span := trace.SpanFromContext(ctx); span.IsRecording() {
				span.SetAttributes(attribute.String("correlation.id", id))
			}
			w.Header().Set(header, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CorrelationIDFromContext extracts the correlation ID from context.
func CorrelationIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDKey{}).(string); ok {
		return id
	}
	return ""
}
