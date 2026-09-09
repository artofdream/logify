---
id: WI-20260909-gitignore-readme-hygiene
type: work-item
status: review
owner: cursor-agent
created: 2026-09-09T20:26:00Z
updated: 2026-09-09T20:30:00Z
lease_expires: 2026-09-10T02:00:00Z
scope:
  - .gitignore
  - README.md
  - docs/collaboration/work-items/WI-20260909-gitignore-readme-hygiene.md
requirements: []
depends_on: []
supersedes: []
---

# Gitignore and README hygiene

## Goal

Confirm generated binaries and HTML reports cannot be committed, document that
Go 1.22+ must be on PATH, and add a one-line release checklist stub. Open a
draft PR against `main`. Do not merge, tag, or cut a GitHub Release.

## Non-goals

- Product features or code behavior changes (except a missing gitignore line).
- Installing Go on anyone's machine.
- Cutting `v0.1.0` or any tag/Release.
- FR/NFR debt, FR-024 (PR #5), or stale WI close-out (PR #6).

## Acceptance criteria

1. `.gitignore` keeps `/logify`, `/logify.exe`, `/*-report.html`, and
   `.obsidian/`, and names common report files explicitly (`sample-report.html`,
   `report.html`).
2. README Build notes Go 1.22+ on PATH (Windows tip) and that generated
   `logify.exe` / `*-report.html` are gitignored.
3. README or docs has a one-line release stub (`v*` tag → existing workflow).
4. Draft PR is open; this task does not merge it.

## Planned files and ownership

- `.gitignore` — explicit report names; keep existing patterns.
- `README.md` — Build PATH/ignore note and release stub.

## Evidence and assumptions

- `git check-ignore` on `origin/main` (`01ae8bf`): `sample-report.html`,
  `logify.exe`, `logify`, and `incident-report.html` match; `report.html`
  (README example) does not match `/*-report.html`.
- Open PRs #5 and #6 do not own `.gitignore` or the README Build section.
- Repo CI enables auto-merge on non-draft PRs; this PR stays draft.

## Validation

Observed on this branch:

```text
git check-ignore -v sample-report.html logify.exe logify report.html incident-report.html
# .gitignore:9:/sample-report.html	sample-report.html
# .gitignore:3:/logify.exe	logify.exe
# .gitignore:2:/logify	logify
# .gitignore:10:/report.html	report.html
# .gitignore:8:/*-report.html	incident-report.html
git diff --check   # pass (no output)
```

`testdata/case/foo-report.html` is not ignored (root-anchored patterns only).
Go compiler checks were not run: no `cmd/` or `internal/` changes.

## Activity log

- `2026-09-09T20:26:00Z` — cursor-agent — claimed disjoint docs/ignore scope
  after fetch of `origin/main` (`01ae8bf`).
- `2026-09-09T20:30:00Z` — cursor-agent — added explicit report names, README
  PATH/ignore note, and release stub. `git check-ignore` and `git diff --check`
  observed as above.

## Handoff or completion

Draft PR against `main`. Do not merge (CI auto-merges ready PRs).
