---
description: Pick one item from docs/TODO.md and implement it
argument-hint: (optional) text to narrow down which item
---

Help the user pick exactly one item from the `docs/TODO.md` backlog, then
implement it.

Steps:
1. Read `docs/TODO.md` in full.
2. Parse out each `##` entry (title + `FEAT-NNN` key) in file order.
3. Use AskUserQuestion with one question offering each entry as an option
   (option description = a one-line summary of that entry). If $ARGUMENTS
   is non-empty, use it to highlight the most likely match, but still ask
   for confirmation unless $ARGUMENTS is an exact FEAT-NNN key already in
   the file.
4. Once the user picks an item, check whether its stated dependency (if
   any) is still an open entry elsewhere in the file. If so, tell the user
   and confirm whether to proceed anyway before continuing.
5. Read `docs/ARCHITECTURE.md` and `docs/DECISIONS.md` before touching
   backend/frontend structure, per this repo's CLAUDE.md ground rules.
6. For anything beyond a trivial change, use EnterPlanMode to propose an
   implementation plan for the selected item and get approval before
   writing code.
7. Implement following this repo's ground rules: minimum diff, no
   speculative abstractions, verify before claiming something works.
8. Once verified, update `docs/TODO.md` to remove or strike through the
   completed item, per the file's own "update as items are completed"
   instruction.
9. If implementing it required a new architectural/technical decision, add
   exactly one entry to `docs/DECISIONS.md` per the ground rules — don't
   log things that are already obvious from the code.

Do not start writing code before the user has confirmed which item to work
on.
