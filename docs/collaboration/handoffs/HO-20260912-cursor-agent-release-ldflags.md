---
id: HO-20260912-cursor-agent-release-ldflags
type: handoff
status: proposed
owner: cursor-agent
created: 2026-09-12T10:50:00Z
work_item: WI-20260912-release-ldflags
requirements: [NFR-016]
next_owner: human
---

# Handoff: Release-tag version injection for v0.1.0

## Scope and requirement IDs

NFR-016 follow-up: wire `-ldflags "-X main.version=${GITHUB_REF_NAME}"` in
the Release `build` job and update docs that still said this was deferred.
Do not create tag `v0.1.0` or a GitHub Release from this item.

## Evidence consulted

`docs/principles.md`, collaboration protocol, work-item/dispatch scan on
`origin/main` `b4d6ec0`, NFR-016 notes, ADR-0010, README, `release.yml`,
`WI-20260912-nfr016-cli-compat` (review; PR #17 merged; ldflags was a
non-goal). Overlapping README/NFR leases on NFR-013 (#19 merged) and
NFR-009 rebase (#18 merged) left untouched except the version sentences.

## Changes and artifacts

- `.github/workflows/release.yml` — tagged builds inject `GITHUB_REF_NAME`
- README, NFR-016 notes, ADR-0010, architecture — injection is wired
- Work item + this handoff

No Go source changes. No tag, no GitHub Release.

## Validation and observed results

Observed on `536e2ea` (Go 1.22.2 linux/amd64):

- PyYAML `safe_load_all` of `.github/workflows/release.yml` — 1 doc; build
  step contains
  `go build -trimpath -ldflags "-X main.version=${GITHUB_REF_NAME}" -o "$out" ./cmd/logify`
- `git diff --check` — clean (no uncommitted whitespace issues)
- `go test ./internal/harness -count=1` — pass (new work-item ownership
  fields accepted)
- Local smoke (not a Release run):
  `go build -trimpath -ldflags "-X main.version=v0.1.0" -o /tmp/logify-ver ./cmd/logify`
  then `-version` / `-V` printed `logify v0.1.0` and exited 0

`go test ./cmd/logify/...` was not required (no Go source edits). The
Release workflow itself was not executed (no tag pushed).

## Assumptions and confidence

High for the workflow line matching the requested `go build` invocation.
Coordinator still owns creating `v0.1.0` after merge.

## Failures or conflicting evidence

None at write time.

## Uncommitted or concurrent changes

None intended after the validation commit.

## Open questions and next action

Human review of https://github.com/artofdream/logify/pull/21.
Coordinator tags `v0.1.0` after merge. Do not merge or tag from this item.
