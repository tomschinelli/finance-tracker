---
status: accepted
date: 2026-11-03
---

# Overdue and upcoming payments in the monthly overview

`/overview/monthly` response gains `overdue_payments` and `upcoming_payments` arrays alongside the existing month-scoped `payments`; `generateOccurrences` now projects non-monthly (`interval_months > 1`) definitions through the end of the viewed year (`upcomingLookAheadEnd`) instead of only up to the viewed month, while monthly definitions are unaffected.

Implements FEAT-004 with the refinement that a non-monthly recurring definition's computed due date is only an estimate — the real bill can land earlier or later. Unpaid non-monthly occurrences whose due date has passed keep surfacing every month until paid (`overdue_payments`, highlighted in the UI), and occurrences projected for a later month in the same year are generated but kept out of the current month's main list, surfaced instead as a collapsed, year-scoped `upcoming_payments` the user can reveal on demand (e.g. to mark one paid if it arrives early) — the look-ahead horizon lives in one function (`upcomingLookAheadEnd`) so it's a one-line change to switch to a rolling window later. The reveal toggle is intentionally not persisted — it resets to hidden on every page load/month change, per explicit user preference over remembering it.
