---
status: superseded by ADR-0009
date: 2026-11-03
---

# v1 scope: accounts, categories, transactions only (no budgets)

v1 covers accounts, categories, and transactions only — no budgets yet. Budgeting is a distinct feature with its own data model (period, limits, rollover rules); building it speculatively before the core expense-tracking flow exists would be a speculative abstraction. Add when there's a working vertical slice to build it on top of.
