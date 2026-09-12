---
id: ADR-0006-issue-workflow-usability
type: decision
status: accepted
owner: cursor-agent
created: 2026-09-09
updated: 2026-09-12
requirements: [NFR-021]
supersedes: []
---

# ADR-0006: Issue-workflow keyboard, semantics, and filter probe

Renumbered from a colliding ADR-0004 after FR-015 (#8) and NFR-006 (#10) landed.

## Context and evidence

NFR-021 requires keyboard-operable create/flag/tag/state actions, flag and
state communicated by text (not color alone), immediate visible confirmation,
and responsive filtering of at least 10,000 issues on documented reference
hardware. The Must issue UI already used native controls and a live status
region. Audit of `page.js` found three real AC1–AC3 gaps: card rebuilds dropped
keyboard focus; the tag field had only placeholder text; filter changes had no
visible match count. AC4 had no measured probe.

## Decision

1. Keep rebuilding issue cards after flag, tag, state, and due-date edits (the
   store is the source of truth). Restore focus to the control that triggered
   the rebuild via `data-control` markers.
2. Communicate flag and workflow state with visible text badges (**Flagged**,
   **State: Open**) plus labeled buttons/selects. Color remains decorative.
3. Confirm create/flag/tag/state in `#issue-feedback` (`role="status"`,
   `aria-live="polite"`, `aria-atomic="true"`). Confirm filters with a persistent
   **Showing N of M issue(s)** summary. Immediate `showFeedback` writes cancel
   `detailFeedbackTimer` so a delayed title/owner/notes status cannot overwrite
   the confirmation. Empty `renderIssues` paths apply/clear `pendingFocus`.
4. Measure AC4 with a Node harness that imports 10,000 synthetic issues and
   times `store.filter` only. Document an interactive target of 100 ms median
   and a CI guard of 500 ms. Do not treat the probe host as published reference
   hardware. The 2026-09-09 change did not window the card list; that bound
   landed later (see addendum and ADR-0011).
5. Keep NFR-021 **Partial** until a published reference workstation exists
   (Q-001). Windowing bounds default DOM cost but is not a timed 10,000-card
   paint on reference hardware.

## Alternatives considered

- Virtualizing or capping rendered cards was deferred in the first change.
  Q-002 is now resolved as **pagination of 25 cards** (ADR-0011).
- Adding axe-core or a CDN a11y engine was rejected: NFR-003 / no-CDN, and
  full WCAG tooling is out of the NFR-021 scope. Source-contract tests are the
  in-repo sensor.
- Failing CI on the 100 ms target was rejected: CI hardware varies; the 500 ms
  guard catches only catastrophic regressions.

## Consequences and risks

Keyboard users keep their place after common edits. Screen-reader users hear
status text and named flag/state. Operators with more than 25 matching issues
page the queue. Status must not be advanced to Implemented on the strength of
the Node filter probe alone.

## Addendum (2026-09-12)

The issue queue now slices `store.filter` results to `ISSUE_PAGE_SIZE` (25)
with Previous/Next controls. The filter probe also times a 25-item slice.
Q-002 is answered. Q-001 remains open. See ADR-0011.

## Verification

`nfr021_a11y_test.js` asserts the AC1–AC3 source contracts, including that
`showFeedback` cancels `detailFeedbackTimer` and that empty `renderIssues`
paths call `applyPendingFocus`. `nfr021_filter_probe.js`
builds 10,000 issues, checks filter counts, prints host + timings, times a
25-item window slice, and exits non-zero only if a timed filter exceeds 500 ms.
