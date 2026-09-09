---
id: WI-20260904-must-issue-ui
type: work-item
status: done
owner: cursor-agent
created: 2026-09-04T19:45:00Z
updated: 2026-09-09T20:24:00Z
lease_expires: null
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

- `gofmt -l cmd internal` — clean
- `go test ./... -count=1` — pass (`analyzer`, `report`, including Node
  `followup_node_test.js` and `node --check` on page scripts)
- `go vet ./...` — pass
- `go build -o /tmp/logify ./cmd/logify` — pass
- `/tmp/logify -output /tmp/sample-report.html testdata/case` — 6 events from
  3 files, 0 warnings; embedded `events`/`warnings` are arrays; each event
  has `evidence-v1-…`; no `http://` / `https://`
- `git diff --check` — pass
- Browser (file:///tmp/sample-report.html): create issue, title edit, tags
  (case-insensitive unique), flag, state open→investigating→resolved (evidence
  kept), filters, issue↔evidence navigation, export
  `logify-follow-up-v1` / schemaVersion 1, clear, import restore, mobile 375px
  readable. No page-script errors.
- PR: https://github.com/artofdream/logify/pull/3 (merged vs `main`, on top of #2)
- Merge (verified 2026-09-09, GitHub API + `origin/main`): squash
  `64bbc1c7c42763ee824dd962542ee0835faa2d51` at 2026-09-04T20:28:16Z.
  This close-out checked merge evidence only; it did not re-run the Go/Node
  or in-browser probes above.

## Activity log

- `2026-09-04T19:45:00Z` — cursor-agent — claimed report/docs paths after
  inspecting PR #2 work item (review, crash-fix only) and sequential
  FR-017–FR-024 dispatches. Recorded explicit takeover: human asked for one
  combined Must UI PR on the crash-fix branch.
- `2026-09-04T20:20:00Z` — cursor-agent — implementation, Go/Node tests,
  fixture generation, and in-browser create/export/import verification
  complete; draft PR #3 opened for review.
- `2026-09-09T19:21:31Z` — cursor-agent — PR #3 is on `main` (`64bbc1c`).
  FR-021 editor work transfers to WI-20260909-fr021-follow-up-details /
  DSP-20260909-FR021. This item keeps the merged Must UI scope only.
- `2026-09-09T20:24:00Z` — cursor-agent — marked done. PR #3 squash
  `64bbc1c` (2026-09-04T20:28:16Z) verified on `origin/main`. FR-021 later
  merged as PR #4 (`01ae8bf`, 2026-09-09T19:52:40Z). FR-024 is draft PR #5
  and is out of this item. Lease released.

## Handoff or completion

Done. Must UI (FR-017…020, FR-022, FR-023) merged as PR #3 on
2026-09-04T20:28:16Z (`64bbc1c7c42763ee824dd962542ee0835faa2d51`).
Follow-on: FR-021 merged as PR #4 (`01ae8bf`); FR-024 remains draft PR #5.
