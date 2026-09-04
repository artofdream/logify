---
id: WI-20260903-fr-017-issue-creation
type: work-item
status: review
owner: codex
created: 2026-09-03T22:14:52Z
updated: 2026-09-04T06:08:00Z
lease_expires: null
scope:
  - docs/collaboration/dispatches/DSP-20260904-FR017.md
  - docs/collaboration/work-items/WI-20260903-fr-017-issue-creation.md
  - docs/collaboration/handoffs/HO-20260904-0610-codex-fr-017.md
  - docs/knowledge/decisions/ADR-0001-follow-up-identities.md
  - docs/requirements/functional-requirements.md
  - docs/requirements/non-functional-requirements.md
  - README.md
  - internal/analyzer/model.go
  - internal/analyzer/analyzer.go
  - internal/analyzer/analyzer_test.go
  - internal/report/report.go
  - internal/report/report_test.go
requirements: [FR-017, NFR-001, NFR-003, NFR-004, NFR-005, NFR-010, NFR-012, NFR-013, NFR-014, NFR-015, NFR-017, NFR-019, NFR-020, NFR-021, NFR-022, NFR-024, NFR-027]
depends_on: []
supersedes: []
---

# Create issues from timeline evidence

## Goal

Implement FR-017 in the self-contained offline report with stable, versioned
issue and evidence identities; safe editable titles; complete evidence links;
and explicit transient-storage behavior.

## Non-goals

- Tags, flags, workflow state, notes, owners, and due dates (FR-018–FR-021).
- Export, import, or durable browser persistence (FR-022).
- Linking multiple evidence groups to one issue (FR-024).
- Modifying source log bundles or introducing runtime services/dependencies.

## Acceptance criteria

- One issue can be created from any displayed event or deduplicated group.
- Each issue has a stable, display-order-independent identifier and editable title.
- The issue view retains signature, instance, file, line, first/last seen, and count.
- Log and metadata text is rendered without executable markup injection.
- Issue controls are keyboard operable, labeled, responsive, and confirm actions.
- The report explains transient storage and warns before unsaved issues are lost.
- Source fixtures remain unchanged after analysis and report interaction support.

## Planned files and ownership

Codex owns only the paths declared in frontmatter in this isolated cloud checkout.
No active work items or uncommitted changes overlapped when this item was opened.

## Evidence and assumptions

- Started from clean `origin/main` at
  `df27c3884d11ecddc657473a6c46666d7edc93d5`, then refreshed before
  implementation to `319145cd545eb447c815350ff466aff0efd4201b` as required by
  the published cloud dispatch.
- Resolved cloud task ID:
  `6a99f0c2-5058-83ed-87d7-5f22d258f64f`.
- FR-017 does not require persistence across reloads; FR-022 defines that later
  capability. This work makes current in-memory storage explicit and guarded.
- An evidence identity excludes occurrence count and observation times so newly
  observed repeats do not change the identity of the originating evidence group.

## Validation

Interim observations; final required validation remains pending:

- `gofmt -w cmd internal && go test ./...` stopped with exit 127 because `gofmt`
  is not installed or on `PATH`; tests did not run in that chained command.
- Extracting the embedded JavaScript, substituting representative report JSON,
  and running the primary-runtime Node `--check -` passed with exit 0.
- A Playwright interaction probe could not launch: the installed package expected
  `/root/.cache/ms-playwright/chromium_headless_shell-1234/...`, but no browser
  executable exists. No browser interaction claim is based on that attempt.
- `git diff --check` passed after implementation edits.
- GitHub Actions CI run `33842702057` installed Go 1.22.12, then failed its
  formatting gate because `internal/analyzer/model.go` was not gofmt-formatted.
  Vet, build, tests, fixture execution, and PR diff checks were skipped. The
  formatter-only discrepancy was corrected for the next run.

The repository CI installs Go 1.22 and runs formatting, tests, vet, build, fixture
execution, and PR diff checks. Its observed results will be recorded before the
requirement or work item is marked complete.

Final observed evidence:

- GitHub Actions run `33842921381`, commit
  `c4fdcd1fa6f90a182edebf2a6a4b7ae538b66ccc`: `validate` job passed.
  - gofmt reported no unformatted files.
  - `go vet ./...` passed.
  - `go build -o /tmp/logify ./cmd/logify` passed.
  - `go test ./... -v` passed: analyzer tests `ok` in 0.004s and report tests
    `ok` in 0.003s, including the new FR-017/NFR-017/NFR-019 cases.
  - fixture execution wrote `/tmp/sample-report.html` with 6 events from 3 files
    and 0 warnings; the report existence/doctype probes passed.
  - PR diff check passed.
- The same workflow's `enable auto-merge` job failed with `GraphQL: Pull Request
  is still a draft (mergePullRequest)`. This is expected evidence that draft PR
  #1 remained unmerged; it is not a validation-job failure.
- A Node DOM execution probe passed: one stable issue was created, complete
  evidence remained available, an unsafe title was edited as text, duplicate
  creation reused the identity, focus/status and unload guards fired, and no
  network call occurred.
- Remote commit `61f62dd...` and local committed tree both resolved to
  `4c1f20d2314ffbcf47d57b1018834c27c43f4390`; the subsequent formatter/evidence
  tree also matched locally and remotely at
  `d7986fb2be3fc206b58a7a8c68fad221bb062923`.
- `git diff --check` and staged diff checks passed throughout.

## Activity log

- `2026-09-03T22:14:52Z` — codex — read mandated project guidance, requirements,
  collaboration protocol, knowledge index, relevant architecture/open questions,
  prior work items, and incident; confirmed clean Git state and no active overlap;
  created branch `codex/FR-017-issue-creation` and claimed bounded cloud scope.
- `2026-09-04T05:46:46Z` — codex — preserved the documentation-first edits,
  fetched and rebased onto latest `origin/main` at `319145c`, read the FR-017
  dispatch and updated collaboration protocol, confirmed the assignment matches
  this branch, recorded resolved cloud task ID
  `6a99f0c2-5058-83ed-87d7-5f22d258f64f`, and renewed the advisory lease.
- `2026-09-04T05:57:00Z` — codex — implemented versioned evidence identities,
  first/last observation tracking, safe offline issue creation and editing,
  responsive/keyboard-native controls, transient-storage disclosure, unload
  warning, focused Go tests, and user documentation. JavaScript syntax and diff
  checks passed; Go and browser runtime gaps are recorded under Validation.
- `2026-09-04T06:04:00Z` — codex — opened draft PR #1 to prevent the repository's
  ready-PR auto-merge job from violating the unmerged-PR constraint. CI run
  `33842702057` identified only `internal/analyzer/model.go` as unformatted; all
  later validation steps were skipped. Corrected the struct-column formatting.
- `2026-09-04T06:08:00Z` — codex — CI run `33842921381` passed every required
  validation step in its `validate` job. The expected draft-only auto-merge job
  failure preserved the unmerged constraint. Node DOM behavior probe also passed;
  requirement statuses were updated from observed evidence, the lease was
  cleared, and the work item entered review.

## Handoff or completion

Review. See
[`HO-20260904-0610-codex-fr-017`](../handoffs/HO-20260904-0610-codex-fr-017.md)
and draft PR [#1](https://github.com/artofdream/logify/pull/1). Do not merge as
part of this handoff.
