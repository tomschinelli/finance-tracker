---
status: accepted
date: 2026-11-03
---

# Migrations apply automatically on backend startup

Migrations apply automatically on backend startup (`golang-migrate` as a library over an embedded FS), not via a manual CLI step. Flyway-style, requested directly; removes a manual setup step and keeps a container image self-contained.
