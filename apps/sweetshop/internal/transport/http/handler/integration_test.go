//go:build testing

package handler_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path"
	"runtime"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	coredomain "github.com/bbsbb/go-edge/core/domain"
	"github.com/bbsbb/go-edge/core/fx/middlewarefx"
	"github.com/bbsbb/go-edge/core/fx/psqlfx"
	"github.com/bbsbb/go-edge/core/fx/rlsfx"
	coretesting "github.com/bbsbb/go-edge/core/testing"
	transporthttp "github.com/bbsbb/go-edge/core/transport/http"
	"github.com/bbsbb/go-edge/sweetshop/internal/config"
	"github.com/bbsbb/go-edge/sweetshop/internal/infrastructure/persistence"
	transportroutes "github.com/bbsbb/go-edge/sweetshop/internal/transport/http"
)

type IntegrationSuite struct {
	suite.Suite
	Cfg    *config.AppConfiguration
	DB     *coretesting.DB
	Logger *slog.Logger
	Router *chi.Mux
	OrgID  uuid.UUID

	orgRepo *persistence.OrganizationRepo
	tx      pgx.Tx
	org     *coredomain.Organization
}

func (s *IntegrationSuite) SetupSuite() {
	_, filename, _, _ := runtime.Caller(0)
	configDir := path.Join(path.Dir(filename), "../../../..", "resources", "config")

	cfg, err := config.NewAppConfiguration(context.Background(), configDir)
	s.Require().NoError(err)

	s.Cfg = cfg
	s.Logger = coretesting.NewNoopLogger()
	s.DB = coretesting.NewDB(s.T(), cfg.PSQL)

	rlsDB, err := rlsfx.NewDB(s.DB.Pool, s.Cfg.RLS, s.Logger)
	s.Require().NoError(err)

	s.orgRepo = persistence.NewOrganizationRepo(s.DB.Pool)

	s.Router = chi.NewRouter()
	s.Router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := psqlfx.ContextWithTx(r.Context(), s.tx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	app := fxtest.New(s.T(),
		cfg.AsFx(),
		fx.Supply(s.Router),
		fx.Supply(s.DB.Pool),
		fx.Supply(rlsDB),
		fx.Supply(s.Logger),
		middlewarefx.Module,
		persistence.Module,
		transportroutes.RouteModule,
		fx.Invoke(func() {
			s.Router.Get("/healthz", transporthttp.LivenessHandler())
			s.Router.Get("/readyz", transporthttp.ReadinessHandler(s.DB.Pool, 0, s.Logger))
		}),
	)
	app.RequireStart()
	s.T().Cleanup(func() { app.RequireStop() })
}

func (s *IntegrationSuite) BeforeTest(_, _ string) {
	s.DB.WithTx(s.T(), func(ctx context.Context) {
		s.OrgID = uuid.Must(uuid.NewV7())
		s.org = &coredomain.Organization{ID: s.OrgID, Slug: "test-shop"}
		s.tx = psqlfx.TxFromContext(ctx)

		s.Require().NoError(s.orgRepo.Create(ctx, s.org))
	})
}

func (s *IntegrationSuite) Do(req *http.Request) *httptest.ResponseRecorder {
	req.Header.Set("X-Organization-Slug", s.org.Slug)
	rec := httptest.NewRecorder()
	s.Router.ServeHTTP(rec, req)
	return rec
}

func (s *IntegrationSuite) CreateProduct(name, category string, priceCents int) map[string]any {
	req := coretesting.JSONRequest(s.T(), http.MethodPost, "/products", map[string]any{
		"name": name, "category": category, "price_cents": priceCents,
	})
	rec := s.Do(req)
	s.Require().Equal(http.StatusCreated, rec.Code)

	var resp map[string]any
	coretesting.DecodeJSON(s.T(), rec, &resp)
	return resp
}

func (s *IntegrationSuite) OpenOrder() map[string]any {
	req := httptest.NewRequest(http.MethodPost, "/orders", nil)
	rec := s.Do(req)
	s.Require().Equal(http.StatusCreated, rec.Code)

	var resp map[string]any
	s.Require().NoError(json.NewDecoder(rec.Body).Decode(&resp))
	return resp
}
