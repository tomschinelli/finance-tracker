---
status: superseded by ADR-0019
date: 2026-11-03
---

# Monthly budget set-aside deposited as a noted balance adjustment

"Set this month's budget" deposits the set-aside amount into a chosen account as a regular `balance_adjustments` row, tagged with a conventioned note (`"Monthly budget set-aside YYYY-MM"`) used to detect it was already applied. Reuses the existing adjustment mechanism instead of adding a budgets table; the note text is the idempotency marker, checked per month when building the monthly overview.

The single-sided deposit was replaced by transfers (ADR-0019), and the single marker split into costs/buffer markers (ADR-0025).
