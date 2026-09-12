# Functional requirements

## Input and discovery

### FR-001 — Analyze a directory tree

- **Priority:** Must
- **Status:** Implemented
- **Rationale:** Support bundles commonly contain logs from several nested server
  instances.
- **Acceptance criteria:**
  1. The CLI accepts exactly one input directory.
  2. It recursively examines all accessible descendants.
  3. A missing path or non-directory path returns a non-zero exit code and a
     useful error.

### FR-002 — Discover supported log files

- **Priority:** Must
- **Status:** Implemented
- **Rationale:** Avoid treating arbitrary bundle artifacts as logs.
- **Acceptance criteria:**
  1. Files ending in `.log` or `.out` are considered.
  2. Apache-style `access_log` and `error_log` names are considered.
  3. Unsupported files are ignored without failing the analysis.
  4. A non-empty line in a discovered access log that does not match the supported
     format is retained as an untimestamped event rather than silently discarded.
- **Note:** Numeric/date rotation suffixes and `.gz` on these same name
  conventions are specified by FR-016. They do not expand FR-002 to arbitrary
  archives or unrecognized decorations.

### FR-003 — Identify source instance and type

- **Priority:** Must
- **Status:** Implemented
- **Rationale:** Operators must know which runtime emitted an event.
- **Acceptance criteria:**
  1. Each event includes an instance identifier derived from the first directory
     below the analysis root, or `root` for a top-level file.
  2. Each event identifies its source as Tomcat/Java, Apache access, or Apache
     error.
  3. Each event retains its relative source path and starting line number.

## Parsing and normalization

### FR-004 — Parse Tomcat and Java application logs

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. Common ISO-like `YYYY-MM-DD HH:MM:SS` and `T`-separated timestamps are
     recognized with optional fractional seconds and offsets.
  2. Comma and period fractional-second separators are accepted.
  3. TRACE, DEBUG, INFO, WARN/WARNING, ERROR/SEVERE, and FATAL levels map to the
     normalized severity model.
  4. A recognized timestamp, severity, and message are retained in one event.

### FR-005 — Preserve multiline Java failures

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. Exception or error headers immediately following a leading event are joined
     to it.
  2. Indented stack frames, `at` frames, `Caused by:` lines, and elided frames are
     joined in their original order.
  3. A subsequent timestamped event ends the current multiline event.

### FR-006 — Parse Apache HTTPD access logs

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. Common/combined-style client, timestamp, request, status, and byte fields are
     recognized.
  2. Request, status, and bytes are represented in the normalized message.
  3. 2xx/3xx responses map to INFO, 4xx to WARN, and 5xx to ERROR.
  4. The numeric status code is retained.

### FR-007 — Parse Apache HTTPD error logs

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. Bracketed Apache timestamps with optional microseconds are recognized.
  2. Apache severity labels are mapped into the normalized severity model.
  3. PID/client prefixes do not obscure the event message.
  4. Facility-qualified labels such as `[proxy:error]` use the severity component.

### FR-008 — Normalize all records

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:** Every emitted event has severity, source type,
  instance, file, line, message, signature, occurrence count, and parse
  confidence; it also records whether a timestamp exists. Parse confidence is
  `high` when a format-specific parser matched the record and `low` when the
  line was retained as unrecognized. Unparsed-record and correlation-confidence
  counts belong on the analysis result and report summary (NFR-028).

## Correlation and filtering

### FR-009 — Order a cross-instance timeline

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. Timestamped events from all instances are sorted chronologically.
  2. Untimestamped events appear after timestamped events.
  3. Events with equal timestamps retain deterministic discovery order.

### FR-010 — Group repeated events

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. A deterministic signature is produced from source type, severity, and a
     normalized first message line.
  2. Volatile large numbers, UUID-like values, and hexadecimal addresses do not
     prevent otherwise identical failures from grouping.
  3. Grouping occurs only within the same instance.
  4. A group reports occurrence count and first/last observed time.

### FR-011 — Filter by time range

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. `-from` and `-to` accept RFC3339 timestamps and are inclusive bounds.
  2. Invalid values cause a non-zero exit and useful error.
  3. Untimestamped events are excluded when either time bound is active.

### FR-012 — Correlate related events beyond chronological proximity

- **Priority:** Should
- **Status:** Implemented
- **Rationale:** Shared request IDs, client addresses, exception causes, or short
  time windows can expose one incident spanning HTTPD and Tomcat.
- **Acceptance criteria:**
  1. Correlation rules are deterministic and documented.
  2. Events remain individually inspectable.
  3. The report displays correlation groups and their supporting evidence.
  4. False-positive-prone heuristics are distinguishable from exact identifiers.
  5. Each correlation identifies its rule and confidence; no correlation is
     preferable to an unsupported relationship.
- **Notes:** Rules, kind/confidence labels, and non-goals are in
  [ADR-0008](../knowledge/decisions/ADR-0008-event-correlation.md). Verified:
  `shared-request-id` (exact/high) and `client-ip-window` (heuristic/low, 5s,
  one access+Tomcat pair, non-loopback IPv4). Heuristic pairs are not unioned
  across time or distinct IPs. Identifiers and client addresses from FR-010
  collapsed occurrences still participate; the timeline stays one row per
  (instance, signature). Java `Caused by:` chains stay one observed event and
  are not a cross-event rule. `testdata/case` produces zero groups.
  `testdata/correlate` produces one exact group (4 members) and one heuristic
  group (2 members); unlabeled, loopback, out-of-window, and singleton IDs
  stay ungrouped. Report embeds groups as a JSON array and names rule, kind,
  confidence, and evidence. Known limits (not AC gaps): no unlabeled
  UUID/number matching; IPv6 is not harvested from Tomcat free text; Apache
  error `[client]` alone does not satisfy the heuristic HTTPD side.

## Report and CLI

### FR-013 — Generate a self-contained HTML report

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. `-output` selects the destination and defaults to `logify-report.html`.
  2. HTML, CSS, JavaScript, and event data are contained in one file.
  3. The report requires no network access or external assets.
  4. Log content is escaped and cannot inject executable markup.

### FR-014 — Explore the report interactively

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. A user can search message, file, and instance text.
  2. A user can filter by severity, instance, and source type.
  3. Each row shows time state, severity, instance, source, message, provenance,
     signature, and repeat information where applicable.
  4. The report summarizes unique events, scanned files, and warnings.

### FR-015 — Report recoverable scan problems

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. An unreadable supported file or scanner failure produces a warning where
     the scanner can continue.
  2. Recoverable file errors do not discard successfully parsed files.
  3. The CLI and report expose the warning count.
  4. Each warning identifies the affected file, failure category, and line or
     range where available.
  5. Summary counts reconcile processed, skipped, and failed inputs.
- **Warning model:** Each warning is a structured record with `file`,
  `category`, optional `line` / `lineEnd`, and `message`. Categories are
  deterministic and limited to the code paths that can emit them:
  - `walk-error`: directory walk could not visit a path (file or directory).
  - `open-error`: a supported file could not be opened.
  - `scan-overflow`: a line exceeded the 4 MiB incremental scanner limit
    (NFR-008). `line` is the overflowing line (last successfully read line + 1).
  - `scan-error`: any other incremental read failure after open. `line` is the
    last successfully read line when known.
- **Input counts:**
  - `filesScanned`: supported files the scanner attempted to parse.
  - `filesProcessed`: supported files that were opened (including files whose
    mid-file overflow or scan error kept earlier events).
  - `filesFailed`: supported files that could not be opened.
  - `filesSkipped`: walk paths that could not be visited. Children inside an
    unreadable directory are unknown and are not invented.
  - Identity: `filesScanned = filesProcessed + filesFailed`.
    `filesScanned + filesSkipped = filesProcessed + filesFailed + filesSkipped`.
  Unsupported filenames are a discovery filter, not skipped inputs.
- **Verification:** `TestOpenErrorKeepsOtherFiles`, `TestWalkErrorSkippedAndOtherFilesKept`,
  `TestOverflowKeepsEarlierEvents`, `TestWarningFromScanCategories`,
  `TestWriteStructuredWarningsAndCounts`, plus `go test ./...`, `go vet`,
  and `./logify.exe -output sample-report.html testdata/case` →
  `6 events from 3 files; processed=3 skipped=0 failed=0; 0 warnings`.
  A mixed unreadable/overflow bundle produced
  `processed=2 skipped=1 failed=1; 3 warnings` with categories `walk-error`,
  `scan-overflow`, and `open-error`; the HTML report listed those warnings and
  kept the `before` and `visible` events.

### FR-016 — Support compressed and rotated logs

- **Priority:** Should
- **Status:** Implemented
- **Rationale:** Support bundles often include logrotate and Tomcat dated files,
  including `.gz` copies. Operators should not have to unpack those by hand
  before analysis.
- **Acceptance criteria:**
  1. Common numeric/date rotated suffixes are discovered.
  2. Gzip-compressed supported logs can be streamed without manual extraction.
  3. Rotation does not create duplicate events beyond normal signature grouping.
- **Discovery:** After lowercasing the basename, Logify strips one trailing
  `.gz` and then one rotation suffix before applying the FR-002 / FR-003
  `looks()` / `detect()` conventions. Recognized rotation suffixes are `.N`,
  `.YYYY-MM-DD` with an optional time fragment, `.YYYYMMDD` with optional
  hour/minute/second digits, optional trailing `.txt` (Tomcat AccessLogValve),
  and logrotate `dateext` `-YYYYMMDD`. Unsupported names (`notes.gz`,
  `archive.tar.gz`, `.zip`, `.bz2`) remain a discovery filter.
- **Gzip:** A discovered `*.gz` file is decoded with `compress/gzip` as a
  single stream over the opened file. Contents are never written to disk.
  There is no tar/zip extraction and no magic-byte detection on names that
  lack `.gz`. An invalid gzip header after a successful `os.Open` is a
  `scan-error`; the file counts as processed. Mid-stream gzip failures keep
  events already parsed (NFR-007 / NFR-008).
- **Dedup:** Rotation copies use the existing FR-010 instance+signature
  grouping. Distinct messages stay separate. The same signature on two
  instances is not merged.
- **Verification:** `TestLooksRotatedAndGzipNames`,
  `TestCanonicalLogNameStripsOneDecoration`, `TestDetectAccessNames` rotated
  cases, `TestAnalyzeRotatedGzipFixtures` (9 grouped events from 6 files;
  `tomcat-a` shared ERROR occurrences=2; `tomcat-b` not merged),
  `TestGzipStreamParsesWithoutExtraction`,
  `TestInvalidGzipIsScanErrorAndKeepsSiblings`,
  `TestTruncatedGzipKeepsEarlierEvents`. Repository validation:
  `gofmt -l cmd internal` clean; `go test ./...` pass; `go build -o logify.exe
  ./cmd/logify` pass; `go vet ./...` pass; `git diff --check` clean.
  `./logify.exe -output sample-report.html testdata/case` →
  `6 events from 3 files; processed=3 skipped=0 failed=0; 0 warnings`.
  `./logify.exe -output rotated-report.html testdata/rotated` →
  `9 events from 6 files; processed=6 skipped=0 failed=0; 0 warnings`.

## Issue follow-up

### FR-017 — Create an issue from timeline evidence

- **Priority:** Must
- **Status:** Implemented
- **Rationale:** A diagnostic finding must become an explicit unit of follow-up
  work without losing its supporting evidence.
- **Acceptance criteria:**
  1. A user can create an issue from one event or one deduplicated event group.
  2. The issue receives a stable identifier and a user-editable title.
  3. The originating event signature, instance, source file, line, first seen,
     last seen, and occurrence count remain linked as evidence.
  4. Creating an issue never changes the source log bundle.
- **Note:** Evidence identity is computed in the report package from signature,
  instance, file, and line (ADR-0001). `firstSeen` is the representative event
  timestamp (first discovered row for that group), not a chronological minimum
  unless discovery order matches time order.

### FR-018 — Tag issues

- **Priority:** Must
- **Status:** Implemented
- **Rationale:** Tags support classification across component, team, symptom, and
  investigation dimensions.
- **Acceptance criteria:**
  1. A user can add and remove multiple free-text tags on an issue.
  2. Tags are trimmed, compared case-insensitively, and displayed consistently.
  3. Duplicate tags cannot be assigned to the same issue.
  4. A user can filter issues by one or more tags.

### FR-019 — Flag issues for attention

- **Priority:** Must
- **Status:** Implemented
- **Rationale:** Operators need a fast visual marker independent of log severity.
- **Acceptance criteria:**
  1. A user can flag and unflag an issue.
  2. A flag is visually distinct and does not overwrite the source event severity.
  3. A user can show only flagged issues.
  4. Flag state survives report reload through the supported persistence mechanism.

### FR-020 — Track issue workflow state

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. Every issue has one of: `open`, `investigating`, `blocked`, `resolved`, or
     `dismissed`.
  2. New issues default to `open`.
  3. A user can change state and filter by state.
  4. The latest state and its modification time are retained.
  5. Resolving or dismissing an issue does not remove its evidence or history.

### FR-021 — Record follow-up details

- **Priority:** Should
- **Status:** Implemented
- **Rationale:** Operators need to record who owns follow-up, when it is due, and
  investigation notes without leaving the offline report.
- **Acceptance criteria:**
  1. An issue can store notes, an optional owner, and an optional due date.
  2. Notes are treated as untrusted text and rendered safely.
  3. Overdue unresolved issues are visibly identified.
  4. A user can filter by owner and overdue state.
- **Note:** Due dates are calendar dates (`YYYY-MM-DD`). An issue is overdue when
  `due` is strictly before the UTC calendar date of the report clock
  (`Date#toISOString` date prefix, or an injected test clock) and the state is
  not `resolved` or `dismissed`. A due date of today is not overdue. Empty owner,
  due, and notes normalize to `null`. Fields persist in `logify-follow-up-v1`
  (no schema bump; ADR-0002 / ADR-0003). Rendering uses DOM text APIs only.

### FR-022 — Persist issue tracking data

- **Priority:** Must
- **Status:** Implemented
- **Rationale:** Follow-up data must outlive a browser session while the generated
  HTML report remains portable and offline.
- **Acceptance criteria:**
  1. Tags, flags, workflow states, notes, owners, and due dates can be exported to
     a documented machine-readable file.
  2. A user can import that file into the corresponding report.
  3. Import validates schema version and reports invalid or unmatched records.
  4. Issue records link by stable issue/evidence identifiers rather than display
     position.
  5. Export and import require no network or server.
  6. Browser-local persistence may improve convenience but is not the only way to
     preserve or transfer follow-up data.
- **Note:** An evidence id has one owning issue. Import skips a record whose
  originating or linked evidence is already owned by a different issue, with
  the same reason as `linkEvidence` (`evidence already linked to <id>`).
  Re-importing the same issue id may replace that issue's refs.

### FR-023 — Present an issue work queue

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. The report exposes a dedicated issue list separate from the raw timeline.
  2. Each item shows title, flag, state, tags, owner, due date, linked evidence,
     and last modification time where available.
  3. A user can combine text, tag, flag, state, owner, severity, instance, and
     overdue filters.
  4. A user can return from an issue to its evidence in the timeline.
  5. Counts distinguish raw events, deduplicated event groups, and tracked issues.
  6. The UI distinguishes observed log evidence, inferred correlations, and
     operator-authored issue metadata.
- **Note:** Owner, due date, and notes are displayed when present (including
  after import). Editors for those fields are FR-021. Observed-record counts are
  the sum of group `occurrences`. Inferred correlations are listed separately
  with rule, kind, confidence, and evidence (FR-012 / ADR-0008); they are not
  operator issue metadata.

### FR-024 — Merge recurring evidence into an existing issue

- **Priority:** Should
- **Status:** Implemented
- **Rationale:** The same failure often reappears as another event group or in a
  later bundle. Operators need to attach that evidence to the existing issue
  without minting a duplicate or silently changing workflow state.
- **Acceptance criteria:**
  1. A user can link additional event groups to an existing issue.
  2. The issue retains every linked evidence reference.
  3. A newly generated report can match recurring signatures to imported issue
     records and clearly identify newly observed occurrences.
  4. Automatic matches are reviewable before changing issue state.
- **Note:** Originating evidence stays in `evidence`; additional refs are
  `linkedEvidence` on `logify-follow-up-v1` (no schema bump; ADR-0002 /
  ADR-0007). Automatic matches use `signature` + `instance`. Newly observed
  occurrences are a higher live occurrence count or newer `lastSeen` on a
  linked evidence id. Import, link, dismiss, and acknowledge do not change
  workflow state. The originating snapshot cannot be unlinked. Rendering uses
  DOM text APIs only. The recurring review panel titles lastSeen-only rows
  (equal occurrence counts, newer `lastSeen`) as a newer last-seen time.
