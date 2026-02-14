package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/bbsbb/go-edge/core/domain"
	coretesting "github.com/bbsbb/go-edge/core/testing"
	transporthttp "github.com/bbsbb/go-edge/core/transport/http"
)

type mockOrganizationLoader struct {
	orgs map[string]*domain.Organization
}

func (m *mockOrganizationLoader) LoadOrganizationBySlug(_ context.Context, slug string) (*domain.Organization, error) {
	org, ok := m.orgs[slug]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return org, nil
}

type OrganizationSuite struct {
	suite.Suite
	logger *slog.Logger
	loader *mockOrganizationLoader
	orgID  uuid.UUID
}

func (s *OrganizationSuite) SetupTest() {
	s.orgID = uuid.Must(uuid.NewV7())
	s.logger = coretesting.NewNoopLogger()
	s.loader = &mockOrganizationLoader{
		orgs: map[string]*domain.Organization{
			"acme": {ID: s.orgID, Slug: "acme"},
		},
	}
}

func (s *OrganizationSuite) TestExtractSlug() {
	tests := []struct {
		name           string
		host           string
		xForwardedHost string
		slugHeader     string
		expectedSlug   string
	}{
		{
			name:         "prefers X-Organization-Slug header",
			host:         "localhost:8080",
			slugHeader:   "acme",
			expectedSlug: "acme",
		},
		{
			name:         "X-Organization-Slug takes precedence over subdomain",
			host:         "other.example.com",
			slugHeader:   "acme",
			expectedSlug: "acme",
		},
		{
			name:         "falls back to subdomain from host",
			host:         "acme.example.com",
			expectedSlug: "acme",
		},
		{
			name:           "falls back to subdomain from x-forwarded-host",
			host:           "localhost:8080",
			xForwardedHost: "acme.example.com",
			expectedSlug:   "acme",
		},
		{
			name:         "handles localhost without subdomain",
			host:         "localhost:8080",
			expectedSlug: "localhost:8080",
		},
		{
			name:         "handles single part host",
			host:         "acme",
			expectedSlug: "acme",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Host = tt.host
			if tt.xForwardedHost != "" {
				req.Header.Set("X-Forwarded-Host", tt.xForwardedHost)
			}
			if tt.slugHeader != "" {
				req.Header.Set("X-Organization-Slug", tt.slugHeader)
			}

			s.Assert().Equal(tt.expectedSlug, extractSlug(req))
		})
	}
}

func (s *OrganizationSuite) TestAddsOrganizationToContext() {
	var capturedOrg *domain.Organization

	handler := WithOrganization(WithOrganizationConfig{
		Logger: s.logger,
		Loader: s.loader,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		org, err := domain.OrganizationFromContext(r.Context())
		s.Require().NoError(err)
		capturedOrg = org
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Host = "acme.example.com"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Assert().Equal(http.StatusOK, rec.Code)
	s.Require().NotNil(capturedOrg)
	s.Assert().Equal(s.orgID, capturedOrg.ID)
	s.Assert().Equal("acme", capturedOrg.Slug)
}

func (s *OrganizationSuite) TestResolvesOrganizationFromSlugHeader() {
	var capturedOrg *domain.Organization

	handler := WithOrganization(WithOrganizationConfig{
		Logger: s.logger,
		Loader: s.loader,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		org, _ := domain.OrganizationFromContext(r.Context())
		capturedOrg = org
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Host = "localhost:8080"
	req.Header.Set("X-Organization-Slug", "acme")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Assert().Equal(http.StatusOK, rec.Code)
	s.Require().NotNil(capturedOrg)
	s.Assert().Equal("acme", capturedOrg.Slug)
}

func (s *OrganizationSuite) TestReturnsProblemDetailsForUnknownOrganization() {
	handler := WithOrganization(WithOrganizationConfig{
		Logger: s.logger,
		Loader: s.loader,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Host = "unknown.example.com"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Assert().Equal(http.StatusNotFound, rec.Code)
	s.Assert().Equal("application/problem+json", rec.Header().Get("Content-Type"))

	var errResp transporthttp.ErrorShape
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&errResp))
	s.Assert().Equal("organization not found", errResp.Detail)
	s.Assert().Equal(string(domain.CodeNotFound), errResp.Code)
}

func (s *OrganizationSuite) TestSkipsConfiguredPaths() {
	handler := WithOrganization(WithOrganizationConfig{
		Logger:    s.logger,
		Loader:    s.loader,
		SkipPaths: []string{"/healthcheck"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := domain.OrganizationFromContext(r.Context())
		s.Assert().ErrorIs(err, domain.ErrMissingOrganization)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
	req.Host = "unknown.example.com"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Assert().Equal(http.StatusOK, rec.Code)
}

func (s *OrganizationSuite) TestUsesXForwardedHostOverHost() {
	var capturedOrg *domain.Organization

	handler := WithOrganization(WithOrganizationConfig{
		Logger: s.logger,
		Loader: s.loader,
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		org, _ := domain.OrganizationFromContext(r.Context())
		capturedOrg = org
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Host = "localhost:8080"
	req.Header.Set("X-Forwarded-Host", "acme.example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	s.Assert().Equal(http.StatusOK, rec.Code)
	s.Require().NotNil(capturedOrg)
	s.Assert().Equal("acme", capturedOrg.Slug)
}

func (s *OrganizationSuite) TestContextReturnsErrorWhenNotSet() {
	_, err := domain.OrganizationFromContext(context.Background())
	s.Assert().ErrorIs(err, domain.ErrMissingOrganization)
}

func (s *OrganizationSuite) TestContextReturnsOrganizationWhenSet() {
	orgID := uuid.Must(uuid.NewV7())
	org := &domain.Organization{ID: orgID, Slug: "test"}
	ctx := domain.ContextWithOrganization(context.Background(), org)

	result, err := domain.OrganizationFromContext(ctx)
	s.Require().NoError(err)
	s.Assert().Equal(orgID, result.ID)
	s.Assert().Equal("test", result.Slug)
}

func TestOrganizationSuite(t *testing.T) {
	suite.Run(t, new(OrganizationSuite))
}
