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
  Probe hosts are not a published operator workstation.

## Q-002 — Issue-queue card windowing at 10k matches

- **Status:** Open
- **Owner:** human
- **Opened:** 2026-09-09
- **Requirements:** NFR-021
- **Evidence needed:** Operator decision on whether the queue should window,
  paginate, or virtualize when many issues match. ADR-0006 deferred this.
- **Notes:** `store.filter` at n=10,000 is a few milliseconds on the probe
  host. Rendering every matching card is unmeasured and expected to dominate.
