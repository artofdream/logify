---
id: ADR-0015-directory-bundle-input-model
type: decision
status: accepted
owner: grok-bot
created: 2026-10-07
updated: 2026-10-07
requirements: [FR-001, FR-002, FR-003, NFR-004, NFR-005]
supersedes: []
---

# ADR-0015: Directory-tree bundle input with filename-based discovery

**Status:** Accepted (retroactive, recorded 2026-10-07)
**Introduced in:** `cbc0c24` feat: establish log analysis foundation
**Refined in:** `de88717` fix: happy-path report crash and review quick wins (#2)
**Related requirements:** FR-001, FR-002, FR-003, NFR-004, NFR-005

This is a **Documented** record reconstructed after the fact from repository
history. It records a decision that was already in force; it does not change
behavior. Where no written rationale exists, the gap is marked `OPEN` rather
than filled in.

Scope note: this record covers the **base** input model. Rotated names and
single-stream `.gz` files are a later extension recorded in ADR-0009 (FR-016).
Recoverable scan failures are covered by ADR-0004 (FR-015). The CLI flag
surface is covered by ADR-0010 (NFR-016).

## Context and evidence

- `cbc0c24` added FR-001 to FR-003 and the analyzer that implements them:
  - FR-001: the CLI accepts **exactly one input directory**, walks it
    recursively, and fails non-zero with a useful error for a missing path or
    a non-directory. Rationale (as written): "Support bundles commonly contain
    logs from several nested server instances."
  - FR-002: discover files ending in `.log` / `.out` and Apache-style
    `access_log` / `error_log` names. Ignore everything else without failing.
    Rationale: "Avoid treating arbitrary bundle artifacts as logs."
  - FR-003: the instance is the first directory below the root, or `root` for
    a top-level file. The source type is Tomcat/Java, Apache access, or Apache
    error. Every event keeps its relative path and starting line. Rationale:
    "Operators must know which runtime emitted an event."
- Code in `cbc0c24`: `analyzer.Analyze` resolves the root with
  `filepath.Abs`, rejects non-directories (`"%s is not a directory"`), walks
  with `filepath.WalkDir`, filters with `looks()` (basename only), and
  classifies with `detect()` (basename only). `instance()` takes the first
  path segment of the relative path. Files are opened read-only with `os.Open`.
- `cbc0c24` also folded in `WI-20260903-claude-parser-findings`, made before
  the first commit. That change made a conventional `error.log` under any
  instance directory count as Apache error, instead of depending on
  `httpd`/`apache` in the path.
- `de88717` (#2) narrowed access-log detection from "basename contains
  `access`" to conventional names (`isApacheAccessName`), so files such as
  `AccessControl.log` stay Tomcat/Java.
- The README "Current limits" section says format detection is
  "filename/path based and intentionally conservative".

## Decision

1. Input is a single local **directory**, the unpacked support bundle. Logify
   does not accept a single file, an archive (`.zip`, `.tar`, `.tar.gz`), or a
   remote location as the input argument. ADR-0009 also keeps multi-member
   archives out of scope.
2. Discovery walks the whole tree recursively and decides by **basename
   only**. It does not sniff content. Unrecognized files are silently ignored,
   and that is not an error.
3. Source type comes from conventional basenames: Apache access names,
   Apache error names (`error.log`, `error_log`, `ssl_error.log`,
   `ssl_error_log`), and everything else discovered counts as Tomcat/Java.
4. The **instance** label is the first directory below the analysis root, or
   `root` for files directly in it. Every event keeps its slash-normalized
   relative path and starting line number as provenance.
5. Input is read-only and untrusted (NFR-004, NFR-005). Logify never writes,
   renames, extracts beside, or deletes anything in the bundle.

## Alternatives considered

- **Content sniffing to detect log type or gzip:** for gzip, ADR-0009
  rejected magic-byte detection. For log type in general: `OPEN`. No record
  of evaluation was found.
- **Accept archives directly:** ADR-0009 rejects tar/zip support (path-escape
  risk, a different feature). For the original FR-001 scope: `OPEN`. No
  record of evaluation was found.
- **Configurable instance mapping** (for example deeper directory levels or
  hostname from file content): no record of evaluation was found. Why the
  first directory level specifically was chosen is `OPEN` beyond FR-001's
  "nested server instances" rationale.

## Consequences and risks

- Operators must unpack a bundle archive before running Logify. Files that
  are already `.gz` are streamed (ADR-0009).
- Bundles whose layout does not put one instance per top-level directory get
  misleading instance labels (for example all `root`, or a date folder
  treated as an instance). This limit is not documented in the README.
- Logs with non-conventional names are silently skipped. They show up only as
  missing data, not as a warning, because FR-002 AC3 says ignored files do not
  fail analysis.
- Any discovered file that is not Apache access or error is parsed as
  Tomcat/Java. Unrecognized lines are kept as low-confidence untimestamped
  events rather than dropped (FR-002 AC4, NFR-028 observability counts).

## Verification

- `TestRunUsageWhenDirectoryMissing` (`cmd/logify`): no directory argument
  gives exit 2 and usage.
- `TestAnalyzeEmptyDirectorySlices`, `TestAnalyzeFixtures`,
  `TestDetectConventionalApacheErrorLog`, `TestDetectAccessNames`, and
  `TestLooksRotatedAndGzipNames` in `internal/analyzer/analyzer_test.go`.
- CI fixture smoke runs `logify -output sample-report.html testdata/case`,
  which has `tomcat-a/` and `httpd-b/` instance directories.
