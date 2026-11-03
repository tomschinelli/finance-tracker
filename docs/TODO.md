# TODO / Backlog

Each item is named by a readable title and carries a feature key
(`FEAT-NNN`) in parentheses — reference the key in commit messages and
PRs for traceability (e.g. `feat(backend): allow adjustments on any
account [FEAT-001]`). Ordered for implementation (not the order first
given) — work top to bottom unless told otherwise. Bank integration and
the desktop app are deliberately last: bank import depends on account
number, and the desktop app is low priority with open design questions.
Update this file as items are completed (strike through or remove) or
re-scoped.

## Account number per account (FEAT-003)

Add an `account_number` field to `accounts` (migration + `sqlc generate` +
API + frontend). Groundwork for bank statement import — bank import needs
to match imported transactions/statements to the right account.

No dependencies, but must land before bank statement import (FEAT-005).

## Bank statement import, first: KBC Belgium (FEAT-005)

Introduce a per-import (or per-account) "bank type" so multiple bank
formats can be added over time; first supported importer is KBC
(Belgium). Depends on **account number per account** (FEAT-003) for
matching imported rows to the right account. Format/parsing details to be
provided later — don't build a generic importer abstraction before the
first concrete format is known.

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
