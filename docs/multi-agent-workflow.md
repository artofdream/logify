# Multi-agent workflow

Use multiple agents only when work can be divided into bounded, independently
verifiable tasks. A coordinating agent owns scope, integration, validation, and
the final claim of completion.

When Codex, Claude, Cursor Agent, Antigravity, Grok, or another autonomous tool
may work in the same checkout, `docs/collaboration/README.md` is the canonical
shared-worktree protocol. Register a work item and advisory path lease there
before editing. This scales to any number of concurrent agents, not just two.

## Roles

- **Coordinator:** identifies affected requirements, partitions work, prevents
  overlapping edits, integrates results, and runs final validation.
- **Researcher/reviewer:** gathers evidence, challenges assumptions, or reviews a
  proposed design without editing the same files as an implementer.
- **Implementer:** makes a bounded change tied to named FR/NFR IDs and supplies
  tests and evidence.
- **Verifier:** independently checks acceptance criteria, failure paths, and the
  final diff when risk or disagreement warrants it.

One agent may fill several roles for small changes, but it must not describe its
own unverified output as independent verification.

## Protocol

1. **Ground:** read `docs/principles.md`, affected requirements, relevant decision
   records, code, and tests.
2. **Frame:** state the problem, scope, assumptions, affected FR/NFR IDs,
   acceptance criteria, and ownership boundaries.
3. **Delegate:** assign bounded outputs. Avoid concurrent edits to the same file;
   prefer research and review tasks alongside implementation.
4. **Handoff:** each agent reports evidence consulted, files changed, commands
   run, failures, assumptions, and unresolved questions.
5. **Integrate:** the coordinator reviews actual artifacts and diffs; summaries
   are leads, not proof.
6. **Verify:** run acceptance tests and repository validation. Resolve conflicting
   conclusions through evidence or record the issue as unresolved.
7. **Learn:** convert significant discoveries and failures into requirements,
   tests, decision records, fixtures, or agent guidance.

## Handoff template

```text
Scope and requirement IDs:
Evidence consulted:
Changes/artifacts:
Validation performed and results:
Assumptions and confidence:
Failures or conflicting evidence:
Open questions and recommended next action:
```

## Guardrails

- Agent count does not replace evidence quality.
- Do not delegate secrets or real customer logs unnecessarily.
- Do not let agents overwrite unrelated or concurrent changes.
- Do not mark requirements implemented from a handoff summary alone.
- Stop integration when essential evidence is missing or two agents have made
  incompatible edits; preserve both findings and resolve explicitly.
