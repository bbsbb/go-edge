package middlewarefx

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"

	"github.com/bbsbb/go-edge/core/fx/otelfx"
)

// Middleware is an application-provided middleware that gets appended to the
// core stack. Apps supply instances via the "middleware" FX value group.
type Middleware struct {
	Name    string
	Handler func(http.Handler) http.Handler
}

// RequestID wraps Chi's RequestID middleware to additionally set the
// request.id attribute on the active OTel span.
func RequestID(next http.Handler) http.Handler {
	return chimw.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if span := trace.SpanFromContext(r.Context()); span.IsRecording() {
			if reqID := chimw.GetReqID(r.Context()); reqID != "" {
				span.SetAttributes(attribute.String("request.id", reqID))
			}
		}
		next.ServeHTTP(w, r)
	}))
}

type params struct {
	fx.In

	Mux        *chi.Mux
	Logger     *slog.Logger
	Config     *Configuration
	OTelConfig *otelfx.Configuration `optional:"true"`
	Extra      []Middleware          `group:"middleware"`
}

func register(p params) {
	if p.Config.Recovery.Enabled {
		p.Mux.Use(Recovery(p.Logger))
	}
	if p.Config.MaxBytes.Enabled {
		p.Mux.Use(MaxBytes(p.Config.MaxBytes.MaxBytes))
	}
	if p.Config.RequestID.Enabled {
		p.Mux.Use(RequestID)
	}
	if p.Config.CorrelationID.Enabled {
		p.Mux.Use(CorrelationID(p.Config.CorrelationID))
	}
	if p.Config.OTelHTTP.Enabled && p.OTelConfig != nil {
		p.Mux.Use(otelhttp.NewMiddleware(p.OTelConfig.ServiceName))
	}
	if p.Config.RequestLog.Enabled {
		p.Mux.Use(RequestLogger(p.Logger))
	}
	for _, mw := range p.Extra {
		p.Mux.Use(mw.Handler)
	}
}

var Module = fx.Module(
	"middleware",
	fx.Provide(provideConfiguration),
	fx.Invoke(register),
)
