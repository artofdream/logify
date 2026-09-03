---
id: aea-framework-adoption
type: framework-mapping
status: active
owner: human
updated: 2026-09-03
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
| Antifragility | The same miss twice indicates a missing sensor or gate. Significant failures create durable safeguards rather than prompt-only reminders. | Fixtures, regression tests, incidents, guides, automated checks |

The detailed behavioral contract remains in [core principles](principles.md).

## Outer harness adoption ledger

Status here is evidence-based and intentionally conservative.

| Layer | Logify implementation | Status | Current probe/evidence | Gap |
|---|---|---|---|---|
| Guides | `AGENTS.md`, `CLAUDE.md`, requirements, collaboration protocol | Adopted locally | Files exist and cross-link | Repository is not yet committed; running agents must reload guidance |
| Sensors | Go tests, fixtures, `go vet`, build, `git diff --check` | Partial | Commands pass locally | No CI, cross-platform matrix, docs/link checker, benchmark, or accessibility sensor |
| Loop | Interpret → Act → Verify → Remember | Adopted as protocol | Collaboration steps and handoff gates are documented | No automation enforces every transition |
| Memory | Git-reviewable Obsidian-compatible Markdown vault | Adopted locally | Vault structure and templates validate | No freshness/index automation; uncommitted memory is not shared history |
| Permissions | Read-only source handling, scoped agent work items, advisory leases | Partial | Protocol and work-item schema exist | Leases are advisory; no merge coordinator or technical path lock exists |
| Observability | File/line provenance, warnings, test output, requirement status, work logs | Partial | Current analyzer/report expose core provenance and counts | Parser confidence, unparsed-line accounting, correlation evidence, and automated status ledger remain incomplete |

`Adopted locally` means present in the working tree, not shipped or shared through
repository history. A commit, CI result, release, and live behavior are different
claims and require different probes.

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
