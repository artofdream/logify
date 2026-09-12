---
id: HO-20260912-cursor-agent-nfr028
type: handoff
status: proposed
owner: cursor-agent
created: 2026-09-12T06:20:00Z
work_item: WI-20260912-nfr028-outer-harness
requirements: [NFR-028, FR-008, NFR-015, NFR-022]
next_owner: human
---

# Handoff: NFR-028 outer harness probes

## Scope and requirement IDs

NFR-028 (stays Partial), FR-008 (parse confidence), NFR-015 / NFR-022 (named
probes for Implemented claims). PR #16 against `main`. Do not merge from this
handoff.

## Evidence consulted

`docs/principles.md`, collaboration protocol, work-item/dispatch scan,
`docs/framework-adoption.md` on `origin/main` `5580cd5`, NFR-028 ACs.

## Changes and artifacts

- `internal/harness/nfr028_test.go` — docs/link, ledger mirror, Implemented
  probe map, open-WI ownership + CODEOWNERS
- Analyzer/report parse confidence and unparsed / correlation-confidence counts
- `.github/CODEOWNERS`, `docs/collaboration/permissions.md`
- Adoption ledger and NFR-028 gap narrowed; status remains Partial

## Validation and observed results

Ran:

- `gofmt -w cmd internal` — clean (`gofmt -l` empty)
- `go test ./...` — pass (all packages)
- `go build -o logify.exe ./cmd/logify` — pass
- `go vet ./...` — pass
- `git diff --check` — pass
- `./logify.exe -output sample-report.html testdata/case` — 6 events,
  `unparsed=1`, correlations 0
- correlate fixture — `unparsedRecords:0`, `highConfidenceCorrelations:1`,
  `lowConfidenceCorrelations:1`

CI on PR #16 was not observed in this handoff (local only).

## Assumptions and confidence

High for the new probes and counts. Permissions remain advisory except where
an operator enables GitHub required reviews / `validate` on `main` (Unknown
from this tree).

## Failures or conflicting evidence

First harness run failed on two `review` hygiene WIs with `requirements: []`.
The probe was corrected: empty requirements allowed except on `active`/`blocked`.

## Uncommitted or concurrent changes

None at handoff time after the validation commit.

## Open questions and next action

Human review of PR #16. Do not mark NFR-028 Implemented until Permissions
has a real merge/path lock or the AC is rewritten, and Sensors still lack
WCAG / in-browser DOM / NFR-009.
