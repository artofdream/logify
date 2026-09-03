---
id: WI-20260903-initialize-github-main
type: work-item
status: active
owner: codex
created: 2026-09-03T22:06:45Z
updated: 2026-09-03T22:06:45Z
lease_expires: 2026-09-03T23:06:45Z
scope:
  - repository Git index and initial main history
  - all canonical untracked files except AGENTS-1.md and ignored build outputs
requirements: [NFR-001, NFR-014, NFR-015, NFR-022, NFR-025, NFR-027]
depends_on: [WI-20260903-claude-parser-findings, WI-20260903-correct-lease-trail]
supersedes: []
---

# Initialize the empty GitHub repository

## Goal

Create and push the first validated `main` commit so subsequent changes can use
ordinary feature branches and pull requests.

## Non-goals

- Commit the unexplained `AGENTS-1.md` conflict artifact.
- Commit ignored executables or generated reports.
- Claim that a pull request exists when the remote has no base branch.

## Acceptance criteria

- The remote is confirmed empty before initialization.
- Canonical files are staged explicitly and reviewed.
- Tests, vet, build, fixture execution, and diff checks pass.
- Initial commit is pushed to `origin/main`.
- Local `AGENTS-1.md` remains untracked and untouched for Claude/human review.

## Planned files and ownership

This task owns Git staging/history only. It does not claim authorship of every
file, and preserves Claude's session note attribution.

## Evidence and assumptions

`git ls-remote --heads origin` returned no refs. A PR cannot be opened without a
base branch, so initializing `main` is the required bootstrap operation.

## Validation

Application validation passed immediately before this work item. Staged-tree and
remote verification remain pending.

## Activity log

- `2026-09-03T22:06:45Z` — codex — opened bootstrap work item after resolving and
  validating Claude's parser findings.

## Handoff or completion

Pending initial commit and push.
