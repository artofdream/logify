---
id: WI-20260911-correlation-dedup-window
type: work-item
status: active
owner: cursor-agent
created: 2026-09-11T19:34:00Z
updated: 2026-09-11T19:39:00Z
lease_expires: 2026-09-12T03:39:00Z
scope:
  - internal/analyzer/model.go
  - internal/analyzer/analyzer.go
  - internal/analyzer/correlate.go
  - internal/analyzer/correlate_test.go
  - docs/knowledge/decisions/ADR-0008-event-correlation.md
  - docs/collaboration/work-items/WI-20260911-correlation-dedup-window.md
requirements: [FR-012]
depends_on: []
supersedes: []
---

# Fix FR-012 correlation after dedup and heuristic window

## Goal

Keep collapsed-duplicate identifiers available to correlation, and emit
`client-ip-window` groups only for access/Tomcat pairs inside the 5s bound
(FR-012 AC5 / ADR-0008).

## Non-goals

Changing report UI, redaction, CLI flags, or adding new correlation rules.

## Acceptance criteria

- Recurring same-signature events still match later request IDs and client
  addresses (exact and heuristic).
- Heuristic groups do not transitively exceed 5s or union distinct IPs.

## Planned files and ownership

cursor-agent owns the paths in frontmatter. Overlaps
`WI-20260911-fr012-event-correlation` (same owner); this item is the
bugfix follow-up on `cursor/analyzer-correlation-issues-74ae`.

## Evidence and assumptions

- Dedup keeps the first walked Message/ClientAddr; signatures strip numeric
  tokens so later labeled IDs never reach `Correlate`.
- Union-find over all in-window pairs implements transitive closure.

## Validation

Observed:

- `gofmt -w cmd internal`
- `go test ./internal/analyzer ./internal/redact ./cmd/logify` — pass
- `go build -o logify.exe ./cmd/logify` — pass
- `go vet ./...` — pass
- `git diff --check` — pass
- `./logify.exe -output sample-report.html testdata/case` — 6 events, 0 warnings

Pre-existing (same failure on `b8c0ad7` before this change):
`go test ./internal/report` — `TestWriteCorrelateGroupsExactAndHeuristic` looks for the literal `Heuristic (confidence: low)` which is assembled in `page.js`, not present as that substring.

## Activity log

- `2026-09-11T19:34:00Z` — cursor-agent — claimed bugfix follow-up.
- `2026-09-11T19:39:00Z` — cursor-agent — retained collapsed-occurrence hints for correlation; heuristic groups are now in-window pairs.
