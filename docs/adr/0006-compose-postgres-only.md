---
status: superseded by ADR-0014
date: 2026-11-03
---

# Docker Compose scope: Postgres only

Docker Compose runs Postgres only. Go and Nuxt both run fast natively for local dev; containerizing them adds build/rebuild overhead with no benefit at this stage. Compose is a dev convenience, not the deployment path.
