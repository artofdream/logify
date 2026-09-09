---
id: DSP-20260904-MUST-ISSUE-UI
type: cloud-dispatch
status: done
owner: cursor-agent
created: 2026-09-04
updated: 2026-09-09
requirements: [FR-017, FR-018, FR-019, FR-020, FR-022, FR-023]
branch: cursor/must-issue-ui-4f51
task_ref: cursor-cloud-must-issue-ui
depends_on: [PR-2-branch cursor/report-crash-quick-wins-decc]
pr: https://github.com/artofdream/logify/pull/3
supersedes: [DSP-20260904-FR017, DSP-20260904-FR018-FR019, DSP-20260904-FR020-FR021, DSP-20260904-FR022, DSP-20260904-FR023-FR024]
---

# Combined Must issue UI

Human-requested takeover of the sequential FR-017…FR-024 cloud reservations.
This dispatch implemented the Must follow-up UI in one PR on top of the PR #2
crash-fix branch. FR-021 (editable notes/owner/due) and FR-024 (merge recurring
evidence) were out of scope for this PR.

`2026-09-09` — PR #3 squash-merged to `main` as `64bbc1c` on
2026-09-04T20:28:16Z. FR-021 later merged as PR #4 (`01ae8bf`,
2026-09-09T19:52:40Z; DSP-20260909-FR021). FR-024 is draft PR #5 and is
not part of this dispatch.
