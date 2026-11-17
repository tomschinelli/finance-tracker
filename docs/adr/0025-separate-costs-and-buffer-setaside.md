---
status: accepted
date: 2026-11-03
---

# Costs and buffer set aside as separate transfers

The monthly budget set-aside is deposited as two separate transfers/markers — "costs" (predicted costs) and "buffer" (smoothed non-monthly costs) — instead of one combined transfer, each independently idempotency-checked via its own note. Lets the user set aside (or re-check) the current month's fixed costs and the irregular-cost buffer independently, since they fund different things; `monthlyBudgetOverview` now exposes `costs_already_set`/`buffer_already_set` and `costs_marker_note`/`buffer_marker_note` instead of a single combined pair.
