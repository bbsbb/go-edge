<!-- last-reviewed: 2026-02-15 content-hash: 1a92019c -->
# Core Beliefs

Agent-first operating principles that govern how this codebase is built and maintained. These are not aspirational — they describe the rules we actually follow and enforce.

## Repository is the system of record

If it's not in the repository, it doesn't exist. Design decisions, architectural constraints, quality grades, technical debt, and operating procedures are all versioned in-repo as markdown. Slack conversations, meeting notes, and tribal knowledge must be encoded into the codebase to be actionable.

## Mechanical enforcement over documentation

When a rule matters, encode it in tooling. Documentation describes intent; linters, CI jobs, and structural tests enforce it. A rule that exists only in prose will be violated — a rule backed by a failing build will not. When documentation falls short, promote the rule into code.

## Boring technology

Prefer stable, well-known technologies with broad ecosystem support. Go stdlib, pgx, Chi, slog, SQLC — these are composable, API-stable, and well-represented in training data. Agents reason more reliably about boring technology. Only introduce novel dependencies when the alternative is significantly worse.

## Boundary validation, not defensive coding

Validate at system boundaries (user input, external APIs, configuration loading). Trust internal code and framework guarantees. Don't add error handling, fallbacks, or nil checks for scenarios that can't happen within the application's own call graph. The right amount of defensive code is the minimum needed at the edges.

## Progressive disclosure

CLAUDE.md is the map, not the manual. It points to ARCHITECTURE.md, which points to docs/, which points to code. Each layer adds detail without repeating the layer above. Agents start with a small, stable entry point and navigate deeper as needed. Never front-load all context — teach where to look.

## Strict layers, local autonomy

Enforce architectural boundaries centrally (dependency direction, forbidden imports, structured logging). Allow freedom within those boundaries. The layered architecture (domain → service → infrastructure → transport) is mechanically enforced via depguard. How solutions are expressed within a layer is flexible, as long as boundaries are respected.

## Conscious debt over accidental drift

When work is deferred, record it explicitly with origin, reason, and revisit conditions. All deferred work lives in [`docs/exec-plans/tech-debt.md`](./exec-plans/tech-debt.md), categorized by type. Untracked shortcuts compound silently; tracked deferrals are manageable.
