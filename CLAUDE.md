# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## finance-management

Personal expense/budget tracker. Nuxt 3 (TypeScript) frontend, Go JSON API
backend, PostgreSQL storage. Single user, no auth in v1.

Read `docs/ARCHITECTURE.md` before touching backend or frontend structure.
Read `docs/DECISIONS.md` before re-deciding something that's already settled.

## Ground rules

1. **Verify, don't assume.** Check actual code, docs, or tool output before
   claiming something is true, works, or is done. "Should work" is not a
   result — run it, read the output, then report.
2. **Minimum diff.** Write the smallest change that resolves the task in
   front of you. No unrelated cleanup, no "while I'm here" extras.
3. **No speculative abstractions.** No config flags, interfaces, plugin
   points, or generic helpers for requirements that don't exist yet. Three
   similar lines beat a premature abstraction.
4. **Touch only what you must.** Don't refactor, reformat, or reorganize
   code outside the scope of the current task.
5. **Define success criteria before changing state.** Before writing code
   or running a migration, state in one or two sentences how you'll verify
   it worked (a test, a curl response, a query result). Check it
   afterward and report what you actually observed.
6. **Common patterns over clever ones.** Especially in Go: standard
   library first, a small set of well-known, widely-used packages second,
   no ORM/framework magic. Prefer the boring, idiomatic choice.
7. **Keep decisions documented, keep the log small.** When an architectural
   or technical decision is made, add one entry to `docs/DECISIONS.md`.
   Only essential decisions go there — not implementation detail that's
   obvious from reading the code.
8. **Verify at the lowest cost that proves the fix.** Prefer a `curl`/`psql`
   check or a unit test over a full container rebuild, and a rebuild over a
   browser round-trip. Only drive the browser for UI-specific behavior that
   can't be confirmed another way, and batch those actions into as few
   tool calls as possible instead of one screenshot per step.

## Stack

| Layer    | Choice                                   |
|----------|-------------------------------------------|
| Frontend | Nuxt 3, TypeScript                        |
| Backend  | Go, stdlib `net/http`, no web framework   |
| DB layer | `pgx` + `sqlc` (no ORM)                   |
| Migrations | `golang-migrate`                        |
| Storage  | PostgreSQL 16                             |

Rationale for each is in `docs/DECISIONS.md`.

## Commands

### Local dev — Postgres only (fastest iteration)

```
docker compose up -d postgres     # Postgres on localhost:5432
cd backend && DATABASE_URL=postgres://finance:finance@localhost:5432/finance?sslmode=disable PORT=8081 go run ./cmd/api
cd frontend && NUXT_PUBLIC_API_BASE=http://localhost:8081 npm run dev
```

Migrations apply automatically on backend startup — there is no separate
`migrate up` step. See `docs/INTELLIJ_SETUP.md` for equivalent IDE run
configs.

### Local dev — full stack via Compose (no native toolchains needed)

```
docker compose up -d     # postgres + backend (:8081) + frontend (:3000)
```

Slower to iterate on (image rebuild per change) but useful for a quick
end-to-end check.

There are no bind mounts, so backend/frontend source is baked into the
image at build time — a code change is invisible to a running container
until it's rebuilt. If the full stack is up via Compose, rebuild and
recreate the affected service(s) after implementing a feature, before
verifying it end-to-end:

```
docker compose up -d --build backend frontend
```

### Backend (from `backend/`)

```
go build ./...                    # compile check
go vet ./...
go test ./...                     # unit tests + internal/api integration tests (needs Docker)
go run ./cmd/api                  # requires DATABASE_URL, optional PORT (default 8081)
sqlc generate                     # regenerate internal/store/*.go after editing queries.sql or adding a migration
```

`sqlc generate` reads `sqlc.yaml`: queries from `internal/store/queries.sql`,
schema from `migrations/`, output into `internal/store` (`queries.sql.go`,
`models.go`). Always run it after changing either source — never hand-edit
the generated files.

`internal/api`'s `*_test.go` files are integration tests: `TestMain`
(`testmain_test.go`) starts one ephemeral Postgres container via
testcontainers-go for the whole package run, applies the real embedded
migrations, and each test truncates all tables for isolation (see
`helpers_test.go`). They exercise the real HTTP handlers and a real
Postgres — not mocks — so Docker must be running locally; `internal/finance`
has plain unit tests with no DB dependency. Add new backend tests next to
the code under test, following whichever of these two patterns fits.

To add a schema change: write a new `backend/migrations/NNNN_name.up.sql` +
`.down.sql` pair (sequential number after the last one), then `sqlc generate`.
Migrations are embedded via `go:embed` (`migrations/embed.go`) and applied
automatically by `main.go` on backend startup — no manual `migrate` CLI step
in the normal workflow.

### Frontend (from `frontend/`)

```
npm run dev        # dev server, http://localhost:3000
npm run build       # production build
npm run generate    # static generation
npm run preview     # preview a production build
```

No lint/test scripts are configured yet. TypeScript strict mode is on
(`nuxt.config.ts`); Nuxt's type-check runs as part of `npm run build`.

## Architecture

```
Browser --(HTTP/JSON)--> Nuxt 3 (frontend) --(HTTP/JSON)--> Go API (backend) --(SQL)--> PostgreSQL
```

Frontend and backend are separate processes/deployables, communicating only
over HTTP/JSON. There's no shared code or shared types — `frontend/types/api.ts`
is hand-kept in sync with the Go API's JSON shapes; update both sides
together when a response shape changes.

### Backend layout (`backend/`)

```
cmd/api/main.go     entrypoint: env config, auto-migration, pgxpool, http.Server
internal/api/        http handlers + routing (one file per resource area:
                      accounts.go, categories.go, payments.go, recurring.go,
                      transfers.go, overview.go, admin.go; server.go wires
                      routes, json.go has shared JSON/error helpers)
internal/store/      sqlc-generated code (queries.sql.go, models.go, db.go) —
                      DO NOT hand-edit; regenerate from queries.sql
internal/finance/     domain logic with no HTTP/DB dependency: occurrence
                      generation (occurrence.go), balance calc (balance.go),
                      budget calc (budget.go)
migrations/           golang-migrate SQL files, embedded via go:embed
```

- Routing is `net/http`'s builtin method+path mux (`mux.HandleFunc("POST
  /accounts", ...)`) — no third-party router.
- Config is environment variables only (`DATABASE_URL`, `PORT`); no config
  files.
- JSON in/out; errors are `{"error": "message"}` with a matching HTTP
  status. `handleDBError` in `json.go` maps common Postgres error codes
  (FK violation, unique violation, check violation) to 400s and `pgx.ErrNoRows`
  to 404 — reuse it rather than writing ad hoc error mapping in a handler.
- CORS is wide open (`withCORS` in `server.go`) since there's no auth and
  no credentials/cookies in play.

### Frontend layout (`frontend/`)

Flat `pages/` + `components/` — no nested feature folders yet. `useApi()`
(`composables/useApi.ts`) is the one place that knows the API base URL and
wraps `$fetch`, surfacing the backend's `{"error": ...}` body as a thrown
`Error`. The API base URL differs for SSR vs. browser fetches (`apiBaseServer`
vs. `public.apiBase` in `nuxt.config.ts`) because the Nuxt server process and
the browser resolve `localhost`/`backend` differently in Docker Compose —
always use `useApi()` rather than calling `$fetch` directly with a hardcoded host.

### Domain model (v1)

- **Account** — purpose-based (Savings, Fix Costs, Living, ...). Exactly one
  account is `is_default` (partial unique index) — only it can be funded
  directly from outside the ledger (e.g. salary, via a balance adjustment);
  every other account is funded only via a **Transfer**. Current balance is
  *derived*, never a stored counter: `starting_balance_cents + Σ(signed paid
  payment amounts) + Σ(balance_adjustments)`.
- **Category** — `kind` is `income` or `expense`; determines the sign used
  when deriving balances.
- **RecurringDefinition** — one model covers monthly/quarterly/yearly
  frequency and income/expense kind (rent, bills, salary, irregular costs
  all share this shape).
- **Payment** — the actual ledger, one row per occurrence.
  `recurring_definition_id IS NULL` = ad hoc; `paid_date IS NULL` =
  expected/unpaid. Expected occurrences for active recurring definitions are
  generated **lazily** when an overview endpoint is read for a period
  (idempotent upsert on `(recurring_definition_id, due_date)`) — there's no
  cron/scheduler.
- **BalanceAdjustment** — sign-agnostic manual entry. Direct creation
  (`POST /accounts/{id}/adjustments`) is only valid for the default account;
  the API rejects it on any other account.
- **Transfer** — not a table: `POST /transfers` inserts two
  `balance_adjustments` rows (debit source, credit destination) in one DB
  transaction. This is how the monthly budget set-aside funds a target
  account, and the only way a non-default account's balance changes.

Money is always integer cents, never floating point.

Computed, not stored: the monthly set-aside budget (fixed monthly expenses +
quarterly/3 + yearly/12) and the yearly overview (paid vs. remaining,
whether each account's balance covers remaining yearly cost). See
`internal/finance/budget.go` and `docs/ARCHITECTURE.md` for the exact
formulas and the "predicted costs vs. buffer" split.

Out of scope for v1 (don't design for these speculatively): multi-user/auth,
multi-currency conversion, CSV/bank import/export.

## Commit messages

[Conventional Commits](https://www.conventionalcommits.org/), because the
type maps directly to a semver bump:

```
<type>[optional scope][!]: <summary>

[optional body]

[optional footer(s)]
```

| Type | Semver bump | Use for |
|------|-------------|---------|
| `fix` | PATCH | Bug fix |
| `feat` | MINOR | New capability |
| `feat!` / `fix!` / footer `BREAKING CHANGE: ...` | MAJOR | Incompatible change (API shape, schema, config) |
| `docs`, `chore`, `refactor`, `test`, `build`, `ci`, `style`, `perf` | none | No version bump |

Examples:

```
feat(backend): add POST /transactions endpoint
fix(frontend): correct amount sign on transfer rows
feat(store)!: rename accounts.balance to accounts.balance_cents

BREAKING CHANGE: column renamed; requires migration 0003.
```

Scope is the touched area (`backend`, `frontend`, `store`, `docs`, ...).
One logical change per commit.
