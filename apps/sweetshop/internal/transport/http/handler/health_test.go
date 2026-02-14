//go:build testing

package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	transporthttp "github.com/bbsbb/go-edge/core/transport/http"
)

type HealthSuite struct {
	IntegrationSuite
}

func (s *HealthSuite) TestLivenessReturns200() {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	s.Router.ServeHTTP(rec, req)

	s.Assert().Equal(http.StatusOK, rec.Code)

	var resp transporthttp.HealthResponse
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&resp))
	s.Assert().Equal(transporthttp.HealthStatusPass, resp.Status)
}

func (s *HealthSuite) TestReadinessReturns200() {
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	s.Router.ServeHTTP(rec, req)

	s.Assert().Equal(http.StatusOK, rec.Code)

	var resp transporthttp.HealthResponse
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&resp))
	s.Assert().Equal(transporthttp.HealthStatusPass, resp.Status)
	s.Require().Contains(resp.Checks, "postgres")
	s.Assert().Equal(transporthttp.HealthStatusPass, resp.Checks["postgres"][0].Status)
}

func TestHealthSuite(t *testing.T) {
	suite.Run(t, new(HealthSuite))
}
