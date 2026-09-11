---
id: ADR-0008-event-correlation
type: decision
status: accepted
owner: cursor-agent
created: 2026-09-11
updated: 2026-09-11
requirements: [FR-012, FR-023]
supersedes: []
---

# ADR-0008: Deterministic event correlation rules

## Context and evidence

FR-012 asks Logify to group related events beyond sitting next to each other on
the timeline. Shared request IDs, client addresses, exception causes, and short
time windows can describe one incident that spans Apache HTTPD and Tomcat.

Acceptance criteria require documented deterministic rules, individually
inspectable events, visible groups with supporting evidence, a visible
distinction between exact identifiers and false-positive-prone heuristics, and
an explicit preference for **no** correlation when evidence is weak.

On `main` before this change, Java `Caused by:` / stack frames were already
joined into one event. That preserves a cause chain as observed evidence; it is
not a cross-event inference. Access-log client addresses were parsed and then
dropped, so an IP+window rule had no HTTPD-side address unless the analyzer
retained it.

## Decision

1. Correlation runs after discovery, parsing, time filtering, instance-local
   deduplication, and chronological sort. It never merges, hides, or reorders
   timeline events. Labeled identifiers and client addresses from occurrences
   collapsed by dedup still participate; groups point at the surviving rows.
2. Only two rules are implemented. Both are deterministic: same input events
   produce the same groups, rule names, confidence labels, evidence strings,
   member order, and correlation IDs.

   | Rule ID | Kind | Confidence | Members | Evidence required |
   |---|---|---|---|---|
   | `shared-request-id` | `exact` | `high` | 2+ events | Same labeled identifier value |
   | `client-ip-window` | `heuristic` | `low` | one access + one Tomcat | Same non-loopback client address on an `apache-access` event and a `tomcat-java` event whose timestamps differ by at most 5s. Each pair is its own group; pairs are not unioned across time or across distinct IPs. |

3. **Exact identifiers** are taken only from labeled forms in the event
   message: `request-id` / `requestId` / `reqId` / `x-request-id`,
   `correlation-id` / `correlationId` / `corrId` / `x-correlation-id`,
   `trace-id` / `traceId` / `x-trace-id`, and W3C `traceparent` (the 32-hex
   trace-id field). Separators are `=` or `:`. Values are 8–128 characters of
   `[A-Za-z0-9._:-]`. Grouping is case-insensitive on the value. Placeholders
   `null`, `none`, `undefined`, `unknown`, `requestid`, and `traceid` are
   ignored. Bare numbers (`Request 12345`), unlabeled UUIDs, and hex addresses
   are **not** used.
4. **Heuristic `client-ip-window`** uses the access-log remote address (retained
   as `Event.ClientAddr`) and IPv4 literals found in Tomcat messages. Loopback,
   unspecified, multicast, and link-local addresses are rejected. Apache error
   `[client …]` is retained for provenance and may appear in exact groups when
   a labeled ID is also present; it does not satisfy the HTTPD side of this
   heuristic by itself. IPv6 literals in free-text Tomcat lines are not mined.
5. An event may belong to more than one group when different rules fire or
   when distinct heuristic pairs share a member. A single matching event never
   creates a group. Two access events that share only a client address, with
   no in-window Tomcat counterpart, stay ungrouped.
6. Each group has a stable `corr-v1-…` id derived from rule, normalized key,
   and member identity (`signature`, `instance`, slash-normalized file, line) —
   the same provenance tuple as evidence IDs, not display order.
7. The report shows groups in a dedicated **Correlation groups** section. Each
   card names the rule, kind (`exact` vs `heuristic`), confidence, evidence
   string, and member events with navigation back to the timeline. Timeline
   rows that participate stay fully visible and gain a text badge. The legend
   treats correlations as inferences, distinct from observed records and
   operator issue metadata.
8. Optional redaction rewrites correlation `evidence` and `ClientAddr` the same
   way it rewrites other log-derived display strings. Correlation IDs, rule
   names, kind, and confidence are not rewritten.

Intentionally not implemented (prefer no link):

- Grouping by exception class, `Caused by:` text, or stack frames across events
- Unlabeled tokens, session IDs, or memory addresses
- Client-IP grouping of HTTPD-only or Tomcat-only sets
- A configurable window or additional heuristic families

## Alternatives considered

- **Auto-merge correlated events into one timeline row** was rejected: AC2
  requires each event to remain inspectable.
- **Matching unlabeled UUIDs / long hex / `Request N`** was rejected: those
  tokens are already treated as volatile for signatures and are
  false-positive-prone (AC4, AC5).
- **Using 127.0.0.1 / `::1` for the IP window** was rejected: loopback is
  shared by unrelated local traffic. The existing `testdata/case` bundle uses
  `127.0.0.1` and must stay ungrouped.
- **Treating Apache error `[client]` as sufficient HTTPD evidence** was
  rejected for the heuristic: the documented case is access + Tomcat. Error
  lines still join exact identifier groups.
- **Schema/issue auto-create from a correlation** was rejected: follow-up
  identity stays operator-driven (ADR-0001 / FR-017).

## Consequences and risks

Operators can see a high-confidence request-id incident that spans HTTPD and
Tomcat, and a separately labeled low-confidence IP+window hint. Quiet bundles
produce an empty list rather than speculative edges. A later parser that
surfaces more labeled IDs will form more exact groups without changing rule
IDs. The 5s window and IPv4-only Tomcat harvest are documented limits; widening
them needs a new ADR.

## Verification

Analyzer tests on `testdata/correlate` and `testdata/case` must prove: exact
groups form for a shared labeled id; the IP-window group is `heuristic`/`low`
and distinct; unlabeled, loopback, out-of-window, and singleton identifiers
stay ungrouped; `testdata/case` has zero groups. Report tests must embed a
JSON array of groups (never `null`), render rule/kind/confidence/evidence, and
keep events individually listed. Redaction must rewrite evidence text after
identities are computed.
