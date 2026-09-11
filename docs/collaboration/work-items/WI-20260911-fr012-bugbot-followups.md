---
id: WI-20260911-fr012-bugbot-followups
type: work-item
status: active
owner: cursor-agent
created: 2026-09-11T20:26:00Z
updated: 2026-09-11T20:26:00Z
lease_expires: 2026-09-12T04:26:00Z
scope:
  - internal/analyzer/model.go
  - internal/analyzer/analyzer.go
  - internal/analyzer/correlate.go
  - internal/analyzer/correlate_test.go
  - README.md
  - docs/requirements/functional-requirements.md
  - docs/knowledge/architecture.md
  - docs/knowledge/glossary.md
  - docs/knowledge/decisions/ADR-0008-event-correlation.md
  - docs/knowledge/incidents/INC-20260911-macos-go122-lc-uuid.md
  - docs/collaboration/work-items/WI-20260911-fr012-event-correlation.md
  - docs/collaboration/work-items/WI-20260911-fr012-bugbot-followups.md
  - docs/collaboration/handoffs/HO-20260911-cursor-agent-fr012-bugbot-followups.md
requirements: [FR-012, FR-010]
depends_on: []
supersedes: []
---

# Port FR-012 Bugbot follow-ups after #13

## Goal

PR #13 squash-merged FR-012 to `main` as `76cae98` before Bugbot follow-up
commits on `cursor/fr012-event-correlation-1995` (`df9e817` tip). Port only
those analyzer fixes onto current `main`. Do not replay the whole feature.
Do not merge.

## Non-goals

Replaying FR-012, reopening #13, merging this PR, changing follow-up schema,
adding new correlation rules.

## Acceptance criteria

1. Dedup keeps occurrence hints (message, client IP, timestamp); correlation
   uses all collapsed occurrences for exact request-id and heuristic IP rules.
   Regression: `TestCorrelateUsesCollapsedDuplicateIdentifiers`.
2. Heuristic `client-ip-window` is pairwise only (one access + one Tomcat,
   same non-loopback IP, ≤5s) — no transitive mega-groups. Tests:
   `TestHeuristicDoesNotChainBeyondWindow`,
   `TestHeuristicDoesNotUnionDistinctIPs`. ADR-0008 updated.
3. macOS `-ldflags=-B=gobuildid` remains on `main` CI (already present in
   `76cae98`); do not re-add unless missing.

## Planned files and ownership

Listed in `scope`. Owner: `cursor-agent`.

## Evidence and assumptions

- `origin/main` at `76cae98` includes the FR-012 feature, empty-correlation
  fix, and LC_UUID CI workaround. It does not include `02bd242`.
- Feature-branch tip `df9e817` is docs-only after `02bd242`.
- No active/queued/review dispatch owns FR-012.

## Validation

`gofmt`, `go test ./...`, `go vet`, `go build`, fixture smoke, `git diff --check`.

## Activity log

- `2026-09-11T20:26:00Z` — cursor-agent — claimed paths on
  `cursor/fr012-bugbot-followups-134f` from `origin/main` (`76cae98`).
  Porting `02bd242` only.
- `2026-09-11T20:28:00Z` — cursor-agent — cherry-picked `02bd242` cleanly
  onto `76cae98` as `bd4e805`. CI LC_UUID already on `main`; not replayed.
  Original FR-012 item left in review.

## Handoff or completion

Pending validation and PR.
