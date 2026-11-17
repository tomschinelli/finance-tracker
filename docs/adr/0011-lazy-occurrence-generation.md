---
status: accepted
date: 2026-11-03
---

# Expected payment occurrences generated lazily on read

Expected payment occurrences are generated lazily on read (idempotent upsert when an overview is fetched), not via a scheduled job. No cron/scheduler needed at this scale; keeps the backend stateless between requests and the stack boring.
