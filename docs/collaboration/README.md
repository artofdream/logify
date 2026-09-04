---
id: collaboration-protocol
type: protocol
status: active
owner: human
updated: 2026-09-04
tags: [collaboration, multi-agent, second-brain]
---

# Multi-agent collaboration protocol

This protocol coordinates Codex, Claude, Cursor Agent, Antigravity, Grok, humans,
and any other autonomous tool that may share the same checkout, working
sequentially or in parallel. It is document based, local first, Git reviewable,
and compatible with Obsidian. Files are the coordination API.

It adapts two ideas from the [Adaptive Experience Architecture](https://architecture.artof.link/):
the loop **Interpret → Act → Verify → Remember**, and the rule that status words
are claims requiring a probe. For Logify:

1. **Interpret:** ground work in principles, requirements, durable knowledge, and
   current work-item ownership.
2. **Act:** make a bounded change inside declared path scope.
3. **Verify:** inspect artifacts and run acceptance probes; do not trust status
   labels or agent summaries alone.
4. **Remember:** promote durable learning into requirements, ADRs, fixtures,
   tests, or knowledge notes.

## Canonical reading order

1. `docs/principles.md`
2. `docs/requirements/README.md` and affected FR/NFRs
3. `docs/knowledge/index.md`
4. Relevant ADRs and open questions
5. Active/queued cloud dispatches and work items with overlapping scope
6. Relevant code, tests, and recent handoffs

Requirements define intended behavior. Code and tests show implemented reality.
ADRs preserve rationale. Work items coordinate temporary ownership. Handoffs and
session notes are evidence, not higher-authority truth.

Cloud assignments are published under `docs/collaboration/dispatches/` on
`main`. They reserve requirement scope across machines when local worktrees and
chat context are unavailable. Fetch before planning, and do not duplicate an
`active`, `queued`, or `review` dispatch without a documented takeover.

## Before editing: claim bounded work

1. Search active ownership:

   ```text
   rg -n "^status: (active|review)$|^scope:" docs/collaboration/work-items
   rg -n "^status: (active|queued|review)$|^requirements:" docs/collaboration/dispatches
   ```

2. Inspect any potentially overlapping work item and the actual working-tree
   status.
3. Create one `work-items/WI-YYYYMMDD-short-slug.md` from the template. Declare
   exact owned paths, affected requirements, owner (a stable per-agent
   identifier, e.g. `codex`, `claude`, `cursor-agent`, `antigravity`,
   `grok-bot`, or a human's name), and an ISO-8601 UTC lease expiry.
4. Use a short advisory lease and renew it by updating `updated` and
   `lease_expires`. A lease communicates intent; it is not a filesystem lock.
5. If scope overlaps, perform read/review work only, select disjoint paths, or
   record an explicit transfer in both work items. Never use last-writer-wins.

An expired lease is not permission to overwrite. Inspect Git and the working tree,
then document takeover or request a transfer.

This protocol does not assume exactly two participants. Any number of agents may
hold non-overlapping leases at the same time; the coordination cost is in
declaring and checking scope, not in the count of agents involved.

## During work

- One agent owns each file or narrowly defined area at a time.
- Update the work item when status, scope, assumptions, or blockers change.
- Use lifecycle: `proposed → active → blocked → review → done|abandoned`.
- Record consequential choices as ADRs under `docs/knowledge/decisions/`.
- Preserve another tool's changes even when they are uncommitted.
- If collision is detected, stop overlapping edits and preserve both versions.

## Handoff and completion

Create a handoff file for a change of owner or asynchronous review. It must name
the work item, exact files/diff or commit, requirement IDs, observed commands and
results, assumptions, failures, and next action. The receiver verifies artifacts.

Before setting a work item to `done`:

- acceptance criteria have observed evidence;
- final validation was run by the coordinator/integrator;
- durable learning was promoted out of transient notes;
- unresolved issues are linked, not hidden;
- the lease is cleared.

## Conflict protocol

1. All agents with overlapping paths stop editing those paths.
2. Preserve every diff or artifact in conflict; do not discard any silently.
3. Compare requirements, ADRs, tests, and direct evidence.
4. A coordinator records the resolution and rationale, or creates an open question.
5. Reassign path ownership explicitly before work resumes.

Consensus is not proof. Conflicts resolved without evidence remain `unknown`.
