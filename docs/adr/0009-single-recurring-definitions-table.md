---
status: accepted
date: 2026-11-03
---

# Budgeting in scope; one `recurring_definitions` table for all frequencies and kinds

Budgeting is now in scope (supersedes ADR-0008). One `recurring_definitions` table covers all frequencies (monthly/quarterly/yearly) and kinds (income/expense). Rent, bills, salary, and irregular costs don't differ structurally — only frequency changes. Two parallel models (e.g. "irregular" vs "monthly") would duplicate the same shape for no benefit.

The fixed frequency enum was later generalized to `interval_months` (ADR-0022).
