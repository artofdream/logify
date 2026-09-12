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
- Status: numbers below are filled after the implementing agent VM run.
  They are probe-host evidence, not a published operator SKU (Q-001).

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
| Input bytes | *pending 1 GiB run* | AC1 ≥ 1 GiB |
| Instances | 4 (`node-a`, `node-b`, `httpd-a`, `httpd-b`) | AC1 |
| Unique events | *pending* | — |
| Peak RSS (`VmHWM`) | *pending* | AC2 < 512 MiB |
| Analyze+write duration | *pending* | AC3 ≤ 5 min |
| `BenchmarkNFR009ScaleSmoke` | *pending 8 MiB CI-style bench* | AC4 |

## Conflicts and uncertainty

- The fixture is a repeating storm, not millions of distinct access URLs.
  AC2 success on this corpus does not imply AC2 on unique-heavy bundles.
- `runtime.MemStats` is reported alongside RSS; AC2 is RSS when `/proc` exists.
- If the 1 GiB run is not executed in this change, NFR-009 stays Partial
  with AC2/AC3 Unknown on this host.

## Recommendation and promotion targets

Keep NFR-009 **Partial**. Promote measured rows into this note and the NFR
gap text. Do not mark Implemented unless AC2 and AC3 both pass on the
documented class.
