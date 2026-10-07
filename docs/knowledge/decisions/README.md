---
id: logify-decisions-index
type: index
status: active
owner: grok-bot
updated: 2026-10-07
---

# Architecture decision records (ADR index)

All Logify ADRs live in this directory, `docs/knowledge/decisions/`. There is
no separate `docs/adr/` folder. New ADRs use [TEMPLATE.md](TEMPLATE.md), take
the next free `ADR-NNNN` number, and keep the outcome immutable: a later
change supersedes by link (see [vault structure](../_STRUCTURE.md)). Several
ADRs were renumbered when concurrent PRs collided (noted inside ADR-0005,
-0006, -0007, -0011, and -0012). Check `main` and open PRs before picking a
number.

"Recorded in" is the commit that added the ADR file. ADR-0013 to ADR-0016 are
**retroactive**: they were written on 2026-10-07 from git history for
decisions already in force since the cited commits. They are Documented
records, not new decisions, and any rationale without evidence is marked
`OPEN` inside them.

## Index

| ADR | Title | Status | Requirements | Summary | Recorded in |
|---|---|---|---|---|---|
| [ADR-0001](ADR-0001-follow-up-identities.md) | Version evidence-derived follow-up identities in the report layer | Accepted | FR-017, FR-022, FR-024, NFR-010, NFR-017, NFR-018 | Report layer derives versioned `evidence-v1-…` IDs from analyzer fields, without display order or counts. | `64bbc1c` (#3) |
| [ADR-0002](ADR-0002-follow-up-export-schema.md) | Portable follow-up export schema v1 | Accepted | FR-022, FR-024, NFR-018, NFR-019, NFR-020 | `logify-follow-up-v1` JSON export/import with schema version and validation limits. | `64bbc1c` (#3) |
| [ADR-0003](ADR-0003-follow-up-details.md) | Editable follow-up details and overdue rule | Accepted | FR-021, FR-022, NFR-019 | Notes, owner, and due-date editors on schema v1 without a bump, with safe rendering and a UTC overdue rule. | `01ae8bf` (#4) |
| [ADR-0004](ADR-0004-scan-warnings.md) | Structured scan warnings and input counts | Accepted | FR-015, NFR-007, NFR-008 | Categorized warnings (`walk-error`, `open-error`, `scan-overflow`, `scan-error`) and reconciled file counts. | `261eaf7` (#8) |
| [ADR-0005](ADR-0005-optional-report-redaction.md) | Optional report-time redaction | Accepted (guarantee review **OPEN**, see below) | NFR-006, NFR-010, NFR-016, NFR-017 | Opt-in `-redact` / `-redact-file` rules applied to report-embedded strings after evidence IDs, plus sensitivity warnings. | `e55e333` (#10) |
| [ADR-0006](ADR-0006-issue-workflow-usability.md) | Issue-workflow keyboard, semantics, and filter probe | Accepted | NFR-021 | Focus restore after card rebuilds, text-named flag and state, visible match count, and a 10k filter probe. | `9869fb4` (#11) |
| [ADR-0007](ADR-0007-recurring-evidence-merge.md) | Multi-evidence issues and reviewable signature matches | Accepted | FR-024, FR-022, NFR-017, NFR-018, NFR-019 | Issues link extra evidence groups. Import surfaces recurring signatures for review and never auto-changes state. | `7a96fb1` (#5) |
| [ADR-0008](ADR-0008-event-correlation.md) | Deterministic event correlation rules | Accepted | FR-012, FR-023 | Named exact versus heuristic correlation rules with evidence. Weak evidence means no correlation. | `76cae98` (#13) |
| [ADR-0009](ADR-0009-rotated-gzip-logs.md) | Rotated and gzip log discovery | Accepted | FR-016, FR-002, FR-003, FR-010, NFR-001, NFR-004, NFR-008 | Canonicalize rotation and `.gz` basenames and stream gzip via stdlib. No tar or zip support. | `5580cd5` (#15) |
| [ADR-0010](ADR-0010-cli-compatibility.md) | Backward-compatible CLI evolution | Accepted | NFR-016, NFR-011 | SemVer flag-compatibility policy, the public flag surface, and `-version` with release-tag ldflags injection. | `0ca7be7` (#17) |
| [ADR-0011](ADR-0011-accessible-report-checks.md) | Stdlib accessibility checks and explicit report labels | Accepted | NFR-013, NFR-021, NFR-003 | Explicit labels, severity as text, a contrast fix, a stdlib a11y checker in CI, and 25-card issue paging. | `273293d` (#19) |
| [ADR-0012](ADR-0012-scale-benchmark.md) | Generated scale fixture and online merge | Accepted | NFR-009, NFR-008, FR-010, FR-012 | Generated (not committed) 1 GiB fixture, online dedup while scanning, and a CI smoke bench. | `0423728` (#18) |
| [ADR-0013](ADR-0013-stdlib-only-go-executable.md) | Dependency-free, standard-library-only Go executable | Accepted (retroactive, recorded 2026-10-07) | NFR-001, NFR-002 | The shipped binary uses only the Go stdlib, and `go.mod` has no requires. Original rationale is `OPEN`. | this PR (decision since `cbc0c24`) |
| [ADR-0014](ADR-0014-single-self-contained-html-report.md) | Single self-contained, offline HTML report as the only output | Accepted (retroactive, recorded 2026-10-07) | FR-013, FR-014, NFR-003, NFR-004, NFR-020 | One HTML file with embedded data, CSS, and JS. No server and no network. Original rationale is `OPEN`. | this PR (decision since `cbc0c24`, refined `64bbc1c`) |
| [ADR-0015](ADR-0015-directory-bundle-input-model.md) | Directory-tree bundle input with filename-based discovery | Accepted (retroactive, recorded 2026-10-07) | FR-001, FR-002, FR-003, NFR-004, NFR-005 | One local directory, recursive walk, basename-only discovery and typing, instance from the first directory, read-only. | this PR (decision since `cbc0c24`, refined `de88717`) |
| [ADR-0016](ADR-0016-tag-triggered-release-workflow.md) | Tag-triggered cross-platform release workflow | Accepted (retroactive, recorded 2026-10-07) | NFR-001, NFR-002, NFR-016 | A `v*` tag push runs validation, cross-compiles six static targets, and publishes a GitHub Release with checksums. | this PR (decision since `5a97f17`, refined `319145c`, `cef5e20`) |

## Coverage notes

- **Version injection via ldflags** is covered by ADR-0010 (decision 4).
  ADR-0016 covers only the release mechanism around it.
- **Overall CLI shape** (public flags, defaults, `-version`, deprecation
  policy) is covered by ADR-0010. The single positional directory argument is
  covered by ADR-0015 (with FR-001).
- **Redaction** is covered by ADR-0005 and is not re-recorded here.

## Open review items

- **OPEN: redaction guarantee not yet reviewed.** The draft Logify SSDD in
  `artofdream/dso` (`03-architecture/ssdd/logify.md`, Draft v0.1, not
  baselined) flags the redaction guarantee as not reviewed. ADR-0005 records
  the decision and its known limits (best-effort, opt-in, incomplete presets,
  follow-up fields not covered). No independent review of whether those
  limits are acceptable for the sponsor's data-handling expectations has been
  recorded. Until one is, treat the guarantee as OPEN. This note does not
  change ADR-0005's status.
- Rationale gaps marked `OPEN` inside ADR-0013 to ADR-0016 (for example why
  stdlib-only and single-file HTML were chosen originally, release target and
  flag choices, Action pinning, and tag protection) are waiting on the
  sponsor or the original authors.
