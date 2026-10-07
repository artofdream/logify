---
id: ADR-0016-tag-triggered-release-workflow
type: decision
status: accepted
owner: grok-bot
created: 2026-10-07
updated: 2026-10-07
requirements: [NFR-001, NFR-002, NFR-016]
supersedes: []
---

# ADR-0016: Tag-triggered cross-platform release workflow

**Status:** Accepted (retroactive, recorded 2026-10-07)
**Introduced in:** `5a97f17` ci: add validation + release workflows
**Refined in:** `319145c` ci: add arm64 targetes; `cef5e20` ci: inject -version from release tag for v0.1.0 (#21)
**Related requirements:** NFR-001, NFR-002, NFR-016

This is a **Documented** record reconstructed after the fact from repository
history. It records a decision that was already in force; it does not change
behavior. Where no written rationale exists, the gap is marked `OPEN` rather
than filled in.

Scope note: the **version string policy** (SemVer for the CLI,
`var version = "dev"`, and `-ldflags "-X main.version=${GITHUB_REF_NAME}"`
on tagged builds) is already recorded in ADR-0010, decision 4. This record
covers the release *mechanism* only and does not restate that policy.

## Context and evidence

- `5a97f17` added `.github/workflows/release.yml`. Its header comment gives
  the only written rationale: "Per AGENTS.md's 'Working rules', generated
  executables are never committed to the repo — this workflow is the
  sanctioned way to distribute them." That AGENTS.md rule ("Do not commit
  generated executables or HTML reports") dates from `cbc0c24`.
- As introduced, the workflow ran on `push` of tags matching `v*` with
  `permissions: contents: write`. A `validate` job ran gofmt, `go vet`, and
  `go test` on ubuntu. A `build` matrix cross-compiled on ubuntu with
  `CGO_ENABLED=0` and `go build -trimpath` for linux/amd64, darwin/amd64,
  darwin/arm64, and windows/amd64, producing
  `logify-<goos>-<goarch>[.exe]`. A `release` job wrote `checksums.txt`
  (`sha256sum`) and published through `softprops/action-gh-release@v2` with
  `generate_release_notes: true`.
- `319145c` added linux/arm64 and windows/arm64, for six targets total.
- `cef5e20` (#21) added `-ldflags "-X main.version=${GITHUB_REF_NAME}"` to the
  build step (policy in ADR-0010).
- NFR-002 evidence cites this workflow for Windows `.exe` cross-compilation.
- Observed outcome: GitHub Releases `v0.1.0` (published 2026-09-12 12:52 CEST)
  and `v0.1.1` (published 2026-09-13 11:09 CEST) were both created by
  `github-actions[bot]`. Tag `v0.1.0` points to `cef5e20` and `v0.1.1` to
  `076e2ba`. This record did not check the release asset lists.
- ADR-0010 (alternatives) says the coordinator creates the tag after the
  relevant PR merges. A "tagged release cut" skill is proposed in issue #23
  and is labeled Documented/Simulated, not Live, in `docs/principles.md`.

## Decision

1. Releases are cut by **pushing a git tag** matching `v*`. Nothing on
   `main` or on a PR publishes binaries. A human or coordinator creates the
   tag; the workflow does not create tags.
2. On a tag push, the Release workflow re-runs gofmt, `go vet`, and `go test`,
   then cross-compiles static (`CGO_ENABLED=0`, `-trimpath`) binaries from
   ubuntu for linux, darwin, and windows on amd64 and arm64.
3. The workflow attaches the binaries plus `checksums.txt` (SHA-256) to a
   GitHub Release with auto-generated notes. Binaries are never committed to
   the repository.
4. Tagged binaries embed the tag as their version (ADR-0010).

## Alternatives considered

- **Commit built executables to the repository:** rejected by the AGENTS.md
  working rule (evidence above).
- **Native per-OS release runners** instead of ubuntu cross-compilation: no
  record of evaluation found (`OPEN`). Pure-Go stdlib code (ADR-0013) makes
  cross-compilation possible. The CI `test` matrix, not the Release
  workflow, runs native builds and tests on windows and macos.
- **Release automation tools** (GoReleaser, release-please) or signed
  artifacts and provenance: no record of evaluation found (`OPEN`).
- **Why this exact target list and these flags** (`-trimpath`,
  `CGO_ENABLED=0`, checksums): `OPEN`. The commits do not explain them.

## Consequences and risks

- Anyone with tag-push rights can publish a release. The repository has no
  record of tag protection (`OPEN`).
- The Release `validate` job is lighter than CI: ubuntu only, with no fixture
  smoke, no multi-OS matrix, and no NFR-009 bench. It relies on the tagged
  commit having passed CI `validate` on `main`. Whether that gap is
  intentional is `OPEN`.
- The `v*` pattern accepts tags that are not SemVer (for example `vfoo`). The
  workflow does not enforce ADR-0010's SemVer policy.
- Third-party Actions are pinned by major tag (`@v4`, `@v5`, `@v2`), not by
  commit SHA. That is a supply-chain risk, and nothing on record says it was
  accepted (`OPEN`).
- `checksums.txt` gives integrity against corruption, not authenticity. There
  are no signatures or provenance attestations.
- The SSDD draft in `artofdream/dso` lists "release baseline" as an open
  item. Which tag is the sponsor-approved baseline is `OPEN` and outside this
  record.

## Verification

- `.github/workflows/release.yml` at `076e2ba` matches the decision above:
  `on.push.tags: ['v*']`, a six-entry matrix, `CGO_ENABLED: 0`,
  `-trimpath -ldflags "-X main.version=${GITHUB_REF_NAME}"`, `sha256sum`, and
  `softprops/action-gh-release@v2`.
- Releases `v0.1.0` and `v0.1.1` exist and were published by
  `github-actions[bot]`.
- `TestNFR016VersionFlag` covers the `-version` / `-V` surface that the
  ldflags injection targets.
