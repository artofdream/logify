---
id: WI-20260903-fr-017-issue-creation
type: work-item
status: active
owner: codex
created: 2026-09-03T22:14:52Z
updated: 2026-09-04T05:46:46Z
lease_expires: 2026-09-04T07:46:46Z
scope:
  - docs/collaboration/dispatches/DSP-20260904-FR017.md
  - docs/collaboration/work-items/WI-20260903-fr-017-issue-creation.md
  - docs/collaboration/handoffs/HO-20260903-fr-017-issue-creation.md
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

The repository CI installs Go 1.22 and runs formatting, tests, vet, build, fixture
execution, and PR diff checks. Its observed results will be recorded before the
requirement or work item is marked complete.

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

## Handoff or completion

Active; no handoff yet.
