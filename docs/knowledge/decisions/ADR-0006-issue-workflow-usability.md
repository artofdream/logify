---
id: ADR-0006-issue-workflow-usability
type: decision
status: accepted
owner: cursor-agent
created: 2026-09-09
updated: 2026-09-10
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
   **Showing N of M issue(s)** summary.
4. Measure AC4 with a Node harness that imports 10,000 synthetic issues and
   times `store.filter` only. Document an interactive target of 100 ms median
   and a CI guard of 500 ms. Do not treat the probe host as published reference
   hardware. Do not window or virtualize the card list in this change.
5. Keep NFR-021 **Partial** until a published reference workstation exists and
   DOM rendering of a 10,000-issue match set is measured or bounded.

## Alternatives considered

- Virtualizing or capping rendered cards was deferred: it would change what
  operators can see without a filter and is a larger product decision
  (see Q-002).
- Adding axe-core or a CDN a11y engine was rejected: NFR-003 / no-CDN, and
  full WCAG tooling is out of the NFR-021 scope. Source-contract tests are the
  in-repo sensor.
- Failing CI on the 100 ms target was rejected: CI hardware varies; the 500 ms
  guard catches only catastrophic regressions.

## Consequences and risks

Keyboard users keep their place after common edits. Screen-reader users hear
status text and named flag/state. Operators with 10,000 issues will still
freeze the tab if they render every card; they should narrow filters. Status
must not be advanced to Implemented on the strength of the Node filter probe
alone.

## Verification

`nfr021_a11y_test.js` asserts the AC1–AC3 source contracts. `nfr021_filter_probe.js`
builds 10,000 issues, checks filter counts, prints host + timings, and exits
non-zero only if a timed filter exceeds 500 ms.
