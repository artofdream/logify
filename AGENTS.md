# Logify agent guide

Logify is a dependency-free Go CLI that turns mixed Tomcat/Java and Apache
HTTPD log bundles into a self-contained interactive HTML incident timeline.

## Source of truth and workflow

Read and apply `docs/principles.md` first. Honesty, antifragility, and knowledge
first are non-negotiable decision gates. For delegated work, follow
`docs/multi-agent-workflow.md`; the coordinating agent remains responsible for
integration and final verification.

Other agents (Codex, Claude, Cursor Agent, Antigravity, Grok, or others) may be
working concurrently in this checkout. Before editing, read
`docs/collaboration/README.md`, inspect active work items and `git status`, then
register bounded path ownership. Use `owner: <your-agent-id>` in your own
work/session files (for example `codex`, `claude`, `cursor-agent`,
`antigravity`, or `grok-bot`) — pick one stable identifier and use it
consistently across your work items, handoffs, and commit trailers.

Product behavior and quality constraints are defined in `docs/requirements/`.
Use the documentation-first workflow described there: update FR/NFRs before
changing observable behavior, derive tests from acceptance criteria, then
implement. Cite affected requirement IDs in change summaries. Do not mark a
requirement implemented until every acceptance criterion is verified.

## Repository map

- `cmd/logify/`: CLI entry point and flag handling.
- `internal/analyzer/`: discovery, parsing, normalization, signatures,
  deduplication, filtering, and timeline ordering.
- `internal/report/`: embedded HTML/CSS/JavaScript report generation.
- `testdata/case/`: representative Tomcat and HTTPD fixtures.
- `README.md`: user-facing behavior, build instructions, and known limits.
- `docs/requirements/`: canonical functional and non-functional requirements.
- `docs/principles.md`: non-negotiable decision and delivery principles.
- `docs/multi-agent-workflow.md`: delegation, handoff, and verification protocol.
- `docs/collaboration/`: multi-agent work ownership and handoff records.
- `docs/knowledge/`: Obsidian-compatible durable Second Brain.

## Working rules

- Preserve the single-native-executable design. Prefer the Go standard library;
  justify any new dependency before adding it.
- Keep parser-specific logic separate from the normalized `Event` model.
- Treat input logs as untrusted. Never execute their contents, and preserve HTML
  escaping when embedding event data in reports.
- Keep scanning resilient: an unreadable or malformed file should produce a
  warning where practical rather than aborting the entire bundle.
- Do not silently discard recognized records. Untimestamped records belong after
  timestamped events unless an active time filter excludes them.
- Keep signatures deterministic. Normalize volatile identifiers carefully and
  deduplicate within an instance, not across unrelated instances.
- Preserve source file and line provenance for every event.
- Do not commit generated executables or HTML reports.
- Avoid destructive Git operations and do not overwrite unrelated local changes.

## Multi-agent coordination

- Any number of agents (human or autonomous) may be active on this repository at
  once. Nothing below assumes exactly two.
- Delegate only bounded, independently verifiable work with named FR/NFR IDs and
  explicit file or area ownership.
- Avoid concurrent edits to the same files. All agents share one working tree, so
  preserve user and peer changes.
- Require handoffs to state scope, evidence, artifacts or files, validation,
  assumptions, failures, and unresolved questions.
- Treat agent summaries as leads. The coordinator inspects the actual diff and
  artifacts, resolves conflicts, and runs final validation.
- Never mark status or validation complete solely because another agent asserted
  it or because multiple agents agree.

## Parser changes

When adding or changing a parser:

1. Add the smallest representative fixture under `testdata/`.
2. Test successful parsing, timestamp and severity mapping, multiline behavior
   where relevant, and malformed-input behavior.
3. Check interactions with deduplication and chronological ordering.
4. Document newly supported formats or remaining ambiguity in `README.md`.

Java continuation detection must retain exception headers, stack frames,
`Caused by:` lines, and elided frames with their leading event. Apache access
status classes map to severity as follows: 2xx/3xx `INFO`, 4xx `WARN`, and 5xx
`ERROR`.

## Validation

Run these commands before handing off code changes:

```text
gofmt -w cmd internal
go test ./...
go build -o logify.exe ./cmd/logify
go vet ./...
git diff --check
```

For parser or report changes, also exercise the fixture bundle:

```text
./logify.exe -output sample-report.html testdata/case
```

On PowerShell, use `.\logify.exe` instead of `./logify.exe`. If Go is not on
`PATH`, report that limitation explicitly; do not claim compiler validation.
Report commands and observed outcomes exactly, including anything not run or
failed.

## Scope discipline

- Prefer focused packages and table-driven tests over broad rewrites.
- Keep public CLI behavior backward compatible unless the task explicitly calls
  for a breaking change.
- Update this guide when repository structure or required validation changes.
