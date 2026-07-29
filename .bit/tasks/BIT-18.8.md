---
id: BIT-18.8
title: The README shows three directories
status: todo
phase: 3
phase_label: The session says complete
---
## **Verse 3**

The README teaches a reader the `.bit/` layout and the command list, and both are now wrong:
it shows a two-directory tree, lists `archive` among the subcommands, and has a whole
*Archive instead of destroy* section built on the single destination. Same as the contract
bar — nothing asserts prose, so the checks are greps.

## Scope
- `README.md` — the Storage tree, the command list, the *Archive instead of destroy* feature
  paragraph, and the *Viewing the archive* future-work item.

## Edits
- [ ] Storage tree: three directories — `tasks/` (live work), `completed/` (signed off),
      `archive/tasks/` (soft-deleted). Say that all three reserve their IDs, which is the
      fact the old one-line comment carried.
- [ ] Command list: `create/read/list/update/move/complete/delete`.
- [ ] *Archive instead of destroy* becomes the two-destination story: `task complete` files a
      finished track and its steps under `.bit/completed/`, `task delete` soft-deletes into
      `.bit/archive/tasks/`, both ride the same relocate primitive, and neither frees the ID.
      Keep the two existing facts that still hold — the step drops from its parent's order,
      and a track only relocates once every step is `done`. That guard is absolute for
      `complete`; `delete` keeps its `--force`.
- [ ] *Viewing the archive* future item: it now names two read surfaces that don't exist
      (completed work and the archive). Reword it; don't expand it into a proposal — the
      scope deliberately left the read side out.
- [ ] Leave the `bit …` versus `bp …` naming alone throughout. The README predates the
      rename and fixing it is a separate job.

## Claude verifies
- [ ] `grep -n "task archive" README.md` returns nothing.
- [ ] `grep -q "\.bit/completed" README.md` and `grep -q "archive/tasks" README.md`.

## User verifies
- [ ] none — deterministic.

## Commit (user)
`docs: document completed and archive/tasks in the README`