# Architecture

## Overview

A personal expense/budget tracker. Single implicit user, no auth in v1.
A Nuxt 3 frontend calls a Go JSON API; PostgreSQL is the system of record.

```
Browser --(HTTP/JSON)--> Nuxt 3 (frontend) --(HTTP/JSON)--> Go API (backend) --(SQL)--> PostgreSQL
```

Frontend and backend are separate processes/deployables. They communicate
only over HTTP/JSON — no shared code, no shared types beyond what's
hand-kept in sync.

## Backend (Go)

- Go 1.22+ `net/http` with its built-in method+path routing
  (`mux.HandleFunc("POST /accounts", ...)`). No third-party router —
  stdlib is sufficient at this scope.
- `pgx/v5` (`pgxpool`) as the Postgres driver/connection pool.
- `sqlc` generates typed Go query functions from hand-written SQL. No ORM —
  SQL stays explicit and inspectable.
- `golang-migrate` for versioned, SQL-first schema migrations, applied
  automatically on backend startup (see "Migrations" below) — not a
  separate manual step.
- Config via environment variables only (`DATABASE_URL`, `PORT`). No config
  files, no framework.
- JSON in/out. Errors returned as `{"error": "message"}` with a matching
  HTTP status code.

Layout:

```
backend/
  cmd/api/main.go        # entrypoint: wiring, auto-migration, http.Server
  internal/
    api/                 # http handlers + routing
    store/                # sqlc-generated code, queries.sql
    finance/              # domain logic: occurrence generation, balance calc, budget calc
  migrations/             # golang-migrate SQL files, embedded into the binary
  Dockerfile
```

### Migrations

`backend/migrations/*.sql` is embedded into the binary with `go:embed`.
`main.go` runs `golang-migrate` as a library (`database/postgres` +
`source/iofs` over the embedded FS) against `DATABASE_URL` before starting
the HTTP server — Flyway-style, automatic on every startup, whether run
natively or in a container. There is no separate `migrate up` CLI step in
the normal workflow.

## Frontend (Nuxt)

- Nuxt 3, TypeScript (strict mode).
- API base URL injected via Nuxt runtime config
  (`NUXT_PUBLIC_API_BASE`), not hardcoded.
- No UI/component library chosen yet — add one only when a real need
  shows up, not preemptively.
- Flat `pages/`/`components/` structure until it actually needs folders.

## Domain model (v1)

- **Account** — `id, name, starting_balance_cents, currency, created_at`.
  Multiple purpose-based accounts (e.g. Savings, Fix Costs, Living).
  Current balance is **derived**, not a stored mutable counter:
  `starting_balance_cents + Σ(signed amount of paid payments for that
  account) + Σ(balance_adjustments for that account)`. Sign comes from the
  payment's category `kind` (income `+`, expense `-`).
- **Category** — `id, name, kind (income|expense)`.
- **RecurringDefinition** — `id, label, category_id, account_id,
  amount_cents, frequency (monthly|quarterly|yearly), anchor_date, active,
  created_at`. One shape covers rent/bills/salary (monthly) and irregular
  costs (quarterly/yearly) — frequency is just a field, not a separate
  model, since they don't differ structurally.
- **Payment** — `id, recurring_definition_id (nullable), label,
  category_id, account_id, amount_cents, due_date, paid_date (nullable),
  created_at`. The actual ledger: one row per occurrence. `paid_date IS
  NULL` = expected/not yet paid. `recurring_definition_id IS NULL` = ad hoc
  (unplanned, non-repeating cost). Expected occurrences for active
  recurring definitions are generated lazily — whenever an overview
  endpoint is read for a period, missing occurrences in range are inserted
  as unpaid payments (idempotent on `(recurring_definition_id, due_date)`).
  No cron/scheduler.
- **BalanceAdjustment** — `id, account_id, amount_cents (signed), note,
  occurred_on, created_at`. Manual correction/transfer, independent of the
  payment ledger — the override mechanism for when an account's real-world
  balance doesn't match what the ledger implies.

Money is stored as integer cents, never floating point, to avoid rounding
errors — standard practice for financial data.

Computed, not stored:

- **Monthly set-aside budget** = Σ(active monthly-expense definitions) +
  Σ(active quarterly-expense definitions)/3 + Σ(active yearly-expense
  definitions)/12 — the amount to transfer out so fixed + amortized
  irregular costs are covered without touching it.
- **Yearly overview** — this year's yearly-frequency payments, paid vs
  remaining totals, and whether each account's current balance covers its
  remaining yearly cost.

### Explicitly out of scope for v1

Deferred until there's a real need, not designed for speculatively:

- Multi-user accounts / authentication
- Multi-currency conversion
- CSV/bank import or export

## Local development

`docker-compose.yml` runs the full stack — `postgres`, `backend`, and
`frontend` — as a one-command quick start (`docker compose up -d`).
Migrations apply automatically on backend startup either way (see
"Migrations" above), so there's no extra setup step.

For active development, running backend and frontend natively (`go run
./cmd/api`, `npm run dev`) against the Compose Postgres is faster to
iterate on — no image rebuild, hot reload. See `docs/INTELLIJ_SETUP.md` for
IDE run configurations.

## Verifying changes

Before changing state (schema, data, running services), state how success
will be checked — then check it against real output, not assumption:

- Schema change → starting the backend against a fresh database applies
  the migration automatically (check logs / exit code), and
  `psql \d <table>` shows the expected columns.
- API change → a `curl` request against the running server returns the
  expected status/JSON body.
- Frontend change → page renders and the target interaction works in a
  real browser session, not just a type-check.
