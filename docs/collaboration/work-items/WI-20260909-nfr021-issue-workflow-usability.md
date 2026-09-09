---
id: WI-20260909-nfr021-issue-workflow-usability
type: work-item
status: active
owner: cursor-agent
created: 2026-09-09T20:54:28Z
updated: 2026-09-09T20:54:28Z
lease_expires: 2026-09-10T08:54:28Z
scope:
  - internal/report/page.html
  - internal/report/page.css
  - internal/report/page.js
  - internal/report/followup_node_test.js
  - internal/report/nfr021_a11y_test.js
  - internal/report/nfr021_filter_probe.js
  - internal/report/report_test.go
  - README.md
  - docs/requirements/non-functional-requirements.md
  - docs/knowledge/architecture.md
  - docs/knowledge/glossary.md
  - docs/knowledge/open-questions.md
  - docs/knowledge/decisions/ADR-0006-issue-workflow-usability.md
  - docs/knowledge/research/RES-20260909-nfr021-issue-filter-probe.md
  - docs/collaboration/work-items/WI-20260909-nfr021-issue-workflow-usability.md
  - docs/collaboration/work-items/WI-20260909-fr021-follow-up-details.md
  - docs/collaboration/dispatches/DSP-20260909-NFR021.md
requirements: [NFR-021]
depends_on: [WI-20260909-fr021-follow-up-details]
supersedes: []
---

# NFR-021 issue-workflow usability

## Goal

Audit the offline HTML issue UI against NFR-021 AC1–AC3, fix real keyboard,
semantics, and confirmation gaps, and add a documented 10,000-issue filter
probe for AC4. Keep status honest: do not mark Implemented without measured
evidence on documented reference hardware.

## Non-goals

FR-012/016, merging PRs, release tags, axe-core or other full WCAG audit
tooling, virtualizing the issue card list, broadening NFR-013.

## Acceptance criteria

- AC1: create, flag, tag, and state controls are keyboard operable; focus is
  restored after card rebuilds; tag field has a programmatic label.
- AC2: flag and state use visible text (not color alone).
- AC3: common actions update a live status region; filter changes show a count.
- AC4: a Node probe builds ~10k issues, measures `store.filter` time, and
  records host + numbers with a reference-hardware disclaimer.

## Planned files and ownership

cursor-agent owns the paths in frontmatter. Analyzer packages are read-only.
WI-20260909-fr021-follow-up-details is review/merged (PR #4) for FR-021; this
item takes the report UI paths for NFR-021 usability only.

## Evidence and assumptions

- Current NFR-021 gap: native controls/labels and a live status region exist;
  no measured 10k filter probe or automated a11y audit.
- `renderAll()` after flag/state/tag/due rebuilds cards and drops keyboard focus.
- Tag add uses placeholder text only (no `<label>` / `aria-label`).
- `store.filter` at 10k is the AC4 probe surface; full DOM render of 10k cards
  is out of this change's windowing scope and remains a documented gap.

## Validation

Pending official `AGENTS.md` commands after implementation.

## Activity log

- `2026-09-09T20:54:28Z` — cursor-agent — claimed NFR-021 after inspecting
  work items and dispatches. No active dispatch owns NFR-021. FR-021 report
  paths are in review on a merged PR; recorded transfer for NFR-021 usability.

## Handoff or completion

In progress.
