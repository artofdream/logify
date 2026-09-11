---
id: HO-20260911-cursor-agent-fr012-bugbot-followups
type: handoff
status: review
owner: cursor-agent
created: 2026-09-11T20:30:00Z
work_item: WI-20260911-fr012-bugbot-followups
requirements: [FR-012, FR-010]
next_owner: human
---

# Port FR-012 Bugbot follow-ups after #13

## Scope and requirement IDs

FR-012 / FR-010 analyzer follow-ups that missed the #13 squash merge.
Branch `cursor/fr012-bugbot-followups-134f`. PR
https://github.com/artofdream/logify/pull/14 vs `main`. Do not merge.

## Evidence consulted

- `origin/main` `76cae98` (squash of #13: feature + empty-slice + LC_UUID)
- `origin/cursor/fr012-event-correlation-1995` tip `df9e817`
- Cherry-pick source `02bd242` (analyzer + ADR-0008 + FR-012 notes)
- Confirmed `.github/workflows/ci.yml` on `main` already has
  `-ldflags=-B=gobuildid` (golang/go#68678)

## Changes and artifacts

Cherry-picked `02bd242` onto `76cae98` as `bd4e805`:

- Dedup keeps `occHints` (message, client IP, timestamp); correlation
  reads every collapsed occurrence
- `client-ip-window` emits pairwise access+Tomcat groups only
- Tests: `TestCorrelateUsesCollapsedDuplicateIdentifiers`,
  `TestHeuristicDoesNotChainBeyondWindow`,
  `TestHeuristicDoesNotUnionDistinctIPs`
- ADR-0008, FR-012 notes, README, architecture, glossary updated
- CI workflow unchanged (already on `main`)

## Validation and observed results

- `gofmt -w cmd internal` — no extra diff
- `go test ./... -count=1` — pass (`cmd/logify`, `analyzer`, `redact`, `report`)
- Targeted analyzer tests — pass
- `go build -o logify.exe ./cmd/logify` — pass
- `go vet ./...` — pass
- `git diff --check` — pass
- `./logify.exe -output sample-report.html testdata/case` — 6 events, 3 files,
  0 warnings, `"correlations":[]`
- `./logify.exe -output correlate-report.html testdata/correlate` — 14 events,
  3 files, 0 warnings, `shared-request-id` (exact) + `client-ip-window`
  (heuristic)

## Assumptions and confidence

`df9e817` was docs-only on the old PR branch and was not cherry-picked.
Work-item notes for this port replace that record.

## Failures or conflicting evidence

None observed in the commands above.

## Uncommitted or concurrent changes

None intended at handoff. Generated HTML reports are gitignored and were
not committed.

## Open questions and next action

Review PR #14. Do not merge from this task.
