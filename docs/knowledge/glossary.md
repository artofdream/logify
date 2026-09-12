---
id: logify-glossary
type: glossary
status: active
owner: human
updated: 2026-09-12
---

# Glossary

- **Evidence:** an observed log record or verifiable artifact with provenance.
- **Inference:** a conclusion derived from evidence and labeled with its rule or
  uncertainty.
- **Correlation group:** an inferred set of timeline events that share a
  documented identifier or a labeled heuristic (FR-012 / ADR-0008). Kind is
  `exact` or `heuristic`; confidence is `high` or `low`. Events stay
  individually inspectable. Heuristic `client-ip-window` groups are one
  in-window access/Tomcat pair and are not transitively unioned.
- **Issue:** a follow-up unit linked to one or more evidence groups.
- **Linked evidence:** additional event-group snapshots attached to an issue
  besides the originating `evidence` (`linkedEvidence` in `logify-follow-up-v1`).
- **Signature match:** a timeline group with the same `signature` and `instance`
  as linked evidence, offered for review rather than auto-linked (FR-024).
- **Overdue issue:** an issue whose `due` calendar date is strictly before the
  UTC calendar date of the report clock and whose state is not `resolved` or
  `dismissed` (FR-021 / ADR-0003).
- **Rotated log:** a supported log name with one numeric or date suffix
  (`catalina.out.1`, `access.log.2026-09-03`, `error.log.1.gz`) recognized
  after FR-016 canonicalization (ADR-0009).
- **Gzip stream:** a discovered `*.gz` file decoded incrementally with
  `compress/gzip` over the opened file. Not tar/zip extraction.
- **Scan warning:** a recoverable discovery or read problem that does not abort
  the bundle. Categories are `walk-error`, `open-error`, `scan-overflow`, and
  `scan-error` (FR-015 / ADR-0004). Invalid gzip after open is `scan-error`.
- **Processed / skipped / failed inputs:** supported files that were opened,
  walk paths that could not be visited, and supported files that could not be
  opened. `filesScanned = filesProcessed + filesFailed`.
- **Redaction:** optional, operator-supplied string replacement applied to
  report-embedded log-derived text before HTML write (NFR-006 / ADR-0005). Off
  by default; not secret detection.
- **Signature:** deterministic normalized identity used to group recurring events.
- **Work item:** a bounded unit of agent/human work with advisory path ownership.
- **Lease:** time-limited declaration of edit intent; not a filesystem lock.
- **Parse confidence:** `high` when a format-specific parser matched the
  record; `low` when the line was retained as unrecognized (FR-008 / NFR-028).
- **Unparsed record:** a retained unrecognized line. Counted on the analysis
  result and report summary; distinct from a scan warning.
- **Probe:** a direct observation, test, or command that supports a status claim.
- **Issue-filter probe:** Node harness that imports 10,000 synthetic issues and
  times `store.filter` for NFR-021 AC4. It does not render issue cards.
- **Promotion:** moving durable learning from session/handoff context into a
  canonical requirement, ADR, test, fixture, or knowledge note.
