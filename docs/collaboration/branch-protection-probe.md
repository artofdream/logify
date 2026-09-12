---
id: branch-protection-probe
type: probe
status: verified
owner: grok-bot-logify
probed_at: 2026-09-12T23:16:00Z
machine: evo-x2
requirements: [NFR-028]
source: "GET /repos/artofdream/logify/branches/main/protection"
---

# Branch protection probe (`main`)

Operator-side GitHub settings are not enforceable from application code. This
file is a **dated read-only probe** of those settings so Permissions evidence
is not left as Unknown.

## Method

```text
gh api repos/artofdream/logify/branches/main/protection
```

Probed on **(evo-x2)** with the sponsor GitHub credentials available to the
Logify coordinator. Re-run the command and update `probed_at` plus the table
below when settings change.

## Observed (2026-09-12)

| Setting | Value | Honesty |
|---|---|---|
| Required status checks | `validate` (strict=false) | Verified — merge blocked without green `validate` |
| Required approving review count | `0` | Verified |
| Require code owner reviews | `false` | Verified — CODEOWNERS is routing only |
| Enforce admins | `false` | Verified — admins may bypass |
| Allow force pushes | `false` | Verified |
| Allow deletions | `false` | Verified |
| Required signatures | `false` | Verified |
| Rulesets | none | Verified (`GET /repos/.../rulesets` → `[]`) |

## What this does **not** claim

- CODEOWNERS is not a merge lock (require_code_owner_reviews=false).
- Work-item leases remain advisory.
- Admins can still bypass protection (enforce_admins=false).
- Sensors still do not include a WCAG engine or CI 1 GiB NFR-009 run; those
  stay explicit gaps / manual paths, not NFR-028 blockers once mechanical
  sensors for supported behavior exist.
