---
id: WI-20260912-nfr016-cli-compat
type: work-item
status: review
owner: cursor-agent
created: 2026-09-12T10:27:11Z
updated: 2026-09-12T10:50:00Z
lease_expires: null
scope:
  - cmd/logify/main.go
  - cmd/logify/main_test.go
  - README.md
  - docs/requirements/non-functional-requirements.md
  - docs/framework-adoption.md
  - docs/knowledge/decisions/ADR-0010-cli-compatibility.md
  - docs/knowledge/architecture.md
  - docs/knowledge/glossary.md
  - .github/workflows/release.yml
  - docs/collaboration/work-items/WI-20260912-nfr016-cli-compat.md
  - docs/collaboration/handoffs/HO-20260912-cursor-agent-nfr016.md
requirements: [NFR-016]
depends_on: []
supersedes: []
---

# NFR-016 backward-compatible CLI evolution

## Goal

Record a SemVer CLI compatibility policy, expose `-version` / `-V` from an
ldflags-friendly `version` var, and add a flag-stability probe so NFR-016 can
be marked Implemented.

## Non-goals

- Wiring `.github/workflows/release.yml` `-ldflags -X` (transferred to
  `WI-20260912-release-ldflags` for the v0.1.0 cut)
- Removing or renaming existing flags
- Analyzer or report package edits

## Acceptance criteria

1. ADR + README state major-version SemVer for breaking flag/default changes
   and a documented deprecation warning before removal.
2. CLI prints an ldflags-friendly `version` (default `dev`) via `-version` / `-V`.
3. A test asserts `-from`, `-to`, `-output`, `-redact`, `-redact-file` still
   exist and help exits 0.
4. NFR-016 is Implemented only after those three pieces exist.

## Planned files and ownership

cursor-agent owns the paths in frontmatter. No overlap with `active` work
items or dispatches. Review-status items are inspection-only.

## Evidence and assumptions

Starting from `origin/main` `4a6d7ae`. Release workflow already notes the
missing `var version` and defers tag injection.

## Validation

Local `AGENTS.md` commands passed. See
`docs/collaboration/handoffs/HO-20260912-cursor-agent-nfr016.md`.
PR https://github.com/artofdream/logify/pull/17 (do not merge from this item).

## Activity log

- `2026-09-12T10:27:11Z` — cursor-agent — claimed after `git fetch origin main`
  and work-item/dispatch scan.
- `2026-09-12T10:30:00Z` — cursor-agent — policy, `-version`, and probes
  landed; local validation passed; lease released for review.
- `2026-09-12T10:50:00Z` — cursor-agent — deferred Release ldflags wiring
  transferred to `WI-20260912-release-ldflags`. This item stays review;
  lease remains released.

## Handoff or completion

`docs/collaboration/handoffs/HO-20260912-cursor-agent-nfr016.md`
