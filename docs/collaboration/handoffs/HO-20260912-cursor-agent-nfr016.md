---
id: HO-20260912-cursor-agent-nfr016
type: handoff
status: proposed
owner: cursor-agent
created: 2026-09-12T10:30:00Z
work_item: WI-20260912-nfr016-cli-compat
requirements: [NFR-016]
next_owner: human
---

# Handoff: NFR-016 CLI compatibility and -version

## Scope and requirement IDs

NFR-016 marked Implemented: policy (ADR-0010 + README), `-version` / `-V`,
and flag-stability probes. PR https://github.com/artofdream/logify/pull/17
against `main`. Do not merge from this handoff.

## Evidence consulted

`docs/principles.md`, collaboration protocol, work-item/dispatch scan on
`origin/main` `4a6d7ae`, NFR-016 ACs, `cmd/logify/main.go`,
`.github/workflows/release.yml` deferred-ldflags note.

## Changes and artifacts

- `docs/knowledge/decisions/ADR-0010-cli-compatibility.md` — SemVer major for
  breaking flags/defaults; deprecation warning before removal
- README CLI compatibility section and `-version` / `-V`
- `cmd/logify`: `var version = "dev"`, `-version` / `-V`, `registerCLI`
- `TestNFR016StableFlagsExist`, `TestNFR016VersionFlag`
- NFR-016 Implemented; adoption status mirror no longer lists it as Proposed
- `release.yml` comment only (ldflags still deferred to v0.1.0)

No analyzer or report edits. Existing flags were not renamed or removed.

## Validation and observed results

Ran (Go 1.22.2 linux/amd64):

- `gofmt -w cmd internal` — clean (no post-commit diff)
- `go test ./...` — pass (all packages, including `cmd/logify` and `internal/harness`)
- `go build -o logify.exe ./cmd/logify` — pass
- `go vet ./...` — pass
- `git diff --check` — pass
- `./logify.exe -version` / `-V` — `logify dev`, exit 0
- `./logify.exe -h` — Usage + stable flags, exit 0
- `./logify.exe -output sample-report.html testdata/case` — 6 events, unparsed=1
- `go build -ldflags "-X main.version=v0.1.0"` — `logify v0.1.0`

CI on PR #17 was not observed in this handoff (local only at write time).

## Assumptions and confidence

High for policy text, version surface, and the flag probe. Release-tag
injection remains unwired by design.

## Failures or conflicting evidence

None locally.

## Uncommitted or concurrent changes

None after the validation commit.

## Open questions and next action

Human review of PR #17. Do not merge from this item. Wire
`-ldflags "-X main.version=${GITHUB_REF_NAME}"` when tagging `v0.1.0`.
