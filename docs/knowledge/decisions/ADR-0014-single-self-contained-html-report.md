---
id: ADR-0014-single-self-contained-html-report
type: decision
status: accepted
owner: grok-bot
created: 2026-10-07
updated: 2026-10-07
requirements: [FR-013, FR-014, NFR-003, NFR-004, NFR-020]
supersedes: []
---

# ADR-0014: Single self-contained, offline HTML report as the only output

**Status:** Accepted (retroactive, recorded 2026-10-07)
**Introduced in:** `cbc0c24` feat: establish log analysis foundation
**Refined in:** `64bbc1c` feat: Must issue-tracking UI (builds on #2) (#3)
**Related requirements:** FR-013, FR-014, NFR-003, NFR-004, NFR-020

This is a **Documented** record reconstructed after the fact from repository
history. It records a decision that was already in force; it does not change
behavior. Where no written rationale exists, the gap is marked `OPEN` rather
than filled in.

## Context and evidence

- `cbc0c24` added FR-013 ("Generate a self-contained HTML report"): `-output`
  defaults to `logify-report.html`. HTML, CSS, JavaScript, and event data live
  in one file. The report needs no network access or external assets, and log
  content cannot inject executable markup. The same commit added NFR-003
  (offline operation, no remote scripts, styles, fonts, or images) and NFR-004
  (untrusted input, and embedded JSON cannot end the script context).
- In `cbc0c24`, `internal/report/report.go` was one `html/template` string
  with inline `<style>` and `<script>`. Analyzer output went in as
  `template.JS(json.Marshal(result))`.
- `64bbc1c` (#3) split the page source into `page.html`, `page.css`,
  `page.js`, and `followup.js`. Each file is compiled in with `//go:embed` and
  inlined into the one output file at write time. The source layout changed;
  the single-file output did not.
- The same PR added the issue queue, with follow-up state kept in browser
  `localStorage` (`page.js`) and a portable `logify-follow-up.json`
  export/import. ADR-0001 and ADR-0002 cover that state; NFR-020 requires
  that no follow-up data is sent over a network.
- There is no server component anywhere in the repository. The DSO SSDD draft
  (`artofdream/dso`, `03-architecture/ssdd/logify.md`, Draft v0.1) also
  places "any server/network component" out of scope. That draft is context,
  not authority.

## Decision

1. Each run writes exactly one HTML file (`-output`, default
   `logify-report.html`). That file holds all markup, styles, scripts, and the
   JSON analysis payload.
2. The report makes no network requests and references no remote assets.
   Logify ships no server, daemon, or hosted viewer. Opening the file in a
   browser is the whole runtime.
3. Report sources live as separate files under `internal/report/` and are
   compiled in with `go:embed`. At runtime the binary needs no template or
   asset files next to it.
4. The payload goes in through `html/template` as `template.JS`. The page
   script renders log-derived text with escaping and DOM text APIs (NFR-004).
5. Interactive follow-up state stays on the operator's machine: browser
   storage as a cache, plus explicit JSON export and import (ADR-0002).

## Alternatives considered

- **Local HTTP server or hosted viewer:** not present in the code and excluded
  by NFR-003 and FR-013 AC2–AC3. No record of it being evaluated was found.
- **Multi-file report (HTML plus asset directory) or a non-HTML format
  (JSON/CSV only):** no record of evaluation found.
- **Why a single offline HTML file was chosen in the first place:** `OPEN`.
  FR-013 has no Rationale field. Likely motives include easy hand-off with a
  support bundle, air-gapped use, and no install for the viewer. These are
  plausible but not written down, so this record does not claim them.

## Consequences and risks

- The report is a **sensitive copy** of parsed log text, and anyone who gets
  the file can read it. ADR-0005 (optional redaction) and the NFR-006
  warnings follow from this.
- Report size and browser memory grow with the number of unique timeline rows.
  ADR-0012 bounds analyzer memory, not report size. A report-size or
  browser-render budget is `OPEN`.
- Follow-up data in `localStorage` is tied to one browser and one report.
  Portability depends on explicit export (ADR-0002, NFR-020).
- Checks that would normally use a browser engine or CDN tool are done with
  stdlib Go and Node scripts instead (ADR-0011).

## Verification

- `TestWriteSelfContained` (`internal/report/report_test.go`) checks for
  `<!doctype html>`, an escaped `</script>` payload, and no `http://` or
  `https://` references.
- `TestWriteEscapesOperatorAndLogText` covers escaping of operator and log
  text.
- The CI fixture smoke step runs `logify -output sample-report.html
  testdata/case` and checks that the file is non-empty and contains
  `<!doctype html>`.
