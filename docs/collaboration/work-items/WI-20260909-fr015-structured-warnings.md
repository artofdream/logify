---
id: WI-20260909-fr015-structured-warnings
type: work-item
status: review
owner: cursor-agent
created: 2026-09-09T20:29:00Z
updated: 2026-09-09T20:35:00Z
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

Observed on this branch:

```text
gofmt -w cmd internal
go test ./...
# ? cmd/logify [no test files]
# ok internal/analyzer
# ok internal/report
go build -o logify.exe ./cmd/logify
go vet ./...
git diff --check
./logify.exe -output sample-report.html testdata/case
# Wrote sample-report.html (6 events from 3 files; processed=3 skipped=0 failed=0; 0 warnings)
```

Generated `logify.exe` and `sample-report.html` were deleted after the probe
and were not committed.

Mixed-bundle CLI (temp dir: readable log, 000-mode dir, 000-mode `.log`,
overflow `.log`):

```text
Wrote /tmp/fr015-warning-report.html (2 events from 3 files; processed=2 skipped=1 failed=1; 3 warnings)
  blocked [walk-error] open …/blocked: permission denied
  huge.log:2 [scan-overflow] bufio.Scanner: token too long
  locked.log [open-error] open …/locked.log: permission denied
```

Browser probe of that report: stats 3/2/1/1/3, Scan warnings listed all three
categories with `huge.log:2`, timeline kept `before` and `visible`.

## Activity log

- `2026-09-09T20:29:00Z` — cursor-agent — claimed FR-015 structured warnings
  after fetching `origin/main` at `01ae8bf`. No active/queued dispatch overlaps
  FR-015. Draft PRs #5–#7 left untouched.
- `2026-09-09T20:35:00Z` — cursor-agent — validation commands above passed.
  Draft PR https://github.com/artofdream/logify/pull/8. FR-015 marked
  Implemented from those probes.

## Handoff or completion

PR #8 is open against `main` and must not be merged by this agent.
