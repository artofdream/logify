---
id: ADR-0010-cli-compatibility
type: decision
status: accepted
owner: cursor-agent
created: 2026-09-12
updated: 2026-09-12
requirements: [NFR-016, NFR-011]
supersedes: []
---

# ADR-0010: Backward-compatible CLI evolution

## Context and evidence

NFR-016 requires existing flags and defaults to remain supported throughout a
major version, or a documented deprecation warning to precede removal.

On `main` before this change the public flags were `-output` (default
`logify-report.html`), `-from`, `-to`, `-redact`, `-redact-file`, and
`-h` / `-help`. There was no version surface. `.github/workflows/release.yml`
builds on `v*` tags but does not inject a version into the binary; that wiring
is deferred to the first tagged release (`v0.1.0`).

Unreleased local builds need a sensible default version string so `-version`
works before ldflags exist.

## Decision

1. Logify versions the operator CLI with [SemVer](https://semver.org/). A
   **major** increment is required to remove, rename, change the default of, or
   reject a previously accepted value of a documented public flag. Additive
   flags and new accepted values are **minor**. Bug fixes that restore
   documented behavior are **patch**.
2. The public CLI surface is the flags documented in README and `-h`:
   `-output`, `-from`, `-to`, `-redact`, `-redact-file`, `-h` / `-help`, and
   `-version` / `-V`. Their documented defaults (including redaction off and
   empty time bounds) stay supported for the current major version.
3. Before a breaking removal or rename, the CLI must print a deprecation
   warning on stderr when the old form is used, and README must name the
   successor, for at least one released minor version in the same major.
   Removal happens only in the next major.
4. `cmd/logify` keeps `var version = "dev"` for unreleased builds. Release
   builds may overwrite it with `-ldflags "-X main.version=vX.Y.Z"` (not wired
   in this change). `-version` and `-V` print `logify <version>` on stdout and
   exit 0 without requiring a directory.
5. While the product is still `0.y.z`, major is 0. Documented flags are still
   a compatibility contract: do not silently break them in 0.x. A break after
   the deprecation window ships as `1.0.0` (or a later major).

## Alternatives considered

- Treat 0.x as unconstrained. Rejected: NFR-016's acceptance criterion is
  major-version stability or an explicit warning, not a silent 0.x exception.
- Wire release ldflags in the same change. Rejected: the first tag is `v0.1.0`
  and is a separate release task; the variable is ldflags-ready now.
- Use Cobra or another CLI kit for aliases and deprecation helpers. Rejected:
  NFR-001 is stdlib-only; `flag` already supports aliases via `BoolVar`.

## Consequences and risks

Operators can keep existing scripts across minor/patch releases. Authors must
not change `-output`'s default or time-bound grammar without a major bump.
Scripts can call `-version` before a directory exists. Unreleased binaries
report `dev` until a release build injects a tag.

## Verification

`TestNFR016StableFlagsExist` (flag registration, `-output` default, help exit
0) and `TestNFR016VersionFlag` (`-version` / `-V`). README documents the
policy and version flags.
