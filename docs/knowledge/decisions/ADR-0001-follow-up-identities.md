---
id: ADR-0001-follow-up-identities
type: decision
status: accepted
owner: codex
created: 2026-09-03
updated: 2026-09-03
requirements: [FR-017, FR-022, FR-024, NFR-010, NFR-017, NFR-018]
supersedes: []
---

# ADR-0001: Version evidence-derived follow-up identities

## Context and evidence

FR-017 requires stable issue identity and complete links to originating timeline
evidence. NFR-017 prohibits display-order identity and requires versioned schema
or signature changes. The analyzer already produces a deterministic signature,
but that signature alone is shared by matching messages in different instances
and does not identify the representative source location of a deduplicated group.

Issue count, first seen, and last seen can change when a later bundle contains
additional occurrences. Including them in identity would prevent FR-024 from
recognizing recurring evidence. Browser randomness and timeline indexes have the
same stability problem.

## Decision

Each normalized event group receives an evidence identifier with this form:

```text
evidence-v1-<lowercase SHA-256 hex>
```

The digest input is the UTF-8 encoding of these NUL-delimited fields in order:

```text
logify-evidence-v1, signature, instance, slash-normalized relative file, line
```

NUL is an unambiguous separator because operating-system paths cannot contain a
NUL byte. Timeline position, report generation time, message display text,
occurrence count, first seen, and last seen are excluded. The evidence record
still carries signature, instance, file, line, first seen, last seen, and count.

FR-017 maps one issue deterministically to one evidence group:

```text
issue-v1-<the evidence digest suffix>
```

Creating an issue for that evidence again in the same report selects the existing
issue. Additional evidence-to-issue relationships remain FR-024 scope.

The existing event signature is treated as signature schema v1 for this mapping;
this change does not alter its algorithm. A future signature or evidence-identity
change must use a new explicit version prefix. Import code introduced under
FR-022 must retain readers for known prior versions or reject them with an
actionable migration error; it must never silently reinterpret an older ID.

## Alternatives considered

- Timeline indexes were rejected because filters and sorting change them.
- Random UUIDs were rejected because repeated analysis could not reconnect the
  same evidence without a separate durable mapping.
- Count and observation times were rejected as identity inputs because recurring
  occurrences legitimately change those values.
- Signature alone was rejected because grouping is intentionally instance-local
  and the evidence link must retain source provenance.

## Consequences and risks

Issue creation is deterministic, offline, and dependency-free. Moving otherwise
unchanged evidence to a different relative file or line intentionally creates a
new evidence identity. A future import feature must implement the documented
version behavior. SHA-256 is provided by the Go standard library; the full digest
is retained to make accidental collisions impractical.

## Verification

Analyzer tests must prove stability across repeated construction, independence
from display-only/count/time fields, and sensitivity to provenance. Report tests
must prove that the issue identity and complete evidence fields are embedded and
that unsafe log/title strings are not installed as executable markup.
