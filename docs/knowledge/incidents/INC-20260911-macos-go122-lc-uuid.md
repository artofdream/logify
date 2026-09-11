---
id: INC-20260911-macos-go122-lc-uuid
type: incident
status: mitigated
owner: cursor-agent
created: 2026-09-11T19:39:00Z
updated: 2026-09-11T20:20:00Z
requirements: [NFR-001, NFR-002]
work_item: WI-20260911-fr012-event-correlation
---

# macOS CI aborts Go 1.22 test binaries without LC_UUID

## Trigger and impact

PR #13 CI run
[34639793270](https://github.com/artofdream/logify/actions/runs/34639793270)
failed `test (macos-latest)` and the required `validate` aggregator. Ubuntu
and Windows legs passed. Every macOS package failed before tests ran:

```text
dyld[…]: missing LC_UUID load command in …/analyzer.test
signal: abort trap
```

`go vet` and `go build` on that job succeeded; `go test` executes the
resulting binaries and is what dyld rejected. Fixture smoke did not run.

## Evidence and diagnosis

- Job used `macos-26-arm64` (macOS 26.6.2) and `go1.22.12 darwin/arm64`.
- The same OS family and Go 1.22.12 succeeded on `main` the day before
  ([34502726419](https://github.com/artofdream/logify/actions/runs/34502726419)).
  The abort is a toolchain/runner constraint, not an FR-012 assertion failure.
- Go issue [golang/go#68678](https://github.com/golang/go/issues/68678): the
  default Go 1.22 Darwin linker omits `LC_UUID`. Go 1.24 emits it by default.
  Go 1.22/1.23 backports require `-ldflags=-B=gobuildid`.

## Recovery

macOS `go build` and `go test` in `.github/workflows/ci.yml` pass
`-ldflags=-B=gobuildid`. Product `go.mod` stays `go 1.22` (NFR-001).
Linux and Windows flags are unchanged.

## Durable safeguard

CI comment plus this incident. A later Go 1.24+ CI bump can drop the flag
once that toolchain is the documented baseline.

## Verification and remaining risk

PR #13 run
[34640118520](https://github.com/artofdream/logify/actions/runs/34640118520)
showed `test (macos-latest)`, `test (ubuntu-latest)`, `test (windows-latest)`,
and `validate` successful after `-ldflags=-B=gobuildid`. Cross-compiled
Darwin release artifacts from Ubuntu are not executed on macOS in this
workflow; they may still lack `LC_UUID` until a tag build is verified on a
Mac (open).
