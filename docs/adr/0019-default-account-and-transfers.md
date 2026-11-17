---
status: accepted
date: 2026-11-03
---

# Default account and transfers between accounts

Accounts get an `is_default` flag (exactly one, enforced by a partial unique index). Only the default account can be adjusted directly (`POST /accounts/{id}/adjustments`, sign-agnostic, no counterpart); every other account's balance only changes via `POST /transfers` — a pair of `balance_adjustments` rows (debit source, credit destination) inserted in one transaction, with both accounts chosen explicitly by the caller, not inferred. The monthly budget set-aside and the accounts-page "Add adjustment" form both use this (supersedes the single-sided deposit/correction on any account, ADR-0015). The default account is the only one fundable from outside the ledger (e.g. salary); crediting any other account without an offsetting debit would invent money, so funding a non-default account must be a real transfer with an explicitly chosen source — never silently assumed.

The direct-adjustment restriction was later lifted (ADR-0023); transfers are unchanged.
