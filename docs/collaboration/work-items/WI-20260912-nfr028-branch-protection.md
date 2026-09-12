---
id: WI-20260912-nfr028-branch-protection
type: work-item
status: review
owner: grok-bot-logify
lease_expires: 2026-09-13T06:00:00Z
scope:
  - docs/collaboration/permissions.md
  - docs/collaboration/branch-protection-probe.md
  - docs/framework-adoption.md
  - docs/requirements/non-functional-requirements.md
  - docs/collaboration/work-items/WI-20260912-nfr028-branch-protection.md
  - internal/harness/nfr028_test.go
requirements: [NFR-028]
---

# WI: NFR-028 Permissions probe + harness closeout

## Intent
Record a dated branch-protection probe proving `validate` is a required status check on `main`, update Permissions/Sensors honesty in the adoption ledger, and move NFR-028 to Implemented when ACs hold — without claiming CODEOWNERS merge-lock or WCAG/1 GiB CI.

## Machine
(evo-x2) — CloudAgent held (on-demand cap).

## Done when
- Probe doc committed and harness-tested
- permissions.md + framework-adoption updated
- NFR-028 Implemented only if ledger layers are Adopted with honest gaps
- `go test ./internal/harness` passes
