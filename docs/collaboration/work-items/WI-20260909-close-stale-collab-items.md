---
id: WI-20260909-close-stale-collab-items
type: work-item
status: review
owner: cursor-agent
created: 2026-09-09T20:24:00Z
updated: 2026-09-09T20:26:00Z
lease_expires: 2026-09-10T08:24:00Z
scope:
  - docs/collaboration/work-items/WI-20260909-close-stale-collab-items.md
  - docs/collaboration/work-items/WI-20260904-report-crash-quick-wins.md
  - docs/collaboration/work-items/WI-20260904-must-issue-ui.md
  - docs/collaboration/work-items/WI-20260909-fr021-follow-up-details.md
  - docs/collaboration/dispatches/DSP-20260904-MUST-ISSUE-UI.md
  - docs/collaboration/dispatches/DSP-20260909-FR021.md
  - docs/collaboration/dispatches/DSP-20260904-FR017.md
  - docs/collaboration/dispatches/DSP-20260904-FR018-FR019.md
  - docs/collaboration/dispatches/DSP-20260904-FR020-FR021.md
  - docs/collaboration/dispatches/DSP-20260904-FR022.md
  - docs/collaboration/handoffs/HO-20260909-1949-cursor-agent-fr021.md
requirements: []
depends_on: []
supersedes: []
---

# Close stale collaboration items after merged PRs

## Goal

Mark collaboration work items and dispatches `done` when their PRs are already
on `main`. Do not change product behavior.

## Non-goals

FR-024 implementation, releases/tags, `.gitignore` hygiene, NFR debt, merging
any PR. Do not edit `DSP-20260904-FR023-FR024.md` (owned by draft PR #5).

## Acceptance criteria

- WI-20260904-report-crash-quick-wins and WI-20260904-must-issue-ui are `done`
  with merge evidence (PR number, date, squash SHA).
- Any other review/active WI whose PR is merged on `origin/main` is also `done`
  (FR-021 / PR #4).
- Related Must-UI dispatches no longer claim unfinished review work.
- FR-024 remains open (draft PR #5); this change does not mark it done.

## Planned files and ownership

cursor-agent owns only the coordination paths in frontmatter.

## Evidence and assumptions

Merge evidence observed `2026-09-09T20:23:12Z` against `origin/main` and the
GitHub pull-request API (`merged: true`):

| PR | Squash on `main` | `merged_at` (UTC) |
| --- | --- | --- |
| [#2](https://github.com/artofdream/logify/pull/2) | `de887178fd0c800cdb5ecda11ad43b8d9acd44ab` | 2026-09-04T20:27:11Z |
| [#3](https://github.com/artofdream/logify/pull/3) | `64bbc1c7c42763ee824dd962542ee0835faa2d51` | 2026-09-04T20:28:16Z |
| [#4](https://github.com/artofdream/logify/pull/4) | `01ae8bf57cc57a02ed610acbeb229feda4d44519` | 2026-09-09T19:52:40Z |

[#5](https://github.com/artofdream/logify/pull/5) was open and draft at that
probe (`draft: true`, `merged: false`). This close-out verifies merge state
only; it does not re-run the original Go/Node/browser probes recorded on those
items.

## Validation

- GitHub `pull_request_read` / `list_pull_requests`: #2 #3 #4 `merged: true`;
  #5 open draft.
- `git fetch origin main`; `origin/main` tip `01ae8bf` is the PR #4 squash.
- `git diff --check` after the docs edit.
- PR: https://github.com/artofdream/logify/pull/6 (draft vs `main`)

## Activity log

- `2026-09-09T20:24:00Z` — cursor-agent — claimed coordination-doc paths after
  fetching `origin/main` (`01ae8bf`). No overlapping active WI on these files.
  Draft PR #5 owns FR-024 paths including `DSP-20260904-FR023-FR024.md`; left
  those files untouched.
- `2026-09-09T20:25:00Z` — cursor-agent — stale merged items marked done with
  PR/SHA/date evidence. Docs PR opening for review.
- `2026-09-09T20:26:00Z` — cursor-agent — draft PR #6 opened vs `main`. Do not
  merge.

## Handoff or completion

In review. Merge evidence recorded; product probes were not re-run. Do not
merge from this task.
