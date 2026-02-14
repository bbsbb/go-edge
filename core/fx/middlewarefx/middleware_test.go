package middlewarefx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/bbsbb/go-edge/core/fx/otelfx"
	coretesting "github.com/bbsbb/go-edge/core/testing"
)

type MiddlewareSuite struct {
	suite.Suite
}

type recordingSpan struct {
	noop.Span
	attrs []attribute.KeyValue
}

func (s *recordingSpan) IsRecording() bool { return true }

func (s *recordingSpan) SetAttributes(attrs ...attribute.KeyValue) {
	s.attrs = append(s.attrs, attrs...)
}

func (s *MiddlewareSuite) TestRequestID_SetsSpanAttribute() {
	span := &recordingSpan{}
	ctx := trace.ContextWithSpan(context.Background(), span)

	var capturedID string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = chimw.GetReqID(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	s.Assert().NotEmpty(capturedID)
	s.Assert().Contains(span.attrs, attribute.String("request.id", capturedID))
}

func (s *MiddlewareSuite) TestRequestID_SkipsWhenNotRecording() {
	noopSpan := &noop.Span{}
	ctx := trace.ContextWithSpan(context.Background(), noopSpan)

	var called bool
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	s.Assert().True(called)
}

func (s *MiddlewareSuite) TestRegister_DefaultConfig_AllMiddlewares() {
	mux := chi.NewMux()
	logger := coretesting.NewNoopLogger()
	cfg := DefaultConfiguration()
	otelCfg := &otelfx.Configuration{ServiceName: "test-svc"}

	register(params{
		Mux:        mux,
		Logger:     logger,
		Config:     &cfg,
		OTelConfig: otelCfg,
	})

	s.Assert().Len(mux.Middlewares(), 6)
}

func (s *MiddlewareSuite) TestRegister_DisableRequestLogging() {
	mux := chi.NewMux()
	logger := coretesting.NewNoopLogger()
	cfg := DefaultConfiguration()
	cfg.RequestLog.Enabled = false
	otelCfg := &otelfx.Configuration{ServiceName: "test-svc"}

	register(params{
		Mux:        mux,
		Logger:     logger,
		Config:     &cfg,
		OTelConfig: otelCfg,
	})

	s.Assert().Len(mux.Middlewares(), 5)
}

func (s *MiddlewareSuite) TestRegister_DisableRecovery() {
	mux := chi.NewMux()
	logger := coretesting.NewNoopLogger()
	cfg := DefaultConfiguration()
	cfg.Recovery.Enabled = false
	otelCfg := &otelfx.Configuration{ServiceName: "test-svc"}

	register(params{
		Mux:        mux,
		Logger:     logger,
		Config:     &cfg,
		OTelConfig: otelCfg,
	})

	s.Assert().Len(mux.Middlewares(), 5)
}

func (s *MiddlewareSuite) TestRegister_DisableMaxBytes() {
	mux := chi.NewMux()
	logger := coretesting.NewNoopLogger()
	cfg := DefaultConfiguration()
	cfg.MaxBytes.Enabled = false
	otelCfg := &otelfx.Configuration{ServiceName: "test-svc"}

	register(params{
		Mux:        mux,
		Logger:     logger,
		Config:     &cfg,
		OTelConfig: otelCfg,
	})

	s.Assert().Len(mux.Middlewares(), 5)
}

func (s *MiddlewareSuite) TestRegister_NilOTelConfig_NoPanic() {
	mux := chi.NewMux()
	logger := coretesting.NewNoopLogger()
	cfg := DefaultConfiguration()

	s.Assert().NotPanics(func() {
		register(params{
			Mux:        mux,
			Logger:     logger,
			Config:     &cfg,
			OTelConfig: nil,
		})
	})

	s.Assert().Len(mux.Middlewares(), 5)
}

func (s *MiddlewareSuite) TestRegister_ExtraMiddleware() {
	mux := chi.NewMux()
	logger := coretesting.NewNoopLogger()
	cfg := DefaultConfiguration()

	var called bool
	extra := Middleware{
		Name: "test-extra",
		Handler: func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				next.ServeHTTP(w, r)
			})
		},
	}

	register(params{
		Mux:    mux,
		Logger: logger,
		Config: &cfg,
		Extra:  []Middleware{extra},
	})

	s.Assert().Len(mux.Middlewares(), 6)

	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	s.Assert().True(called)
}

func TestMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareSuite))
}
