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
the network. Import validates the schema version, skips invalid records
(including an evidence id already owned by a different issue), and reports
unmatched evidence IDs. Notes, owner, and due date are editable on each
issue and persist through the same schema. An issue can link additional event
groups; import of a later report surfaces matching signatures and newly
observed occurrence counts (or a newer last-seen time with the same count)
for review and does not change workflow state.

## Sensitive reports

Logify does not phone home: there is no telemetry and no automatic upload.

The HTML report is a **sensitive copy of parsed log text**. It embeds messages,
relative file paths, instance directory names, correlation evidence, and the
analysis root. Share it only as you would the original support bundle. Do not
publish the file or attach it to an unrestricted ticket. Exported follow-up JSON
is also sensitive: it contains operator titles, tags, notes, owners, and due
dates.

Every successful run prints a one-line reminder on stderr. The report itself
opens with a **Sensitive copy** banner.

Optional redaction is **off by default**. When you opt in, Logify replaces
matches in the embedded root, warnings, each event message, file, instance, and
client address, and each correlation evidence string **before** those strings are
written into the HTML file. Source logs
are never modified. Evidence IDs stay bound to the original provenance so
Import validates the schema version, skips invalid records (including an
evidence id already owned by a different issue), and reports unmatched
evidence IDs. Notes, owner, and due date are editable on each
issue and persist through the same schema. An issue can link additional event
groups; import of a later report surfaces matching signatures and newly
observed occurrence counts (or a newer last-seen time with the same count)
for review and does not change workflow state.

```powershell
.\logify.exe -redact email -redact ipv4 -redact "literal:change-me" -output report.html C:\logs
.\logify.exe -redact-file testdata\redact\rules.txt -output report.html C:\logs
```

Rule forms (repeat `-redact`; or put one rule per line in `-redact-file`, with
`#` comments and blank lines ignored):

| Form | Meaning |
|---|---|
| `email`, `ipv4`, `uuid`, `bearer`, `jwt` | Documented presets (case-insensitive names) |
| `literal:<text>` | Exact substring |
| `regex:<pattern>` | Go RE2 regular expression |
| `<pattern>` | Treated as a Go RE2 regular expression |

Matches become `[REDACTED]`. This is best-effort string replacement, not secret
detection. It does **not** cover:

- Unknown or novel secret formats you did not list
- Values split across lines or encoded (base64, hex dumps) unless your rule matches that form
- IPv6, phone numbers, or other identifiers unless you add a rule
- Operator-typed follow-up titles, notes, owners, tags, or due dates
- Signatures, evidence IDs, timestamps, severity, status codes, source type, or correlation IDs / rule / kind / confidence labels
- The original log bundle (only the HTML copy is rewritten)

Presets are incomplete on purpose. `ipv4` also matches dotted numbers that are
not addresses (for example `1.2.3.4` version strings). Invalid rules and
patterns that match the empty string are rejected.

## Build and run

Requires Go 1.22 or newer on `PATH`. On Windows, if `go` is not recognized after install, close and reopen the terminal so the updated PATH applies.

Generated `logify.exe` and `*-report.html` files (including `sample-report.html` and `report.html`) are gitignored — do not commit them.

```powershell
go test ./...
go build -o logify.exe ./cmd/logify
.\logify.exe -h
.\logify.exe -version
.\logify.exe -output report.html C:\path\to\support-bundle
```

`-h` / `-help` print flags to stderr and exit 0.

`-version` / `-V` print `logify <version>` on stdout and exit 0 without a
directory. Unreleased builds report `dev`. Tagged releases can overwrite that
string later with `go build -ldflags "-X main.version=vX.Y.Z"` (not wired in
the release workflow until `v0.1.0`).

Optional RFC3339 bounds filter timestamped events (untimestamped events are excluded when a bound is active). Bounds are compared as absolute instants. Java and Apache error timestamps without an offset are parsed as UTC, so they will not stay in a `+02:00` "morning" window unless that window still covers the UTC instant. Apache access stamps that include an offset are compared using that offset.

```powershell
.\logify.exe -from 2026-09-03T08:00:00+02:00 -to 2026-09-03T12:00:00+02:00 C:\logs
```

On `testdata/case`, that example keeps the two Apache access events (`10:00:03+02:00` and `10:00:04+02:00`) and drops Tomcat/Apache error lines whose timezone-less `10:00:00,123` / `10:00:05` stamps become `10:00Z` and fall after `12:00+02:00` (`10:00:00.000Z`). `-from` / `-to` require a zone or `Z`; a value such as `2026-09-03T08:00:00` is rejected.

Release: push a `v*` tag to run [`.github/workflows/release.yml`](.github/workflows/release.yml); do not commit the resulting binaries.

## CLI compatibility

Logify versions the operator CLI with [SemVer](https://semver.org/)
(NFR-016 / [ADR-0010](docs/knowledge/decisions/ADR-0010-cli-compatibility.md)).

The public surface is the flags documented here and on `-h`: `-output`
(default `logify-report.html`), `-from`, `-to`, `-redact`, `-redact-file`,
`-h` / `-help`, and `-version` / `-V`. Those names and defaults stay
supported throughout a major version. Removing, renaming, changing a default,
or rejecting a previously accepted value is a breaking change and requires a
new major version. Before a removal or rename, the CLI prints a deprecation
warning on stderr and this README names the successor, for at least one
released minor version in the same major.

Additive flags are minor. Bug fixes that restore documented behavior are
patch. While the product is still `0.y.z`, documented flags are still a
compatibility contract; a break after that warning window ships as `1.0.0`
(or a later major).

## Behavior

- Recursively discovers `.log`, `.out`, `access_log`, and `error_log` files, including common numeric/date rotation suffixes (`catalina.out.1`, `access.log.2026-09-03`, `error.log.1.gz`) and `.gz` on those same names. Gzip is streamed with the Go standard library; tar/zip/bz2 archives are not unpacked.
- Uses the first directory beneath the input root as the instance name.
- Recognizes common Java/log4j-style timestamps and levels, Apache combined/common access logs, and Apache error logs.
- Joins Java stack frames, `Caused by`, and elided-frame lines to their leading event.
- Normalizes every record into a common model, assigns HTTP severity from status class, and sorts timestamped events chronologically.
- Generates stable signatures from normalized first lines and aggregates repeats per instance while retaining first/last occurrence times.
- Surfaces recoverable scan problems as structured warnings (`walk-error`, `open-error`, `scan-overflow`, `scan-error`) with file, category, optional line or range, and message. The CLI and report show the warning count plus processed / skipped / failed input counts (`filesScanned = filesProcessed + filesFailed`). Mid-file overflow keeps events already parsed from that file. Unreadable directories are skipped; children inside them are not invented. Unrecognized lines are retained with low parse confidence and counted as `unparsed` on the summary; correlation groups are counted as high- vs low-confidence.
- Groups related events when a documented rule finds shared evidence (FR-012). Exact labeled request/correlation/trace IDs are high confidence, including identifiers on later occurrences collapsed by FR-010. A client-IP + 5s window across one Apache access event and one Tomcat event is labeled heuristic/low, skips loopback, and is not transitively unioned across clients or times. Events stay on the timeline; groups appear in **Correlation groups** with the rule, confidence, and supporting evidence. No group is invented when evidence is weak.
- Embeds all data, styles, and JavaScript in the report. Timeline filters work offline by text, severity, instance, and source. The issue queue adds combined text, tag, flag, state, owner, severity, instance, and overdue filters.
- Treat the HTML file as sensitive. See [Sensitive reports](#sensitive-reports).

## Issue follow-up in the report

1. Open the generated HTML file (no server required).
2. On a timeline row, choose **Create issue**. The issue id is stable for that evidence group (`issue-v1-…`). **Link to existing issue** attaches this group to another issue without changing that issue's workflow state.
3. Use **Issue queue** to edit the title, notes, optional owner, and optional due date; add or remove tags; flag or unflag; and change workflow state.
4. Unresolved issues whose due date is before today (UTC calendar date of the report clock) show an **Overdue** badge. Filter the queue by owner or **Overdue only**. Resolved and dismissed issues are never overdue.
5. **Show evidence** returns to a matching timeline row. **Open issue** goes the other way. Additional linked groups can be unlinked; the originating evidence cannot.
6. Create, flag, tag, and change state with the keyboard: **Tab** reaches labeled controls, **Enter** adds a tag, and **Left/Right** switches the Timeline and Issue queue tabs. Flag and workflow state are named in text (**Flagged**, **State: Open**), not color alone. The status line confirms each action; the queue shows **Showing N of M issue(s)** when filters change.
7. **Export follow-up JSON** writes a portable file. **Import follow-up JSON** loads it into this report. A later report with the same `signature` + `instance` lists candidate matches and newly observed occurrence counts (or a newer last-seen time) for review; nothing auto-changes state. Duplicate evidence ownership is skipped. **Clear local follow-up data** drops the browser copy after a confirmation.

The export schema is documented in [`docs/knowledge/decisions/ADR-0002-follow-up-export-schema.md`](docs/knowledge/decisions/ADR-0002-follow-up-export-schema.md). Identity rules are in [`docs/knowledge/decisions/ADR-0001-follow-up-identities.md`](docs/knowledge/decisions/ADR-0001-follow-up-identities.md). Overdue and detail-field rules are in [`docs/knowledge/decisions/ADR-0003-follow-up-details.md`](docs/knowledge/decisions/ADR-0003-follow-up-details.md). Scan warning categories and input-count identity are in [`docs/knowledge/decisions/ADR-0004-scan-warnings.md`](docs/knowledge/decisions/ADR-0004-scan-warnings.md). Redaction rules are in [`docs/knowledge/decisions/ADR-0005-optional-report-redaction.md`](docs/knowledge/decisions/ADR-0005-optional-report-redaction.md). Keyboard, confirmation, and the 10,000-issue filter probe are in [`docs/knowledge/decisions/ADR-0006-issue-workflow-usability.md`](docs/knowledge/decisions/ADR-0006-issue-workflow-usability.md) and [`docs/knowledge/research/RES-20260909-nfr021-issue-filter-probe.md`](docs/knowledge/research/RES-20260909-nfr021-issue-filter-probe.md). Multi-evidence merge and review rules are in [`docs/knowledge/decisions/ADR-0007-recurring-evidence-merge.md`](docs/knowledge/decisions/ADR-0007-recurring-evidence-merge.md). Correlation rules are in [`docs/knowledge/decisions/ADR-0008-event-correlation.md`](docs/knowledge/decisions/ADR-0008-event-correlation.md). Rotated and gzip discovery is in [`docs/knowledge/decisions/ADR-0009-rotated-gzip-logs.md`](docs/knowledge/decisions/ADR-0009-rotated-gzip-logs.md). CLI compatibility is in [`docs/knowledge/decisions/ADR-0010-cli-compatibility.md`](docs/knowledge/decisions/ADR-0010-cli-compatibility.md).

## Current limits

Format detection is filename/path based and intentionally conservative. Logs with custom date formats, multi-line messages that do not resemble Java stack traces, rotation or compression schemes outside the documented suffixes (for example `.zip`, `.bz2`, `.xz`, or tar archives), and timezone-less timestamps may need additional parser profiles. Timezone-less Java and Apache error timestamps are treated as UTC by Go's parser. Optional `-redact` rules are best-effort substring/regex replacement and do not detect unknown secrets. Correlation does not use unlabeled numbers, bare UUIDs, exception class names, or loopback client addresses; the IP-window heuristic is IPv4, 5 seconds, and access+Tomcat only.

The issue queue renders every matching card. `store.filter` at 10,000 issues is
measured in milliseconds on the documented probe host; painting 10,000 full
cards is not windowed and will not stay interactive. Narrow filters before
working a large imported set. The probe host is not a published operator
workstation (NFR-021 remains Partial).
