---
id: WI-20260903-adopt-aea-principles
type: work-item
status: done
owner: codex
created: 2026-09-03T21:50:31Z
updated: 2026-09-03T21:51:50Z
lease_expires: null
scope:
  - docs/framework-adoption.md
  - docs/principles.md
  - docs/knowledge/architecture.md
  - docs/requirements/non-functional-requirements.md
  - README.md
requirements: [NFR-022, NFR-023, NFR-024, NFR-026, NFR-028]
depends_on: []
supersedes: []
---

# Adopt Adaptive Experience Architecture principles

## Goal

Map the referenced framework's formula, core principles, loop, and six outer
harness layers into Logify with an honest adoption ledger.

## Non-goals

- Copy the source framework wholesale.
- Claim runtime or CI controls that Logify does not yet have.

## Acceptance criteria

- Canonical source pages are linked.
- Each adopted concept has a Logify translation and current evidence/status.
- Partial and unknown capabilities remain labeled.

## Planned files and ownership

Codex owns only the paths declared in frontmatter for this work item.

## Evidence and assumptions

The framework homepage, schema, comparison, and glossary were fetched directly
from `architecture.artof.link` on 2026-09-03.

## Validation

- 52 unique FR/NFR IDs validated.
- Required framework documents confirmed present.
- `git diff --check` passed.
- `go test ./...` passed.
- `go vet ./...` passed.

## Activity log

- `2026-09-03T21:50:31Z` — codex — work item created after confirming no active
  work-item overlap; repository has extensive pre-existing uncommitted work.
- `2026-09-03T21:51:50Z` — codex — adoption map integrated and validation probes
  passed; lease released.
- `2026-09-03T22:03:28Z` — retrospective correction — Claude correctly identified
  that earlier Codex bootstrap edits were outside this work item's declared scope.
  See `WI-20260903-correct-lease-trail` and
  `../../knowledge/incidents/INC-20260903-incomplete-lease-trail.md`. The original
  scope is preserved rather than expanded retroactively.

## Handoff or completion

Complete. No ownership transfer required. All repository changes remain
uncommitted, so adoption is local working-tree state rather than shared Git
history. This completion statement applies only to this work item's declared
scope; it did not cover the earlier collaboration-protocol bootstrap edits.
