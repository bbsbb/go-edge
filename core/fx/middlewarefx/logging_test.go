package middlewarefx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	coretesting "github.com/bbsbb/go-edge/core/testing"
)

type LoggingSuite struct {
	suite.Suite
}

func (s *LoggingSuite) TestRequestLogger_LogsExpectedFields() {
	capture := coretesting.NewLogCapture(&s.Suite)

	ctx := context.WithValue(context.Background(), correlationIDKey{}, "test-corr-id")

	handler := RequestLogger(capture.Logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	output := capture.Output()
	s.Assert().Contains(output, "http request")
	s.Assert().Contains(output, "method=GET")
	s.Assert().Contains(output, "path=/health")
	s.Assert().Contains(output, "status=200")
	s.Assert().Contains(output, "duration_ms=")
	s.Assert().Contains(output, "correlation_id=test-corr-id")
}

func TestLoggingSuite(t *testing.T) {
	suite.Run(t, new(LoggingSuite))
}
