---
id: WI-20260909-fr021-follow-up-details
type: work-item
status: review
owner: cursor-agent
created: 2026-09-09T19:21:31Z
updated: 2026-09-09T19:45:00Z
lease_expires: 2026-09-10T07:21:31Z
scope:
  - internal/report/followup.js
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
  - docs/knowledge/decisions/ADR-0003-follow-up-details.md
  - docs/collaboration/work-items/WI-20260909-fr021-follow-up-details.md
  - docs/collaboration/work-items/WI-20260904-must-issue-ui.md
  - docs/collaboration/dispatches/DSP-20260909-FR021.md
  - docs/collaboration/dispatches/DSP-20260904-FR020-FR021.md
  - docs/collaboration/dispatches/DSP-20260904-MUST-ISSUE-UI.md
  - docs/collaboration/handoffs/HO-20260909-1949-cursor-agent-fr021.md
requirements: [FR-021]
depends_on: [WI-20260904-must-issue-ui]
supersedes: [DSP-20260904-FR020-FR021]
---

# FR-021 record follow-up details

## Goal

Add offline HTML editors for issue notes, optional owner, and optional due date;
persist through existing `logify-follow-up-v1` localStorage and export/import;
visibly mark overdue unresolved issues; and filter the issue queue by owner and
overdue state.

## Non-goals

FR-024 merge recurring evidence, FR-012 correlation, FR-016 compression/rotation,
desktop wrapper, OS CI matrix, release tagging, schema v2 unless a compatibility
break is proven necessary.

## Acceptance criteria

FR-021 AC1–AC4, verified by Node store tests plus fixture report generation.
Notes/owner render through DOM text APIs only.

## Planned files and ownership

cursor-agent owns the paths in frontmatter. Analyzer packages are read-only.
WI-20260904-must-issue-ui is merged on `main` (PR #3) and left FR-021 Proposed;
this work item takes the report UI paths for FR-021 only.

## Evidence and assumptions

- DSP-20260904-FR020-FR021 (cancelled) and the FR-023 note state that imported
  owner/due/notes already display and round-trip; editors were the remaining gap.
- ADR-0002 already includes `owner`, `due`, and `notes` in schema v1. No bump
  unless editors need a new field.
- Overdue = due calendar date strictly before the UTC calendar date of the report
  clock, and state is not `resolved` or `dismissed` (existing `isOverdue`).

## Validation

- `gofmt -l cmd internal` — clean
- `go test ./... -count=1` — pass (`analyzer`, `report`, including
  `followup_node_test.js` FR-021 block and `node --check` on page scripts)
- `go vet ./...` — pass
- `go build -o logify.exe ./cmd/logify` — pass
- `./logify.exe -output sample-report.html testdata/case` — 6 events from 3
  files, 0 warnings; report contains Issue queue, Clear due date, Overdue only,
  `logify-follow-up-v1`; no `http://` / `https://`
- `git diff --check` — pass
- Browser `file:///workspace/sample-report.html`: create issue; edit owner
  (Ada Lovelace), due 2020-01-15, notes `<script>alert(1)</script>` as text
  (no execution); Overdue badge while open; owner and Overdue-only filters;
  resolved hides overdue; Clear due date; ~375px fields readable
- PR: https://github.com/artofdream/logify/pull/4 (draft vs `main`)

## Activity log

- `2026-09-09T19:21:31Z` — cursor-agent — claimed FR-021 after fetching `main`
  (`64bbc1c`). Inspected active/review work items: MUST-ISSUE-UI is review/merged
  and explicitly left FR-021 out of scope. No other active dispatch owns FR-021.
- `2026-09-09T19:26:00Z` — cursor-agent — added store editors, overdue badge,
  owner/overdue filters (existing chrome), ADR-0003, and Node tests. Official
  validation still pending; FR-021 status remains Proposed until probes pass.
- `2026-09-09T19:49:00Z` — cursor-agent — Go/Node validation, fixture report,
  and in-browser editor/overdue/filter/XSS probes passed. Fixed a missing
  `</label>` that broke `html/template`. FR-021 marked Implemented. Draft PR #4.

## Handoff or completion

In review on PR #4. Do not mark done until the PR is reviewed. Do not merge
from this task.
