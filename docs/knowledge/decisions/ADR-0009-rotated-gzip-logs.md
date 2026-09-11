---
id: ADR-0009-rotated-gzip-logs
type: decision
status: accepted
owner: cursor-agent
created: 2026-09-11
updated: 2026-09-11
requirements: [FR-016, FR-002, FR-003, FR-010, NFR-001, NFR-004, NFR-008]
supersedes: []
---

# ADR-0009: Rotated and gzip log discovery

## Context and evidence

FR-016 requires common numeric/date rotated suffixes to be discovered, gzip
copies of supported logs to be readable without manual extraction, and rotation
not to invent extra timeline rows beyond FR-010 grouping.

On `main` before this change, `looks()` required a `.log` / `.out` suffix or an
`access_log` / `error_log` substring. Names such as `catalina.out.1`,
`access.log.2026-09-03`, and `error.log.1.gz` were ignored. `detect()`
classified from the raw basename, so a later `looks()` widening without
stripping decorations would have typed `error.log.1` as Tomcat/Java.

NFR-001 forbids a third-party archive library. NFR-004 / NFR-005 require
untrusted, read-only input. Zip-slip and multi-member archive extraction are
out of scope: the product analyzes files already present in the bundle.

## Decision

1. Canonicalize a basename before `looks()` and `detect()`: lowercase, strip
   one trailing `.gz`, then strip one rotation suffix. Reuse the existing
   access/error/Java name rules on that canonical form.
2. Recognized rotation suffixes are `.N`, `.YYYY-MM-DD` with an optional time
   fragment, `.YYYYMMDD` with optional hour/minute/second digits, an optional
   trailing `.txt` (Tomcat `AccessLogValve`), and logrotate `dateext`
   `-YYYYMMDD`.
3. Open `.gz` files with `compress/gzip.NewReader` over the already-opened
   file. Stream into the existing `bufio.Scanner`. Do not write a decompressed
   sidecar. Do not unpack tar, zip, bz2, or xz. Do not sniff gzip magic on
   names that lack `.gz`.
4. Invalid gzip after a successful `os.Open` is `scan-error` and counts as
   processed (the file was opened; decode failed). Mid-stream gzip errors keep
   events already parsed.
5. Deduplication stays FR-010: same instance + signature collapses rotation
   copies; distinct messages and other instances do not merge.

## Alternatives considered

- Magic-byte gzip detection on every discovered file. Rejected: would reinterpret
  plaintext `.log` files that happen to start with `1f 8b`, and FR-016 names
  `.gz` as the compression signal.
- Extract `.gz` to a temp file, then parse. Rejected: extra disk writes, leftover
  copies, and no streaming bound (NFR-008).
- Support `.tar.gz` / `.zip`. Rejected: path-escape risk and archive inventory
  are a different feature; FR-016 is single-stream gzip on supported log names.
- New warning category for gzip. Rejected: header/decode failures fit
  `scan-error` (ADR-0004).

## Consequences and risks

Operators who previously unpacked `*.log.1.gz` before running Logify can point
at the bundle as-is. Unrecognized schemes remain ignored. A file named
`error.log.gz` that is not gzip produces a scan warning and no events from
that file; siblings still process.

## Verification

`TestLooksRotatedAndGzipNames`, `TestCanonicalLogNameStripsOneDecoration`,
`TestDetectAccessNames` rotated cases, `TestAnalyzeRotatedGzipFixtures`,
`TestGzipStreamParsesWithoutExtraction`, `TestInvalidGzipIsScanErrorAndKeepsSiblings`,
`TestTruncatedGzipKeepsEarlierEvents`, plus repository validation in `AGENTS.md`
and `./logify.exe -output rotated-report.html testdata/rotated`.
