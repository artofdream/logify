# Claude Code project guidance

@AGENTS.md

The imported `AGENTS.md` is the shared source of truth for architecture,
invariants, and validation. Apply these Claude-specific conventions as well:

- Apply `docs/principles.md` before planning or delegating work, and use the
  handoff contract in `docs/multi-agent-workflow.md` for subagents.
- Codex may be working concurrently. Before editing, follow
  `docs/collaboration/README.md`, inspect active work items and `git status`, and
  register bounded path ownership with `owner: claude`. Read relevant Codex
  handoffs and verify their artifacts directly.
- Delegate only independent, bounded work with explicit ownership. Do not assign
  multiple agents to edit the same file; personally inspect shared-tree changes
  and run final validation before accepting a handoff.
- Read `docs/requirements/README.md`, the affected FR/NFRs, and relevant package
  tests before editing. Update requirements before changing behavior.
- Use focused edits and inspect the complete diff before finishing.
- Do not broaden permission settings or add hooks, MCP servers, dependencies, or
  network calls without an explicit project need.
- When changing the embedded report, verify both safe JSON/HTML embedding and
  offline operation with no external assets.
- State exactly which validation commands ran and distinguish failures in the
  code from missing local tooling.
- Never include log fixture contents copied from real support bundles if they may
  contain credentials, tokens, personal data, hostnames, or customer data.
