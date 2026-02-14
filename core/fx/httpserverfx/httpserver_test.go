package httpserverfx

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/fx/fxtest"

	coretesting "github.com/bbsbb/go-edge/core/testing"
)

type HTTPServerSuite struct {
	suite.Suite
}

func (s *HTTPServerSuite) TestConfiguration_Validate() {
	tests := []struct {
		name       string
		config     Configuration
		wantErr    bool
		wantFields []string
	}{
		{
			name: "valid config",
			config: Configuration{
				Port:           8080,
				RequestTimeout: 30,
			},
			wantErr: false,
		},
		{
			name: "invalid port zero",
			config: Configuration{
				Port:           0,
				RequestTimeout: 30,
			},
			wantErr:    true,
			wantFields: []string{"Port"},
		},
		{
			name: "invalid request timeout zero",
			config: Configuration{
				Port:           8080,
				RequestTimeout: 0,
			},
			wantErr:    true,
			wantFields: []string{"RequestTimeout"},
		},
		{
			name: "invalid request timeout too high",
			config: Configuration{
				Port:           8080,
				RequestTimeout: 121,
			},
			wantErr:    true,
			wantFields: []string{"RequestTimeout"},
		},
		{
			name: "invalid read header timeout too high",
			config: Configuration{
				Port:              8080,
				RequestTimeout:    30,
				ReadHeaderTimeout: 121,
			},
			wantErr:    true,
			wantFields: []string{"ReadHeaderTimeout"},
		},
		{
			name: "invalid read timeout too high",
			config: Configuration{
				Port:           8080,
				RequestTimeout: 30,
				ReadTimeout:    121,
			},
			wantErr:    true,
			wantFields: []string{"ReadTimeout"},
		},
		{
			name: "invalid idle timeout too high",
			config: Configuration{
				Port:           8080,
				RequestTimeout: 30,
				IdleTimeout:    601,
			},
			wantErr:    true,
			wantFields: []string{"IdleTimeout"},
		},
		{
			name: "valid with custom timeouts",
			config: Configuration{
				Port:              8080,
				RequestTimeout:    30,
				ReadHeaderTimeout: 10,
				ReadTimeout:       15,
				IdleTimeout:       300,
			},
			wantErr: false,
		},
		{
			name: "invalid cors max age too high",
			config: Configuration{
				Port:           8080,
				RequestTimeout: 30,
				Cors:           CorsConfiguration{MaxAge: 86401},
			},
			wantErr:    true,
			wantFields: []string{"MaxAge"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := tt.config.Validate()
			if tt.wantErr {
				s.Require().Error(err)
				var validationErrs validator.ValidationErrors
				s.Require().ErrorAs(err, &validationErrs)

				fields := make([]string, len(validationErrs))
				for i, fe := range validationErrs {
					fields[i] = fe.Field()
				}
				s.Assert().ElementsMatch(tt.wantFields, fields)
			} else {
				s.Assert().NoError(err)
			}
		})
	}
}

type testAppConfig struct {
	httpConfig *Configuration
}

func (c *testAppConfig) HTTPServerConfiguration() *Configuration {
	return c.httpConfig
}

type IntegrationSuite struct {
	suite.Suite
	mux     *chi.Mux
	testApp *fxtest.App
}

func (s *IntegrationSuite) SetupTest() {
	s.testApp = fxtest.New(
		s.T(),
		fx.Supply(
			coretesting.NewNoopLogger(),
			fx.Annotate(
				&testAppConfig{
					httpConfig: &Configuration{
						Port:           15123,
						RequestTimeout: 1,
					},
				},
				fx.As(new(WithHTTPServer)),
			),
		),
		Module,
		fx.WithLogger(func() fxevent.Logger {
			return fxevent.NopLogger
		}),
		fx.Populate(&s.mux),
	)
	s.testApp.RequireStart()
}

func (s *IntegrationSuite) TearDownTest() {
	s.testApp.RequireStop()
}

func (s *IntegrationSuite) TestRequestOk() {
	s.mux.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, "OK")
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	s.mux.ServeHTTP(rr, req)

	s.Assert().Equal(http.StatusOK, rr.Code)
	s.Assert().Equal("OK", rr.Body.String())
}

func (s *IntegrationSuite) TestRequestTimeout() {
	s.mux.Get("/slow", func(_ http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	s.mux.ServeHTTP(rr, req)

	s.Assert().Equal(http.StatusGatewayTimeout, rr.Code)
}

type CorsIntegrationSuite struct {
	suite.Suite
	mux     *chi.Mux
	testApp *fxtest.App
}

func (s *CorsIntegrationSuite) SetupTest() {
	s.testApp = fxtest.New(
		s.T(),
		fx.Supply(
			coretesting.NewNoopLogger(),
			fx.Annotate(
				&testAppConfig{
					httpConfig: &Configuration{
						Port:           15124,
						RequestTimeout: 1,
						Cors: CorsConfiguration{
							AllowedOrigins: []string{"https://example.com"},
						},
					},
				},
				fx.As(new(WithHTTPServer)),
			),
		),
		Module,
		fx.WithLogger(func() fxevent.Logger {
			return fxevent.NopLogger
		}),
		fx.Populate(&s.mux),
	)
	s.testApp.RequireStart()
}

func (s *CorsIntegrationSuite) TearDownTest() {
	s.testApp.RequireStop()
}

func (s *CorsIntegrationSuite) TestPreflightRequest() {
	s.mux.Get("/api", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, "OK")
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	s.mux.ServeHTTP(rr, req)

	s.Assert().Equal("https://example.com", rr.Header().Get("Access-Control-Allow-Origin"))
	s.Assert().Contains(rr.Header().Get("Access-Control-Allow-Methods"), "GET")
}

func (s *CorsIntegrationSuite) TestSimpleRequest() {
	s.mux.Get("/api", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, "OK")
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("Origin", "https://example.com")
	s.mux.ServeHTTP(rr, req)

	s.Assert().Equal(http.StatusOK, rr.Code)
	s.Assert().Equal("https://example.com", rr.Header().Get("Access-Control-Allow-Origin"))
}

func (s *CorsIntegrationSuite) TestDisallowedOrigin() {
	s.mux.Get("/api", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, "OK")
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("Origin", "https://evil.com")
	s.mux.ServeHTTP(rr, req)

	s.Assert().Empty(rr.Header().Get("Access-Control-Allow-Origin"))
}

func (s *IntegrationSuite) TestNoCorsHeaders() {
	s.mux.Get("/api", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, "OK")
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set("Origin", "https://example.com")
	s.mux.ServeHTTP(rr, req)

	s.Assert().Equal(http.StatusOK, rr.Code)
	s.Assert().Empty(rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestHTTPServerSuite(t *testing.T) {
	suite.Run(t, new(HTTPServerSuite))
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationSuite))
}

func TestCorsIntegrationSuite(t *testing.T) {
	suite.Run(t, new(CorsIntegrationSuite))
}
