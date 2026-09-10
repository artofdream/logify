---
id: ADR-0007-recurring-evidence-merge
type: decision
status: accepted
owner: cursor-agent
created: 2026-09-09
updated: 2026-09-10
requirements: [FR-024, FR-022, NFR-017, NFR-018, NFR-019]
supersedes: []
---

# ADR-0007: Multi-evidence issues and reviewable signature matches

Renumbered from a colliding ADR-0004 after FR-015, NFR-006, and NFR-021 landed.

## Context and evidence

FR-024 requires linking additional event groups to an existing issue, retaining
every evidence reference, matching recurring signatures when a new report is
imported, and keeping those automatic matches reviewable before issue state
changes. ADR-0001 already excludes occurrence counts and observation times from
`evidence-v1-…` so the same group can be recognized later. ADR-0002 stores one
originating `evidence` object on `logify-follow-up-v1`.

Creating a second issue from a later occurrence of the same failure would split
follow-up. Auto-reopening a resolved issue when the signature reappears would
change workflow state without an operator decision.

## Decision

1. Do not bump the export schema. `evidence` remains the originating snapshot
   and the required v1 field. Additional refs are an optional `linkedEvidence`
   array of the same snapshot shape. Dismissed automatic matches persist as
   optional `ignoredEvidence` (array of `evidence-v1-…` ids). Older files
   without these fields import as empty arrays.
2. Evidence ids stay `evidence-v1-…`. Linking does not mint a new issue id.
   Creating from an event that is already linked selects that issue.
3. The originating `evidence` cannot be unlinked. Additional refs can.
4. Automatic match key is `signature` + `instance` (instance-local, same as
   analyzer grouping). A timeline event is a **candidate** when that key matches
   any linked snapshot, the evidence id is not already linked to any issue, and
   the id is not in `ignoredEvidence`. Manual **Link to existing issue** may
   attach any event group, including a different signature.
5. **Newly observed occurrences** compare the stored snapshot to the live
   timeline row with the same evidence id: `live.occurrences > stored.occurrences`
   or a newer `lastSeen`. Import and hydrate do not overwrite stored counts, so
   the delta survives reload until the operator acknowledges.
6. Import, candidate listing, linking, dismissing, and acknowledging **must not**
   change workflow state. A resolved issue with new occurrences stays resolved
   until the operator changes state.
7. Treat linked file paths, signatures, and titles as untrusted text. The page
   script assigns them with DOM `textContent` / `value` APIs only.

8. An evidence id has one owning issue. `importJSON` skips a record whose
   originating or linked evidence is already owned by a different issue, using
   the same reason as `linkEvidence` / `createFromEvent`
   (`evidence already linked to <id>`). Re-importing the same issue id may
   replace that issue's refs. First accepted record in a file wins when two
   imported issues claim the same id.
9. The recurring review panel must show the newer `lastSeen` when the row is
   lastSeen-only (equal occurrence counts). Issue cards already branch on this
   case.

Limits: at most 50 linked evidence snapshots per issue (originating + additional).

## Alternatives considered

- Schema v2 with `evidence` as an array was rejected: it would break v1 readers
  that expect an object. Optional additive fields preserve FR-022 files.
- Auto-link on signature match was rejected: AC4 requires review before state
  (and implicit membership) changes.
- Matching on signature alone (across instances) was rejected: ADR-0001 and
  FR-010 group only within an instance.
- Overwriting stored occurrence counts on hydrate (v1 import behavior for
  display freshness) was rejected for FR-024 because it would hide new
  occurrences after the first persist.

## Consequences and risks

Operators can merge recurring groups in one issue and carry that set across
reports. A later log rotation that changes relative path or line creates a new
evidence id; it appears as a candidate rather than an automatic link. Export
writes stored snapshots, so unacknowledged occurrence counts travel as the last
accepted baseline.

## Verification

Node tests must link and unlink additional evidence, round-trip `linkedEvidence`,
report occurrence updates and signature candidates on import, and prove that
import/link/acknowledge/dismiss leave workflow state unchanged. The page script
must not assign `innerHTML`.
