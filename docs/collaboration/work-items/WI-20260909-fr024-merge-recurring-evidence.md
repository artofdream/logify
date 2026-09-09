---
id: WI-20260909-fr024-merge-recurring-evidence
type: work-item
status: review
owner: cursor-agent
created: 2026-09-09T19:53:28Z
updated: 2026-09-09T20:17:00Z
lease_expires: 2026-09-10T07:53:28Z
scope:
  - internal/report/followup.js
  - internal/report/followup.go
  - internal/report/page.js
  - internal/report/page.html
  - internal/report/page.css
  - internal/report/followup_node_test.js
  - internal/report/report_test.go
  - README.md
  - docs/requirements/functional-requirements.md
  - docs/knowledge/architecture.md
  - docs/knowledge/glossary.md
  - docs/knowledge/decisions/ADR-0002-follow-up-export-schema.md
  - docs/knowledge/decisions/ADR-0007-recurring-evidence-merge.md
  - docs/collaboration/work-items/WI-20260909-fr024-merge-recurring-evidence.md
  - docs/collaboration/dispatches/DSP-20260909-FR024.md
  - docs/collaboration/dispatches/DSP-20260904-FR023-FR024.md
  - docs/collaboration/handoffs/HO-20260909-cursor-agent-fr024.md
  - docs/collaboration/work-items/WI-20260909-fr021-follow-up-details.md
requirements: [FR-024]
depends_on: [WI-20260909-fr021-follow-up-details]
supersedes: [DSP-20260904-FR023-FR024]
---

# FR-024 merge recurring evidence into an existing issue

## Goal

Let operators link additional timeline event groups to an existing issue, retain
every linked evidence reference across export/import, and surface signature
matches plus newly observed occurrences for review without silently changing
workflow state.

## Non-goals

FR-012 correlation, FR-016 compression/rotation, desktop wrapper, OS CI matrix,
release tagging, merging PRs, schema v2 unless a compatibility break is proven
necessary.

## Acceptance criteria

FR-024 AC1–AC4, verified by Node store tests plus fixture report generation.
Automatic matches must not change `open|investigating|blocked|resolved|dismissed`.

## Planned files and ownership

cursor-agent owns the paths in frontmatter. Analyzer packages are read-only.
FR-021 is on `main` via PR #4 (`01ae8bf`). This work item takes the report
follow-up paths for FR-024 only.

## Evidence and assumptions

- ADR-0001 keeps occurrence counts and observation times out of `evidence-v1-…`
  so a later report can recognize the same group.
- ADR-0002 `evidence` remains the originating snapshot. Additional refs are an
  optional `linkedEvidence` array; schema stays `logify-follow-up-v1` / version 1.
- Recurring automatic match key is `signature` + `instance` (grouping is
  instance-local). Manual link may attach any event group.
- DSP-20260904-FR023-FR024 is cancelled; FR-024 remained Proposed.

## Validation

- `gofmt -l cmd internal` — clean after alignment of `followup.go`
- `go test ./... -count=1` — pass
- `go vet ./...` — pass
- `go build -o logify.exe ./cmd/logify` — pass
- `./logify.exe -output sample-report.html testdata/case` — 6 events, 3 files, 0 warnings
- `git diff --check` — pass
- Browser probes on `file:///workspace/sample-report.html` for link/unlink and
  import review without state change
- PR: https://github.com/artofdream/logify/pull/5 (draft vs `main`)

## Activity log

- `2026-09-09T19:53:28Z` — cursor-agent — claimed FR-024. Fetched `origin/main`
  (`01ae8bf`, PR #4 squash-merged FR-021). No other active/queued/review dispatch
  owns FR-024. Branch `cursor/fr024-merge-recurring-evidence-c99d`.
- `2026-09-09T20:17:00Z` — cursor-agent — Go/Node validation, fixture report,
  and in-browser link/import/review/XSS probes passed. FR-024 marked
  Implemented. Draft PR #5.

## Handoff or completion

In review on PR #5. Do not mark done until the PR is reviewed. Do not merge
from this task.
