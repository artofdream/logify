---
id: WI-20260909-nfr006-redaction
type: work-item
status: review
owner: cursor-agent
created: 2026-09-09T20:40:00Z
updated: 2026-09-09T20:48:00Z
lease_expires: 2026-09-10T08:00:00Z
scope:
  - cmd/logify/
  - internal/redact/
  - internal/report/report.go
  - internal/report/report_test.go
  - internal/report/page.html
  - internal/report/page.css
  - internal/report/page.js
  - testdata/redact/
  - README.md
  - docs/requirements/non-functional-requirements.md
  - docs/knowledge/architecture.md
  - docs/knowledge/decisions/ADR-0005-optional-report-redaction.md
  - docs/collaboration/work-items/WI-20260909-nfr006-redaction.md
  - docs/collaboration/handoffs/HO-20260909-2048-cursor-agent-nfr006.md
requirements: [NFR-006]
depends_on: []
supersedes: []
---

# NFR-006 optional redaction and sensitivity warning

## Goal

Close the remaining NFR-006 gaps: optional configurable redaction before report
embedding, and a more prominent user-facing sensitivity warning. Verify that
Logify still has no telemetry or automatic uploads.

## Non-goals

NFR-021 10k/a11y audit, FR-012/016, merging other PRs, release tags, telemetry
backends, changing source log files.

## Acceptance criteria

1. No telemetry or automatic uploads (verify still true).
2. Documentation and user-facing surfaces warn that reports contain copied log
   data.
3. Optional redaction can remove configurable secrets and personal identifiers
   before report generation. Default behavior is unchanged (no redaction unless
   opted in).

## Planned files and ownership

See `scope`. Owner: `cursor-agent`.

## Evidence and assumptions

- NFR-006 was Partial; AC3 and the weak warning were the documented gaps.
  Status is now Implemented with observed validation on this branch.
- WI-20260904-report-crash-quick-wins listed redaction as a non-goal and is in
  `review` with an expired lease.
- No active/queued dispatch claims NFR-006.
- Evidence IDs do not include message text (ADR-0001), so display redaction can
  run after identity is computed.

## Validation

`gofmt`, `go test ./...`, `go build`, `go vet`, `git diff --check`, and a
fixture report run.

## Activity log

- `2026-09-09T20:40:00Z` — cursor-agent — claimed NFR-006 redaction/warning work.
- `2026-09-09T20:55:00Z` — cursor-agent — implemented opt-in redaction, banner, and docs; validation pending.
- `2026-09-09T20:48:00Z` — cursor-agent — `go test ./...`, `go vet`, fixture CLI runs passed; PR #10.

## Handoff or completion

See `docs/collaboration/handoffs/HO-20260909-2048-cursor-agent-nfr006.md`.
