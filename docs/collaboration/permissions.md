---
id: collaboration-permissions
type: protocol
status: active
owner: human
updated: 2026-09-12
tags: [collaboration, permissions, nfr-028]
---

# Permissions and merge gates

This note records the **real** constraints on who or what may change source
evidence, code, follow-up data, and integration state. It does not invent
locks the repository does not have.

## Source evidence (input logs)

The CLI and report **read** the bundle. They do not rewrite log files. Optional
`-redact` changes only the generated HTML copy (NFR-005, NFR-006).

## Follow-up data

Issue metadata lives in the browser local cache and in operator-exported
`logify-follow-up-v1` JSON. It is not written back into the log tree and is
not uploaded by Logify.

## Code and docs (shared checkout)

Agents and humans follow the [collaboration protocol](README.md):

- Inspect active work items and Git state before editing.
- Declare bounded path ownership and an advisory `lease_expires` in a work item.
- A lease communicates intent. It is **not** a filesystem or Git lock.
- Overlap stops edits; do not last-writer-wins.

`internal/harness` fails `go test` (and therefore CI) when an open work item
(`proposed`, `active`, `blocked`, `review`) is missing `id`, `type`, `status`,
`owner`, `lease_expires`, at least one `scope` path, or at least one
requirement ID.

## Review routing (CODEOWNERS)

[`.github/CODEOWNERS`](../../.github/CODEOWNERS) names `@artofdream` for the
tree. That is review routing for this personal repository.

CODEOWNERS becomes a merge constraint only if GitHub branch protection is
configured with **Require review from Code Owners**. This repository does not
claim that setting from application code. Do not treat a green agent PR as a
self-approval of an irreversible action.

## CI and branch protection checklist

[`.github/workflows/ci.yml`](../../.github/workflows/ci.yml) already documents
the intended human/CI gate. Operator-side GitHub settings (not enforceable
from this tree) for `main`:

1. Settings â†’ General â†’ Pull Requests â†’ allow or deny auto-merge as policy.
2. Settings â†’ Branches â†’ protect `main`.
3. Require status checks to pass before merging.
4. Select the single aggregator job named `validate` (covers the OS `test`
   matrix: `gofmt` on Linux, `go vet`, `go test ./...`, native `go build`,
   fixture smoke).
5. Optionally enable **Require review from Code Owners** if review routing
   should block merges. Leave "Require approvals" at an explicit human policy;
   do not rely on a bot rubber-stamp.
6. Do not grant agents permission to push directly to `main`.

A dated read-only probe of the live settings lives in
[branch-protection-probe.md](branch-protection-probe.md) (NFR-028). As of that
probe, **Require status checks** is enabled with the aggregator job `validate`
required, so GitHub refuses merges to `main` without a green `validate`.
**Require review from Code Owners** is still off. Re-probe when settings change;
do not treat this paragraph as live without an updated `probed_at`.

## What remains advisory

- Work-item leases
- CODEOWNERS without required-review branch protection
- Agent consensus or a closed work item (acceptance probes decide)
