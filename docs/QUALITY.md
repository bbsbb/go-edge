<!-- last-reviewed: 2026-02-15 -->
# Quality

Quality grades per domain and architectural layer. For the current database schema, see [`generated/db-schema.md`](./generated/db-schema.md) (`make docs-schema`).

## Grading Framework

Each domain area and architectural layer is graded on a simple scale:

- **A** — Production-ready. Tests, docs, error handling, and observability are all in place.
- **B** — Functional and tested. Minor gaps in edge cases, docs, or observability.
- **C** — Works but incomplete. Missing tests, limited error handling, or no docs.
- **D** — Scaffolding only. Structure exists but implementation is minimal or stubbed.

## Current Grades

### Core Framework (`core/`)

| Area | Grade | Notes |
|------|-------|-------|
| HTTP server (httpserverfx) | A | Timeouts, graceful shutdown, lifecycle hooks. |
| PostgreSQL (psqlfx) | A | Connection pooling, health checks, lifecycle hooks, OTel tracing via otelpgx, pgx→domain error translation. |
| RLS (rlsfx) | A | Row-level security, transaction helper, tested. |
| OTel (otelfx) | B | TracerProvider + MeterProvider, OTLP HTTP exporters. No local collector yet. |
| Configuration | A | YAML + env overlay, secret:// resolution, validated. |
| Logging (loggerfx) | A | Structured slog, FX event logging. |
| Boot (bootfx) | A | Application lifecycle, FX composition, signal handling. |
| Middleware (middlewarefx) | A | Configurable stack via `WithMiddleware` with nested per-middleware config structs: panic recovery, max request body size, request ID, correlation ID (configurable header), OTel HTTP, request logging. All middleware uses `Enabled` flags (`DefaultConfiguration()` enables all). App middleware injection via FX value group. |
| Domain errors | B | Code-based classification, Is/As/Unwrap. No dedicated tests yet. |
| Error response writer | B | RFC 9457 problem details (`application/problem+json`) via chi/render. Domain code mapping, multi-error extraction, request ID correlation. Tested in core, used by organization middleware. |

### Sweetshop (`apps/sweetshop/`)

| Area | Grade | Notes |
|------|-------|-------|
| Domain | B | Product/Order entities, value enums, repository interfaces. Business rules (order lifecycle) tested via integration. No dedicated domain unit tests. |
| Service | B | ProductService, OrderService with structured logging. Tested via integration tests, not isolated unit tests. |
| Persistence | B | SQLC-generated queries, RLS via rlsfx, mappers. Delete/Update return not-found correctly. Tested via integration. |
| Transport (HTTP) | B | Chi handlers, RFC 9457 errors, route module with FX wiring. Tested via integration. |
| Configuration | A | Full `With*` interface coverage, development + testing YAML. |
| Migrations | A | Schema, organizations, products, orders/items, app user. RLS on tenant-owned tables only. |
| Architecture tests | A | Forbidden imports, file size limits, test coverage completeness. |
| Integration tests | A | 21 tests against real Postgres, transaction-per-test isolation, full stack (handler → service → repo → DB). |
