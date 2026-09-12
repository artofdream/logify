---
id: HO-20260912-cursor-agent-nfr009
type: handoff
status: proposed
owner: cursor-agent
created: 2026-09-12T10:35:00Z
work_item: WI-20260912-nfr009-scale-bench
requirements: [NFR-009, NFR-008, NFR-014, NFR-028]
next_owner: human
---

# Handoff: NFR-009 scale bench and online merge

## Scope and requirement IDs

NFR-009 (Implemented for the documented storm fixture). PR #18 against `main`.
Do not merge from this handoff.

## Evidence consulted

`docs/principles.md`, collaboration protocol, work-item/dispatch scan (no
active NFR-009 owner), NFR-009 ACs, analyzer parse/dedup/occHints path.

## Changes and artifacts

- Online `eventMerger`; `parseFile` emits records; extra `occHints` capped
- Generated storm fixture (`internal/analyzer/scale_fixture.go`)
- `TestNFR009ScaleSmoke`, `TestNFR009Scale1GiB`, `BenchmarkNFR009ScaleSmoke`
- `Makefile`, Linux CI smoke bench, `testdata/scale/README.md`
- ADR-0010, RES-20260912, NFR-009 Implemented with measured numbers

## Validation and observed results

Ran:

- `gofmt -w cmd internal` — clean
- `go test ./...` — pass
- `go build -o logify.exe ./cmd/logify` — pass
- `go vet ./...` — pass
- `git diff --check` — pass
- `./logify.exe -output sample-report.html testdata/case` — 6 events
- `BenchmarkNFR009ScaleSmoke` — 8 MiB, 44 unique events, 2.164 s/op
- `LOGIFY_NFR009_FULL=1` `TestNFR009Scale1GiB` — input 1,073,741,887 bytes,
  4 instances, 44 events, peak RSS 14,299,136 (13.6 MiB), 59.259 s,
  report 211,043 bytes. AC1–AC3 PASS on this host.

## Assumptions and confidence

High for the storm fixture and the streaming merge. Unique-heavy access
logs were not measured and are documented as outside this fixture.
Reference hardware is a class; the VM is a probe host (Q-001 still open
for a SKU).

## Failures or conflicting evidence

None on the storm fixture. 8 MiB smoke Analyze (~2.2 s) is slower per byte
than the 1 GiB run (~59 s / 1 GiB) because `-benchtime=1x` is one cold
iteration; do not scale the smoke ns/op linearly.

## Uncommitted or concurrent changes

None at handoff after the measurement commit.

## Open questions and next action

Human review of PR #18. Do not merge from this agent. Re-run
`make bench-nfr009` if scanning/merge/report embedding changes.
