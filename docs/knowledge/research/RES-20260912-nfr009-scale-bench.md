---
id: RES-20260912-nfr009-scale-bench
type: research
status: active
owner: cursor-agent
created: 2026-09-12
updated: 2026-09-12
sources: [../../internal/analyzer/scale_fixture.go, ../../internal/analyzer/nfr009_full_test.go, ../../testdata/scale/README.md]
---

# NFR-009 1 GiB storm-fixture measurement

## Question and scope

Does profile `storm` at ≥1 GiB / four instances stay under 512 MiB peak RSS
and five minutes on the documented Linux class? This is analysis (`Analyze` +
report write) in a child process, not fixture generation.

## Sources and freshness

- Generator: `internal/analyzer/scale_fixture.go`
- Probe: `TestNFR009Scale1GiB` (`LOGIFY_NFR009_FULL=1`)
- How to re-run: [`testdata/scale/README.md`](../../../testdata/scale/README.md)
- Status: numbers below are from the implementing agent VM run
  (`LOGIFY_NFR009_FULL=1`, 2026-09-12T10:31Z). Probe-host evidence, not a
  published operator SKU (Q-001).

## Findings

**Host (disclaimer):** Linux 6.12.94+ KVM, 4× Intel Xeon, 16 GiB RAM
(`MemAvailable` at probe time ~7 GiB), Go 1.22.2 linux/amd64. This is the
cloud-agent VM that implemented ADR-0010, not documented product reference
hardware. It meets the documented **class** (Linux amd64, 2+ CPU, 4+ GiB).

**Method:** `EnsureScaleBundle` writes `/tmp/logify-nfr009`; a child `go test`
process runs `Analyze` then `report.Write`. Peak RSS is Linux `VmHWM`.
Generation RSS is excluded.

| Metric | Value | AC |
|---|---|---|
| Input bytes | 1,073,741,887 (1.00 GiB) | AC1 ≥ 1 GiB |
| Instances | 4 (`node-a`, `node-b`, `httpd-a`, `httpd-b`) | AC1 |
| Files / physical records | 5 files; ~12.5 M records | — |
| Unique events after FR-010 | 44 | — |
| Peak RSS (`VmHWM`) | 14,299,136 (13.6 MiB) | AC2 < 512 MiB |
| HeapAlloc / MemStats.Sys | 1.26 MiB / 12.6 MiB | — |
| Analyze+write duration | 59.259 s | AC3 ≤ 5 min |
| Report size | 211,043 bytes | — |
| `BenchmarkNFR009ScaleSmoke` | 8 MiB, 44 unique events, 2.164 s/op (`-benchtime=1x`) | AC4 |

## Conflicts and uncertainty

- The fixture is a repeating storm, not millions of distinct access URLs.
  AC2 success on this corpus does not imply AC2 on unique-heavy bundles.
- `runtime.MemStats` is reported alongside RSS; AC2 is RSS when `/proc` exists.
- AC2/AC3 passed on this storm fixture. They are not a claim about
  unique-heavy access logs.

## Recommendation and promotion targets

NFR-009 is **Implemented** for the documented storm fixture. Re-run
`make bench-nfr009` after merge-path changes that touch scanning, merge,
correlation, or report embedding. Do not treat this as a 512 MiB guarantee
for millions of distinct signatures.
