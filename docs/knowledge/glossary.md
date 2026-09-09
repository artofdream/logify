---
id: logify-glossary
type: glossary
status: active
owner: human
updated: 2026-09-09
---

# Glossary

- **Evidence:** an observed log record or verifiable artifact with provenance.
- **Inference:** a conclusion derived from evidence and labeled with its rule or
  uncertainty.
- **Issue:** a follow-up unit linked to one or more evidence groups.
- **Overdue issue:** an issue whose `due` calendar date is strictly before the
  UTC calendar date of the report clock and whose state is not `resolved` or
  `dismissed` (FR-021 / ADR-0003).
- **Scan warning:** a recoverable discovery or read problem that does not abort
  the bundle. Categories are `walk-error`, `open-error`, `scan-overflow`, and
  `scan-error` (FR-015 / ADR-0004).
- **Processed / skipped / failed inputs:** supported files that were opened,
  walk paths that could not be visited, and supported files that could not be
  opened. `filesScanned = filesProcessed + filesFailed`.
- **Redaction:** optional, operator-supplied string replacement applied to
  report-embedded log-derived text before HTML write (NFR-006 / ADR-0005). Off
  by default; not secret detection.
- **Signature:** deterministic normalized identity used to group recurring events.
- **Work item:** a bounded unit of agent/human work with advisory path ownership.
- **Lease:** time-limited declaration of edit intent; not a filesystem lock.
- **Probe:** a direct observation, test, or command that supports a status claim.
- **Issue-filter probe:** Node harness that imports 10,000 synthetic issues and
  times `store.filter` for NFR-021 AC4. It does not render issue cards.
- **Promotion:** moving durable learning from session/handoff context into a
  canonical requirement, ADR, test, fixture, or knowledge note.
