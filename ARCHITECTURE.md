<!-- last-reviewed: 2026-02-15 -->
# Architecture

This document is the authoritative reference for the codebase structure. Read this before making changes.

## Module Map

The repository is a Go multi-module monorepo:

```
core/          Shared framework: configuration, FX modules (bootfx, httpserverfx, loggerfx, middlewarefx, otelfx, psqlfx, rlsfx), testing utilities
apps/<name>/   Application modules (auto-discovered by Makefiles)
```

Each application lives under `apps/` as a separate Go module and depends on `core/` via a `replace` directive. `core/` has no knowledge of applications.

## Core Packages

### FX Modules

Each module exposes `var Module` and expects the application config to implement its `With*` interface.

| Package | Provides | Config Interface |
|---------|----------|-----------------|
| `bootfx` | Application bootstrap — composes core modules, starts FX | `WithFx` — `AsFx() fx.Option` |
| `httpserverfx` | `*http.Server`, `*chi.Mux` with timeouts and lifecycle | `WithHTTPServer` — port, request timeout, CORS |
| `loggerfx` | `*slog.Logger` with configurable level and format | `WithLogging` — level, format (text/JSON) |
| `otelfx` | Global TracerProvider + MeterProvider, OTLP HTTP exporters | `WithOTel` — endpoint, service name, sample rate |
| `psqlfx` | `*pgxpool.Pool` with health checks, OTel tracing, `TranslateError()` for pgx→domain error mapping (generic messages, no entity context), `TxFromContext()`/`ContextWithTx()` for ambient transactions | `WithPSQL` — host, port, database, credentials, pool |
| `middlewarefx` | Configurable HTTP middleware stack — recovery, max body size, request ID, correlation ID, OTel, logging. All middleware has `Enabled` flags (`DefaultConfiguration()` enables all). App middleware injection via FX value group `"middleware"`. | `WithMiddleware` — nested per-middleware config structs (enabled flags, correlation header, max bytes) |
| `rlsfx` | `*rlsfx.DB` — `Tx()` enforces RLS via `SET LOCAL`; `Query[T]()`/`Exec()` generic helpers combining RLS transaction + error translation | `WithRLS` — schema, field |

### Utility Packages

| Package | Provides |
|---------|----------|
| `configuration` | `LoadConfiguration[T]()` — YAML + env overlay + secret resolution + validation |
| `domain` | `Organization` context helpers; `Error` model with code-based classification and sentinel errors; `ID` type wrapping UUID v7 with `ParseID()` returning domain errors |
| `datatype` | `StringEnum`, `DBStringEnum` — generic enum types with DB support; `ScanJSON` for JSON columns |
| `secretstore` | `Service` interface — `GetSecretValue(name)` for pluggable secret backends |
| `transport/http` | `LivenessHandler()` for k8s liveness (static 200); `ReadinessHandler()` for k8s readiness (checks Postgres); `NewSecureCookie()`; `WriteError()` for domain→RFC 9457 problem details; `NoOpBinder`/`NoOpRenderer` embeddable defaults; `RenderOrLog()`/`RenderListOrLog()` for logged render calls |
| `transport/http/middleware` | `WithOrganization()` — extracts org from subdomain, adds to context |
| `migrations` | `MigrateUp()`, `MigrateReset()`, `VerifyVersion()`, `CreateMigration()` — parameterized Goose wrapper; apps supply `embed.FS`, version table name, and relative dir |
| `testing` | `NewDB()`, `DB.WithTx()` for transaction-isolated tests; `MockRLS()` for RLS session variables; `JSONRequest()`/`DecodeJSON()` for HTTP test helpers |

## Package Layering (application template)

```
apps/<name>/
├── cmd/                        Entry points (CLI commands, server bootstrap)
├── internal/
│   ├── domain/                 Pure business types, interfaces, errors, events
│   ├── service/                Application services + Registry (orchestrate domain + ports)
│   ├── infrastructure/
│   │   └── persistence/        SQLC-generated code, mappers, repository implementations
│   │       ├── queries/        SQL query files (input to SQLC)
│   │       └── sqlcgen/        Generated Go code (committed to VCS)
│   ├── transport/
│   │   └── http/
│   │       ├── handler/        HTTP handlers (inbound adapters)
│   │       ├── middleware/      HTTP middleware (request ID, correlation, logging)
│   │       ├── dto/            Request/response types, validation, mappers
│   │       ├── routes.go       FX route module composing middleware + handlers
│   │       └── errors.go       Domain error → HTTP status translation
│   ├── config/                 Application configuration + FX supply
│   └── migrations/             Goose SQL migrations
```

## Dependency Direction

The application follows hexagonal architecture (ports and adapters) with DDD tactical patterns:

- **Domain** — pure business types, interfaces, errors. No external dependencies.
- **Service** — orchestrates domain operations. Depends on domain interfaces, not implementations.
- **Infrastructure** — implements domain interfaces (persistence, external services).
- **Transport** — adapts external protocols (HTTP) to service calls.
- **Cmd** — composition root that wires everything via FX.

Dependencies flow inward. Outer layers depend on inner layers, never the reverse. This is mechanically enforced via `depguard` rules in `.golangci.yml`.

```
transport/  ──→  service/  ──→  domain/
infrastructure/  ──→  domain/
config/  ──→  (supplies values to all layers via FX)
cmd/  ──→  (wires everything together via FX)
```

### Forbidden Import Matrix

| Package | Must NOT import |
|---------|----------------|
| `domain/` | `pgx`, `database/sql`, `net/http`, `chi`, `infrastructure/`, `transport/`, `service/`, `config/` |
| `service/` | `pgx`, `database/sql`, `net/http`, `chi`, `infrastructure/`, `transport/` |
| `infrastructure/` | `net/http`, `chi`, `transport/`, `service/` |
| `internal/**` (except `config/`, `migrations/`, organization repos, tests) | `psqlfx`, `pgxpool` — must use `rlsfx`; only config, migrations, and organization context repos may import these directly |
| `transport/` | `pgx`, `database/sql`, `infrastructure/` |

### Allowed Dependencies

| Package | May import |
|---------|-----------|
| `domain/` | stdlib, `uuid` |
| `service/` | `domain/`, `config/`, stdlib, external libraries |
| `infrastructure/` | `domain/`, `core/fx/rlsfx`, `pgx`, `sqlcgen/`, stdlib. Organization repos also use `core/fx/psqlfx` and `pgxpool` (outside RLS). |
| `transport/` | `domain/`, `service/`, `chi`, `go-chi/render`, stdlib |
| `config/` | `core/*`, stdlib |
| `cmd/` | everything (this is the composition root) |

## Domain-Driven Design

Applications are structured around Domain-Driven Design (DDD) tactical patterns. We adopt the vocabulary and layering because it produces clear boundaries and readable code — not as dogma. Use the patterns where they fit naturally; skip them where they add ceremony without value.

### Vocabulary

| Term | Where it lives | What it means |
|------|---------------|---------------|
| **Entity** | `domain/` | Object with identity that persists over time (e.g., `User`, `Organization`). Identified by a UUID, mutable. |
| **Value Object** | `domain/` | Immutable object defined by its attributes, not identity (e.g., `Email`, `Money`). Compared by value. |
| **Repository** | Interface in `domain/`, implementation in `infrastructure/persistence/` | Abstracts data access for an aggregate. Domain defines the contract, infrastructure provides the implementation. |
| **Application Service** | `service/` | Orchestrates use cases by coordinating domain objects and repositories. Contains no business rules — those belong in the domain. A `service.Registry` provides all services to the transport layer via a single injection point. |
| **Adapter** | `transport/` (inbound), `infrastructure/` (outbound) | Translates between external protocols and the domain. HTTP handlers are inbound adapters; repository implementations are outbound adapters. |
| **Domain Error** | `domain/errors.go` | Typed errors with a `Code` for classification. Services produce them; transport adapters translate them to protocol-specific responses. |
| **Domain Event** | `domain/` (if appropriate) | Record of something significant that happened in the domain. Introduce only when the domain genuinely needs event-driven behavior. |

### How the layers map to DDD

```
domain/            The domain model — entities, value objects, repository interfaces, errors, events
service/           Application services — use case orchestration, no business logic
infrastructure/    Outbound adapters — repository implementations, external service clients
transport/         Inbound adapters — HTTP handlers, CLI commands
cmd/               Composition root — wires ports to adapters via FX
```

The dependency rule (outer layers depend on inner, never the reverse) is the Dependency Inversion Principle applied to DDD: `service/` depends on repository *interfaces* defined in `domain/`, not on `infrastructure/` implementations. FX wires the concrete implementations at startup.

### Naming conventions

- Domain types use business language, not technical jargon: `User`, `Organization`, not `UserModel` or `UserEntity`
- Repository interfaces are named after the aggregate: `UserRepository`, not `UserStore` or `UserDAO`
- Service methods describe use cases: `CreateUser`, `AuthenticateWithCode`, not `InsertUser` or `ProcessLogin`
- Handlers are grouped by resource: `handler/user.go`, `handler/organization.go`

### Validation ownership

The transport layer parses and decodes (JSON deserialization, path parameter parsing). The service layer validates business rules (valid category, positive price, order is open). Do not split validation across layers — it creates ambiguity about where the source of truth is.

- **Transport:** `render.Bind()` decodes the request body. `ParseID()` parses path parameters. Malformed input (bad JSON, unparseable UUID) is rejected here.
- **Service:** All business validation happens here. Required fields, value ranges, enum membership, invariant checks. Services return `CodeValidation` errors for input violations and `CodeInvariant` errors for business rule violations.
- **DTOs** carry no validation logic. They are pure data shapes with `NoOpBinder`/`NoOpRenderer` for chi/render compatibility.

### RLS boundary: what sits inside vs outside

Tables that *establish* tenant context sit **outside** RLS. Tables that *contain* tenant-scoped data sit **inside** RLS. The dependency type declares intent:

- `*rlsfx.DB` → RLS-enforced. For tenant-scoped data (products, orders, etc.).
- `*pgxpool.Pool` → direct access. For tables queried to resolve the tenant (organizations) or for cross-tenant operations.

Example: the organizations table is queried in middleware to resolve a slug → org before any RLS transaction exists. Its repository uses `*pgxpool.Pool` and picks up ambient transactions via `psqlfx.TxFromContext()`.

## FX Module Pattern

Each package with FX integration exposes a `var Module = fx.Module(...)`.

- **`fx.Provide`** — constructors that return types/interfaces
- **`fx.Invoke`** — side-effect functions (register routes, register middleware)
- **Modules compose** — `RouteModule` includes `middleware.Module`

Configuration flows through the `With*` interface pattern:

```go
// Core defines the interface
type WithHTTPServer interface {
    HTTPServerConfiguration() *Configuration
}

// App config implements it
func (c *AppConfiguration) HTTPServerConfiguration() *Configuration { return c.HTTPServer }

// AsFx() supplies the config with type annotations
func (c *AppConfiguration) AsFx() fx.Option {
    return fx.Supply(c, fx.Annotate(c, fx.As(new(httpserverfx.WithHTTPServer)), ...))
}
```

This decouples FX modules from the concrete application config type — modules depend only on their `With*` interface.

The composition root is `cmd/server.go`:

```
bootfx.BootFx(cfg, ...)    Core modules (logging, HTTP, Postgres pool, RLS)
persistence.Module          Repository implementations → domain interfaces
internalhttp.RouteModule    Middleware + handler registration
```

## Configuration Loading

```
YAML base config → environment variable overlay → secret:// resolution → struct validation
```

Each layer overrides the previous. Environment-specific YAML files (`development.yaml`, `testing.yaml`) provide defaults. Environment variables override for deployment flexibility. The `secret://` prefix defers to a pluggable secret store (see [docs/SECURITY.md](./docs/SECURITY.md) for details).

Configuration files live in `apps/<name>/resources/config/`. The loading pipeline is handled by `core/configuration/`.

## Domain Error Model

Domain errors carry a `Code` for classification. The error model lives in `core/domain/errors.go` and is shared across all applications.

| Code | HTTP Status | Meaning |
|------|-------------|---------|
| `NOT_FOUND` | 404 | Entity does not exist |
| `CONFLICT` | 409 | Duplicate or version conflict |
| `VALIDATION` | 400 | Input validation failure |
| `FORBIDDEN` | 403 | Not authorized |
| `INVARIANT_VIOLATED` | 422 | Business rule violation |

Services and repositories never deal with HTTP concepts — they produce domain errors. Translation to HTTP responses happens via `core/transport/http.WriteError()`, which produces RFC 9457 problem details (`application/problem+json`) with `status`, `code`, `detail`, `instance`, `request_id`, and optional `errors` fields. Persistence errors with clear domain meaning (no rows, unique violation) are translated to domain errors via `psqlfx.TranslateError()`. Errors without domain meaning (connection failures, unexpected pgx errors) pass through as plain errors — `WriteError()` treats them as 500 with a generic message (real error is logged, not exposed to clients).

## Database

### Schema

Application tables live in the `app` schema. Define tables as needed for your domain. See [`docs/generated/db-schema.md`](./docs/generated/db-schema.md) for the auto-generated schema reference (`make docs-schema`).

### Row-Level Security (RLS)

See [docs/SECURITY.md](./docs/SECURITY.md) for the full RLS model, auth pattern, and secret management.

RLS enforces tenant isolation at the database level. Each RLS-protected table gets a policy:

```sql
CREATE POLICY organization_isolation_policy ON app.<table>
USING (organization_id = current_setting('app.current_organization')::UUID);
```

The `rlsfx.DB.Tx()` method sets `app.current_organization` via `SET LOCAL` before each transaction. Queries outside this method do not have RLS applied. Non-RLS repositories use `*pgxpool.Pool` directly and participate in ambient transactions via `psqlfx.TxFromContext()`.

**DO NOT** use `*pgxpool.Pool` to access RLS-protected tables. The dependency type declares intent: `*rlsfx.DB` = tenant-isolated, `*pgxpool.Pool` = direct access (admin/cross-tenant). Using `*pgxpool.Pool` on an RLS-protected table bypasses tenant isolation and is a security vulnerability. When adding a new repository, choose the dependency type based on whether the table has an RLS policy.

### Migrations

Managed by Goose in `apps/<name>/internal/migrations/`. Migrations are SQL files, numbered sequentially.

**Forward-only.** Down migrations are not supported. To revert a migration, write a new forward migration that undoes the changes. This is a deliberate decision: down migrations create a false sense of reversibility, are rarely tested, and diverge from the actual production state. A new forward migration is explicit, reviewable, and goes through the same CI pipeline as any other change.

## Health Probes

See [docs/RELIABILITY.md](./docs/RELIABILITY.md) for timeout chains, graceful shutdown, and error handling details.

- **`/healthz`** (liveness) — always returns 200. No external checks. Kubernetes uses this to decide if the process is alive.
- **`/readyz`** (readiness) — checks Postgres connectivity. Returns 200 or 503. Kubernetes uses this to decide if the pod should receive traffic.

## Growth Paths

### Adding a new application

1. Create `apps/<name>/` with a `go.mod`:
   ```
   module github.com/bbsbb/go-edge/<name>

   replace github.com/bbsbb/go-edge/core => ../../core
   ```
2. Create the package layout following the template in [Package Layering](#package-layering-application-template) above
3. Add application configuration in `internal/config/` implementing the `With*` interfaces from core
4. Wire the composition root in `cmd/server.go` using `bootfx.BootFx()`
5. Register health probes in the composition root:
   ```go
   p.Mux.Get("/healthz", transporthttp.LivenessHandler())
   p.Mux.Get("/readyz", transporthttp.ReadinessHandler(p.Pool, 0, p.Logger))
   ```
6. Add migrations in `internal/migrations/` — embed the `versions/` FS, set a unique version table name (e.g. `public.<app>_goose_db_version`), delegate to `core/migrations`. See `apps/sweetshop/internal/migrations/` for the thin wrapper pattern and `apps/sweetshop/cmd/migrate.go` for the CLI template.
7. Add resource files: `resources/config/development.yaml`, `resources/config/testing.yaml`
8. Add a `Makefile` with standard targets (`test`, `lint`, `build`)
9. Add structural tests in `architecture_test.go` at the module root — verify forbidden imports (backup to depguard), file size limits (backup to revive), and test coverage completeness (`TestAllPackagesHaveTests`)
10. The app is auto-discovered by root Makefiles via `$(wildcard apps/*)`

### Adding a new repository

1. Add interface to `domain/repository.go`
2. Add SQL queries to `infrastructure/persistence/queries/<entity>.sql`
   - Use `:execrows` (not `:exec`) for UPDATE and DELETE mutations so the repo can detect zero-rows-affected and return `pgx.ErrNoRows` → `CodeNotFound`
3. Run `make sqlc-generate` to regenerate `sqlcgen/` (iterates all apps)
4. Add domain mappers to `infrastructure/persistence/mappers.go` following the naming convention:
   - `<type>ToDomain(sqlcgen.<Type>) *domain.<Type>` — converts SQLC row to domain entity
   - `<type>CreateParams(*domain.<Type>) sqlcgen.Create<Type>Params` — converts domain entity to SQLC insert params
   - `<type>UpdateParams(*domain.<Type>) sqlcgen.Update<Type>Params` — converts domain entity to SQLC update params
5. Implement repository in `infrastructure/persistence/<entity>.go`
   - RLS-protected tables: depend on `*rlsfx.DB`, use `rlsfx.Query[T]()`/`rlsfx.Exec()` helpers
   - Non-RLS tables: depend on `*pgxpool.Pool`, pick up ambient tx via `psqlfx.TxFromContext()`
   - For `:execrows` mutations: check `n == 0` → return `pgx.ErrNoRows` (translated to `CodeNotFound` by `TranslateError`)
6. Add provider to `infrastructure/persistence/module.go`

### Adding a new HTTP endpoint

1. Add request/response DTOs to `transport/http/dto/` — embed `transporthttp.NoOpBinder`/`NoOpRenderer` for chi/render compatibility
2. Create handler in `transport/http/handler/` following the established patterns:
   - **Path parameters:** `id, err := coredomain.ParseID(chi.URLParam(r, "id"))` — returns domain error on invalid UUID
   - **Request body:** `render.Bind(r, &req)` — decodes JSON into DTO; on error, return `CodeValidation` domain error
   - **Success response:** `render.Status(r, http.StatusOK)` then `transporthttp.RenderOrLog(w, r, resp, h.logger)` (or `RenderListOrLog` for slices)
   - **Error response:** `transporthttp.WriteError(w, r, err, h.logger)` — translates domain errors to RFC 9457 problem details
   - **No content:** `w.WriteHeader(http.StatusNoContent)` for DELETE operations
3. Register route in `transport/http/routes.go`

### Adding a new transport (gRPC, CLI, etc.)

Create a new package under `transport/` (e.g., `transport/grpc/`). It follows the same pattern: depends on `domain/` and `service/`, never on `infrastructure/`.

### Adding new infrastructure (cache, message queue, etc.)

Create a new package under `infrastructure/` (e.g., `infrastructure/cache/`). Define the interface in `domain/`, implement in `infrastructure/`.

## Technical Debt

Deferred work is tracked in [`docs/exec-plans/tech-debt.md`](./docs/exec-plans/tech-debt.md). Each entry records what was deferred, why, and when to revisit. Check this file before starting new initiatives — some deferred items may now be relevant.
