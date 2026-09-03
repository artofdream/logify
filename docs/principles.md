# Core principles

These principles govern product behavior, engineering decisions, and all
human/agent collaboration in Logify. They are ordered for reference, not rank;
all three must hold at the same time.

They adopt the core principles documented by the
[Adaptive Experience Architecture](https://architecture.artof.link/schema.html),
with the Logify-specific mapping maintained in
[`framework-adoption.md`](framework-adoption.md).

## 1. Honesty

Make the state of knowledge and work inspectable. Distinguish verified facts,
evidence-based inferences, assumptions, proposals, and unknowns. Surface failures
and conflicting evidence. Never claim completion without proof against the
documented acceptance criteria.

In practice:

- Cite the source artifact, command, test, or observation behind material claims.
- Label assumptions, uncertainty, missing evidence, and unverified behavior.
- Report attempted work, actual changes, validation, failures, and remaining gaps.
- Preserve conflicting evidence until it is resolved; do not manufacture
  consensus.
- Correct earlier conclusions explicitly and retain traceability to the reason.
- Treat status words such as `verified`, `implemented`, `done`, and `shipped` as
  claims. Probe them or use `Unknown`, `Partial`, or `Proposed`.

## 2. Antifragility

Use failures, surprises, disagreement, and changing requirements to make the
system and workflow more reliable. A significant failure should leave durable
learning, not merely a one-off fix.

In practice:

- Record the trigger, impact, diagnosis, and recovery for significant failures.
- Add a regression test, validation rule, runbook improvement, or equivalent
  safeguard for confirmed defects when practical.
- Keep changes focused, reversible, versioned, and recoverable.
- Isolate work so a failed task does not invalidate unrelated results.
- Compare evidence or request independent verification when agents disagree.
- Link recurring incidents to prior occurrences and measure whether safeguards
  reduce recurrence, detection time, or recovery time.
- Treat the same miss twice as evidence of a missing sensor or gate, not a need
  for another reminder.

## 3. Knowledge first

Establish and document shared factual and decision context before implementation.
Consult requirements, architecture, constraints, and prior decisions first. Turn
new knowledge into a durable project asset instead of leaving it in transient
agent context.

In practice:

- Begin with a documented problem, scope, constraints, and acceptance criteria.
- Read authoritative project documents and relevant prior decisions before edits.
- Identify source authority and freshness; record unresolved questions rather than
  inventing facts.
- Record consequential decisions with rationale, alternatives, and consequences.
- Update affected documentation in the same change as implementation.
- Reference canonical artifacts in prompts and handoffs instead of copying
  divergent summaries.
- Pause implementation or create a research task when essential knowledge is
  missing.
- Treat committed repository history as shared memory. Chat, an uncommitted note,
  and agent consensus are useful context but not durable shared knowledge.

## Decision gate

Agreement between agents is not sufficient evidence. Accept a result only when it
is traceable and candid (**honesty**), captures learning and improves resilience
(**antifragility**), and rests on durable shared evidence (**knowledge first**).

Operational work follows **Interpret → Act → Verify → Remember**. Agents and
parsers may interpret; authoritative evidence, explicit operator decisions, and
acceptance probes decide.
