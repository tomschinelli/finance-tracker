# Glossary alignment — no PRD, `/implement` directly

## Align UI labels with the glossary (FEAT-014)

Rename user-facing labels to the `GLOSSARY.md` terms, no logic change:
"default" badge → Main account; "Predicted costs" → Fixed costs; "Add
budget to account" → set aside Monthly budget; "Unpaid" → Expected; "Add
recurring cost" / "Recurring definitions" → Recurring item; "Add ad hoc
cost" → Ad hoc payment.

## Docs cleanup after glossary (CHORE-001)

- Trim `CLAUDE.md` "Domain model (v1)" to a pointer at `GLOSSARY.md`, keep
  code-level rules (integer cents, derived balance, lazy occurrence
  generation); it still says only the default account takes direct
  adjustments, reversed by ADR 0023.
- Add `GLOSSARY.md` link under "Docs" in `README.md`.
- Mark ADR 0019 superseded (in part) by 0023, and 0012 by 0025.
