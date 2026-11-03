# TODO / Backlog

Each item has a feature key (`FEAT-NNN`) — reference it in commit messages
and PRs for traceability (e.g. `feat(backend): allow adjustments on any
account [FEAT-001]`). Ordered for implementation (not the order first
given) — work top to bottom unless told otherwise. Bank integration and
the desktop app are deliberately last: bank import depends on FEAT-003,
and the desktop app is low priority with open design questions. Update
this file as items are completed (strike through or remove) or re-scoped.

## FEAT-003 — Account number per account

Add an `account_number` field to `accounts` (migration + `sqlc generate` +
API + frontend). Groundwork for FEAT-005 — bank import needs to match
imported transactions/statements to the right account.

No dependencies, but must land before FEAT-005.

## FEAT-005 — Bank statement import (first: KBC, Belgium)

Introduce a per-import (or per-account) "bank type" so multiple bank
formats can be added over time; first supported importer is KBC
(Belgium). Depends on **FEAT-003** (account number) for matching imported
rows to the right account. Format/parsing details to be provided later —
don't build a generic importer abstraction before the first concrete
format is known.

## FEAT-006 — (Low priority) Desktop app: single Go binary

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
