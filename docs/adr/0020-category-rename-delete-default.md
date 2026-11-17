---
status: accepted
date: 2026-11-03
---

# Category rename/delete and protected default category

Categories get a per-record `PATCH /categories/{id}` (rename) and `DELETE /categories/{id}`, plus a protected `is_default` category ("Others", `kind = expense`) that's lazily (re)created on `GET /categories` whenever none exists and can never be deleted. User-requested category management UI needs rename/delete; this narrowly supersedes the "per-record delete endpoints declined in favor of edit-and-reimport" part of ADR-0017 for categories only (not accounts/payments/recurring definitions). Deleting a category still fails with 400 if any payment or recurring definition references it (existing FK + `handleDBError` mapping, no new check needed) — "Others" exists so there's always a safe fallback category to re-point things at first.
