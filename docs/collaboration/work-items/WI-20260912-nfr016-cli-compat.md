---
id: WI-20260912-nfr016-cli-compat
type: work-item
status: active
owner: cursor-agent
created: 2026-09-12T10:27:11Z
updated: 2026-09-12T10:27:11Z
lease_expires: 2026-09-12T16:00:00Z
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

- Wiring `.github/workflows/release.yml` `-ldflags -X` (deferred to v0.1.0)
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

`AGENTS.md` commands plus `./logify.exe -version` and `-h`.

## Activity log

- `2026-09-12T10:27:11Z` — cursor-agent — claimed after `git fetch origin main`
  and work-item/dispatch scan.

## Handoff or completion
