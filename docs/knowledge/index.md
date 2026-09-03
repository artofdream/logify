---
id: logify-knowledge-index
type: index
status: active
owner: human
updated: 2026-09-03
tags: [knowledge, second-brain, obsidian]
---

# Logify knowledge vault

This directory is an Obsidian-compatible, plain-Markdown Second Brain for durable
project knowledge. It is intentionally useful without Obsidian, a database, or an
agent vendor. Git is the history and review mechanism.

## Maps

- [Structure](_STRUCTURE.md)
- [Architecture](architecture.md)
- [Glossary](glossary.md)
- [Open questions](open-questions.md)
- [Requirements](../requirements/README.md)
- [Core principles](../principles.md)
- [Collaboration protocol](../collaboration/README.md)

## Collections

- `decisions/` — durable architecture decision records.
- `research/` — sourced investigations that may inform a decision.
- `incidents/` — significant failures and the safeguards learned from them.
- `sessions/` — short episodic memory, one agent per file.

## Knowledge rules

- Prefer links to canonical sources over copied summaries.
- Every claim records evidence, source authority, freshness, and uncertainty where
  material.
- Session notes are temporary context. Promote reusable facts into requirements,
  ADRs, architecture, fixtures, tests, or glossary entries.
- Update an existing canonical note when knowledge changes; record the correction
  rather than creating silent contradictions.
- Never place secrets, credentials, or unsanitized customer logs in the vault.
- Unknown remains a valid state. Do not fill gaps with plausible invention.
