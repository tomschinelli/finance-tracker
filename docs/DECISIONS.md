# Decisions

Essential, hard-to-reverse decisions only. Not a changelog — if it's
obvious from reading the code, it doesn't belong here.

| Date | Decision | Why |
|------|----------|-----|
| 2026-11-03 | Backend: Go, stdlib `net/http`, no web framework | Go 1.22+'s built-in router covers method+path routing; a framework would add a dependency and indirection with no current benefit. |
| 2026-11-03 | DB access: `pgx` + `sqlc`, no ORM | Type-safe query results without ORM magic; SQL stays explicit, reviewable, and easy to reason about. |
| 2026-11-03 | Migrations: `golang-migrate`, SQL-first | De facto standard for Go; plain up/down SQL files, no DSL to learn. |
| 2026-11-03 | Frontend: Nuxt 3 + TypeScript | User requirement. |
| 2026-11-03 | No auth/multi-user in v1 | Single-user personal tool; auth is real complexity (sessions, password storage, authorization checks) that isn't needed yet — add it when a real multi-user need shows up. |
| 2026-11-03 | Docker Compose scope: Postgres only | Go and Nuxt both run fast natively for local dev; containerizing them adds build/rebuild overhead with no benefit at this stage. Compose is a dev convenience, not the deployment path. |
| 2026-11-03 | Money stored as integer cents, not floating point | Avoids floating-point rounding errors in financial amounts — standard practice. |
| 2026-11-03 | v1 scope: accounts, categories, transactions only (no budgets yet) | Budgeting is a distinct feature with its own data model (period, limits, rollover rules); building it speculatively before the core expense-tracking flow exists would be a speculative abstraction. Add when there's a working vertical slice to build it on top of. |
| 2026-11-03 | Budgeting is now in scope; one `recurring_definitions` table for all frequencies (monthly/quarterly/yearly) and kinds (income/expense) | Rent, bills, salary, and irregular costs don't differ structurally — only frequency changes. Two parallel models (e.g. "irregular" vs "monthly") would duplicate the same shape for no benefit. |
| 2026-11-03 | Account current balance is derived (starting balance + paid payments + manual adjustments), never a stored mutable counter | Avoids dual-write drift between a counter and the ledger; stays auditable; a `balance_adjustments` table gives the user an explicit, logged way to override/correct a balance without corrupting the derivation. |
| 2026-11-03 | Expected payment occurrences are generated lazily on read (idempotent upsert when an overview is fetched), not via a scheduled job | No cron/scheduler needed at this scale; keeps the backend stateless between requests and the stack boring. |
| 2026-11-03 | Monthly set-aside budget = fixed monthly expenses + (quarterly expenses / 3) + (yearly expenses / 12) | Gives a single "transfer this much" figure that amortizes irregular costs into a monthly buffer, matching how the user actually budgets. |
| 2026-11-03 | Migrations apply automatically on backend startup (`golang-migrate` as a library over an embedded FS), not via a manual CLI step | Flyway-style, requested directly; removes a manual setup step and keeps a container image self-contained. |
| 2026-11-03 | Docker Compose now also runs `backend` and `frontend` (supersedes "Postgres only") | Added as a one-command quick start; native `go run`/`npm run dev` remains the faster path for active development, so this is additive. |
