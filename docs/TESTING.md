<!-- last-reviewed: 2026-02-15 content-hash: e5d37460 -->
# Testing

Testing strategy, patterns, and review checklist.

## Test at the Right Level

Focus testing effort where it provides the most value:

| Layer | Test? | How |
|-------|-------|-----|
| **Handlers + Services + Persistence** | Always | Integration tests against a real database. Full stack: HTTP request → handler → service → repo → DB. |
| **Domain** | When non-trivial | Unit tests for validation, computed properties, business rules. |
| **Config** | Rarely | Only if custom loading or validation logic exists. |

**Integration tests are the primary strategy.** They exercise the full stack with a real database, catching bugs that mocked tests miss (SQL errors, RLS behavior, transaction semantics). Mock only external services that you don't control. If time is limited, write integration tests first.

## Patterns

### Always Use Test Suites

Every test file uses `testify/suite`. Suites provide `SetupSuite`, `SetupTest`, `TearDownTest`, `BeforeTest` hooks for shared fixtures and cleanup.

### Integration Tests (Primary Pattern)

The base integration suite lives in the handler test package (`handler_test`). It wires the full FX stack against a real Postgres instance with transaction-per-test isolation. See `apps/sweetshop/internal/transport/http/handler/integration_test.go` for the canonical implementation.

```go
// handler/integration_test.go — base suite
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

    // Transaction-injection middleware: every request sees the test transaction.
    // The transaction rolls back on cleanup, providing full test isolation.
    s.Router = chi.NewRouter()
    s.Router.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx := psqlfx.ContextWithTx(r.Context(), s.tx)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    })

    // Wire the full FX stack — same modules as cmd/server.go
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

// BeforeTest runs before each test — seeds a fresh organization inside
// the test transaction so each test gets isolated tenant context.
func (s *IntegrationSuite) BeforeTest(_, _ string) {
    s.DB.WithTx(s.T(), func(ctx context.Context) {
        s.OrgID = uuid.Must(uuid.NewV7())
        s.org = &coredomain.Organization{ID: s.OrgID, Slug: "test-shop"}
        s.tx = psqlfx.TxFromContext(ctx)

        s.Require().NoError(s.orgRepo.Create(ctx, s.org))
    })
}

// Do executes a request with the test organization's slug header.
func (s *IntegrationSuite) Do(req *http.Request) *httptest.ResponseRecorder {
    req.Header.Set("X-Organization-Slug", s.org.Slug)
    rec := httptest.NewRecorder()
    s.Router.ServeHTTP(rec, req)
    return rec
}
```

```go
// handler/product_test.go — test suite embeds the base
type ProductSuite struct {
    IntegrationSuite
}

func (s *ProductSuite) TestCreateProduct() {
    req := coretesting.JSONRequest(s.T(), http.MethodPost, "/products", map[string]any{
        "name": "Vanilla Scoop", "category": "ice_cream", "price_cents": 350,
    })
    rec := s.Do(req)
    s.Assert().Equal(http.StatusCreated, rec.Code)
}

func TestProductSuite(t *testing.T) {
    suite.Run(t, new(ProductSuite))
}
```

Key properties:
- **Real database** — exercises SQL, RLS policies, constraints, and transaction semantics
- **Transaction-per-test isolation** — `WithTx` begins a transaction, rolls back on cleanup. No test data leaks between tests.
- **Full stack** — HTTP request → chi router → handler → service → repo → rlsfx → Postgres. Catches integration bugs that mocked tests miss.
- **Fast** — each test adds ~50ms. No container startup per test — the DB runs via Docker Compose.

### Mocks via Mockery (External Services Only)

Use `mockery` only for external service interfaces that you don't control (payment gateways, email providers, etc.). Do not mock internal interfaces (repositories, services) — use integration tests instead.

```go
s.paymentGateway.EXPECT().Charge(mock.Anything, amount).Return(nil).Once()
```

Mock configuration lives in `.mockery.yaml` per module. Generated mocks go to `tests/mocks/`.

### Table-Driven Tests for Parameterized Cases

When testing the same behavior with different inputs, use table-driven tests inside a suite method:

```go
func (s *MySuite) TestValidation() {
    cases := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid email", "a@b.com", false},
        {"missing @", "invalid", true},
    }
    for _, tc := range cases {
        s.Run(tc.name, func() {
            err := validate(tc.input)
            if tc.wantErr {
                s.Assert().Error(err)
            } else {
                s.Assert().NoError(err)
            }
        })
    }
}
```

### Handler Tests via Integration Suite

Handler tests use the integration suite's real router — no mocked services:

```go
func (s *UserSuite) TestGetUser_NotFound() {
    req := httptest.NewRequest(http.MethodGet, "/users/019505e0-0000-7000-8000-000000000000", nil)
    rec := httptest.NewRecorder()
    s.Router.ServeHTTP(rec, req)
    s.Assert().Equal(http.StatusNotFound, rec.Code)
}
```

### Use go-cmp for Complex Comparisons

When comparing structs with nested fields or when `Equal` produces hard-to-read diffs, use `go-cmp`:

```go
s.Assert().Empty(cmp.Diff(expected, actual))
```

### Use gofakeit for Test Data

Generate realistic test data instead of hand-crafting strings:

```go
user := domain.User{
    Name:  gofakeit.Name(),
    Email: gofakeit.Email(),
}
```

## Core Testing Fixtures

`core/testing` provides test utilities behind a `//go:build testing` tag. Importing it from any test file is always safe — both `make test` and `golangci-lint` build with `-tags=testing`. Use the import alias `coretesting "github.com/bbsbb/go-edge/core/testing"`.

- **`coretesting.NewNoopLogger()`** — returns a `*slog.Logger` that discards all output. Use this instead of `slog.New()` which is forbidden in tests by forbidigo.
- **`coretesting.NewDB(t, dbConfig)`** — creates a `pgxpool.Pool` with automatic cleanup
- **`coretesting.DB.WithTx(t, fn)`** — runs `fn` inside a transaction that rolls back on cleanup, providing test isolation
- **`coretesting.MockRLS(ctx, t, schema, field, value)`** — sets a PostgreSQL session variable to simulate RLS context without going through `rlsfx`

Always check `core/testing/` for existing fixtures before writing custom test setup.

## Before Writing Tests

1. Check if the package already has a test suite — add to it, don't create a parallel one
2. Check `core/testing/` for fixtures that match your needs
3. Check sibling packages for patterns to follow (table-driven structure, mock setup, assertion style)

## After Writing Tests — Review Checklist

Every test file must pass this review before it's done:

- **No superfluous comments.** Test method names should be self-describing. Remove comments that restate what the code does.
- **No duplication across test cases.** If setup or assertions repeat, extract them into suite helpers or use table-driven tests.
- **Parameterize where possible.** If multiple tests differ only in input/expected values, collapse them into a table-driven test.
- **Extract fixtures.** If the same test objects are constructed in multiple suites, consider adding a fixture to `core/testing/` or a local `testdata` helper.
