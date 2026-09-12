---
id: WI-20260912-keep-learning-and-apply
type: work-item
status: active
owner: cursor-agent
created: 2026-09-12T14:18:00Z
updated: 2026-09-12T14:18:00Z
lease_expires: 2026-09-12T18:30:00Z
scope:
  - docs/principles.md
  - docs/framework-adoption.md
  - AGENTS.md
  - README.md
  - docs/requirements/README.md
  - docs/collaboration/work-items/WI-20260912-keep-learning-and-apply.md
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

- `git diff --check`
- `go test ./...` if harness/docs probes require it

## Activity log

- `2026-09-12T14:18:00Z` — cursor-agent — work item created; no other active
  item owns `docs/principles.md` or `AGENTS.md`.

## Handoff or completion

In progress. Do not merge from this item.
