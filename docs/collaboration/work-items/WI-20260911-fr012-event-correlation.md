---
id: WI-20260911-fr012-event-correlation
type: work-item
status: review
owner: cursor-agent
created: 2026-09-11T19:25:43Z
updated: 2026-09-11T20:27:00Z
lease_expires: 2026-09-12T04:27:00Z
scope:
  - internal/analyzer/model.go
  - internal/analyzer/analyzer.go
  - internal/analyzer/correlate.go
  - internal/analyzer/correlate_test.go
  - internal/analyzer/analyzer_test.go
  - testdata/correlate/
  - internal/redact/redact.go
  - internal/redact/redact_test.go
  - internal/report/report.go
  - internal/report/report_test.go
  - internal/report/page.html
  - internal/report/page.js
  - internal/report/page.css
  - README.md
  - docs/requirements/functional-requirements.md
  - docs/requirements/non-functional-requirements.md
  - docs/knowledge/architecture.md
  - docs/knowledge/glossary.md
  - docs/knowledge/decisions/ADR-0005-optional-report-redaction.md
  - docs/knowledge/decisions/ADR-0008-event-correlation.md
  - docs/collaboration/work-items/WI-20260911-fr012-event-correlation.md
  - .github/workflows/ci.yml
  - docs/knowledge/incidents/INC-20260911-macos-go122-lc-uuid.md
requirements: [FR-012, FR-023, NFR-002]
depends_on: []
supersedes: []
---

# FR-012 correlate related events

## Goal

Add deterministic, documented correlation groups so operators can see related
HTTPD and Tomcat events that share exact identifiers, plus an optional labeled
heuristic, without inventing unsupported links.

## Non-goals

FR-016 compression/rotation, changing follow-up export schema, auto-creating
issues from correlations, unlabeled number/UUID matching, IPv6 client-window
extraction from free text, merging PRs.

## Acceptance criteria

FR-012 AC1–AC5. Events stay individually inspectable. Heuristics are labeled
and lower confidence. Prefer no group over a weak unsupported relationship.

## Planned files and ownership

cursor-agent owns the paths in frontmatter. Follow-up store files are
read-only.

## Evidence and assumptions

- No active/queued/review dispatch owns FR-012. `main` at `4c8da68` is clean.
- Exception `Caused by:` chains are already one event; they are not a
  cross-event correlation rule.
- Access-log client addresses are not in the normalized message, so the
  analyzer must retain `ClientAddr` for the heuristic.

## Validation

Observed on `02bd242` (Linux):

- `gofmt -l cmd internal` — clean
- `go test ./...` — pass (`cmd/logify`, `analyzer`, `redact`, `report`)
- `go vet ./...` — pass
- `go build -o logify.exe ./cmd/logify` — pass
- `git diff --check` — pass
- `./logify.exe -output sample-report.html testdata/case` — 6 events, 3 files, 0 warnings
- `./logify.exe -output correlate-report.html testdata/correlate` — 14 events, 3 files, 0 warnings, 2 groups (exact ×4, heuristic ×2)

Prior GitHub CI on the merge commit (`940f6f1` / squash `76cae98`), run
[34640118520](https://github.com/artofdream/logify/actions/runs/34640118520):
ubuntu, windows, macos, and validate succeeded. PR #13 is already merged;
this follow-up commit is on the same branch and was not re-run by CI
(workflow is `pull_request` + `push` to `main` only). Do not merge from
this task. Do not open a duplicate PR.

## Activity log

- `2026-09-11T19:25:43Z` — cursor-agent — claimed FR-012 on branch
  `cursor/fr012-event-correlation-1995` from `main` (`4c8da68`).
- `2026-09-11T19:40:00Z` — cursor-agent — validation passed. FR-012 marked
  Implemented with documented limits. Draft PR #13.
- `2026-09-11T20:20:00Z` — cursor-agent — sponsor follow-up on the same
  branch: keep collapsed-occurrence identifiers for correlation (FR-010
  still one row) and emit pairwise `client-ip-window` groups only.
  macOS LC_UUID CI fix already on this branch (`940f6f1`). Do not merge.
- `2026-09-11T20:27:00Z` — cursor-agent — local validation passed on
  `02bd242`. Pushed to `cursor/fr012-event-correlation-1995`. PR #13
  description updated. No duplicate PR. Not merged.

## Handoff or completion

Bugbot follow-up is on PR #13's branch at `02bd242`. Local tests passed.
#13 is already merged to `main` (`76cae98`); these analyzer fixes are
not on `main` until a later PR. Do not open a duplicate from this task.
Do not merge.
