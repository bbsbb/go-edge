<!-- last-reviewed: 2026-02-15 -->
# Execution Plans

Permanent, versioned records of engineering initiatives and deferred work. Everything here is committed to git and discoverable by all agents and humans.

## Completed

| Date | Plan | Summary |
|------|------|---------|
| 2026-02-15 | [Sweetshop](completed/2026-02-15-sweetshop-first-application.md) | First application proving all framework layers end-to-end |

## Active

None.

## Structure

```
docs/exec-plans/
├── active/           In-flight initiatives visible to the team
├── completed/        Permanent decision records for finished work
├── tech-debt.md      Categorized deferred work with revisit conditions
└── index.md          This file
```

## What belongs here

| Artifact | Location | Purpose |
|----------|----------|---------|
| Active plan | `active/<name>.md` | Team-visible initiative in progress — others can discover scope, decisions, and status |
| Completed plan | `completed/<name>.md` | Permanent record of why something was built the way it was — future agents reference these before starting related work |
| Tech debt | `tech-debt.md` | Conscious deferrals with origin, reason, and revisit conditions |

## Lifecycle

1. **Ephemeral planning** happens outside this directory (e.g., `specs/wip/` or ad-hoc). Plans at this stage are lightweight and private to the current working session.
2. **Promote to active** when an initiative is team-visible or spans multiple sessions. Create `active/<name>.md` with the plan content.
3. **Move to completed** when the initiative finishes. Write a summary to `completed/YYYY-MM-DD-<name>.md`. The date prefix makes plans sortable chronologically.
4. **Promote on completion** — when an ephemeral plan finishes, write a summary to `completed/YYYY-MM-DD-<name>.md` so the reasoning is preserved.

Not every initiative needs an exec plan. Small, self-contained changes that complete in a single session don't require one. Use exec plans when the decisions and rationale need to survive beyond the current session.

## Format

### Active plans

No rigid template. Capture the goal, approach, key decisions, and scope.

### Completed plans

Completed plans are **summaries**, not raw copies of the original plan. Distill to what a future agent needs:

- **Goal** — one sentence on what problem this solved
- **Decisions** — key choices made and alternatives rejected (briefly)
- **Outcome** — what was actually built, and any deviations from the original intent
- **References** — links to relevant code, docs, or PRs

Keep completed plans short. Context is a scarce resource — a future agent reading this needs the distilled reasoning, not the full deliberation history.
