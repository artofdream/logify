# Non-functional requirements

## Portability and deployment

### NFR-001 — Native standalone executable

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. `go build ./cmd/logify` succeeds with Go 1.22 or newer.
  2. The application has no required runtime service or database.
  3. The application uses no third-party Go module at the v1 baseline.

### NFR-002 — Cross-platform behavior

- **Priority:** Must
- **Status:** Partial
- **Acceptance criteria:**
  1. Path handling uses Go portability APIs.
  2. Windows builds produce a runnable `.exe`.
  3. CI verifies Windows, Linux, and macOS builds and tests.
- **Gap:** CI `validate` runs tests and a fixture smoke on Ubuntu only. The
  release workflow cross-compiles Linux, Windows, and macOS (amd64/arm64) from
  Ubuntu; it does not run the test suite on Windows or macOS runners.

### NFR-003 — Offline operation

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:** Analysis and report viewing perform no network calls,
  and reports reference no remote scripts, styles, fonts, or images.

## Security and privacy

### NFR-004 — Treat logs as untrusted input

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. Log contents are never executed as commands or templates.
  2. Embedded JSON cannot terminate the report script context.
  3. Rendered messages, paths, and labels are HTML-escaped.

### NFR-005 — Preserve source data

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:** Analysis opens input logs read-only and never modifies,
  renames, or deletes source bundle content.

### NFR-006 — Minimize data exposure

- **Priority:** Should
- **Status:** Partial
- **Acceptance criteria:**
  1. No telemetry or automatic uploads occur.
  2. Documentation warns that reports contain copied log data.
  3. Optional redaction can remove configurable secrets and personal identifiers
     before report generation.
- **Gap:** Configurable redaction is not implemented, and report sensitivity needs
  a more prominent user-facing warning.

## Reliability and scale

### NFR-007 — Isolate recoverable file failures

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:** A failure reading one supported file does not prevent
  processing of other files, and the failure is surfaced as a warning.

### NFR-008 — Bound per-record memory

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. Files are scanned incrementally rather than loaded whole.
  2. An individual scanned line is limited to 4 MiB.
  3. Scanner overflow becomes a warning rather than a process panic.

### NFR-009 — Scale to realistic support bundles

- **Priority:** Should
- **Status:** Proposed
- **Acceptance criteria:**
  1. A documented benchmark fixture represents at least 1 GiB and multiple
     instances.
  2. Peak memory remains below 512 MiB for that fixture.
  3. Analysis completes within five minutes on documented reference hardware.
  4. Performance regressions are measurable with a repeatable benchmark.

### NFR-010 — Deterministic results

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:** Given identical files, CLI options, locale, and
  timezone, event ordering, signatures, severities, and occurrence counts are
  identical; only report generation time may differ.

## Usability and accessibility

### NFR-011 — Actionable CLI failures

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:** Invalid arguments and unrecoverable failures write a
  concise explanation to standard error and return a non-zero exit code.

### NFR-012 — Responsive report

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:** Timeline content remains readable without horizontal
  page scrolling at desktop and narrow/mobile viewport widths.

### NFR-013 — Accessible report interaction

- **Priority:** Should
- **Status:** Partial
- **Acceptance criteria:**
  1. All controls are keyboard operable and have programmatic labels.
  2. Severity is communicated by text, not color alone.
  3. Text and interactive controls meet WCAG 2.2 AA contrast requirements.
  4. Automated accessibility checks cover the generated report.
- **Gap:** Controls rely on placeholder/option text and no automated accessibility
  check exists.

## Maintainability and verification

### NFR-014 — Automated verification

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. `go test ./...`, `go vet ./...`, and `go build ./cmd/logify` pass.
  2. Parser fixtures cover Java multiline, HTTP access, and HTTP error records.
  3. Tests cover signature normalization and safe, offline report generation.

### NFR-015 — Requirements traceability

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. Every planned product behavior or quality constraint has a stable FR/NFR ID.
  2. Code changes cite affected IDs in the change description.
  3. Non-obvious test-to-requirement relationships cite IDs in test comments or
     names.
  4. Requirement status is updated with the implementing change.

### NFR-016 — Backward-compatible CLI evolution

- **Priority:** Should
- **Status:** Proposed
- **Acceptance criteria:** Existing flags and defaults remain supported throughout
  a major version, or a documented deprecation warning precedes removal.

## Follow-up data integrity

### NFR-017 — Stable follow-up identities

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. Issue identifiers do not depend on timeline display order.
  2. Evidence identifiers remain stable for unchanged normalized input across
     repeated analyses.
  3. Signature algorithm or schema changes are versioned and migration behavior is
     documented.

### NFR-018 — Portable follow-up data

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. Exported follow-up data uses a documented, non-proprietary format.
  2. The format carries an explicit schema version.
  3. Export is deterministic apart from documented timestamps and identifier
     generation.
  4. Unknown future fields do not cause silent data loss.

### NFR-019 — Safe issue metadata handling

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. Imported titles, tags, notes, owners, and other metadata are treated as
     untrusted input and cannot inject executable HTML or JavaScript.
  2. Import applies documented size and count limits.
  3. Invalid records are rejected with actionable errors without discarding valid
     records unless atomic import is explicitly selected.

### NFR-020 — Transparent local persistence

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. The UI states where follow-up data is currently stored.
  2. The UI warns before an action that would discard unsaved changes.
  3. A user can explicitly export, import, and clear local follow-up data.
  4. No follow-up data is transmitted over a network.

### NFR-021 — Issue-workflow usability

- **Priority:** Should
- **Status:** Partial
- **Acceptance criteria:**
  1. Creating, flagging, tagging, and changing issue state are keyboard operable.
  2. Flag and state are communicated by text/icon semantics, not color alone.
  3. Common issue actions provide immediate visible confirmation.
  4. Filtering remains responsive with at least 10,000 tracked issues on documented
     reference hardware.
- **Gap:** Controls are native buttons/inputs with text labels for flag and
  state, and actions update a live status region. There is no measured 10,000-issue
  filter probe or automated accessibility audit.

## Engineering principles

### NFR-022 — Honest evidence and status

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. Material completion claims identify validation evidence.
  2. Assumptions, unverified behavior, failures, and unresolved conflicts are
     explicitly reported.
  3. Requirement status cannot be `Implemented` while an acceptance criterion is
     known to be unmet or unverified.
  4. Agent handoffs use the evidence fields defined in
     `docs/multi-agent-workflow.md`.

### NFR-023 — Antifragile delivery

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. A confirmed defect adds a regression test or documents why one is impractical.
  2. Significant failures produce a durable safeguard, fixture, requirement,
     decision, or runbook improvement.
  3. Changes preserve recoverability and do not destroy unrelated work.
  4. Conflicting agent results are resolved using evidence or retained as an
     explicit open question.

### NFR-024 — Knowledge-first engineering

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. Observable behavior and quality changes begin with affected FR/NFRs and
     acceptance criteria.
  2. Agents consult canonical documentation before implementation.
  3. Consequential new knowledge is captured in requirements, decisions, tests,
     fixtures, or guidance in the same change set.
  4. Missing essential knowledge is researched or recorded as an open question,
     not replaced with an invented fact.

### NFR-025 — Verifiable multi-agent coordination

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. Delegated tasks have bounded scope, ownership, and expected evidence.
  2. Concurrent agents do not intentionally edit the same files without an
     explicit integration plan.
  3. The coordinator inspects artifacts and independently runs final validation.
  4. Consensus alone is never treated as verification.

### NFR-026 — Durable cross-agent knowledge

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. Canonical knowledge is stored as portable, Git-reviewable Markdown.
  2. Decisions, research, incidents, and session context use documented templates
     and link to affected requirements.
  3. Reusable discoveries are promoted from transient session/handoff notes into
     canonical documentation, tests, fixtures, or ADRs.
  4. Unknown, stale, or contradictory knowledge is labeled rather than silently
     presented as current fact.

### NFR-027 — Shared-checkout collision control

- **Priority:** Must
- **Status:** Implemented
- **Acceptance criteria:**
  1. Codex and Claude inspect active work items and Git state before editing.
  2. Each concurrent task declares bounded path ownership and a renewable advisory
     lease in its own work-item file.
  3. Ownership transfers and completions use evidence-based handoff records.
  4. Expired leases do not authorize overwriting uncommitted work.
  5. Overlap conflicts stop edits and preserve both artifacts until an explicit,
     documented resolution.
  6. Remote assignments are published on the shared branch with requirement IDs,
     owner, state, dependency, intended branch, and observed task/PR references.

### NFR-028 — Outer harness coverage

- **Priority:** Must
- **Status:** Partial
- **Rationale:** Trustworthy diagnosis needs controls around interpretation, not
  only parser code.
- **Acceptance criteria:**
  1. **Guides** define requirements, boundaries, playbooks, and ownership.
  2. **Sensors** mechanically test supported behavior and fail when evidence is
     stale, missing, or contradictory.
  3. **Loop** applies Interpret → Act → Verify → Remember to product and delivery
     work.
  4. **Memory** stores durable, source-linked decisions and learning outside
     transient chat.
  5. **Permissions** constrain who or what may modify source evidence, code,
     follow-up data, and integration state.
  6. **Observability** provides provenance and a probe for material status claims;
     without one, status is `Unknown`.
  7. `docs/framework-adoption.md` maintains evidence and gaps for every layer.
- **Gap:** Guides, loop, and local memory are present. Sensors, technical
  permissions, and product/delivery observability remain partial as documented in
  the adoption ledger.
