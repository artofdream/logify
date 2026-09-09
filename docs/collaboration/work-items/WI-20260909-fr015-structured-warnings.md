---
id: WI-20260909-fr015-structured-warnings
type: work-item
status: active
owner: cursor-agent
created: 2026-09-09T20:29:00Z
updated: 2026-09-09T20:29:00Z
lease_expires: 2026-09-10T08:00:00Z
scope:
  - cmd/logify/main.go
  - internal/analyzer/analyzer.go
  - internal/analyzer/analyzer_test.go
  - internal/analyzer/model.go
  - internal/report/report.go
  - internal/report/report_test.go
  - internal/report/page.html
  - internal/report/page.js
  - internal/report/page.css
  - README.md
  - docs/requirements/functional-requirements.md
  - docs/knowledge/decisions/ADR-0004-scan-warnings.md
  - docs/knowledge/architecture.md
  - docs/knowledge/glossary.md
  - docs/collaboration/work-items/WI-20260909-fr015-structured-warnings.md
requirements: [FR-015]
depends_on: []
supersedes: []
---

# FR-015 structured scan warnings and input counts

## Goal

Close the remaining FR-015 gaps: structured warnings (file, category, optional
line/range, message) and reconciled processed / skipped / failed input counts
in the analyzer result, CLI summary, and HTML report. Preserve AC1–3 and
overflow-keeps-events.

## Non-goals

NFR-002 OS matrix, NFR-006 redaction, NFR-021 a11y, FR-012/016, release tags,
desktop UI, and merging draft PRs #5–#7.

## Acceptance criteria

1. Unreadable supported file or scanner failure → warning; scan continues.
2. Recoverable errors do not discard successfully parsed files or earlier events.
3. CLI and report expose the warning count.
4. Each warning identifies file, deterministic category, and line/range when known.
5. Summary counts reconcile processed, skipped, and failed inputs.

## Planned files and ownership

Owned paths are listed in `scope`. No other agent should edit those files while
this lease is active.

## Evidence and assumptions

- Main gap text (FR-015): warnings are `path: error` strings; overflow already
  keeps earlier events; counts are only `filesScanned` + warning list.
- Code paths that emit warnings today: WalkDir errors, `os.Open` failures, and
  `bufio.Scanner` errors (including `ErrTooLong`).
- Assumption: unsupported filenames are a discovery filter, not skipped inputs.
  Skipped counts walk paths that could not be visited.

## Validation

```text
gofmt -w cmd internal
go test ./...
go build -o logify.exe ./cmd/logify
go vet ./...
git diff --check
./logify.exe -output sample-report.html testdata/case
```

## Activity log

- `2026-09-09T20:29:00Z` — cursor-agent — claimed FR-015 structured warnings
  after fetching `origin/main` at `01ae8bf`. No active/queued dispatch overlaps
  FR-015. Draft PRs #5–#7 left untouched.

## Handoff or completion
