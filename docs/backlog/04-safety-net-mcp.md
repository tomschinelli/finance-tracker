# Safety net + MCP

FEAT-007 blocks FEAT-008.

## Point-in-time snapshot / rollback (FEAT-007)

Add a way to snapshot the full database state and restore (roll back) to
a prior snapshot. Builds on the existing `GET /data/export` / `POST
/data/import` full-JSON-snapshot mechanism — likely wraps it with stored,
timestamped snapshots (e.g. a `snapshots` table or on-disk storage) and a
restore endpoint, rather than inventing a new backup format. Needed as a
safety net before the MCP API goes live, so a bad write made through an
external AI tool can be undone.

No dependencies, but must land before the MCP API (FEAT-008).

## MCP API for AI tool access (FEAT-008)

Expose an MCP (Model Context Protocol) server/API so an external AI
assistant can read and write data directly — e.g. importing transactions
from a custom spreadsheet. Depends on **point-in-time snapshot/rollback**
(FEAT-007) landing first, since writes made by an external AI need to be
undoable. Scope (which resources/operations are exposed, auth/access
control for the MCP endpoint) to be defined when we get here.
