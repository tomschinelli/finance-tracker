---
status: accepted
date: 2026-11-03
---

# Budget modal: direct adjustment or transfer

`AddBudgetModal` can deposit the set-aside amount either as a direct `POST /accounts/{id}/adjustments` (no source account) or as a `POST /transfers` from a chosen account, mirroring the accounts page's existing direct/transfer radio toggle. The modal previously forced picking a source account even though direct adjustments on any account have been allowed since ADR-0023; the user wants to decide per deposit whether there's a traceable source or not, same as the accounts page already allows.
