<!-- last-reviewed: 2026-02-18 -->
# Tech Debt

Conscious technical debt with context on origin, deferral reason, and conditions for revisiting.

## Code & Design

### Secret store wiring in sweetshop
- **Origin:** Guard fixes review (2026-02-18)
- **Reason:** `NewAppConfiguration` in `apps/sweetshop/internal/config/configuration.go` does not pass `WithSecrets()` to `LoadConfiguration`. `production.yaml` contains `secret://` values for psql host, credentials, and otel endpoint. Without a secret service, these values remain as literal strings. Core now supports YAML-based secret resolution, but the app must wire a `secretstore.Service` for it to take effect.
- **What's needed:** Pass `WithSecrets(secretstore.NewEnvService(""))` (or a production adapter) conditionally based on environment in `NewAppConfiguration`.
- **Revisit:** Before first production deployment

### Pagination helpers
- **Origin:** Template baseline
- **Reason:** Sweetshop's `GET /products` returns all products without pagination. Acceptable for a proof-of-concept.
- **Revisit:** When any list endpoint needs to handle more than ~100 items

### Rate limiting middleware
- **Origin:** Template baseline
- **Reason:** No usage patterns established yet
- **Revisit:** After public API exposure

### Redis caching layer (redis-go)
- **Origin:** Anticipated need
- **Reason:** No caching infrastructure exists yet. As read-heavy endpoints scale (product listings, organization lookups), a caching layer will be required to avoid hitting Postgres on every request.
- **What's needed:**
  - Add `go-redis/redis` dependency
  - Create `core/fx/redisfx/` FX module (pool, health check, configuration via `WithRedis` interface)
  - Cache-aside pattern helpers in a shared package (get-or-fetch, invalidation)
  - Integration with readiness probe (`/readyz` should check Redis connectivity)
- **Revisit:** When any endpoint needs sub-millisecond reads or when Postgres query load becomes a concern

### Circuit breakers
- **Origin:** Template baseline
- **Reason:** No outbound service calls exist
- **Revisit:** When integrating with external services

### Domain events over message queue
- **Origin:** Template baseline
- **Reason:** In-process synchronous publishing is sufficient initially
- **Revisit:** When async processing or cross-service events are needed

### Message queue adapter
- **Origin:** Template baseline
- **Reason:** No async processing needs yet
- **Revisit:** When async processing is needed

### Cross-module dependency validation
- **Origin:** Template baseline
- **Reason:** Only one module exists; premature to enforce
- **Revisit:** When a third module is added

### Protobuf schema package
- **Origin:** Planned feature
- **Reason:** No cross-service communication or schema contracts exist yet
- **Revisit:** When adding gRPC transport, cross-service APIs, or event schemas that benefit from a shared IDL

## Infrastructure & Tooling

### CI image push to registry
- **Origin:** Template baseline
- **Reason:** No deployment pipeline yet
- **Revisit:** Deployment pipeline initiative

### Docker Compose production profile
- **Origin:** Template baseline
- **Reason:** Separate concern from build
- **Revisit:** Deployment pipeline initiative

### Production deployment pipeline
- **Origin:** Template baseline
- **Reason:** No deployment target yet; build and packaging are in place
- **Revisit:** When a staging or production environment is provisioned

### Kubernetes manifests
- **Origin:** Template baseline
- **Reason:** Separate concern from build; depends on deployment pipeline
- **Revisit:** Deployment pipeline initiative

### Custom business metrics
- **Origin:** Template baseline
- **Reason:** Needs domain maturity before meaningful metrics can be defined
- **Revisit:** After core business flows are implemented and running under load

### Alerting rules and SLO definitions
- **Origin:** Template baseline
- **Reason:** Needs observability backend deployed in a persistent environment first
- **Revisit:** After OTel collector is running in staging/production

### Recurring proactive drift detection
- **Origin:** Template baseline
- **Reason:** Background agent tasks that scan for drift and open fix-up PRs require CI/CD maturity and trusted automation
- **Revisit:** After CI pipeline is stable and deployment pipeline exists

### Per-worktree app isolation
- **Origin:** Template baseline
- **Reason:** Isolated app + observability stack per git worktree adds complexity with limited current benefit (single developer)
- **Revisit:** When multiple developers or agents work concurrently on different worktrees

## Agent Workflow

### Agent-to-agent review loop ⚠️ HIGH PRIORITY
- **Origin:** Harness readiness assessment
- **What Harness does:** Agent opens a PR, self-reviews locally, then requests additional agent reviews (local + cloud). Agents respond to feedback inline, push updates, and iterate in a loop until all agent reviewers are satisfied. Humans may review but aren't required to. Agents can squash and merge their own PRs. This is the core throughput multiplier — it removes humans from the critical path on most PRs.
- **What go-edge has today:** Review philosophy documented in CLAUDE.md (corrections are cheap, review for correctness not style, every PR must pass CI). No mechanical review workflow — no agent can post review comments, respond to feedback, approve, or merge.
- **What's needed:**
  - Agent skill or script to run a structured self-review checklist against a diff (architecture violations, concurrency safety, doc impact matrix, test coverage)
  - `gh` CLI integration for agents to post PR review comments and respond to feedback
  - Defined review criteria that agents can evaluate mechanically (CI green, no forbidden imports, tests exist for new code, no files >500 lines)
  - Trust model: which checks must pass before auto-merge is allowed, which require human sign-off
  - Per-worktree isolation (prerequisite — agents need independent environments to review without interference)
- **Reason deferred:** Requires per-worktree isolation, a running application to review against, and CI maturity for trusted auto-merge. The philosophy is in place; the tooling depends on foundational gaps being closed first.
- **Revisit:** After per-worktree isolation and first application are built. This should be the first agent workflow capability added once those prerequisites exist.

### Application-driving capabilities
- **Origin:** Harness readiness assessment
- **Status:** Partially addressed — sweetshop has integration tests that boot the full stack (handler → service → repo → DB) and exercise endpoints via `httptest`. Missing: booting the actual HTTP server and hitting it over the network, agent-driven test execution.
- **Revisit:** When end-to-end testing against a running server is needed (e.g., middleware chain, TLS, real HTTP clients)

### Doc gardening remediation
- **Origin:** Harness readiness assessment
- **Reason:** validate-docs.sh and validate-architecture.sh detect drift but don't fix it. Automated remediation (agent opens fix-up PRs) requires CI maturity and trusted automation
- **Revisit:** After CI pipeline supports automated PR creation and merge

### Custom taste invariants
- **Origin:** Harness readiness assessment
- **Reason:** depguard, forbidigo, and revive enforce architectural and safety rules. Additional taste rules (structured logging format enforcement, file organization conventions) would reduce style drift but current rule set is sufficient for framework stage
- **Revisit:** When pattern deviations become frequent enough to warrant mechanical enforcement
