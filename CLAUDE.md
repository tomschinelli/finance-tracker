# finance-management

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

## Stack

| Layer    | Choice                                   |
|----------|-------------------------------------------|
| Frontend | Nuxt 3, TypeScript                        |
| Backend  | Go, stdlib `net/http`, no web framework   |
| DB layer | `pgx` + `sqlc` (no ORM)                   |
| Migrations | `golang-migrate`                        |
| Storage  | PostgreSQL 16                             |
| Local dev DB | Docker Compose (Postgres only — not a deployment artifact) |

Rationale for each is in `docs/DECISIONS.md`.

## Repo layout

```
finance-management/
  CLAUDE.md
  docker-compose.yml      # local Postgres only
  docs/
    ARCHITECTURE.md
    DECISIONS.md
  frontend/                # Nuxt 3 app (not created yet)
  backend/                 # Go API (not created yet)
```

## Local development

```
docker compose up -d       # starts Postgres on localhost:5432
```

Backend and frontend run natively (`go run ./cmd/api`, `npm run dev`) once
they exist — they are not containerized for local dev.

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

## Status

Git repo initialized, `main` branch. Docs and local-Postgres scaffolding
only — no application code yet. Backend and frontend are created in a
follow-up step, against this architecture.
