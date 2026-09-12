---
id: WI-20260912-nfr021-pager-hidden
type: work-item
status: active
owner: cursor-agent
created: 2026-09-12T10:42:00Z
updated: 2026-09-12T10:42:00Z
lease_expires: 2026-09-12T22:42:00Z
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

Pending.

## Activity log

- `2026-09-12T10:42:00Z` — cursor-agent — #19 auto-merged without the
  pager-hide commit. Claimed a follow-up on current `main`.

## Handoff or completion
