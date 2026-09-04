# Logify

Logify is a dependency-free Go command-line tool that recursively scans mixed Tomcat/Java and Apache HTTPD log bundles and produces one self-contained interactive HTML incident timeline.

Development follows a documentation-first process. The canonical [functional
and non-functional requirements](docs/requirements/README.md) define scope,
acceptance criteria, implementation status, and planned gaps.

Engineering and multi-agent work follows three [core principles](docs/principles.md):
honesty, antifragility, and knowledge first.
Their complete Logify translation—including Shared Understanding, authoritative
evidence, and the six-layer outer harness—is maintained in the
[Adaptive Experience Architecture adoption map](docs/framework-adoption.md).

Codex and Claude coordinate through a [document-based shared-checkout
protocol](docs/collaboration/README.md) backed by an
[Obsidian-compatible project knowledge vault](docs/knowledge/index.md).

The HTML report includes a dedicated **issue queue** next to the timeline. You
can create an issue from one event or one deduplicated group, edit its title,
add tags, flag it for attention, and move it through `open`, `investigating`,
`blocked`, `resolved`, or `dismissed`. Creating an issue never changes the log
bundle.

Follow-up data can be **exported and imported** as `logify-follow-up.json`
(schema `logify-follow-up-v1`). Browser local storage is a convenience cache for
the same generated report; it is not the portable copy and is never sent over
the network. Import validates the schema version, skips invalid records, and
reports unmatched evidence IDs. Notes, owner, and due date are preserved on
import/export but are not editable in this baseline (FR-021).

## Build and run

Requires Go 1.22 or newer.

```powershell
go test ./...
go build -o logify.exe ./cmd/logify
.\logify.exe -output report.html C:\path\to\support-bundle
```

Optional RFC3339 bounds filter timestamped events (untimestamped events are excluded when a bound is active). Bounds are compared as absolute instants. Java and Apache error timestamps without an offset are parsed as UTC, so they will not stay in a `+02:00` "morning" window unless that window still covers the UTC instant. Apache access stamps that include an offset are compared using that offset.

```powershell
.\logify.exe -from 2026-09-03T08:00:00+02:00 -to 2026-09-03T12:00:00+02:00 C:\logs
```

On `testdata/case`, that example keeps the two Apache access events (`10:00:03+02:00` and `10:00:04+02:00`) and drops Tomcat/Apache error lines whose timezone-less `10:00:00,123` / `10:00:05` stamps become `10:00Z` and fall after `12:00+02:00` (`10:00:00.000Z`). `-from` / `-to` require a zone or `Z`; a value such as `2026-09-03T08:00:00` is rejected.

## Behavior

- Recursively discovers `.log`, `.out`, `access_log`, and `error_log` files.
- Uses the first directory beneath the input root as the instance name.
- Recognizes common Java/log4j-style timestamps and levels, Apache combined/common access logs, and Apache error logs.
- Joins Java stack frames, `Caused by`, and elided-frame lines to their leading event.
- Normalizes every record into a common model, assigns HTTP severity from status class, and sorts timestamped events chronologically.
- Generates stable signatures from normalized first lines and aggregates repeats per instance while retaining first/last occurrence times.
- Embeds all data, styles, and JavaScript in the report. Timeline filters work offline by text, severity, instance, and source. The issue queue adds combined text, tag, flag, state, owner, severity, instance, and overdue filters.
- Treat the HTML file as sensitive: it contains a copy of parsed log text (messages, paths, and host identifiers) and should be shared like the original bundle. Exported follow-up JSON contains operator titles, tags, and any imported notes.

## Issue follow-up in the report

1. Open the generated HTML file (no server required).
2. On a timeline row, choose **Create issue**. The issue id is stable for that evidence group (`issue-v1-…`).
3. Use **Issue queue** to edit the title, add or remove tags, flag or unflag, and change workflow state.
4. **Show evidence** returns to the matching timeline row. **Open issue** goes the other way.
5. **Export follow-up JSON** writes a portable file. **Import follow-up JSON** loads it into this report. **Clear local follow-up data** drops the browser copy after a confirmation.

The export schema is documented in [`docs/knowledge/decisions/ADR-0002-follow-up-export-schema.md`](docs/knowledge/decisions/ADR-0002-follow-up-export-schema.md). Identity rules are in [`docs/knowledge/decisions/ADR-0001-follow-up-identities.md`](docs/knowledge/decisions/ADR-0001-follow-up-identities.md).

## Current limits

Format detection is filename/path based and intentionally conservative. Logs with custom date formats, multi-line messages that do not resemble Java stack traces, compressed/rotated logs without a recognized suffix, and timezone-less timestamps may need additional parser profiles. Timezone-less Java and Apache error timestamps are treated as UTC by Go's parser.
