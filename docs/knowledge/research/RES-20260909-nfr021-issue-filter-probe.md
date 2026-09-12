---
id: RES-20260909-nfr021-issue-filter-probe
type: research
status: active
owner: cursor-agent
created: 2026-09-09
updated: 2026-09-12
sources: [../../internal/report/nfr021_filter_probe.js, ../../internal/report/followup.js, ../../internal/report/page.js]
---

# NFR-021 AC4 issue-filter probe

## Question and scope

Does `LogifyFollowUp` `store.filter` stay within an interactive budget when
10,000 issues are loaded? The probe does **not** paint HTML issue cards and
is **not** a published operator workstation specification. Since 2026-09-12
the live queue pages matches at 25 cards; the probe also times a 25-item
slice.

## Sources and freshness

- Harness: `internal/report/nfr021_filter_probe.js` (stdlib Node, no CDN).
- Store: `internal/report/followup.js` `filter()`.
- Run: 2026-09-09T20:54Z on the cloud-agent VM that implemented this change.
- Re-run: `node internal/report/nfr021_filter_probe.js` (also invoked by
  `go test ./internal/report`).

## Findings

**Host (disclaimer):** Linux 6.12.94+ KVM, 4× Intel Xeon, 15 GiB RAM,
Node v22.14.0, Go 1.22.2. This is the probe host, not documented product
reference hardware.

**Method:** import 10,000 synthetic `logify-follow-up-v1` issues (no
localStorage persist), warm each filter once, then five timed rounds with
`process.hrtime`. Interactive target: 100 ms median. CI guard: 500 ms max.

| Filter | Matched | Median ms | Max ms |
|---|---:|---:|---:|
| all | 10000 | 0.754 | 0.925 |
| text-few | 1 | 3.808 | 4.196 |
| text-many | 10000 | 3.673 | 3.821 |
| flagged | 1429 | 0.474 | 0.506 |
| state-open | 2000 | 0.656 | 0.811 |
| tags | 500 | 1.689 | 1.799 |
| owner | 333 | 0.799 | 0.806 |
| overdue | 546 | 0.851 | 1.214 |
| combined | 143 | 1.149 | 1.179 |

Worst median on this host: **3.8 ms** (text-few). The 100 ms interactive
target was met for `store.filter` on this host. The 500 ms CI guard passed.

**2026-09-12 re-run** on the same class of cloud-agent host (Linux 6.12.94+,
4× Xeon KVM, 16 GiB, Node v22.14.0): worst median **5.2 ms** (text-many);
windowed slice stayed 25 items on the `all` filter. Still under 100 ms.
Still not reference hardware.

## Conflicts and uncertainty

- Later CI or developer machines will print different milliseconds. Treat the
  table as a dated snapshot; the live probe output is the current measurement.
- `filter()` concatenates several fields and lowercases them per issue. That is
  why text search is slower than equality filters. It is still far under 100 ms
  at n=10,000 on this host.
- Rendering 10,000 full issue cards (each with multiple inputs) is no longer
  the default path: `renderIssues` pages at `ISSUE_PAGE_SIZE = 25`. The probe
  still does not paint those 25 cards in a browser. NFR-021 AC4 remains
  gapped for published reference hardware (Q-001).

## Recommendation and promotion targets

Keep NFR-021 **Partial**. Q-002 is resolved (paginate 25). Do not mark AC4
Implemented until a published reference workstation is agreed (Q-001) or a
browser probe on that hardware shows the paged queue stays responsive.
