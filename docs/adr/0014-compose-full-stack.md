---
status: accepted
date: 2026-11-03
---

# Docker Compose also runs backend and frontend

Docker Compose now also runs `backend` and `frontend` (supersedes ADR-0006). Added as a one-command quick start; native `go run`/`npm run dev` remains the faster path for active development, so this is additive.
