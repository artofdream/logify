---
id: aea-framework-adoption
type: framework-mapping
status: active
owner: human
updated: 2026-09-12
nfr028_status: Partial
sources:
  - https://architecture.artof.link/
  - https://architecture.artof.link/schema.html
  - https://architecture.artof.link/comparison.html
  - https://architecture.artof.link/glossary.html
tags: [architecture, principles, outer-harness, second-brain]
---

# Adaptive Experience Architecture adoption

Logify adopts the transferable principles of the
[Adaptive Experience Architecture](https://architecture.artof.link/) (AEA),
adapted to an offline log-analysis and issue-follow-up product. This is a mapping,
not a claim that Logify implements the full commerce-oriented framework.

## Formula

AEA defines:

> Adaptive Experience = Shared Understanding + Domain Services + Outer Harness

For Logify, this becomes:

> Trustworthy Diagnosis = Reviewable Incident Understanding + Authoritative
> Evidence + Outer Harness

- **Reviewable Incident Understanding:** normalized timeline, deduplicated event
  groups, correlations, tracked issues, and explicit unknowns.
- **Authoritative Evidence:** immutable source logs plus user-authored follow-up
  facts. Parsers interpret evidence; they do not replace it.
- **Outer Harness:** guides, sensors, loop, memory, permissions, and observability
  that make analysis and agent delivery reviewable.

The governing translation is: **agents and parsers may interpret; source evidence,
explicit operator decisions, and acceptance probes decide.**

## Core principles

| AEA principle | Logify adoption | Enforced through |
|---|---|---|
| Honesty | `verified`, `implemented`, `done`, and parser/correlation conclusions are claims; without a probe they remain `Unknown`, `Partial`, or `Proposed`. | Requirements status, provenance, warnings, validation output, handoffs |
| Knowledge First | Durable repository knowledge is read before work; ephemeral chat and agent consensus are not canonical memory. | Requirements, ADRs, knowledge vault, work items, session promotion |
| Antifragility | The same miss twice indicates a missing sensor or gate. A significant failure should create a durable safeguard rather than a prompt-only reminder. | Fixtures, regression tests, incidents, guides, automated checks |

The detailed behavioral contract remains in [core principles](principles.md).

## Outer harness adoption ledger

Status here is evidence-based and intentionally conservative. `nfr028_status`
in the frontmatter must match [NFR-028](requirements/non-functional-requirements.md).
The [status mirror](#requirement-status-mirror) must list every Partial,
Proposed, or Deferred requirement.

| Layer | Logify implementation | Status | Current probe/evidence | Gap |
|---|---|---|---|---|
| Guides | `AGENTS.md`, `CLAUDE.md`, requirements, collaboration protocol, [permissions](collaboration/permissions.md) | Adopted on `main` | Files exist, cross-link, and are committed; `TestNFR028DocsLinksAndLedger` fails on broken relative links | Running agents must reload guidance |
| Sensors | Go tests, fixtures, `go vet`, build, `git diff --check`, multi-OS CI, Node `--check`, follow-up store tests, NFR-021 source-contract and 10k `store.filter` probe, `internal/harness` docs/link + ledger + work-item probes, NFR-009 generated-fixture smoke + `BenchmarkNFR009ScaleSmoke` | Partial | `test` matrix plus `validate` on `main` push/PR; `go test ./internal/harness` on every OS; `TestNFR028DocsLinksAndLedger`, `TestNFR028ImplementedClaimsHaveNamedProbes`, `TestNFR028OpenWorkItemsHaveOwnershipFields`; Linux CI `BenchmarkNFR009ScaleSmoke`; `TestNFR009ScaleSmoke` on every OS | Full 1 GiB NFR-009 bench is manual/nightly; no WCAG engine / AT run or full in-browser DOM/page execution |
| Loop | Interpret → Act → Verify → Remember | Adopted | Collaboration steps and handoff gates are documented | No automation enforces every transition |
| Memory | Git-reviewable Obsidian-compatible Markdown vault | Adopted on `main` | Vault structure, templates, and relative links validate via the docs/link probe | No dedicated index-freshness job beyond link resolution; uncommitted notes are not shared history |
| Permissions | Read-only source handling, scoped work items, advisory leases, `.github/CODEOWNERS`, CI `validate` checklist | Partial | [permissions.md](collaboration/permissions.md); CODEOWNERS names `@artofdream`; open-WI field probe; `validate` aggregator is the documented merge check | Leases are advisory; CODEOWNERS is review routing, not a merge lock, unless an operator enables required code-owner reviews (that GitHub setting is Unknown from this tree) |
| Observability | File/line provenance, warnings, parse confidence, unparsed-record and correlation-confidence counts, requirement status ledger | Adopted | Analyzer `Observability()` / report payload `unparsedRecords`, `highConfidenceCorrelations`, `lowConfidenceCorrelations`; timeline marks low parse confidence; `TestNFR028ImplementedClaimsHaveNamedProbes` fails Implemented IDs with no named test/CI/JS probe | Binary high/low only; no live runtime telemetry |

`Adopted locally` would mean present only in a working tree. A commit, CI
result, release, and live behavior are different claims and require different
probes.

## Requirement status mirror

Consumed by `TestNFR028DocsLinksAndLedger`. Every Partial, Proposed, or
Deferred requirement must appear. Values must match `docs/requirements/`.

| ID | Status |
|---|---|
| NFR-013 | Partial |
| NFR-021 | Partial |
| NFR-028 | Partial |

## Boundaries

- Logify does not inherit AEA's commerce domain services or stakeholder team.
- Agent roles are responsibilities, not headcount.
- No agent may approve its own irreversible action solely because it implemented
  the change.
- A closed work item does not prove a feature works; acceptance probes do.
- Where evidence is missing, record `Unknown` instead of extrapolating.

## Adoption evolution

Update this ledger when a layer gains or loses a probe. Significant architectural
changes require an ADR. The ledger must link to observed evidence and must not be
advanced because documentation or code merely exists.

NFR-028 remains **Partial**: Permissions still cannot claim an enforced
path/merge lock from this repository, and Sensors still lack WCAG and
in-browser DOM. NFR-009 has a generated-fixture smoke bench in `go test` /
Linux CI; the 1 GiB child-process run stays manual/nightly.
