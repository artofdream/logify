---
id: WI-20260912-nfr028-outer-harness
type: work-item
status: active
owner: cursor-agent
created: 2026-09-12T05:55:00Z
updated: 2026-09-12T05:55:00Z
lease_expires: 2026-09-12T14:00:00Z
scope:
  - .github/CODEOWNERS
  - .github/workflows/ci.yml
  - AGENTS.md
  - README.md
  - docs/framework-adoption.md
  - docs/requirements/functional-requirements.md
  - docs/requirements/non-functional-requirements.md
  - docs/collaboration/README.md
  - docs/collaboration/permissions.md
  - docs/collaboration/work-items/WI-20260912-nfr028-outer-harness.md
  - docs/collaboration/handoffs/HO-20260912-cursor-agent-nfr028.md
  - docs/knowledge/architecture.md
  - docs/knowledge/glossary.md
  - internal/analyzer/model.go
  - internal/analyzer/analyzer.go
  - internal/analyzer/analyzer_test.go
  - internal/report/report.go
  - internal/report/report_test.go
  - internal/report/page.js
  - internal/report/page.html
  - internal/report/page.css
  - internal/harness/
requirements: [NFR-028, FR-008, NFR-015, NFR-022]
depends_on: []
supersedes: []
---

# Close remaining NFR-028 outer-harness gaps

## Goal

Add mechanical sensors and product observability so NFR-028's remaining
Partial gaps are narrower and evidence-based: docs/link + ledger freshness,
unparsed/low-confidence accounting, and honest permissions documentation
(CODEOWNERS + CI/work-item field checks). Do not mark NFR-028 Implemented
unless every AC is honestly met.

## Non-goals

- Full WCAG engine or in-browser DOM/AT execution
- NFR-009 scale benchmark
- Fake merge locks or purchased branch-protection features
- Desktop wrapper or new product dependencies

## Acceptance criteria

1. Critical relative docs links and the adoption-ledger vs requirement-status
   mirror fail `go test` / CI when broken or contradictory.
2. Analyzer/report expose parse confidence and unparsed / correlation-confidence
   counts; Implemented claims without a named probe fail or stay Unknown.
3. CODEOWNERS and a permissions note document real gates; leases stay advisory.
4. `docs/framework-adoption.md` ledger is updated from the new probes.

## Planned files and ownership

cursor-agent owns the paths in frontmatter. Overlap with `review` work items
that already landed on `main` is inspection-only; those leases are not active
edits.

## Evidence and assumptions

Starting from `origin/main` `5580cd5`. No `active`/`queued` dispatch overlaps
NFR-028. Review-status WIs exist but are not editing.

## Validation

Commands in `AGENTS.md`.

## Activity log

- `2026-09-12T05:55:00Z` — cursor-agent — claimed after `git fetch origin main`
  and work-item/dispatch scan.

## Handoff or completion
