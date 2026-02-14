//go:build testing

package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	coretesting "github.com/bbsbb/go-edge/core/testing"
)

type HealthProbeSuite struct {
	suite.Suite
}

func (s *HealthProbeSuite) TestLivenessReturnsPass() {
	handler := LivenessHandler()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	handler.ServeHTTP(rr, req)

	s.Assert().Equal(http.StatusOK, rr.Code)
	s.Assert().Equal("application/json", rr.Header().Get("Content-Type"))

	var resp HealthResponse
	s.Require().NoError(json.NewDecoder(rr.Body).Decode(&resp))
	s.Assert().Equal(HealthStatusPass, resp.Status)
}

func (s *HealthProbeSuite) TestReadinessFailsWithNilPool() {
	handler := ReadinessHandler(nil, 0, coretesting.NewNoopLogger())
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	handler.ServeHTTP(rr, req)

	s.Assert().Equal(http.StatusServiceUnavailable, rr.Code)
	s.Assert().Equal("application/json", rr.Header().Get("Content-Type"))

	var resp HealthResponse
	s.Require().NoError(json.NewDecoder(rr.Body).Decode(&resp))
	s.Assert().Equal(HealthStatusFail, resp.Status)
	s.Require().Contains(resp.Checks, "postgres")
	s.Assert().Equal(HealthStatusFail, resp.Checks["postgres"][0].Status)
	s.Assert().Equal(ComponentTypeDatastore, resp.Checks["postgres"][0].ComponentType)
}

func (s *HealthProbeSuite) TestReadinessCustomTimeout() {
	handler := ReadinessHandler(nil, 100*time.Millisecond, coretesting.NewNoopLogger())
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	handler.ServeHTTP(rr, req)

	s.Assert().Equal(http.StatusServiceUnavailable, rr.Code)
}

func TestHealthProbeSuite(t *testing.T) {
	suite.Run(t, new(HealthProbeSuite))
}
