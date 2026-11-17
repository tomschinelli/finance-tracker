---
status: accepted
date: 2026-11-03
---

# Direct balance adjustments allowed on any account

`POST /accounts/{id}/adjustments` is now allowed on any account, not just the default one (supersedes the restriction in ADR-0019); the accounts-page "Add adjustment" form posts a direct adjustment for every account, with no transfer-counterpart workaround. Non-default accounts' real-world (bank) balances drift for reasons outside the app's transfer model and need direct correction; routing that correction through a fake transfer from another account was a workaround, not a real transfer. `POST /transfers` is unchanged and remains how the monthly budget set-aside actually moves money between accounts.
