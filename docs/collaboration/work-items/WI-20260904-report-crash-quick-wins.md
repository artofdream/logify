---
id: WI-20260904-report-crash-quick-wins
type: work-item
status: review
owner: cursor-agent
created: 2026-09-04T19:40:00Z
updated: 2026-09-04T19:50:00Z
lease_expires: 2026-09-05T00:00:00Z
scope:
  - cmd/logify/
  - internal/analyzer/analyzer.go
  - internal/analyzer/analyzer_test.go
  - internal/report/report.go
  - internal/report/report_test.go
  - README.md
  - LICENSE
  - docs/requirements/functional-requirements.md
  - docs/requirements/non-functional-requirements.md
  - docs/framework-adoption.md
  - .github/workflows/ci.yml
  - docs/collaboration/work-items/WI-20260904-report-crash-quick-wins.md
requirements: [FR-011, FR-013, FR-014, FR-015, NFR-002, NFR-004, NFR-006, NFR-007, NFR-008, NFR-014, NFR-022]
depends_on: []
supersedes: []
---

# Report crash and review quick wins

## Goal

Fix the happy-path HTML report crash (`warnings`/`events` JSON null) and the
other bounded review quick wins: FR-011 test + README timezone honesty,
tighter `detect()`, keep events on mid-file overflow, honest FR-015/NFR-002
status, draft-PR auto-merge skip, LICENSE, and a report-sensitivity note.

## Non-goals

FR-017 issue tracking, parser rewrite, OS CI matrix, redaction, gzip/rotation.

## Acceptance criteria

- Fixture report JSON arrays are non-null; `warnings.length` does not throw.
- Tests cover time-filter README window, detect false positive, overflow keep.
- Docs status matches observed behavior.

## Planned files and ownership

cursor-agent owns the paths in frontmatter.

## Evidence and assumptions

Source: default-branch review of `main` @ `4ff06e0`. MIT chosen; repo had no
license. FR-017 dispatch is active on other files and is not in this scope.

## Validation

- `gofmt -l cmd internal` — clean
- `go test ./... -count=1` — pass (`analyzer`, `report`)
- `go vet ./...` — pass
- `go build -o /tmp/logify ./cmd/logify` — pass
- `/tmp/logify -output /tmp/sample-report.html testdata/case` — 6 events, 3 files, 0 warnings
- Node parse of embedded JSON: `events` and `warnings` are arrays; `warnings.length` is 0 (no throw)
- `git diff --check` — pass
- PR: https://github.com/artofdream/logify/pull/2

## Activity log

- `2026-09-04T19:40:00Z` — cursor-agent — claimed bounded quick-win paths after
  confirming no overlapping active work item on these files.
- `2026-09-04T19:50:00Z` — cursor-agent — implementation and local validation
  complete; PR #2 opened for review.

## Handoff or completion

In review on PR #2. Do not mark done until the PR is reviewed and merged.
