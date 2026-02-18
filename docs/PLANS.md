<!-- last-reviewed: 2026-02-15 content-hash: 9b41218f -->
# Plans

Manually maintained roadmap. Updated when initiatives start or complete. See also [PRODUCT_SENSE.md](./PRODUCT_SENSE.md) for product context and the [product-specs catalogue](./product-specs/index.md) for shipped feature specifications.

## Completed

| Initiative | Summary |
|---|---|
| Core framework | FX modules (boot, HTTP server, logger, middleware, OTel, PSQL, RLS), configuration, domain model, testing utilities |
| CI pipeline | 5-job GitHub Actions: lint, build, test, vulncheck, validate |
| Observability stack | Vector → VictoriaLogs/Metrics/Tempo, query scripts, OTel instrumentation |
| Mechanical enforcement | depguard layers, forbidigo rules, revive limits, validation scripts (docs, architecture, naming) |
| Documentation structure | CLAUDE.md → ARCHITECTURE.md → docs/, progressive disclosure, doc impact matrix |
| Harness readiness assessment | Gap analysis against OpenAI Harness patterns; identified 9 gaps, addressed documentation and tooling gaps |
| First application (sweetshop) | Built sweetshop service proving all framework layers end-to-end: domain, service, persistence, transport, RLS, SQLC, FX wiring, integration tests |
| Post-sweetshop improvements | Fixed /healthz liveness bug (LivenessHandler/ReadinessHandler), extracted core/migrations package, documented handler patterns, mapper conventions, integration test template, and org context flow |

## Active

| Initiative | Summary |
|---|---|

## Planned

| Initiative | Summary |
|---|---|
| Automated quality scanning | Recurring agent tasks to scan for pattern deviations and maintain quality grades |
| Per-worktree isolation | Boot isolated app + observability stack per git worktree for parallel agent work |
| Application-driving capabilities | Agent-accessible integration test harness: boot app, hit endpoints, validate behavior |
| Doc gardening automation | Automated remediation for stale docs, not just detection |
| Custom taste invariants | Additional linter rules for structured logging format, naming conventions, file organization |
