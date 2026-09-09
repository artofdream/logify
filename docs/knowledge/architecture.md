---
id: logify-architecture
type: architecture
status: active
owner: human
updated: 2026-09-09
sources: [../../README.md, ../../cmd/logify/main.go, ../../internal/analyzer, ../../internal/redact, ../../internal/report]
---

# Current architecture

Logify is a dependency-free Go CLI. `cmd/logify` accepts a directory and options;
`internal/analyzer` discovers and normalizes logs, builds signatures, groups
repeats, and orders the timeline; `internal/redact` compiles optional
operator-supplied replacement rules; `internal/report` emits one offline HTML file
from `page.html`, `page.css`, `page.js`, and `followup.js` (embedded at build
time). Optional `-redact` / `-redact-file` rules run after evidence IDs are
computed and only rewrite report-embedded log-derived strings (ADR-0005 /
NFR-006). The analyzer is not responsible for issue identity: the report package
derives `evidence-v1-…` IDs from existing event fields and the page script owns
the follow-up store, local cache, and JSON export/import. Issue notes, owner,
and due date are operator metadata (FR-021 / ADR-0003): edited in the page
script, persisted in `logify-follow-up-v1`, and compared for overdue against
the UTC calendar date of the report clock. An issue may link additional
evidence groups (`linkedEvidence`); import surfaces signature matches and new
occurrence counts for review and does not change workflow state (FR-024 /
ADR-0007).

Recoverable scan problems are structured warnings (`walk-error`, `open-error`,
`scan-overflow`, `scan-error`) with file, category, optional line/range, and
message. `filesScanned` equals `filesProcessed + filesFailed`; `filesSkipped`
counts walk paths that could not be visited (FR-015 / ADR-0004).

Issue-workflow usability (NFR-021 / ADR-0006): flag and state are named in
text; common actions restore focus after a card rebuild and confirm in a live
status region. `nfr021_filter_probe.js` times `store.filter` at 10,000 issues;
the page still renders every matching card (see Q-002).

This note describes observed structure. Requirements remain authoritative for
intended behavior, and tests/compiler output remain evidence of implementation.

The target trust model follows the [AEA adoption mapping](../framework-adoption.md):
reviewable incident understanding is derived from authoritative log evidence and
protected by an outer harness of guides, sensors, loop, memory, permissions, and
observability.
