---
status: accepted
date: 2026-11-03
---

# Per-payment `excluded_from_budget` flag

`payments` gets an `excluded_from_budget boolean not null default false` column, settable on create and via `PATCH /payments/{id}`; the monthly overview's "paid against fixed costs" and "remaining expected" figures skip any expense payment with the flag set, regardless of which account it's posted against. Ad hoc costs were being posted to the fixed-cost account as a workaround, which silently distorted "paid against fixed costs" (computed from all expense payments in the month, not just recurring ones) even though the set-aside itself was already derived purely from `recurring_definitions`. A per-payment flag (vs. deriving "ad hoc" from `recurring_definition_id IS NULL`) lets a specific ad hoc cost still count toward the budget when that's actually wanted, and lets an already-posted payment be retro-flagged via PATCH.
