---
id: WI-20260912-release-ldflags
type: work-item
status: review
owner: cursor-agent
created: 2026-09-12T10:50:00Z
updated: 2026-09-12T11:00:00Z
lease_expires: 2026-09-12T18:00:00Z
scope:
  - .github/workflows/release.yml
  - README.md
  - docs/requirements/non-functional-requirements.md
  - docs/knowledge/decisions/ADR-0010-cli-compatibility.md
  - docs/knowledge/architecture.md
  - docs/collaboration/work-items/WI-20260912-release-ldflags.md
  - docs/collaboration/work-items/WI-20260912-nfr016-cli-compat.md
  - docs/collaboration/handoffs/HO-20260912-cursor-agent-release-ldflags.md
requirements: [NFR-016]
depends_on: [WI-20260912-nfr016-cli-compat]
supersedes: []
---

# Wire Release-tag version injection for v0.1.0

## Goal

Inject `main.version` from `GITHUB_REF_NAME` in `.github/workflows/release.yml`
so tagged binaries report the tag via `-version` / `-V`. Update the README,
NFR-016 notes, and ADR-0010 so they no longer say this is deferred.

## Non-goals

- Creating, pushing, or cutting tag `v0.1.0`
- Opening a GitHub Release
- Merging this PR
- Changing `cmd/logify` Go sources (already ldflags-ready)

## Acceptance criteria

1. Release `build` job uses
   `go build -trimpath -ldflags "-X main.version=${GITHUB_REF_NAME}" -o "$out" ./cmd/logify`
   with `CGO_ENABLED=0` and the existing OS/arch matrix.
2. Docs that said injection is deferred until v0.1.0 now say Release tags
   inject `main.version` from `GITHUB_REF_NAME`.
3. No git tag or GitHub Release is created from this item.

## Planned files and ownership

cursor-agent owns the paths in frontmatter. README / NFR notes edits are
limited to the version-injection sentences (not NFR-009 scale or NFR-013
a11y content). `release.yml` transfer from
`WI-20260912-nfr016-cli-compat` (review; lease released; PR #17 merged;
deferred this wiring as a non-goal).

## Evidence and assumptions

- `origin/main` `b4d6ec0` after fetch.
- Active items that also list README / NFR docs: NFR-013 a11y (PR #19
  merged) and NFR-009 rebase (PR #18 merged). This item does not edit
  their remaining scope.
- Coordinator tags `v0.1.0` after this PR merges.

## Validation

Observed on `536e2ea`: YAML parse OK; harness work-item probe pass;
local ldflags smoke prints `logify v0.1.0`. See the handoff.

## Activity log

- `2026-09-12T10:50:00Z` — cursor-agent — claimed after `git fetch origin
  main` (`b4d6ec0`), work-item/dispatch scan, and principles/NFR-016/ADR-0010
  read. Transferred deferred ldflags wiring from
  `WI-20260912-nfr016-cli-compat`.
- `2026-09-12T11:00:00Z` — cursor-agent — PR
  https://github.com/artofdream/logify/pull/21 opened (ready for review).
  YAML parse, harness probe, and local ldflags smoke passed. Status
  moved to review; do not tag from this item.

## Handoff or completion

In review on https://github.com/artofdream/logify/pull/21. Do not merge
or tag from this item.

[HO-20260912-cursor-agent-release-ldflags](../handoffs/HO-20260912-cursor-agent-release-ldflags.md)
