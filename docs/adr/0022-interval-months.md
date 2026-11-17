---
status: accepted
date: 2026-11-03
---

# Replace frequency enum with `interval_months`

`recurring_definitions.frequency` (enum `monthly`/`quarterly`/`yearly`) is replaced by `interval_months integer` (any value >= 1); `quarterly` had no special status beyond being `interval_months = 3`. User wanted an arbitrary cadence (e.g. "every 2 months"), which a fixed 3-value enum can't express. `interval_months` already matched the internal occurrence-generation step size and the budget-buffer divisor, so this is a rename-and-generalize, not a new concept: monthly = 1, yearly = 12, any other positive integer is a custom interval, with no separate "quarterly" case in code.
