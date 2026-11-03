# IntelliJ setup

Quickest path: `docker compose up -d` from the repo root runs Postgres,
backend, and frontend together with no IDE setup at all. Use the run
configurations below instead when actively developing — native processes
reload faster and let you attach a debugger.

## Plugins

- **Go plugin** — bundled in GoLand; in IntelliJ IDEA Ultimate, install
  "Go" from Settings → Plugins.
- **Node.js** — bundled in IntelliJ IDEA Ultimate/WebStorm.

## Opening the project

Open the repo root (`finance-management/`) as the project. If `backend/`
isn't auto-detected as a Go module, right-click it → "Mark Directory as" →
"Go Module Root" (needs `backend/go.mod` to exist, i.e. after the backend
is scaffolded). If `frontend/` isn't auto-detected as a Node project,
right-click it → "Mark Directory as" → "Node Modules Root"/project root —
or just open `frontend/package.json` once to trigger detection.

## Run configurations

**API server** (Go)
- Run/Debug Configurations → `+` → Go Build
- Run kind: Package
- Package path: `./cmd/api`
- Working directory: `backend/`
- Environment variables: `DATABASE_URL=postgres://finance:finance@localhost:5432/finance?sslmode=disable;PORT=8081`

Requires Postgres running (`docker compose up -d postgres` from the repo
root). Migrations apply automatically on startup — no extra step.

**Frontend dev** (npm)
- Run/Debug Configurations → `+` → npm
- package.json: `frontend/package.json`
- Command: `run`, Scripts: `dev`
- Environment variables: `NUXT_PUBLIC_API_BASE=http://localhost:8081`

Start the API server configuration first (or Compose's `backend` service),
then the frontend one. Nuxt dev server listens on `http://localhost:3000`.

Port 8081 (not the more common 8080) is used for the API because 8080 may
already be taken by other local services on this machine.
