---
status: accepted
date: 2026-11-03
---

# Full JSON snapshot export/import

`GET /data/export` / `POST /data/import` exchange a full JSON snapshot (every row of every table, including ids and timestamps, versioned via a top-level `version` field); import unconditionally replaces ALL data, including `accounts` and `categories`, preserving ids across the round-trip and resetting sequences past the highest imported id. Gives the user a real backup/restore mechanism and a way to hand-edit data (e.g. remove a bad row) outside the DB; preserving ids keeps cross-table references valid without needing an id-remapping pass, and per-record delete endpoints were explicitly declined in favor of this edit-and-reimport workflow.

For categories only, per-record delete was later added (ADR-0020).
