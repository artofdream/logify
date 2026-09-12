# NFR-009 scale fixture

The documented support-bundle fixture is generated on demand. **Do not commit
the 1 GiB corpus.**

## What it represents

Profile `storm` (the NFR-009 fixture):

- At least 1 GiB of mixed Tomcat/Java and Apache HTTPD-like lines
- Four instances: `node-a`, `node-b`, `httpd-a`, `httpd-b`
- Repeating incident signatures (error storms, a small URL set, recycled
  request IDs and client addresses)

This matches a realistic support bundle more closely than an adversarial
unique-line access dump. FR-010 still keeps one timeline row per
`(instance, signature)`. A corpus of millions of distinct URLs would retain
millions of rows and is **not** this fixture.

## How to generate and measure

From the repository root:

```text
make bench-nfr009-smoke
make bench-nfr009
```

Equivalent commands:

```text
go test ./internal/analyzer -run TestNFR009ScaleSmoke -count=1
go test ./internal/analyzer -bench BenchmarkNFR009ScaleSmoke -benchtime=1x -count=1

LOGIFY_NFR009_FULL=1 LOGIFY_NFR009_DIR=/tmp/logify-nfr009 \
  go test ./internal/analyzer -run TestNFR009Scale1GiB -count=1 -timeout 15m
```

`TestNFR009Scale1GiB` writes the corpus under `$LOGIFY_NFR009_DIR` (default
`os.TempDir()/logify-nfr009`), then analyzes it in a **child process** so
Linux `VmHWM` does not include generation. It also writes
`os.TempDir()/logify-nfr009-report.html` (gitignored).

Reuse: a matching `.logify-nfr009-meta.json` skips regeneration.

## Reference hardware class

NFR-009 AC3 uses this **class**, not a specific SKU (see Q-001):

| Item | Requirement |
|---|---|
| OS | Linux amd64 |
| CPU | 2+ cores |
| RAM | 4+ GiB available to the process |
| Disk | Local disk with room for a 1 GiB generated tree plus the HTML report |
| Go | 1.22 or newer |

A measured probe-host run is recorded in
[`docs/knowledge/research/RES-20260912-nfr009-scale-bench.md`](../../docs/knowledge/research/RES-20260912-nfr009-scale-bench.md).
That host is evidence, not a published operator workstation.

## CI

`go test ./...` always runs `TestNFR009ScaleSmoke` (a few MiB). Linux CI also
runs `BenchmarkNFR009ScaleSmoke` once. The 1 GiB job is manual/nightly.
