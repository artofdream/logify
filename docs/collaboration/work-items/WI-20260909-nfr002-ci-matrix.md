---
id: WI-20260909-nfr002-ci-matrix
type: work-item
status: active
owner: cursor-agent
created: 2026-09-09T20:40:00Z
updated: 2026-09-09T20:40:00Z
lease_expires: 2026-09-10T08:40:00Z
scope:
  - .github/workflows/ci.yml
  - docs/requirements/non-functional-requirements.md
  - docs/framework-adoption.md
  - AGENTS.md
  - docs/collaboration/work-items/WI-20260909-nfr002-ci-matrix.md
requirements: [NFR-002]
depends_on: []
supersedes: []
---

# Close NFR-002 CI OS matrix gap

## Goal

Make CI run tests and a native build on Ubuntu, Windows, and macOS so NFR-002
AC3 is met. Keep draft auto-merge skip. Document Linux-only gofmt/diff-check.

## Non-goals

NFR-006 redaction, NFR-021 10k probe, FR-012/016, release/tag changes, PRs
#5–#8, stdlib product rewrites unless a real path bug appears.

## Acceptance criteria

1. Path handling already uses `path/filepath` — verify, do not rewrite.
2. Release workflow already emits Windows `.exe` — verify, do not rewrite.
3. CI `test` matrix runs `go test`, `go build`, and fixture smoke on
   `ubuntu-latest`, `windows-latest`, and `macos-latest`.
4. `enable-auto-merge` still `needs: validate` and skips drafts.
5. NFR-002 status/gap text is honest after the matrix is in place.

## Planned files and ownership

Owned by `cursor-agent`. Disjoint from open PRs #5–#8. WI-20260904 listed
`ci.yml` / NFR-002 but marked OS CI matrix as a non-goal; lease expired.

## Evidence and assumptions

- `internal/analyzer/analyzer.go` uses `filepath.Abs`, `WalkDir`, `Rel`,
  `Base`, `ToSlash`.
- `.github/workflows/release.yml` sets `ext=".exe"` for `windows`.
- Assumption: existing `validate` required-check name should stay so
  branch protection and auto-merge keep working; matrix lives in `test`.

## Validation

Local: `gofmt`, `go test ./...`, `go build`, `go vet`, `git diff --check`,
fixture smoke. Remote: PR CI matrix legs.

## Activity log

- `2026-09-09T20:40:00Z` — cursor-agent — claimed NFR-002 CI matrix scope.

## Handoff or completion

Pending.
