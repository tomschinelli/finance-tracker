---
status: accepted
date: 2026-11-03
---

# `POST /data/clear` wipes every table

`POST /data/clear` deletes every row in every table (payments, recurring definitions, balance adjustments, accounts, categories) and resets sequences, leaving a fully empty database. A partial wipe that kept accounts/categories was confusing ("clear data" left visible accounts behind); a full wipe matches user expectation and mirrors `/data/import`'s clear step, so there's only one definition of "empty" in the codebase.
