---
id: WI-20260911-fr012-event-correlation
type: work-item
status: active
owner: cursor-agent
created: 2026-09-11T19:25:43Z
updated: 2026-09-11T19:25:43Z
lease_expires: 2026-09-12T03:25:43Z
scope:
  - internal/analyzer/model.go
  - internal/analyzer/analyzer.go
  - internal/analyzer/correlate.go
  - internal/analyzer/correlate_test.go
  - internal/analyzer/analyzer_test.go
  - testdata/correlate/
  - internal/redact/redact.go
  - internal/redact/redact_test.go
  - internal/report/report.go
  - internal/report/report_test.go
  - internal/report/page.html
  - internal/report/page.js
  - internal/report/page.css
  - README.md
  - docs/requirements/functional-requirements.md
  - docs/requirements/non-functional-requirements.md
  - docs/knowledge/architecture.md
  - docs/knowledge/glossary.md
  - docs/knowledge/decisions/ADR-0005-optional-report-redaction.md
  - docs/knowledge/decisions/ADR-0008-event-correlation.md
  - docs/collaboration/work-items/WI-20260911-fr012-event-correlation.md
requirements: [FR-012, FR-023]
depends_on: []
supersedes: []
---

# FR-012 correlate related events

## Goal

Add deterministic, documented correlation groups so operators can see related
HTTPD and Tomcat events that share exact identifiers, plus an optional labeled
heuristic, without inventing unsupported links.

## Non-goals

FR-016 compression/rotation, changing follow-up export schema, auto-creating
issues from correlations, unlabeled number/UUID matching, IPv6 client-window
extraction from free text, merging PRs.

## Acceptance criteria

FR-012 AC1–AC5. Events stay individually inspectable. Heuristics are labeled
and lower confidence. Prefer no group over a weak unsupported relationship.

## Planned files and ownership

cursor-agent owns the paths in frontmatter. Follow-up store files are
read-only.

## Evidence and assumptions

- No active/queued/review dispatch owns FR-012. `main` at `4c8da68` is clean.
- Exception `Caused by:` chains are already one event; they are not a
  cross-event correlation rule.
- Access-log client addresses are not in the normalized message, so the
  analyzer must retain `ClientAddr` for the heuristic.

## Validation

Pending `gofmt`, `go test ./...`, `go vet`, build, fixture smoke.

## Activity log

- `2026-09-11T19:25:43Z` — cursor-agent — claimed FR-012 on branch
  `cursor/fr012-event-correlation-1995` from `main` (`4c8da68`).

## Handoff or completion
