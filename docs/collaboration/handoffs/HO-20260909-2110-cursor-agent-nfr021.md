---
id: HO-20260909-2110-cursor-agent-nfr021
type: handoff
status: review
owner: cursor-agent
created: 2026-09-09T21:10:00Z
work_item: WI-20260909-nfr021-issue-workflow-usability
requirements: [NFR-021]
next_owner: human
---

# NFR-021 issue-workflow usability

## Scope and requirement IDs

NFR-021 AC1–AC4 on the offline HTML issue UI. Branch
`cursor/nfr021-issue-workflow-usability-acec`. Draft PR
https://github.com/artofdream/logify/pull/11 vs `main`. Do not merge.

## Evidence consulted

- `docs/requirements/non-functional-requirements.md` NFR-021 (Partial; gap was
  no 10k probe / no a11y audit)
- `internal/report/page.html`, `page.js`, `page.css`, `followup.js`
- WI-20260909-fr021-follow-up-details (merged PR #4); no active NFR-021 dispatch
- Principles: honesty — AC4 not marked Implemented

## Changes and artifacts

- Focus restore after card rebuild; New tag label; Enter-to-tag; tab arrows
- Flagged / State: text badges; live status + Showing N of M
- `nfr021_a11y_test.js` source contracts; `nfr021_filter_probe.js` 10k filter
- ADR-0004, RES-20260909-nfr021-issue-filter-probe, Q-001, Q-002
- NFR-021 remains Partial

Commit: `14a0666` (plus any follow-up docs commit on this branch)

## Validation and observed results

- `gofmt -l cmd internal` — clean
- `go test ./... -count=1` — pass (`analyzer`, `report` 0.301s including
  NFR-021 Node harnesses)
- `go build -o logify.exe ./cmd/logify` — pass
- `go vet ./...` — pass
- `git diff --check` — pass
- `./logify.exe -output sample-report.html testdata/case` — 6 events, 3 files,
  0 warnings; report contains issue-workflow-hint, rememberFocus, no http(s)
- `node internal/report/nfr021_filter_probe.js` — 10,000 issues; worst median
  3.8 ms on 4× Xeon KVM / 15 GiB / Node v22.14.0; 100 ms target met for
  `store.filter`; 500 ms CI guard passed
- Browser `file:///workspace/sample-report.html` (computerUse agent): create
  issue; Flagged + State: Open/Investigating text; Enter adds `db`; status
  line; Showing N of M; Tab reaches flag/state/tag; Right Arrow switches tabs;
  375px readable, no horizontal scroll

## Assumptions and confidence

- High for AC1–AC3 implementation + source tests + one browser walk
- Medium for AC4: store.filter is fast here; DOM render of 10k cards is
  unmeasured; probe host is not published reference hardware

## Failures or conflicting evidence

None. NFR-021 stays Partial by design.

## Uncommitted or concurrent changes

Do not commit `logify.exe` or `sample-report.html`. Other open PRs were not
touched.

## Open questions and next action

Q-001 published reference hardware; Q-002 card windowing. Review draft PR #11;
do not merge; do not mark NFR-021 Implemented.
