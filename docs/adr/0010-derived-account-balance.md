---
status: accepted
date: 2026-11-03
---

# Account balance is derived, never stored

Account current balance is derived (starting balance + paid payments + manual adjustments), never a stored mutable counter. Avoids dual-write drift between a counter and the ledger; stays auditable; a `balance_adjustments` table gives the user an explicit, logged way to override/correct a balance without corrupting the derivation.
