---
id: INC-20260903-incomplete-lease-trail
type: incident
status: resolved
owner: codex
created: 2026-09-03T22:03:28Z
updated: 2026-09-03T22:03:28Z
requirements: [NFR-022, NFR-023, NFR-025, NFR-027]
work_item: WI-20260903-correct-lease-trail
source_review: claude
---

# Incomplete Codex lease trail during protocol bootstrap

## Trigger and impact

Claude observed that Codex's completed AEA-adoption work item did not declare
ownership of `AGENTS.md`, `docs/multi-agent-workflow.md`, or
`docs/collaboration/*`, although Codex had created or edited them earlier.

The issue did not alter application code, but it made the collaboration record
incomplete and weakened the honesty and shared-checkout guarantees the protocol
was intended to establish.

## Evidence and diagnosis

- `WI-20260903-adopt-aea-principles` was opened after the initial collaboration
  protocol and vault edits.
- Its scope covered the later AEA adoption edits only.
- Its activity log acknowledged pre-existing uncommitted work but did not enumerate
  Codex's earlier unleased paths.
- The protocol did not exist when those first files were created. That explains
  sequencing but does not make the resulting trail complete.

Root cause: protocol bootstrap and framework adoption were treated as one
continuous documentation task, but ownership was registered only for the later
phase. The completed record therefore described current scope accurately while
omitting relevant historical edit provenance.

## Recovery

- Accepted Claude's finding without qualification.
- Created `WI-20260903-correct-lease-trail` before making corrective edits.
- Enumerated historically affected paths as retrospective provenance, explicitly
  not as retroactive ownership.
- Linked this incident from the original work item rather than silently changing
  its original scope.
- Avoided `AGENTS.md` and `AGENTS-1.md` because current Git state suggests possible
  concurrent agent output.

## Durable safeguard

When introducing a collaboration protocol into an already-active repository:

1. create a bootstrap work record as the first coordination artifact possible;
2. enumerate files already changed before the protocol existed;
3. label that list retrospective provenance, not a lease;
4. separate bootstrap, adoption, and later implementation into distinct work
   items; and
5. have another agent review the first trail.

This incident itself is the durable example. A future protocol edit should add a
formal bootstrap clause after active ownership of protocol files is confirmed.

## Verification and remaining risk

The record can be mechanically checked for links and formatting, but historical
exclusive ownership cannot be reconstructed. `AGENTS-1.md` remains unexplained
and untouched. A human or current Claude session should decide whether it is a
deliberate alternative, a conflict copy, or a file to reconcile.
