---
id: ADR-0011-accessible-report-checks
type: decision
status: accepted
owner: cursor-agent
created: 2026-09-12
updated: 2026-09-12
requirements: [NFR-013, NFR-021, NFR-003]
supersedes: []
---

# ADR-0011: Stdlib accessibility checks and explicit report labels

Renumbered from a colliding ADR-0010 after NFR-016 (#17) landed
`ADR-0010-cli-compatibility.md` on `main`. Concurrent NFR-009 (#18) is
not on `main` and still uses a colliding ADR-0010 on that branch.

## Context and evidence

NFR-013 required keyboard-operable labeled controls, severity in text, WCAG
2.2 AA contrast, and automated checks on the generated report. The previous
gap was wrapping labels without `for`, placeholder-as-name risk, no CI a11y
probe, and `--line` on `--panel` at 1.43:1 (fails 3:1 non-text contrast).
axe-core or a CDN engine would add a Node dependency and a network asset,
conflicting with NFR-003.

NFR-021 AC4 still lacked a bound on painting 10,000 issue cards. Q-002 asked
whether to window, paginate, or virtualize.

## Decision

1. Use explicit `<label for>` (or `aria-label` for the hidden file input) on
   every static control. Create dynamic fields with `labeledField` so issue
   cards and the link picker are named without relying on placeholder text.
   Do not wrap the due-date clear button inside the date label.
2. Prefix timeline severity with `Severity:` and keep the color bar
   decorative.
3. Give interactive controls a dedicated `--control-border` token that meets
   3:1 against `--panel`. Decorative card edges may keep the darker `--line`.
4. Verify AC3/AC4 with stdlib Go contrast math plus a generated-HTML label
   scan (`internal/report/a11y.go`) and a Node source contract
   (`nfr013_a11y_test.js`). Do not add npm modules or axe-core.
5. Page the issue queue at `ISSUE_PAGE_SIZE = 25`. Keep NFR-021 **Partial**
   until a published reference workstation exists (Q-001). Mark NFR-013
   **Implemented** on the strength of the token-contrast and label probes,
   with the documented caveat that those probes are not a WCAG engine or AT
   run.

## Alternatives considered

- **axe-core in CI** was rejected: new npm dependency, heavier than the
  requirement, and still not an AT run. The task allowed keeping Partial if
  axe were required; it is not required for the written ACs.
- **Implicit wrapping labels only** was rejected: the documented gap called
  out placeholder/option text and asked for proper `for` bindings.
- **Virtualized scrolling** was rejected for the first window: pagination is
  keyboard-native, has a stable card height, and needs no scroll metrics.
- **Painting all 10k cards** remains unsupported as the default path.

## Consequences and risks

Operators get named controls and visible severity. CI fails if a used text
token or `--control-border` drops below AA. Large issue imports no longer
mount every card at once; paging must be used to reach later matches.
Placeholder and disabled-opacity painting still follow the browser. Status
must not treat a green contrast table as a full WCAG certification.

## Verification

`TestNFR013ReportCSSContrast`, `TestNFR013GeneratedReportA11y`,
`TestNFR013PageJSLabeledFields`, `TestNFR013AccessibleReportContracts`, and
`nfr013_a11y_test.js`. NFR-021 windowing is covered by `nfr013_a11y_test.js`,
`nfr021_filter_probe.js` (`issuePageSize: 25`), and generated-report chrome
assertions.
