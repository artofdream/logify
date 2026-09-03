---
id: WI-20260904-cloud-dispatch-registry
type: work-item
status: done
owner: codex
created: 2026-09-03T22:13:16Z
updated: 2026-09-03T22:14:55Z
lease_expires: null
scope:
  - docs/collaboration/README.md
  - docs/collaboration/dispatches/**
  - docs/collaboration/work-items/WI-20260904-cloud-dispatch-registry.md
  - docs/requirements/non-functional-requirements.md
requirements: [FR-017, FR-018, FR-019, FR-020, FR-021, FR-022, FR-023, FR-024, NFR-025, NFR-027]
depends_on: []
supersedes: []
---

# Publish cloud-task dispatch registry

## Goal

Make cloud ownership and queued follow-up work visible to Claude, Codex, humans,
and offline agents through committed Git state.

## Non-goals

- Claim queued tasks are running.
- Start dependent implementation before prerequisite PRs merge.
- Modify the unexplained local `AGENTS-1.md`.

## Acceptance criteria

- The protocol defines dispatch records and their status semantics.
- FR-017 is recorded active with its cloud setup identifier and branch.
- Later batches are recorded queued with dependencies and no fabricated task ID.
- Other agents can discover reservations with one documented command.
- Coordination state is pushed to `origin/main`.

## Planned files and ownership

Codex owns only the declared coordination, dispatch, and NFR paths.

## Evidence and assumptions

The FR-017 cloud task returned client setup identifier
`local-chatgpt:1f4d602a-095b-4476-88b1-7842d2416b21`. No final server task ID or
PR exists yet.

## Validation

- Five dispatch records exist: one observed active setup and four queued batches.
- Each record declares requirements, owner, state, dependency, intended branch,
  task reference, and PR reference.
- `git diff --check` and staged diff checks passed.
- Registry commit `1f0bf7d87cd755b611acdffb8565512d16eaea4c` was pushed to
  `origin/main`.
- `AGENTS-1.md` remained untracked and untouched.

## Activity log

- `2026-09-03T22:13:16Z` — codex — confirmed clean tracked state, no active work
  item, and only the untouched untracked `AGENTS-1.md`.
- `2026-09-03T22:14:55Z` — codex — registry validated and published; lease
  released.

## Handoff or completion

Complete. Dependent dispatches remain queued and must not be described as running
until a task identifier is observed after their prerequisite merges.
