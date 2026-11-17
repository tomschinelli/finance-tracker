# TODO / Backlog

Each item is named by a readable title and carries a feature key
(`FEAT-NNN`) in parentheses — reference the key in commit messages and
PRs for traceability (e.g. `feat(backend): allow adjustments on any
account [FEAT-001]`). Ordered for implementation (not the order first
given) — work top to bottom unless told otherwise. Bank integration and
the desktop app are deliberately last: bank import needs its format
details, and the desktop app is low priority with open design questions.
Update this file as items are completed (strike through or remove) or
re-scoped.

## Bank statement import, first: KBC Belgium (FEAT-005)

Introduce a per-import (or per-account) "bank type" so multiple bank
formats can be added over time; first supported importer is KBC
(Belgium). Accounts now carry an `account_number` to match imported rows
to the right account. Format/parsing details to be provided later — don't
build a generic importer abstraction before the first concrete format is
known.

## Point-in-time snapshot / rollback (FEAT-007)

Add a way to snapshot the full database state and restore (roll back) to
a prior snapshot. Builds on the existing `GET /data/export` / `POST
/data/import` full-JSON-snapshot mechanism — likely wraps it with stored,
timestamped snapshots (e.g. a `snapshots` table or on-disk storage) and a
restore endpoint, rather than inventing a new backup format. Needed as a
safety net before the MCP API goes live, so a bad write made through an
external AI tool can be undone.

No dependencies, but must land before the MCP API (FEAT-008).

## MCP API for AI tool access (FEAT-008)

Expose an MCP (Model Context Protocol) server/API so an external AI
assistant can read and write data directly — e.g. importing transactions
from a custom spreadsheet. Depends on **point-in-time snapshot/rollback**
(FEAT-007) landing first, since writes made by an external AI need to be
undoable. Scope (which resources/operations are exposed, auth/access
control for the MCP endpoint) to be defined when we get here.

## Account type: debit, prepaid credit, credit card (FEAT-009)

Add a `type` field to accounts (e.g. debit card, prepaid credit card,
credit card) to capture more detail about how each account is held.
Independent of the `account_number` field already added for bank import
matching (FEAT-005).

## Recurring costs with multiple occurrences per period (FEAT-010)

Some recurring expenses happen more than once within a single period (e.g.
fuel costs several times a month) but should still be grouped under one
shared name rather than tracked as separate recurring definitions. Each
occurrence should still count toward the budget individually, so the
monthly set-aside calculation needs to sum all occurrences for the period,
not assume at most one per `RecurringDefinition`.

## Reconciliation (FEAT-011)

Let the user set an Account's Current balance to the real bank balance at
any time (see `GLOSSARY.md`: Reconciliation, Untracked difference). The gap
is recorded as a relative entry, labeled distinctly from a normal
Adjustment, and the UI hints that money moved without being tracked. The
plain Adjustment stays alongside it.

**To be discussed:** turning part of an Untracked difference into normal
Payments later, once its source is known (e.g. a €300 bill found inside a
+€1,000 difference). Leaning towards: the reconciled balance never changes;
the remaining difference shifts by the added Payment's signed amount. Name
for the action ("Explain"?) and the sign handling not yet agreed.

## Recurring item amount follows last Paid amount (FEAT-012)

When a Payment of a Recurring item is marked Paid with a Paid amount that
differs from its Expected amount, update the Recurring item's amount to
the Paid amount, so later Expected Payments and budget figures follow
(e.g. salary that varies month to month). Applies to every Recurring item,
no "variable" flag. Open: whether already-generated future Expected
Payments (e.g. projected non-monthly ones) get their amount updated too.

## Yearly overview covers all non-monthly expenses (FEAT-013)

Today the yearly overview only includes Recurring items with an interval
of exactly 12 months, doesn't filter by Kind (yearly income counts as a
cost), and computes its own "monthly buffer" (yearly total / 12) that
differs from the monthly overview's Buffer. Make it cover all non-monthly
*expense* Recurring items, so its buffer figure equals the Buffer (see
`GLOSSARY.md`: Yearly overview, Buffer, Covers).

## Align UI labels with the glossary (FEAT-014)

Rename user-facing labels to the `GLOSSARY.md` terms, no logic change:
"default" badge → Main account; "Predicted costs" → Fixed costs; "Add
budget to account" → set aside Monthly budget; "Unpaid" → Expected; "Add
recurring cost" / "Recurring definitions" → Recurring item; "Add ad hoc
cost" → Ad hoc payment.

## Docs cleanup after glossary (CHORE-001)

- Trim `CLAUDE.md` "Domain model (v1)" to a pointer at `GLOSSARY.md`, keep
  code-level rules (integer cents, derived balance, lazy occurrence
  generation); it still says only the default account takes direct
  adjustments, reversed by ADR 0023.
- Add `GLOSSARY.md` link under "Docs" in `README.md`.
- Mark ADR 0019 superseded (in part) by 0023, and 0012 by 0025.

## Desktop app: single Go binary — low priority (FEAT-006)

- New `cmd/` entrypoint (e.g. `cmd/desktop`) producing a single binary,
  no dock icon.
- Needs a local embedded DB suited to single-binary distribution (SQLite
  is the likely candidate) instead of Postgres — implies either a second
  sqlc target or a storage interface boundary.
- GUI approach (embedded webview around the existing Nuxt frontend vs. a
  native Go GUI toolkit) and "no dock icon" mechanics still need research
  when we get here — not designing this speculatively now.
- **Storage should be configurable**, not hardcoded to a bundled local
  file: a config option should allow either (a) a fully local, self-
  contained DB file — making the whole app portable/shareable as one
  unit — or (b) pointing the desktop app at the *same* database the web
  app's backend uses (e.g. the same Postgres instance), so both can
  operate against shared data. Needs design when we get here; likely
  implies the storage layer can't assume SQLite-only if option (b) must
  also work.

Last item: lowest priority, most open design questions, no dependents.
