---
status: accepted
date: 2026-11-03
---

# Recurring definition tags as a `text[]` column

`recurring_definitions` gets a `tags text[] not null default '{}'` column; `GET /recurring?tag=<tag>` filters to definitions carrying that exact tag. Simplest way to support user-requested filtering of recurring payments by tag; a plain array column avoids a join table and rename/merge-tag endpoints that nothing has asked for yet — add a `tags` table later only if tag management (rename, autocomplete across definitions) becomes an actual need.
