---
id: ADR-0003-follow-up-details
type: decision
status: accepted
owner: cursor-agent
created: 2026-09-09
updated: 2026-09-09
requirements: [FR-021, FR-022, NFR-019]
supersedes: []
---

# ADR-0003: Editable follow-up details and overdue rule

## Context and evidence

FR-021 requires notes, an optional owner, and an optional due date on an issue,
safe rendering, visible overdue unresolved issues, and owner/overdue filters.
ADR-0002 already stores `owner`, `due`, and `notes` on `logify-follow-up-v1` so
imported values display and round-trip. DSP-20260904-FR020-FR021 left editors
out of the Must UI PR; that is the remaining gap.

## Decision

1. Do not bump the export schema. Editors read and write the existing v1 fields.
   Empty owner, due, and notes persist as JSON `null`.
2. Owner is trimmed, at most 200 characters. Notes are stored as typed (not
   trimmed) up to 8000 characters. Due must be a real `YYYY-MM-DD` calendar date
   or empty.
3. Treat all three fields as untrusted text. The page script assigns them with
   DOM `value` / `textContent` APIs only (`el()`, input/textarea values). Do not
   use `innerHTML` for operator metadata.
4. **Overdue** means all of:
   - `due` is present;
   - `due` is strictly before the UTC calendar date of the report clock
     (`(nowIso || Date#toISOString()).slice(0, 10)`);
   - state is not `resolved` or `dismissed`.
   A due date equal to that UTC date is not overdue. `blocked` and
   `investigating` remain overdue when the date has passed.
5. The issue queue already has owner and overdue filter chrome (FR-023). Editors
   populate those fields; the overdue checkbox uses the rule above.

## Alternatives considered

- Schema v2 for editors was rejected: v1 already has the fields.
- Viewer-local calendar "today" is more natural for some operators but would
  make overdue depend on the machine timezone. Logify already treats
  timezone-less log timestamps as UTC; the report clock stays UTC for a
  deterministic, testable rule.
- Inclusive overdue (`due <= today`) would mark work due today as late. Strict
  before today matches "past the due date."

## Consequences and risks

Operators can edit details offline; export/import compatibility with existing
v1 files is preserved. Viewers east of UTC near midnight may see a different
overdue set than viewers still on the previous local date. Invalid calendar
dates such as `2026-02-30` are rejected on import and edit.

## Verification

Node tests must edit owner/due/notes, round-trip them through export/import,
prove the overdue rule (including resolved/dismissed and due-today), and keep
XSS payloads as text. The page script must not assign `innerHTML`.
