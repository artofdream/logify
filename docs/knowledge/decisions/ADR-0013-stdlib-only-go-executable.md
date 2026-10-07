---
id: ADR-0013-stdlib-only-go-executable
type: decision
status: accepted
owner: grok-bot
created: 2026-10-07
updated: 2026-10-07
requirements: [NFR-001, NFR-002]
supersedes: []
---

# ADR-0013: Dependency-free, standard-library-only Go executable

**Status:** Accepted (retroactive, recorded 2026-10-07)
**Introduced in:** `cbc0c24` feat: establish log analysis foundation
**Related requirements:** NFR-001, NFR-002

This is a **Documented** record reconstructed after the fact from repository
history. It records a decision that was already in force; it does not change
behavior. Where no written rationale exists, the gap is marked `OPEN` rather
than filled in.

## Context and evidence

- `cbc0c24` (the first commit on `main`, 2026-09-04) added `go.mod` with only
  `module github.com/artofdream/logify` and `go 1.22`. It has no `require`
  block and the repository has no `go.sum`. `git log -- go.mod` shows no later
  change, so this still holds at `076e2ba` (v0.1.1).
- The same commit added NFR-001 ("Native standalone executable"). Its
  acceptance criteria are: builds with Go 1.22+, no required runtime service or
  database, and no third-party Go module at the v1 baseline.
- The same commit's `AGENTS.md` working rule says: "Preserve the
  single-native-executable design. Prefer the Go standard library; justify any
  new dependency before adding it." README and AGENTS.md both describe Logify
  as "dependency-free".
- `cbc0c24` is a bootstrap commit
  (`WI-20260903-initialize-github-main`). The decision was made before git
  history exists, so no earlier commit or discussion is available.
- Later ADRs apply the constraint when rejecting alternatives: ADR-0005
  (third-party secret libraries), ADR-0009 (third-party archive library),
  ADR-0010 (Cobra or another CLI kit), and ADR-0011 (axe-core / npm).

## Decision

1. The shipped `logify` executable uses only the Go standard library. `go.mod`
   has no `require` entries.
2. A new third-party Go module needs an explicit justification. The
   repository's mechanism is an ADR plus an NFR-001 amendment, because AC3
   forbids such modules at the v1 baseline.
3. The constraint covers Go modules linked into the product. It does not cover
   build and CI tooling. The workflows use third-party GitHub Actions
   (for example `softprops/action-gh-release`), and some tests run Node
   scripts, without adding a Go module.

## Alternatives considered

- **Use third-party modules where convenient** (CLI framework, archive or
  secret-scanning libraries, an accessibility engine). This was rejected case by
  case in ADR-0005, ADR-0009, ADR-0010, and ADR-0011, each time by citing this
  constraint.
- **Why the stdlib-only constraint was chosen in the first place:** `OPEN`.
  The repository records the constraint (NFR-001, AGENTS.md) but not the
  original reasoning, such as supply-chain surface, offline builds, or audit
  cost. NFR-001 has no Rationale field. These motives are plausible but
  unverified, so this record does not claim them.

## Consequences and risks

- Builds need only a Go 1.22+ toolchain. There is no module download and no
  dependency CVE surface for the product binary.
- Features that a library would make easy (archives, secret detection,
  WCAG checks) are either built narrowly on the stdlib or left out of scope.
  ADR-0009 and ADR-0011 accept that trade-off.
- `CGO_ENABLED=0` cross-compilation in the release workflow (ADR-0016) depends
  on the absence of cgo dependencies.
- Risk: the boundary between "product module" and "tooling" is not written
  down anywhere except this record. Whether tooling (Actions, Node probes)
  should be held to a similar bar is `OPEN`.

## Verification

- `go.mod` at `076e2ba` has no `require` directive, and there is no `go.sum`.
- The CI `test` job builds with `go build ./cmd/logify` on ubuntu, windows,
  and macos (NFR-001 / NFR-002 comments in `.github/workflows/ci.yml`).
- No automated probe currently fails if a `require` line is added. `OPEN`:
  whether to add one, such as a harness check on `go.mod`.
