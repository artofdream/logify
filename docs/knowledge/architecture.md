---
id: logify-architecture
type: architecture
status: active
owner: human
updated: 2026-09-12
sources: [../../README.md, ../../cmd/logify/main.go, ../../internal/analyzer, ../../internal/redact, ../../internal/report, ../../testdata/scale/README.md]
---

# Current architecture

Logify is a dependency-free Go CLI. `cmd/logify` accepts a directory and options
(`-output`, `-from`, `-to`, `-redact`, `-redact-file`); `-version` / `-V` print
`var version` (default `dev`; ldflags-overridable) and exit without analyzing
(ADR-0010 / NFR-016). Public flag names and defaults stay supported for a
major SemVer version, or a documented deprecation warning precedes removal.
`internal/analyzer` discovers and normalizes logs (including common rotated
and `.gz` names streamed via `compress/gzip`, ADR-0009 / FR-016), builds signatures, groups
repeats **online while scanning** (ADR-0012 / NFR-009; extra correlation hints
are capped), orders the timeline, and applies documented correlation rules
(ADR-0008 / FR-012); `internal/redact` compiles optional
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
ADR-0007). Correlation groups are inferences (`exact` vs `heuristic`) with a
named rule, confidence, and evidence string; they do not merge timeline rows
or mint issues. Correlation still reads labeled IDs and client addresses from
occurrences collapsed by FR-010. The IP-window heuristic emits pairwise
access/Tomcat groups only.

Recoverable scan problems are structured warnings (`walk-error`, `open-error`,
`scan-overflow`, `scan-error`) with file, category, optional line/range, and
message. `filesScanned` equals `filesProcessed + filesFailed`; `filesSkipped`
counts walk paths that could not be visited (FR-015 / ADR-0004). Each event
carries parse confidence (`high` when a format-specific parser matched, `low`
when the line was retained unrecognized). The result and HTML summary count
unparsed records (by occurrence) and correlation groups by confidence
(NFR-028). `internal/harness` fails CI when critical relative docs links break,
the adoption ledger contradicts requirement statuses, an Implemented
requirement has no named test/CI/JS probe, or an open work item is missing
ownership fields.

Issue-workflow usability (NFR-021 / ADR-0006): flag and state are named in
text; common actions restore focus after a card rebuild and confirm in a live
status region. The queue pages matching cards (`ISSUE_PAGE_SIZE = 25`).
`nfr021_filter_probe.js` times `store.filter` at 10,000 issues plus a 25-item
slice. Q-001 (published reference hardware) remains open, so NFR-021 stays
Partial.

Accessible report interaction (NFR-013 / ADR-0011): static and dynamic
controls use explicit `for=` labels; timeline severity is `Severity: …` text;
`--control-border` meets 3:1 against `--panel`; `internal/report/a11y.go`
checks generated-report labels and CSS token contrast in CI without npm/axe.

NFR-009 scale: `GenerateScaleBundle` streams a multi-instance storm fixture to
`/tmp` (or `testdata/scale/generated/`, gitignored). `TestNFR009ScaleSmoke` and
`BenchmarkNFR009ScaleSmoke` run in ordinary tests/CI; the 1 GiB child-process
run is `make bench-nfr009`. See [ADR-0012](decisions/ADR-0012-scale-benchmark.md)
and [testdata/scale/README.md](../../testdata/scale/README.md).

This note describes observed structure. Requirements remain authoritative for
intended behavior, and tests/compiler output remain evidence of implementation.

The target trust model follows the [AEA adoption mapping](../framework-adoption.md):
reviewable incident understanding is derived from authoritative log evidence and
protected by an outer harness of guides, sensors, loop, memory, permissions, and
observability.
