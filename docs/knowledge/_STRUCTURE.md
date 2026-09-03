---
id: knowledge-structure
type: structure
status: active
owner: human
updated: 2026-09-03
---

# Vault structure

| Path | Purpose | Authority | Write pattern |
|---|---|---|---|
| `index.md` | Navigation and vault constitution | Canonical map | Curated |
| `architecture.md` | Current descriptive system model | Code-verified | Update in place |
| `glossary.md` | Shared terms | Canonical terminology | Update in place |
| `open-questions.md` | Missing/contested knowledge | Honest gap ledger | Append and resolve |
| `decisions/ADR-*.md` | Decision rationale and consequences | Canonical rationale | Immutable outcome; supersede by link |
| `research/*.md` | Sourced investigation | Supporting evidence | One topic per file |
| `incidents/*.md` | Failure, diagnosis, recovery, learning | Operational evidence | One incident per file |
| `sessions/*.md` | Short-lived episodic context | Non-canonical | One writer per file; promote then close |
| `../collaboration/work-items/*.md` | Active scope and advisory lease | Coordination state | One owner per file |
| `../collaboration/handoffs/*.md` | Ownership transfer evidence | Coordination evidence | Append-only handoff |

All notes use YAML frontmatter for machine discovery and ordinary Markdown links
for portability. Obsidian wikilinks may be added, but must not be the only route to
canonical material.
