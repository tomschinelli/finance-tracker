---
description: Add a new backlog item to docs/TODO.md
argument-hint: <short description of the task>
---

Record a new backlog item in `docs/TODO.md` for: $ARGUMENTS

Do NOT implement anything — this command only records the task.

Steps:
1. Read `docs/TODO.md` in full, including its header notes.
2. Determine the next `FEAT-NNN` key: highest existing number + 1, zero-padded
   to 3 digits.
3. Write a new `##` section matching the exact style of existing entries: a
   short, readable title ending in `(FEAT-NNN)`, then 1-3 sentences
   describing the task. If it depends on, or is a dependency of, another
   backlog item, say so explicitly by that item's FEAT key (same way
   existing entries cross-reference each other).
4. Decide where the section belongs: the file is "ordered for
   implementation... work top to bottom" per its header. Insert the new
   entry in the position that respects any dependency you noted in step 3,
   not necessarily at the end. If it has no dependencies and nothing
   depends on it, append it just before the closing "Desktop app" entry
   (which the file's own closing note says should stay last).
5. Match the file's existing voice and brevity — don't over-explain.
6. After editing, show the user the new section you added.
