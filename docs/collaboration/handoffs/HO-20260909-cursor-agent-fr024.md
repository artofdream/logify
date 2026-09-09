---
id: HO-20260909-cursor-agent-fr024
type: handoff
status: proposed
owner: cursor-agent
created: 2026-09-09T20:17:00Z
work_item: WI-20260909-fr024-merge-recurring-evidence
requirements: [FR-024]
next_owner: human
---

# FR-024 recurring evidence merge ready for review

## Scope and requirement IDs

FR-024 AC1–AC4. Branch `cursor/fr024-merge-recurring-evidence-c99d`. PR
https://github.com/artofdream/logify/pull/5 against `main` (FR-021 already on
`main` via PR #4 / `01ae8bf`).

## Evidence consulted

ADR-0001, ADR-0002, ADR-0003, follow-up store/page scripts, FR-024 ACs.
DSP-20260904-FR023-FR024 was cancelled; this work takes FR-024 only.

## Changes and artifacts

Store: `linkEvidence`, `unlinkEvidence`, `listReviews`,
`acknowledgeOccurrences`, `dismissCandidate`. Schema stays
`logify-follow-up-v1` with optional `linkedEvidence` / `ignoredEvidence`
(ADR-0004). UI: timeline link picker, multi-evidence cards, review panel.
Import/link/dismiss/acknowledge do not change workflow state.

## Validation and observed results

- `gofmt -w cmd internal` — alignment-only change in `followup.go`
- `go test ./... -count=1` — pass (analyzer + report, including FR-024 Node tests)
- `go vet ./...` — pass
- `go build -o logify.exe ./cmd/logify` — pass
- `./logify.exe -output sample-report.html testdata/case` — 6 events, 3 files, 0 warnings; report contains Link to existing issue, Recurring evidence review, `logify-follow-up-v1`; no `http://` / `https://`; no `innerHTML`
- `git diff --check` — pass
- Browser `file:///workspace/sample-report.html`: create issue; link POST /api into it (two evidence refs); unlink additional (originating remains, state open); clear; import review JSON; signature match listed; imported state resolved; notes `<script>alert(1)</script>` as text; Link to issue keeps resolved

## Assumptions and confidence

PR #4 was already squash-merged to `main` when this task started; the new PR
stacks on that merge commit rather than replaying FR-021 commits. Occurrence
deltas require the same evidence id with a higher stored-vs-live count; the
browser probe exercised signature-match review (AC3/AC4) plus Node tests for
occurrence updates.

## Failures or conflicting evidence

`RecordScreen` SAVE_RECORDING hit an EIO copy; the polished `recording_full.mp4`
was copied to artifacts manually. No functional failure.

## Uncommitted or concurrent changes

None intended at handoff time beyond the follow-up commit that records this
validation.

## Open questions and next action

Review PR #5. Do not merge from this task.
