---
id: ADR-0002-follow-up-export-schema
type: decision
status: accepted
owner: cursor-agent
created: 2026-09-04
updated: 2026-09-04
requirements: [FR-022, NFR-018, NFR-019, NFR-020]
supersedes: []
---

# ADR-0002: Portable follow-up export schema v1

## Context and evidence

FR-022 requires a machine-readable, offline export/import path that is not
browser-local storage alone. NFR-018 requires an explicit schema version and
documented format. NFR-019 requires validation, size/count limits, and
untrusted-text handling.

## Decision

Exported follow-up data is a JSON object:

```text
{
  "schema": "logify-follow-up-v1",
  "schemaVersion": 1,
  "exportedAt": "<RFC3339>",
  "reportRoot": "<analysis root as embedded in the report>",
  "issues": [ /* FollowUpIssue */ ]
}
```

Each issue includes `id`, `title`, `state`, `flagged`, `tags`, `owner`, `due`,
`notes`, `createdAt`, `modifiedAt`, and an `evidence` object with `id`,
`signature`, `instance`, `file`, `line`, `firstSeen`, `lastSeen`,
`occurrences`, `severity`, and `sourceType`. `owner`, `due`, and `notes` are
present so imported FR-021 fields round-trip; this PR does not provide editors
for them.

Import rules:

1. Reject the file when `schema` / `schemaVersion` is missing or unknown.
2. Reject the file when it exceeds 5 MiB or 10,000 issues.
3. Validate each issue; skip invalid records and keep valid ones (partial
   import). Report unmatched evidence IDs without dropping the issue.
4. Link by `id` and `evidence.id`, never by list position.
5. Preserve unknown future fields on an issue when re-exporting.
6. Treat all operator strings as untrusted text (DOM text APIs only).

Browser `localStorage` is a convenience cache keyed by
`logify-follow-up-v1:<root>:<generatedAt>`. It is not the portable path.

Limits: title 200 characters; at most 50 tags of 64 characters; notes 8000;
owner 200; due `YYYY-MM-DD`.

## Alternatives considered

- localStorage-only persistence fails FR-022 AC6 and is not transferable.
- Atomic-only import would discard valid records when one is bad; partial
  import is the default (NFR-019).
- CSV was rejected as weaker for nested evidence and schema versioning.

## Consequences and risks

Operators can move follow-up data between machines by copying one JSON file
alongside the HTML report. A later schema v2 must keep a v1 reader or fail
with a migration error.

## Verification

Node tests must round-trip create → export → import, reject a bad schema, and
report unmatched evidence. Go tests must embed `schema` in the page script and
keep the report free of network URLs.
