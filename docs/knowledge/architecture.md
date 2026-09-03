---
id: logify-architecture
type: architecture
status: active
owner: human
updated: 2026-09-03
sources: [../../README.md, ../../cmd/logify/main.go, ../../internal/analyzer, ../../internal/report]
---

# Current architecture

Logify is a dependency-free Go CLI. `cmd/logify` accepts a directory and options;
`internal/analyzer` discovers and normalizes logs, builds signatures, groups
repeats, and orders the timeline; `internal/report` emits one offline HTML file.

This note describes observed structure. Requirements remain authoritative for
intended behavior, and tests/compiler output remain evidence of implementation.

The target trust model follows the [AEA adoption mapping](../framework-adoption.md):
reviewable incident understanding is derived from authoritative log evidence and
protected by an outer harness of guides, sensors, loop, memory, permissions, and
observability.
