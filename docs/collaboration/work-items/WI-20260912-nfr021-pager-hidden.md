---
id: WI-20260912-nfr021-pager-hidden
type: work-item
status: review
owner: cursor-agent
created: 2026-09-12T10:42:00Z
updated: 2026-09-12T10:43:00Z
lease_expires: 2026-09-13T10:43:00Z
scope:
  - internal/report/page.css
  - internal/report/a11y_test.go
  - internal/report/nfr013_a11y_test.js
  - docs/collaboration/work-items/WI-20260912-nfr021-pager-hidden.md
requirements: [NFR-021, NFR-013]
depends_on: [WI-20260912-nfr013-nfr021-a11y]
supersedes: []
---

# Hide issue pager when display:flex would show it

## Goal

Land the pager-hide fix that missed #19's squash merge. `.issue-pager
{ display: flex }` overrides the HTML `hidden` attribute, so Previous/Next
stay visible on queues with ≤25 matches.

## Non-goals

Reopening #19, changing NFR-021 status, window-size changes.

## Acceptance criteria

- `.issue-pager[hidden] { display: none; }` is in `page.css` and the
  generated report.
- Source-contract and generated-report tests fail if the override returns.

## Planned files and ownership

cursor-agent owns the paths above. WI-20260912-nfr013-nfr021-a11y remains
review after #19 merged without this CSS rule.

## Evidence and assumptions

Browser walk on #19 before merge showed Previous/Next on a one-issue
queue. The fix was on `ea0bfc9` after merge SHA `72a73c9`.

## Validation

- `gofmt -l cmd internal` — clean
- `go test ./internal/report ./internal/harness -count=1` — pass
- `go build -o logify.exe ./cmd/logify` — pass
- `go vet ./internal/report` — pass
- `git diff --check` — pass
- `node internal/report/nfr013_a11y_test.js` — ok
- Draft PR: https://github.com/artofdream/logify/pull/20

## Activity log

- `2026-09-12T10:42:00Z` — cursor-agent — #19 auto-merged without the
  pager-hide commit. Claimed a follow-up on current `main`.

- `2026-09-12T10:43:00Z` — cursor-agent — CSS + contracts on
  `cursor/nfr021-hide-issue-pager-8155`. Local report/harness tests passed.
  Draft PR #20. Do not merge unless requested.

## Handoff or completion

In review on draft PR #20. Do not merge unless requested.
