---
id: WI-20260910-bugbot-followups
type: work-item
status: active
owner: cursor-agent
created: 2026-09-10T16:30:00Z
updated: 2026-09-10T16:30:00Z
lease_expires: 2026-09-10T22:00:00Z
scope:
  - cmd/logify/main.go
  - cmd/logify/main_test.go
  - internal/report/followup.js
  - internal/report/followup_node_test.js
  - internal/report/page.js
  - internal/report/nfr021_a11y_test.js
  - docs/requirements/functional-requirements.md
  - docs/requirements/non-functional-requirements.md
  - docs/knowledge/decisions/ADR-0006-issue-workflow-usability.md
  - docs/knowledge/decisions/ADR-0007-recurring-evidence-merge.md
  - README.md
  - docs/collaboration/work-items/WI-20260910-bugbot-followups.md
  - docs/collaboration/handoffs/HO-20260910-cursor-agent-bugbot-followups.md
requirements: [NFR-011, FR-022, FR-024, NFR-021]
depends_on: []
supersedes: []
---

# Bugbot follow-ups from merged #5 / #10 / #11

## Goal

Fix the Bugbot findings that landed on `main` with PRs #5, #10, and #11, in
one new PR. Do not merge. Do not push to the old PR branches.

## Non-goals

Reopening #5/#10/#11, merging this PR, new features, schema bump.

## Acceptance criteria

1. `-h` / `-help` exit 0 (`flag.ErrHelp`); tested in `cmd/logify`.
2. `importJSON` rejects duplicate evidence ownership across issues with a
   validation skip and reason matching link/create.
3. Recurring review panel shows newer `lastSeen` when the row is lastSeen-only.
4. `showFeedback` cancels `detailFeedbackTimer`.
5. Empty `renderIssues` early returns clear/apply `pendingFocus`.
6. Tests cover each fix.

## Planned files and ownership

Listed in `scope`. Owner: `cursor-agent`.

## Evidence and assumptions

Bugbot threads on merged PRs #5, #10, and #11. Local `main` is `7a96fb1`
(includes those merges).

## Validation

`gofmt`, `go test ./...`, `go vet`, `go build`, fixture smoke, `git diff --check`.

## Activity log

- `2026-09-10T16:30:00Z` — cursor-agent — claimed paths; implementing on
  `cursor/bugbot-followups-7ba8`.

## Handoff or completion
