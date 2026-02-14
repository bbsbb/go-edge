<!-- last-reviewed: 2026-02-15 -->
# Sweetshop — First Application

## Goal

Build the first application in the go-edge monorepo to prove the framework works end-to-end under real workload pressure.

## Decisions

- **Domain:** Sweetshop selling ice cream and marshmallows. Simple enough to build quickly, complex enough to exercise all layers (two entities with a relationship, business rules, RLS multi-tenancy).
- **No authentication:** Organization resolved from middleware context. Keeps the example focused on framework validation, not auth complexity.
- **Integration tests over mocks:** Initially planned mocked service/handler tests, then pivoted to full-stack integration tests (real Postgres, transaction-per-test isolation). Integration tests caught real bugs (DELETE not returning 404, RLS variable ordering) that mocks would have missed.
- **RLS on tenant-owned tables only:** Organizations table does NOT have RLS — it's queried to establish the tenant context before RLS is set. Only products, orders, and order_items have RLS policies.
- **SQLC `:execrows` for mutations:** DELETE and UPDATE queries use `:execrows` so the repo can detect 0 affected rows and return not-found errors.
- **`rlsfx.NewDB()` added to core:** Extracted from `NewRLS` to allow constructing `*rlsfx.DB` without FX dependency injection, needed by the integration test suite.

## Outcome

Fully functional sweetshop service with:
- 8 HTTP endpoints (product CRUD + order lifecycle)
- 5 SQL migrations (schema, organizations, products, orders/items, app user)
- Hexagonal architecture: domain → service → persistence → transport
- 21 integration tests + 13 architecture tests, all passing
- Full CI validation: lint, test, build from monorepo root

## References

- Application code: `apps/sweetshop/`
- Integration test suite: `apps/sweetshop/tests/suite/integration.go`
- Integration tests: `apps/sweetshop/tests/integration/`
- Architecture tests: `apps/sweetshop/architecture_test.go`
