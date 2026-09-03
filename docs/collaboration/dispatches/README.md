---
id: cloud-dispatch-registry
type: registry
status: active
owner: human
updated: 2026-09-04
---

# Cloud-task dispatch registry

These records notify agents which requirements are assigned to cloud execution.
They are coordination state, not proof that implementation or a PR exists.

Discover reserved work after fetching `main`:

```text
rg -n "^status: (active|queued|blocked|review)$|^requirements:|^branch:" docs/collaboration/dispatches
```

Statuses are `queued`, `active`, `blocked`, `review`, `done`, and `cancelled`.
Only `active` may carry a task identifier; only `review` may claim an observed PR.
An agent must not duplicate `active`, `queued`, or `review` requirements without
an explicit takeover recorded here and in the related work item.

Queued dependencies are dispatched only after their prerequisites merge. Every
transition records observed evidence; never invent a task ID, branch, PR, test
result, or merge state.
