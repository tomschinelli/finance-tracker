# Parked — not ready to grill

FEAT-005 waits on the KBC format details; FEAT-006 is low priority (likely `/wayfinder` when its time comes).

## Bank statement import, first: KBC Belgium (FEAT-005)

Introduce a per-import (or per-account) "bank type" so multiple bank
formats can be added over time; first supported importer is KBC
(Belgium). Accounts now carry an `account_number` to match imported rows
to the right account. Format/parsing details to be provided later — don't
build a generic importer abstraction before the first concrete format is
known.

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
