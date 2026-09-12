---
id: WI-20260912-keep-learning-and-apply
type: work-item
status: review
owner: cursor-agent
created: 2026-09-12T14:18:00Z
updated: 2026-09-12T14:28:00Z
lease_expires: 2026-09-12T18:30:00Z
scope:
  - docs/principles.md
  - docs/framework-adoption.md
  - AGENTS.md
  - README.md
  - docs/requirements/README.md
  - docs/collaboration/work-items/WI-20260912-keep-learning-and-apply.md
  - docs/collaboration/handoffs/HO-20260912-cursor-agent-keep-learning-and-apply.md
requirements: [NFR-023, NFR-026, NFR-028]
depends_on: []
supersedes: []
---

# Adopt keep learning and apply in Logify docs

## Goal

Close [#24](https://github.com/artofdream/logify/issues/24): add **keep
learning and apply** as a peer principle, cite
[AEA #434](https://gitlab.com/artof-group/adaptive-experience-architecture/-/work_items/434),
and keep Logify [#22](https://github.com/artofdream/logify/issues/22) /
[#23](https://github.com/artofdream/logify/issues/23) as Documented/Simulated
evidence. Docs adoption is Documented once merged. Do not mark skills Live.

## Non-goals

- Implementing or live-probing the #22 / #23 skills
- Changing Go code or NFR statuses
- Claiming the principle is Live on architecture.artof.link

## Acceptance criteria

- `docs/principles.md` has four principles; decision gate names all four
- `AGENTS.md` and `README.md` name keep learning and apply
- `docs/framework-adoption.md` maps AEA #434 honestly (sponsor operating
  principle; not a public schema ID)
- PR open against `main`; #24 commented with the PR URL

## Planned files and ownership

`cursor-agent` owns only the paths in frontmatter. Leftover active items
`WI-20260912-nfr013-nfr021-a11y` and `WI-20260912-nfr009-rebase-pr18` still
list `README.md` / `docs/framework-adoption.md` from merged #19 / #18; this
item takes the principles-docs slice only.

## Evidence and assumptions

Public AEA schema still lists three principles. #434 is the AEA work item
(private GitLab; URL from sponsor / zorg-dungeon). AEA status for the
principle is Documented/Planned until a Pages probe.

## Validation

- `git diff --check` — pass
- `go test ./internal/harness ./...` — pass (NFR-028 docs/link + open-WI
  probes included)

## Activity log

- `2026-09-12T14:18:00Z` — cursor-agent — work item created; no other active
  item owns `docs/principles.md` or `AGENTS.md`.
- `2026-09-12T14:22:00Z` — cursor-agent — draft PR #25 vs `main`.
- `2026-09-12T14:25:00Z` — cursor-agent — validation passed. GitHub issue
  comment on #24 returned 403 from the token; PR body already refs #24.
- `2026-09-12T14:28:00Z` — cursor-agent — added related AEA skill-matrix
  [#433](https://gitlab.com/artof-group/adaptive-experience-architecture/-/work_items/433)
  and named glossary/comparison as Documented/Planned (not Live).

## Handoff or completion

In review on draft PR #25. Do not merge from this item. Docs adoption is
Documented once merged. Skills #22/#23 remain Documented/Simulated.
