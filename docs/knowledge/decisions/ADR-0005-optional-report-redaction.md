---
id: ADR-0005-optional-report-redaction
type: decision
status: accepted
owner: cursor-agent
created: 2026-09-09
updated: 2026-09-11
requirements: [NFR-006, NFR-010, NFR-016, NFR-017]
supersedes: []
---

# ADR-0005: Optional report-time redaction

Renumbered from a colliding ADR-0004 when FR-015 (#8) landed first and kept
`ADR-0004-scan-warnings.md`.

## Context and evidence

NFR-006 requires no telemetry, a warning that reports contain copied log data,
and optional redaction of configurable secrets and personal identifiers before
report generation. The previous gap was AC3 plus a weak user-facing warning
(one README bullet). Analysis already normalizes volatile tokens for
signatures; that is not operator-configurable redaction and must stay
deterministic (NFR-010). Evidence IDs are bound to signature, instance, file,
and line (ADR-0001 / NFR-017), not message text.

## Decision

1. Redaction is opt-in via repeatable `-redact` rules and optional
   `-redact-file`. Default analysis and report embedding stay unredacted.
2. Apply redaction in the report package after `EvidenceID` is computed, to
   `root`, structured warning `file`/`message`, each event `message`,
   `file`, `instance`, and `clientAddr`, and each correlation `evidence`
   string. Do not rewrite source logs, signatures, timestamps,
   severity, status codes, source type, or correlation IDs / rule / kind /
   confidence labels.
3. Rules are stdlib `regexp` only: named presets (`email`, `ipv4`, `uuid`,
   `bearer`, `jwt`), `literal:<text>`, `regex:<pattern>`, or a bare pattern.
   Reject empty-matching patterns. Replacement token is `[REDACTED]`.
4. Warn on every successful CLI write (stderr) and with a static HTML banner.
   Document coverage and non-coverage in README. Do not claim the report is
   safe to publish.

## Alternatives considered

- Redact during parse: would change signatures and break NFR-010 / follow-up
  identity unless operators always used the same rules.
- Built-in secret scanning without flags: silent default behavior change and
  false confidence.
- Third-party secret libraries: violates the stdlib-only executable constraint.

## Consequences and risks

Operators who omit rules still embed raw log text. Presets are incomplete and
can over-match (`ipv4`). Follow-up JSON typed after generation is not
redacted. A rule that matches part of `root` also changes the localStorage
key for that report.

## Verification

Package tests in `internal/redact` and `internal/report`, plus `cmd/logify`
CLI tests for default vs opted-in behavior and the stderr warning.
