# Logify requirements

This directory is the product source of truth for Logify. Requirements are
defined before implementation and remain independent of package structure.

All requirements and delivery work are governed by the project
[core principles](../principles.md): honesty, antifragility, and knowledge first.
When work is delegated, follow the [multi-agent workflow](../multi-agent-workflow.md).
Concurrent Codex–Claude work uses the document-based
[collaboration protocol](../collaboration/README.md) and durable
[knowledge vault](../knowledge/index.md).

- [Functional requirements](functional-requirements.md) describe observable
  behavior and user capabilities.
- [Non-functional requirements](non-functional-requirements.md) describe quality,
  security, portability, performance, and maintainability constraints.

## Requirement format

Every requirement has a stable identifier, priority, status, rationale, and
verifiable acceptance criteria.

Priorities follow MoSCoW:

- **Must**: required for the current supported product baseline.
- **Should**: important, but the baseline remains useful without it.
- **Could**: desirable when capacity permits.
- **Won't (now)**: explicitly outside the current release scope.

Statuses are:

- **Proposed**: agreed direction, not yet implemented.
- **Partial**: some acceptance criteria are met.
- **Implemented**: all listed acceptance criteria are met and verified.
- **Deferred**: intentionally postponed.

## Documentation-first workflow

1. Review authoritative requirements, evidence, prior decisions, and known gaps;
   record material assumptions and unknowns.
2. Add or amend an FR/NFR before changing externally observable behavior or a
   quality constraint.
3. Review the requirement for ambiguity, scope, priority, and testability.
4. Add tests that reference the applicable requirement IDs in test names or
   comments when the relationship is not obvious.
5. Implement the smallest change that satisfies the acceptance criteria.
6. Run the validation documented in `AGENTS.md` and preserve observed results.
7. Update requirement status only when every acceptance criterion is verified;
   report anything not run, failed, or still uncertain.

IDs are permanent. Do not renumber or reuse removed IDs; mark obsolete
requirements as deferred and explain the replacement.

## Knowledge maintenance

A discovered failure or format ambiguity should produce durable knowledge: a
sanitized fixture, regression test, requirement clarification, decision record,
or agent-guide improvement. Never advance status speculatively. If evidence is
missing or contradictory, preserve it as an open question until resolved.

## Release baseline

The current v1 baseline consists of all **Must** requirements marked
**Implemented**. Requirements marked **Partial** or **Proposed** identify the
next documented work rather than an implicit promise of delivery.

The intended product is both an incident timeline and a lightweight follow-up
workspace. Operators must be able to turn findings into tracked issues, classify
them with tags and flags, and carry that work forward without modifying the
original log bundle.
