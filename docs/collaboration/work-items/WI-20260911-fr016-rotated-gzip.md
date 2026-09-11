---
id: WI-20260911-fr016-rotated-gzip
type: work-item
status: review
owner: cursor-agent
created: 2026-09-11T21:42:00Z
updated: 2026-09-11T21:50:00Z
lease_expires: 2026-09-12T06:00:00Z
scope:
  - internal/analyzer/analyzer.go
  - internal/analyzer/analyzer_test.go
  - testdata/rotated/
  - README.md
  - AGENTS.md
  - docs/requirements/functional-requirements.md
  - docs/knowledge/architecture.md
  - docs/knowledge/glossary.md
  - docs/knowledge/decisions/ADR-0009-rotated-gzip-logs.md
  - docs/collaboration/work-items/WI-20260911-fr016-rotated-gzip.md
  - docs/collaboration/handoffs/HO-20260911-cursor-agent-fr016.md
requirements: [FR-016, FR-002, FR-003, FR-010, NFR-001, NFR-004, NFR-008]
depends_on: []
supersedes: []
---

# FR-016 compressed and rotated logs

## Goal

Discover common numeric/date rotated log names and stream `.gz` supported logs
without manual extraction, while keeping FR-010 signature grouping as the only
dedup and preserving existing parsers, correlation, redaction, and warnings.

## Non-goals

tar/zip/bz2/xz archives, gzip magic-byte detection on unsuffixed names, new
warning categories, new third-party dependencies, merging the PR.

## Acceptance criteria

FR-016 AC1–AC3. Gzip is a single-file stream (`compress/gzip`). Discovery
reuses tightened `detect()` / `looks()` conventions after stripping one `.gz`
and one rotation suffix.

## Planned files and ownership

cursor-agent owns the paths in frontmatter.

## Evidence and assumptions

- `main` at `79ac9e0` is clean; no active/queued/review dispatch owns FR-016.
- Review-status work items do not overlap these paths.
- Assumption: invalid gzip after a successful `os.Open` is `scan-error` and
  counts as processed (file was opened; decode failed).

## Validation

- `gofmt -l cmd internal` — clean
- `go test ./...` — pass
- `go build -o logify.exe ./cmd/logify` — pass
- `go vet ./...` — pass
- `git diff --check` — clean
- `./logify.exe -output sample-report.html testdata/case` —
  `6 events from 3 files; processed=3 skipped=0 failed=0; 0 warnings`
- `./logify.exe -output rotated-report.html testdata/rotated` —
  `9 events from 6 files; processed=6 skipped=0 failed=0; 0 warnings`
- PR: https://github.com/artofdream/logify/pull/15 (vs `main`). Do not merge.

## Activity log

- `2026-09-11T21:42:00Z` — cursor-agent — claimed FR-016 on branch
  `cursor/fr016-rotated-gzip-logs-9b7b` from `main` (`79ac9e0`).
- `2026-09-11T21:50:00Z` — cursor-agent — validation passed. FR-016 marked
  Implemented. PR #15 ready for review. Do not merge.

## Handoff or completion
