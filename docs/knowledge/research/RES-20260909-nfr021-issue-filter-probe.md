---
id: RES-20260909-nfr021-issue-filter-probe
type: research
status: active
owner: cursor-agent
created: 2026-09-09
updated: 2026-09-09
sources: [../../internal/report/nfr021_filter_probe.js, ../../internal/report/followup.js]
---

# NFR-021 AC4 issue-filter probe

## Question and scope

Does `LogifyFollowUp` `store.filter` stay within an interactive budget when
10,000 issues are loaded? The probe does **not** render HTML issue cards and
is **not** a published operator workstation specification.

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

## Conflicts and uncertainty

- Later CI or developer machines will print different milliseconds. Treat the
  table as a dated snapshot; the live probe output is the current measurement.
- `filter()` concatenates several fields and lowercases them per issue. That is
  why text search is slower than equality filters. It is still far under 100 ms
  at n=10,000 on this host.
- Rendering 10,000 full issue cards (each with multiple inputs) was not
  measured and is expected to be the dominant cost. NFR-021 AC4 therefore
  remains gapped for the on-page workflow.

## Recommendation and promotion targets

Keep NFR-021 **Partial**. Promote this note from the requirement evidence
block. Do not mark AC4 Implemented until a published reference workstation is
agreed (Q-001) and either card rendering is windowed (Q-002) or a browser
probe shows a 10,000-match render stays responsive.
