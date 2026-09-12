# Core principles

These principles govern product behavior, engineering decisions, and all
human/agent collaboration in Logify. They are ordered for reference, not rank;
all four must hold at the same time.

They adopt the core principles documented by the
[Adaptive Experience Architecture](https://architecture.artof.link/schema.html)
and [AEA #434 Keep Learning and Apply](https://gitlab.com/artof-group/adaptive-experience-architecture/-/work_items/434),
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

## 4. Keep learning and apply

Turn repeated friction, misses, and successful patterns into durable assets,
then use those assets on the next similar task. Learning that stays only in
chat is lost. This is Logify's adoption of
[AEA #434 Keep Learning and Apply](https://gitlab.com/artof-group/adaptive-experience-architecture/-/work_items/434)
(GitLab `artof-group/adaptive-experience-architecture` work item
`docs(framework): core principle Keep Learning and Apply`; related
[skill-matrix #433](https://gitlab.com/artof-group/adaptive-experience-architecture/-/work_items/433)).
The public [glossary](https://architecture.artof.link/glossary.html) and
[comparison](https://architecture.artof.link/comparison.html) still list
Honesty, Knowledge first, and Antifragility; AEA marks this fourth principle
Documented/Planned until a Pages probe. Do not claim it is Live on
[architecture.artof.link](https://architecture.artof.link/).

In practice:

- File an issue, agree the change, then write a skill, sensor, fixture, ADR,
  runbook, or doc. Do not leave the lesson in a transcript.
- Label those assets honestly (`Documented`, `Simulated`, or `Live`). Do not
  mark a skill Live without a live probe.
- Apply the existing skill or doc on the next similar run. A second miss of
  the same class is evidence the asset is missing, unused, or too weak.
- Skill proposals from real stacks—[#22](https://github.com/artofdream/logify/issues/22)
  (PR train / parallel-merge) and [#23](https://github.com/artofdream/logify/issues/23)
  (tagged release cut)—are examples of this loop. Both are Documented/Simulated,
  not Live. Docs adoption is tracked in
  [#24](https://github.com/artofdream/logify/issues/24).

## Decision gate

Agreement between agents is not sufficient evidence. Accept a result only when it
is traceable and candid (**honesty**), captures learning and improves resilience
(**antifragility**), rests on durable shared evidence (**knowledge first**), and
turns that learning into assets the next similar task actually uses (**keep
learning and apply**).

Operational work follows **Interpret → Act → Verify → Remember**. Agents and
parsers may interpret; authoritative evidence, explicit operator decisions, and
acceptance probes decide.
