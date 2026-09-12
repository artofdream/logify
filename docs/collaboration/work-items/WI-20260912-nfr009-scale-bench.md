---
id: WI-20260912-nfr009-scale-bench
type: work-item
status: review
owner: cursor-agent
created: 2026-09-12T10:25:00Z
updated: 2026-09-12T10:35:00Z
lease_expires: null
scope:
  - Makefile
  - .gitignore
  - .github/workflows/ci.yml
  - AGENTS.md
  - README.md
  - testdata/scale/
  - docs/requirements/non-functional-requirements.md
  - docs/framework-adoption.md
  - docs/knowledge/architecture.md
  - docs/knowledge/open-questions.md
  - docs/knowledge/decisions/ADR-0012-scale-benchmark.md
  - docs/knowledge/research/RES-20260912-nfr009-scale-bench.md
  - docs/collaboration/work-items/WI-20260912-nfr009-scale-bench.md
  - docs/collaboration/handoffs/HO-20260912-cursor-agent-nfr009.md
  - internal/analyzer/analyzer.go
  - internal/analyzer/model.go
  - internal/analyzer/correlate.go
  - internal/analyzer/merge.go
  - internal/analyzer/merge_test.go
  - internal/analyzer/scale_fixture.go
  - internal/analyzer/scale_measure.go
  - internal/analyzer/nfr009_test.go
  - internal/analyzer/scale_bench_test.go
requirements: [NFR-009, NFR-008, NFR-014, NFR-028]
depends_on: []
supersedes: []
---

# NFR-009 scale benchmark and streaming merge

## Goal

Give NFR-009 a repeatable, generated (not git-committed) 1 GiB multi-instance
fixture, a measurable benchmark, documented reference hardware, and the
smallest streaming / bounded-buffer improvements that can keep peak RSS honest
against the 512 MiB Should budget. Leave status Partial unless every AC is
verified.

## Non-goals

- Committing a 1 GiB fixture
- Third-party dependencies
- Changing the timeline model so unique signatures are discarded
- Marking NFR-009 Implemented without measured AC2/AC3 evidence
- Merging the PR

## Acceptance criteria

NFR-009 ACs 1–4 as written, plus: CI may run only a small smoke; full 1 GiB is
manual/nightly; measured numbers are recorded even when they miss the budget.

## Planned files and ownership

Bounded to the scope list. WI-20260912-nfr028 is `review` with a cleared lease
and already merged on `main`; this item takes a new slice of the shared docs
and analyzer files.

## Evidence and assumptions

Assumption: a realistic support-bundle *storm* (high repeat, bounded unique
signatures, four instances) is the documented fixture. An adversarial
unique-line access corpus would retain one row per signature (FR-010) and is
called out as the remaining AC2 risk rather than hidden.

## Validation

`gofmt`, `go test ./...`, `go build`, `go vet`, `git diff --check`, fixture
smoke, plus `make bench-nfr009` / `LOGIFY_NFR009_FULL=1` on this VM.

## Activity log

- `2026-09-12T10:25:00Z` — cursor-agent — claimed paths; branching
  `cursor/nfr009-scale-bench-fd16` from current `main`.
- `2026-09-12T10:40:00Z` — cursor-agent — online merge, bounded occHints,
  generated storm fixture, docs/CI/Make; pre-test commit next.
- `2026-09-12T10:35:00Z` — cursor-agent — 1 GiB child-process run passed
  AC1–AC3 (13.6 MiB RSS, 59 s). Status Implemented. Handoff written.

## Handoff or completion

[HO-20260912-cursor-agent-nfr009](../handoffs/HO-20260912-cursor-agent-nfr009.md)
— PR #18, do not merge.
