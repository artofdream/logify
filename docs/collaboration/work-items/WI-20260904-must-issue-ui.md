---
id: WI-20260904-must-issue-ui
type: work-item
status: active
owner: cursor-agent
created: 2026-09-04T19:45:00Z
updated: 2026-09-04T19:45:00Z
lease_expires: 2026-09-05T08:00:00Z
scope:
  - internal/report/
  - README.md
  - docs/requirements/functional-requirements.md
  - docs/requirements/non-functional-requirements.md
  - docs/knowledge/architecture.md
  - docs/knowledge/decisions/ADR-0001-follow-up-identities.md
  - docs/knowledge/decisions/ADR-0002-follow-up-export-schema.md
  - docs/framework-adoption.md
  - docs/collaboration/work-items/WI-20260904-must-issue-ui.md
  - docs/collaboration/dispatches/DSP-20260904-FR017.md
  - docs/collaboration/dispatches/DSP-20260904-FR018-FR019.md
  - docs/collaboration/dispatches/DSP-20260904-FR020-FR021.md
  - docs/collaboration/dispatches/DSP-20260904-FR022.md
  - docs/collaboration/dispatches/DSP-20260904-FR023-FR024.md
  - docs/collaboration/dispatches/DSP-20260904-MUST-ISSUE-UI.md
requirements: [FR-017, FR-018, FR-019, FR-020, FR-022, FR-023, NFR-003, NFR-004, NFR-005, NFR-017, NFR-018, NFR-019, NFR-020]
depends_on: [WI-20260904-report-crash-quick-wins]
supersedes: [DSP-20260904-FR017, DSP-20260904-FR018-FR019, DSP-20260904-FR020-FR021, DSP-20260904-FR022, DSP-20260904-FR023-FR024]
---

# Must working issue UI

## Goal

Grow the self-contained offline HTML report into a working issue-tracking UI
for Must requirements FR-017, FR-018, FR-019, FR-020, FR-022, and FR-023.
Branch from `cursor/report-crash-quick-wins-decc` (PR #2). Do not rewrite the
analyzer.

## Non-goals

FR-021 editable notes/owner/due (display imported values only), FR-024 merge
recurring evidence, FR-012 correlation, gzip/rotation, desktop wrapper, OS CI
matrix, analyzer identity/first-seen rewrite.

## Acceptance criteria

The listed Must FR acceptance criteria, plus stdlib-only single-file offline
HTML, XSS discipline, and tests that exercise follow-up JS (create +
export/import), not only Go JSON embedding.

## Planned files and ownership

cursor-agent owns the paths in frontmatter. Analyzer packages are read-only.
PR #1 (`codex/FR-017-issue-creation`) is a learning source only; this work
implements the report-layer design cleanly on the crash-fix branch.

## Evidence and assumptions

- Explicit human request to combine Must follow-up FRs in one PR against
  `main`, taking over the sequential cloud dispatches.
- Evidence IDs can be derived in `internal/report` from existing Event fields
  (signature, instance, file, line) without changing the analyzer.
- firstSeen is the representative event timestamp (first discovered row);
  lastSeen is the existing analyzer `LastSeen`. This is not a chronological
  min unless discovery order matches time order.
- Observed-record counts are the sum of `occurrences` (raw rows grouped into
  the timeline). Event groups are `len(events)`.

## Validation

Pending implementation.

## Activity log

- `2026-09-04T19:45:00Z` — cursor-agent — claimed report/docs paths after
  inspecting PR #2 work item (review, crash-fix only) and sequential
  FR-017–FR-024 dispatches. Recorded explicit takeover: human asked for one
  combined Must UI PR on the crash-fix branch.

## Handoff or completion
