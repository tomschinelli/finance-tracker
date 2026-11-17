---
status: accepted
date: 2026-11-03
---

# Backend integration tests via testcontainers

Backend integration tests (`internal/api/*_test.go`) run against a real Postgres in an ephemeral container via `testcontainers-go`, one container per test binary run, isolated per test by truncating all tables; `internal/finance` gets plain unit tests (no DB). Testcontainers is the standard way to test real DB behavior (constraints, cascades, sequences) in Go without touching the dev database or needing a hand-rolled container lifecycle; a shared container + per-test truncate is much faster than one container per test.
