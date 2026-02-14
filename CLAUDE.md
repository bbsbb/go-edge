<!-- last-reviewed: 2026-02-15 -->
# Claude Code Guidelines

## Core Beliefs

Read **[docs/CORE_BELIEFS.md](./docs/CORE_BELIEFS.md)** first. It defines the operating principles that govern every decision in this codebase: repo as system of record, mechanical enforcement over documentation, boring technology, boundary validation, progressive disclosure, strict layers with local autonomy, and conscious debt over accidental drift. All other docs implement these principles.

## References

**Always read before writing code:**
- **[docs/CORE_BELIEFS.md](./docs/CORE_BELIEFS.md)** — operating principles (linked above)
- **[ARCHITECTURE.md](./ARCHITECTURE.md)** — package layering, dependency rules, growth paths

**Read when relevant** (use the doc impact matrix below to determine which):
- **[docs/TESTING.md](./docs/TESTING.md)** — testing strategy, patterns, review checklist
- **[docs/DESIGN.md](./docs/DESIGN.md)** — technology choices, tradeoffs
- **[docs/design-docs/](./docs/design-docs/index.md)** — design decision records
- **[docs/SECURITY.md](./docs/SECURITY.md)** — RLS model, auth pattern, secret management
- **[docs/RELIABILITY.md](./docs/RELIABILITY.md)** — health probes, timeouts, graceful shutdown
- **[docs/OBSERVABILITY.md](./docs/OBSERVABILITY.md)** — local stack, query endpoints
- **[docs/QUALITY.md](./docs/QUALITY.md)** — quality grades per module
- **[docs/PLANS.md](./docs/PLANS.md)** — roadmap: completed, active, and planned initiatives
- **[docs/exec-plans/](./docs/exec-plans/index.md)** — execution plans (active, completed) and deferred work

## Doc Impact Matrix

After completing a task, scan this table. If your changes match a row, check and update the listed docs.

| What changed | Check / update |
|---|---|
| FX module (new or modified) | `ARCHITECTURE.md` module table, `docs/QUALITY.md` |
| Package structure or dependency direction | `ARCHITECTURE.md` package layering |
| Configuration fields or `With*` interfaces | `ARCHITECTURE.md` config interface table |
| Error model or codes | `ARCHITECTURE.md` domain error model |
| Health probes, timeouts, shutdown | `docs/RELIABILITY.md` |
| Auth, RLS, secrets | `docs/SECURITY.md` |
| OTel, logging, metrics | `docs/OBSERVABILITY.md` |
| Technology choice added/changed | `docs/DESIGN.md` |
| Test strategy or patterns | `docs/TESTING.md` |
| Multi-session initiative started | `docs/PLANS.md`, `docs/exec-plans/active/` |
| Initiative completed | `docs/PLANS.md`, `docs/exec-plans/completed/` |
| Work consciously deferred | `docs/exec-plans/tech-debt.md` |

## Code Style

- **No superfluous comments.** Do not restate what the code already says. Do not add comments to generated code or test files. Comments exist only when the _why_ is non-obvious — tricky logic, surprising constraints, or intentional deviations. If a comment just describes _what_ the next line does, delete it.
- Package-level doc comments are required for public packages.
- Exported functions/types need doc comments only when their purpose isn't obvious from the name.

## Concurrency Safety

- **Always consider race conditions and thread safety.** Every shared mutable state, every goroutine boundary, and every context-carried value is a potential data race. This is a first-class review criterion, not an afterthought.
- Protect shared state with appropriate synchronization (`sync.Mutex`, `sync.RWMutex`, `sync.Once`, channels, or atomics). Document the locking strategy when it's not obvious.
- Prefer immutable data and value types over shared mutable state. Pass copies across goroutine boundaries unless there's a clear reason not to.
- Use `-race` in all test runs (already enabled via `make test`). A race detector failure is a hard blocker — never suppress or ignore it.
- When reviewing code, explicitly ask: "What happens if this runs concurrently?" If the answer isn't "nothing, it's safe by construction," the code needs synchronization or a comment explaining why it's safe.

## Repository Structure

This is a **Go monorepo** with applications in the `apps/` directory. See [ARCHITECTURE.md](./ARCHITECTURE.md) for the full package layering and dependency rules.

Key principles:
- Each application is a **separate Go module** under `apps/<name>/`
- Applications import `core/` via `replace` directive
- Shared fx modules go in `core/fx/<modulename>fx/`
- Makefiles auto-discover apps via `$(wildcard apps/*)`
- Don't create empty directories upfront — create them when needed
- Don't add dependencies until they're actually used

## Testing

See [docs/TESTING.md](./docs/TESTING.md) for full strategy and patterns.

- **Integration tests are the primary strategy** — full stack against a real database, transaction-per-test isolation
- Always use `testify/suite` — no standalone test functions
- Mock only external services you don't control — never mock internal interfaces (repositories, services)
- Before writing tests: check for existing suites, patterns, and `core/testing/` fixtures
- After writing tests: review for superfluous comments, duplication, and parameterization opportunities

## Configuration

- Use `goccy/go-yaml` (not the archived `gopkg.in/yaml.v3`)
- Use `sethvargo/go-envconfig` for environment variable binding
- Support `secret://` prefix for secret store resolution
- FX module configs: add `var _ configuration.WithValidation = (*Configuration)(nil)` compile-time check

## Tools (see Makefile.setup.mk)

- golangci-lint v2.8.0
- gofumpt v0.9.2
- gotestsum v1.13.0
- mockery v2.53.5
- govulncheck v1.1.4

## Dependencies

**NEVER edit `go.mod` or `go.sum` files directly.** Use Makefile targets:

- `make deps-add PKG=<package>` — add a dependency
- `make deps-tidy` — clean up go.mod/go.sum

## Build Commands

**Always use Makefile targets instead of running go/lint commands directly.**

From module directory (e.g., `core/` or `apps/<name>/`):
- `make test` — run tests
- `make lint` — run linter

From repository root:
- `make ci` — run lint, test, and build for all modules
- `make guard` — run all meta-validations (docs, architecture, naming, quality, security)
- `make tidy` — run go mod tidy on all modules

## Review & Merge Philosophy

- **Corrections are cheap, waiting is expensive.** Bias toward merging when the change is directionally correct. Fix-up commits are normal, not a sign of failure.
- **Review for correctness, not style.** Linters enforce style mechanically. Human/agent review focuses on logic errors, missing edge cases, architectural violations, and concurrency safety.
- **Every PR must pass CI.** `make ci` and `make guard` — all green before merge. No exceptions.
- **Agent-generated code gets the same scrutiny as human code.** Review the diff, not the author. Agent output can be subtly wrong in ways that look plausible.
- **Self-review before requesting review.** Before marking work as ready, re-read the diff. Check: does it do what was asked? Are there unintended changes? Does it follow the doc impact matrix?

## Workflow

All work is driven by markdown files in `specs/wip/`:

- **PLAN.md** — architecture and design decisions for the current work
- **TODO.md** — task checklist with atomic, committable units

Task cycle:
1. Read PLAN.md and TODO.md to understand current work
2. Execute the next pending task
3. Verify: lint, tests, and build must all pass
4. Review changes together
5. Commit when approved
6. Mark task complete in TODO.md (✓ prefix)
7. Move to next task

**Definition of done:** A task is only complete when:
1. `make ci` and `make guard` pass from root
2. The doc impact matrix has been scanned and any affected docs updated

### Exec Plans

See [docs/exec-plans/index.md](./docs/exec-plans/index.md) for the full lifecycle. When an initiative completes, write a summary to `docs/exec-plans/completed/YYYY-MM-DD-<name>.md` — goal, key decisions, outcome, not the raw plan. Before starting related work, check completed plans for prior decisions and `tech-debt.md` for deferred items.
