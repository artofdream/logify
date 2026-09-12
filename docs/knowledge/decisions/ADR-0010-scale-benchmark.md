---
id: ADR-0010-scale-benchmark
type: decision
status: accepted
owner: cursor-agent
created: 2026-09-12
updated: 2026-09-12
requirements: [NFR-009, NFR-008, FR-010, FR-012]
supersedes: []
---

# ADR-0010: Generated scale fixture and online merge

## Context and evidence

NFR-009 asks for a 1 GiB multi-instance fixture, peak RSS below 512 MiB,
analysis within five minutes on documented hardware, and a repeatable bench.
Committing a 1 GiB corpus is rejected (git size, secrets risk, no antifragile
value). Before this change, `Analyze` accumulated every pre-dedup event, then
`dedup` kept a full `occHints` copy of every occurrence — so a million repeats
of one signature retained a million message strings.

FR-010 still requires one inspectable timeline row per `(instance, signature)`.
Discarding unique rows to force AC2 would change the product, not just the
scanner.

## Decision

1. **Generate the fixture.** Profile `storm` writes four instances of
   Tomcat/HTTPD-like lines to `/tmp/logify-nfr009` (or `LOGIFY_NFR009_DIR`).
   Document the layout in `testdata/scale/README.md`. Gitignore any in-tree
   `testdata/scale/generated/` tree. Deterministic seed `9`.
2. **Merge online.** `parseFile` emits each completed record into an
   `eventMerger`. Peak working set is unique rows plus a bounded hint slice,
   not the raw line count.
3. **Bound correlation hints.** Extra `occHints` keep only later occurrences
   that still carry a labeled request/trace/correlation ID or a usable client
   IP, up to 64 extras per row. The lead row's own fields remain visible to
   FR-012. Repeats without correlation value store no extra hint.
4. **Measure in a child process.** `TestNFR009Scale1GiB` generates (or reuses)
   the 1 GiB tree, then re-execs so Linux `VmHWM` excludes generation. CI runs
   only `TestNFR009ScaleSmoke` and Linux `BenchmarkNFR009ScaleSmoke`.
5. **Reference hardware is a class.** Linux amd64, 2+ CPU, 4+ GiB RAM. A
   specific SKU stays [Q-001](../open-questions.md).
6. **Status follows measured ACs.** The 2026-09-12 probe-host run met AC2
   and AC3 on the storm fixture (13.6 MiB RSS, 59 s). Do not treat that as
   proof that unique-heavy access logs stay under 512 MiB.

## Alternatives considered

- Commit a 1 GiB fixture: rejected (NFR-009 implementation constraint).
- Spill unique events to disk / change the timeline model: a large design
  change, out of scope for this Should item.
- Keep all `occHints`: fails the 512 MiB budget on any high-repeat storm even
  when unique cardinality is small.
- Enforce AC2/AC3 as `t.Fatal` in CI: the full fixture is too large for the
  PR matrix; a miss would also hide honest Partial evidence.

## Consequences and risks

- Later labeled IDs beyond the 64th distinct extra hint on one signature are
  not visible to correlation. Small fixtures and the documented storm profile
  stay within the cap.
- Unique-heavy bundles can still exceed 512 MiB; that is an FR-010 cardinality
  limit, not a scanner leak.
- `go test -bench=.` without `LOGIFY_NFR009_FULL=1` runs only the 8 MiB smoke
  bench; the 1 GiB `BenchmarkNFR009Scale1GiB` skips.

## Verification

- `TestNFR009ScaleSmoke`, `TestNFR009OccHintsBoundedWithoutCorrelationValue`,
  `TestNFR009OccHintsCapAndKeepLaterRequestIDs`,
  `TestCorrelateUsesCollapsedDuplicateIdentifiers`
- `BenchmarkNFR009ScaleSmoke` (CI Linux)
- `make bench-nfr009` / `LOGIFY_NFR009_FULL=1` (manual)
- [RES-20260912-nfr009-scale-bench](../research/RES-20260912-nfr009-scale-bench.md)
