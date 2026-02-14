package middlewarefx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type CorrelationSuite struct {
	suite.Suite
}

func (s *CorrelationSuite) TestGeneratesUUID_WhenNoHeader() {
	var capturedID string
	handler := CorrelationID(CorrelationIDConfig{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = CorrelationIDFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	s.Assert().NotEmpty(capturedID)
	s.Assert().Equal(capturedID, rec.Header().Get(CorrelationIDHeader))
}

func (s *CorrelationSuite) TestReusesExistingHeader() {
	existing := "existing-corr-id"
	var capturedID string
	handler := CorrelationID(CorrelationIDConfig{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = CorrelationIDFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(CorrelationIDHeader, existing)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	s.Assert().Equal(existing, capturedID)
	s.Assert().Equal(existing, rec.Header().Get(CorrelationIDHeader))
}

func (s *CorrelationSuite) TestSetsSpanAttribute() {
	span := &recordingSpan{}
	ctx := trace.ContextWithSpan(context.Background(), span)

	var capturedID string
	handler := CorrelationID(CorrelationIDConfig{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = CorrelationIDFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	s.Assert().NotEmpty(capturedID)
	s.Assert().Contains(span.attrs, attribute.String("correlation.id", capturedID))
}

func (s *CorrelationSuite) TestCustomHeader() {
	customHeader := "X-Custom-Trace"
	existing := "custom-trace-123"

	var capturedID string
	handler := CorrelationID(CorrelationIDConfig{Header: customHeader})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = CorrelationIDFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(customHeader, existing)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	s.Assert().Equal(existing, capturedID)
	s.Assert().Equal(existing, rec.Header().Get(customHeader))
	s.Assert().Empty(rec.Header().Get(CorrelationIDHeader))
}

func (s *CorrelationSuite) TestFromContext_Empty() {
	s.Assert().Empty(CorrelationIDFromContext(context.Background()))
}

func TestCorrelationSuite(t *testing.T) {
	suite.Run(t, new(CorrelationSuite))
}
