package middlewarefx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	coretesting "github.com/bbsbb/go-edge/core/testing"
	transporthttp "github.com/bbsbb/go-edge/core/transport/http"
)

type RecoverySuite struct {
	suite.Suite
}

func (s *RecoverySuite) TestReturns500OnPanic() {
	handler := Recovery(coretesting.NewNoopLogger())(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("something broke")
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	handler.ServeHTTP(rec, req)

	s.Assert().Equal(http.StatusInternalServerError, rec.Code)
	s.Assert().Equal("application/problem+json", rec.Header().Get("Content-Type"))

	var resp transporthttp.ErrorShape
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&resp))
	s.Assert().Equal(http.StatusInternalServerError, resp.Status)
	s.Assert().Empty(resp.Code)
	s.Assert().Equal("/test", resp.Instance)
}

func (s *RecoverySuite) TestPassesThroughWithoutPanic() {
	handler := Recovery(coretesting.NewNoopLogger())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	s.Assert().Equal(http.StatusOK, rec.Code)
}

func (s *RecoverySuite) TestLogsStackTrace() {
	capture := coretesting.NewLogCapture(&s.Suite)
	handler := Recovery(capture.Logger)(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("test panic")
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	output := capture.Output()
	s.Assert().Contains(output, "panic recovered")
	s.Assert().Contains(output, "test panic")
	s.Assert().Contains(output, "goroutine")
}

func TestRecoverySuite(t *testing.T) {
	suite.Run(t, new(RecoverySuite))
}
