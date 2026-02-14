<!-- last-reviewed: 2026-02-15 -->
# Design

System-wide technology choices and tradeoffs. For structural patterns and dependency rules, see [ARCHITECTURE.md](../ARCHITECTURE.md). For API quick-references, see the [references catalogue](./references/index.md). For design decision records, see the [design-docs catalogue](./design-docs/index.md).

## Architectural Pattern

Applications follow **Domain-Driven Design** (DDD) tactical patterns within a **hexagonal architecture** (ports and adapters). We adopt the vocabulary and layering where they produce clear boundaries — not as dogma. Business logic lives in `domain/`, use case orchestration in `service/`, and all I/O is pushed to adapter layers (`transport/`, `infrastructure/`). Repository interfaces are defined in the domain and implemented in infrastructure.

See [ARCHITECTURE.md — Domain-Driven Design](../ARCHITECTURE.md#domain-driven-design) for the full vocabulary, layer mapping, and naming conventions.

## Core Framework

Technology choices baked into `core/`. These are non-negotiable — all applications inherit them.

| Technology | Role |
|-----------|------|
| Go | Static typing, single-binary deployment, agent-legible stdlib |
| Uber FX | Dependency injection via constructors and lifecycle hooks, no code generation |
| Chi | Lightweight HTTP router, `net/http` compatible, middleware via `Use()` |
| go-chi/render | Response rendering, request binding, content negotiation |
| go-chi/cors | CORS middleware with configurable origins, methods, headers |
| pgx v5 | Native PostgreSQL driver, `pgxpool`, `SET LOCAL` for RLS |
| otelpgx | OpenTelemetry tracing for pgx queries |
| otelhttp | HTTP span instrumentation for handlers and clients |
| OpenTelemetry | OTLP HTTP exporters for traces, metrics, and logs, vendor-neutral |
| slog | Stdlib structured logging, context-aware, JSON/text output |
| tint | Colorized slog output for development |
| goccy/go-yaml | YAML configuration file parsing |
| go-envconfig | Environment variable binding for configuration structs |
| go-playground/validator | Struct validation with tags (`required`, `gte`, `hostname`, etc.) |
| go-viper/mapstructure v2 | Struct-to-map decoding for DSN building |
| Goose | SQL migrations as embedded files, library mode (not CLI) |

## Recommended Application Libraries

Libraries that applications should use when building on top of core. Not in `core/go.mod` — each app adds them as needed.

| Technology | Role | Status |
|-----------|------|--------|
| SQLC | Compile-time SQL → type-safe Go code, no ORM | In use (sweetshop) |
| lestrrat-go/jwx v3 | JWT signing/verification, HMAC-SHA256 | Planned |
| gqlgen | Schema-first code generator for type-safe GraphQL servers | Planned |
| gqlparser v2 | GraphQL query and schema parser | Planned |
| urfave/cli v2 | CLI commands, flags, help generation for `cmd/` | Planned |

## Conscious Tradeoffs

**OTel dependency weight.** `core/go.mod` pulls ~90 indirect dependencies via OpenTelemetry (gRPC, protobuf, Google Cloud types). This is a large supply chain surface for a "boring technology" repo. We accept this because agent-queryable observability (traces, metrics, SLA validation) requires OTel — there is no lighter-weight alternative that provides vendor-neutral OTLP export. `govulncheck` runs on every PR via `make guard` to mitigate supply chain risk.

**Forward-only migrations.** Down migrations are not supported. See [ARCHITECTURE.md — Migrations](../ARCHITECTURE.md#migrations) for rationale.

## Testing Libraries

Required for all test code. See [TESTING.md](./TESTING.md) for strategy and patterns.

| Technology | Role | Status |
|-----------|------|--------|
| testify | Test suites (`suite`), assertions (`assert`, `require`), mocks | In use |
| mockery v2 | Mock generation from interfaces, expecter pattern | In use |
| go-cmp | Struct comparison with readable diffs | Planned |
| gofakeit | Realistic fake data generation for tests | Planned |
| genqlient | Type-safe GraphQL client code generator for API testing | Planned |

