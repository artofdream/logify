---
id: logify-open-questions
type: question-ledger
status: active
owner: human
updated: 2026-09-12
---

# Open questions

Record material unknowns as `Q-NNN` with owner, date, evidence needed, and links to
affected requirements or work items. Resolve in place with the answer, evidence,
date, and any promoted ADR/requirement.

## Q-001 — Published reference hardware

- **Status:** Open
- **Owner:** human
- **Opened:** 2026-09-09
- **Requirements:** NFR-009, NFR-021
- **Evidence needed:** An agreed workstation or CI class (CPU, RAM, OS) that
  requirement text can call "documented reference hardware."
- **Notes:** NFR-009 now names a **hardware class** (Linux amd64, 2+ CPU,
  4+ GiB RAM, local disk for a 1 GiB generated tree) in
  [`testdata/scale/README.md`](../../testdata/scale/README.md). That class is
  enough for AC3 wording; a specific SKU is still unset. NFR-021 AC4 still
  records probe-host timings with a disclaimer
  ([RES-20260909-nfr021-issue-filter-probe](research/RES-20260909-nfr021-issue-filter-probe.md)).
  Probe hosts are not a published operator workstation. Card list paging
  (Q-002) does not close this question.

## Q-002 — Issue-queue card windowing at 10k matches

- **Status:** Resolved
- **Owner:** cursor-agent
- **Opened:** 2026-09-09
- **Resolved:** 2026-09-12
- **Requirements:** NFR-021
- **Answer:** Paginate the matching issue list at 25 cards
  (`ISSUE_PAGE_SIZE`). Previous/Next are labeled buttons. `showIssue` jumps
  to the page that contains the issue. Virtualized scrolling was rejected as
  the first bound (ADR-0011).
- **Evidence:** `internal/report/page.js` `renderIssues` slices the filter
  result; `nfr021_filter_probe.js` records `issuePageSize: 25`. NFR-021
  remains Partial because Q-001 is still open.
