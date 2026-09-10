---
id: ADR-0004-scan-warnings
type: decision
status: accepted
owner: cursor-agent
created: 2026-09-09
updated: 2026-09-09
requirements: [FR-015, NFR-007, NFR-008]
supersedes: []
---

# ADR-0004: Structured scan warnings and input counts

## Context and evidence

FR-015 requires recoverable scan problems to continue, keep successful work,
expose a warning count, identify file / category / line-or-range, and reconcile
processed, skipped, and failed inputs. On `main` before this change, warnings
were plain `path: error` strings. Mid-file `bufio.ErrTooLong` already kept
earlier events. `filesScanned` plus the warning list did not separate skipped
and failed inputs.

Observed warning sources in `internal/analyzer`:

1. `filepath.WalkDir` callback errors
2. `os.Open` failures on discovered supported files
3. `bufio.Scanner` overflow (`ErrTooLong`, NFR-008)
4. any other `Scanner.Err()` after open

No decode/parse warning path exists: unrecognized access lines become fallback
events, not warnings.

## Decision

1. Replace `[]string` warnings with a structured `Warning` (`file`, `category`,
   optional `line` / `lineEnd`, `message`). Categories are only the four names
   that have an emitting code path: `walk-error`, `open-error`, `scan-overflow`,
   `scan-error`.
2. `file` is the slash-normalized path relative to the analysis root when
   relative conversion succeeds.
3. Overflow `line` is the overflowing line (last successfully read line + 1).
   Other scan errors use the last successfully read line. Open and walk errors
   omit line fields.
4. Count semantics:
   - `filesScanned`: supported files the scanner attempted
   - `filesProcessed`: opened successfully, including recoverable mid-file
     failures that keep earlier events
   - `filesFailed`: supported files that could not be opened
   - `filesSkipped`: walk paths that could not be visited
   - Identity: `filesScanned = filesProcessed + filesFailed`
5. Unsupported filenames remain a discovery filter, not skipped inputs. Children
   inside an unreadable directory are unknown and are not counted.
6. The CLI prints `Result.SummaryLine()` and one structured warning per line.
   The HTML report embeds the same objects and lists them in a Scan warnings
   section. Embedded JSON is not a versioned public API; the HTML file is the
   product, so the warning shape changes with this report.

## Alternatives considered

- Keep string warnings and encode category/line in the text. Rejected: AC4
  requires identifiable fields, and the report UI would have to re-parse.
- Count overflow files as failed. Rejected: AC2 keeps their earlier events, so
  calling the input failed would imply it was discarded.
- Count every non-matching filename as skipped. Rejected: those paths are not
  inputs; inflating skipped would hide real walk failures.

## Consequences and risks

Operators who scraped the old `Wrote … (N events from M files, K warnings)`
line must read the new `processed=` / `skipped=` / `failed=` summary. Report
JavaScript is embedded in the same file, so old string-warning scripts cannot
meet new reports.

## Verification

Analyzer tests cover overflow structure, open-error isolation, walk-error skip,
category classification, and count identity. Report tests cover embedded
objects, counts, and HTML escaping of warning text.
