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
  instance, file, line, message, signature, and occurrence count; it also records
  whether a timestamp exists.

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
- **Status:** Proposed
- **Rationale:** Shared request IDs, client addresses, exception causes, or short
  time windows can expose one incident spanning HTTPD and Tomcat.
- **Acceptance criteria:**
  1. Correlation rules are deterministic and documented.
  2. Events remain individually inspectable.
  3. The report displays correlation groups and their supporting evidence.
  4. False-positive-prone heuristics are distinguishable from exact identifiers.
  5. Each correlation identifies its rule and confidence; no correlation is
     preferable to an unsupported relationship.

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

### FR-016 — Support compressed and rotated logs

- **Priority:** Should
- **Status:** Proposed
- **Acceptance criteria:**
  1. Common numeric/date rotated suffixes are discovered.
  2. Gzip-compressed supported logs can be streamed without manual extraction.
  3. Rotation does not create duplicate events beyond normal signature grouping.

## Issue follow-up

### FR-017 — Create an issue from timeline evidence

- **Priority:** Must
- **Status:** Proposed
- **Rationale:** A diagnostic finding must become an explicit unit of follow-up
  work without losing its supporting evidence.
- **Acceptance criteria:**
  1. A user can create an issue from one event or one deduplicated event group.
  2. The issue receives a stable identifier and a user-editable title.
  3. The originating event signature, instance, source file, line, first seen,
     last seen, and occurrence count remain linked as evidence.
  4. Creating an issue never changes the source log bundle.
  5. Creating the same issue again in one report reuses the evidence-derived issue
     identity rather than creating a display-order-dependent duplicate.
  6. Until FR-022 is implemented, the report identifies issue data as transient,
     warns before discarding it, and does not imply that it has been persisted.

### FR-018 — Tag issues

- **Priority:** Must
- **Status:** Proposed
- **Rationale:** Tags support classification across component, team, symptom, and
  investigation dimensions.
- **Acceptance criteria:**
  1. A user can add and remove multiple free-text tags on an issue.
  2. Tags are trimmed, compared case-insensitively, and displayed consistently.
  3. Duplicate tags cannot be assigned to the same issue.
  4. A user can filter issues by one or more tags.

### FR-019 — Flag issues for attention

- **Priority:** Must
- **Status:** Proposed
- **Rationale:** Operators need a fast visual marker independent of log severity.
- **Acceptance criteria:**
  1. A user can flag and unflag an issue.
  2. A flag is visually distinct and does not overwrite the source event severity.
  3. A user can show only flagged issues.
  4. Flag state survives report reload through the supported persistence mechanism.

### FR-020 — Track issue workflow state

- **Priority:** Must
- **Status:** Proposed
- **Acceptance criteria:**
  1. Every issue has one of: `open`, `investigating`, `blocked`, `resolved`, or
     `dismissed`.
  2. New issues default to `open`.
  3. A user can change state and filter by state.
  4. The latest state and its modification time are retained.
  5. Resolving or dismissing an issue does not remove its evidence or history.

### FR-021 — Record follow-up details

- **Priority:** Should
- **Status:** Proposed
- **Acceptance criteria:**
  1. An issue can store notes, an optional owner, and an optional due date.
  2. Notes are treated as untrusted text and rendered safely.
  3. Overdue unresolved issues are visibly identified.
  4. A user can filter by owner and overdue state.

### FR-022 — Persist issue tracking data

- **Priority:** Must
- **Status:** Proposed
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

### FR-023 — Present an issue work queue

- **Priority:** Must
- **Status:** Proposed
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

### FR-024 — Merge recurring evidence into an existing issue

- **Priority:** Should
- **Status:** Proposed
- **Acceptance criteria:**
  1. A user can link additional event groups to an existing issue.
  2. The issue retains every linked evidence reference.
  3. A newly generated report can match recurring signatures to imported issue
     records and clearly identify newly observed occurrences.
  4. Automatic matches are reviewable before changing issue state.
