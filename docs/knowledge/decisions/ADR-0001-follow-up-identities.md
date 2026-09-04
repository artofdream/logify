---
id: ADR-0001-follow-up-identities
type: decision
status: accepted
owner: cursor-agent
created: 2026-09-04
updated: 2026-09-04
requirements: [FR-017, FR-022, FR-024, NFR-010, NFR-017, NFR-018]
supersedes: []
---

# ADR-0001: Version evidence-derived follow-up identities in the report layer

## Context and evidence

FR-017 requires a stable issue identifier and a durable link to originating
timeline evidence. NFR-017 forbids display-order identity and requires versioned
schema or signature changes. The analyzer already produces a deterministic
signature, but that signature is shared across instances and does not identify
the representative source location of a group.

Draft PR #1 computed `evidenceId` inside the analyzer and added `FirstSeen`.
This change keeps the analyzer unchanged: the report package derives identity
from fields the analyzer already emits. Recurring-count and observation times
must stay out of the identity so a later FR-024 can recognize the same group.

## Decision

Each timeline row (already a per-instance signature group) receives:

```text
evidence-v1-<lowercase SHA-256 hex>
```

The digest input is the UTF-8 encoding of these NUL-delimited fields:

```text
logify-evidence-v1, signature, instance, slash-normalized relative file, line
```

Timeline position, generation time, message text, occurrence count, first seen,
and last seen are excluded. Creating an issue maps deterministically:

```text
issue-v1-<the evidence digest suffix>
```

Creating again from the same evidence selects the existing issue. The existing
event signature remains signature schema v1. A future identity change must use a
new prefix. Import (FR-022) must reject unknown prefixes with an actionable
error rather than silently reinterpret them.

`firstSeen` in the report payload is the representative event `timestamp` when
`hasTimestamp` is true (first discovered row for that group). `lastSeen` is the
analyzer `LastSeen`. That is not a chronological minimum unless discovery order
matches time order.

## Alternatives considered

- Analyzer-owned `EvidenceID` / pointer `FirstSeen` (PR #1) was rejected for
  this PR because the task forbids rewriting the analyzer.
- Timeline indexes change under filters.
- Random UUIDs cannot reconnect the same evidence after a new analysis.
- Signature alone is instance-local and drops file/line provenance.

## Consequences and risks

Issue creation is deterministic, offline, and dependency-free. Moving otherwise
unchanged evidence to a different relative path or line creates a new identity.
SHA-256 is provided by the Go standard library and computed at report write
time so the page script does not need Web Crypto on `file://`.

## Verification

Report-package tests must prove stability, independence from display/count/time
fields, and sensitivity to provenance. Follow-up JS tests must reuse the same
embedded `evidenceId` when creating an issue twice.
