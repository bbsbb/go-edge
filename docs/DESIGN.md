<!-- last-reviewed: 2026-02-15 content-hash: a7200056 -->
# Design

System-wide technology choices and tradeoffs. For structural patterns and dependency rules, see [ARCHITECTURE.md](../ARCHITECTURE.md). For API quick-references, see the [references catalogue](./references/index.md). For design decision records, see the [design-docs catalogue](./design-docs/index.md).

## Domain-Driven Design

Applications are structured around DDD tactical patterns within a hexagonal architecture (ports and adapters). We adopt the vocabulary and layering because it produces clear boundaries and readable code — not as dogma. Use the patterns where they fit naturally; skip them where they add ceremony without value.

For structural rules (package layout, dependency direction, forbidden imports), see [ARCHITECTURE.md](../ARCHITECTURE.md).

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

