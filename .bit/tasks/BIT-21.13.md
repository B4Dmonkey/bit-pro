---
id: BIT-21.13
title: Lowercase parent no longer destroys an existing bar
status: todo
phase: 2
phase_label: Recurrence
---
## **Verse 2**

The most destructive symptom, measured during planning: `bp task create --parent bit-1` minted
`bit-1.1` even though `BIT-1.1` already existed, then wrote over it. The file kept its original
name and gained a different task's contents — the original bar was destroyed with exit 0 and no
output. `NextChildID`'s existence check passes (the filesystem is case-insensitive) while its
glob misses (the glob is not), which is what makes the collision invisible.

## Scope
- `task/store.go` — `NextChildID`, and the `parent`/`anchor` arguments of `InsertAfter`,
  `AppendToOrder` and `Move`, all of which `cmd/task_create.go` passes straight through.
- `cmd/task_create_test.go` — the new test.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTaskCreateCmd_LowercaseParentDoesNotDestroyAnExistingBar`
     - **Behavior:** minting a child ID counts the bars that exist, whatever case the parent was
       given in, so a new bar never lands on an occupied ID.
     - **Setup:** `initProject(t, "BIT")`; create track `BIT-1`; create bar `BIT-1.1` with body
       `ORIGINAL BAR ONE`. Run `task create "sneaky" --parent bit-1`.
     - **Assertions:** stdout is `BIT-1.2`; `.bit/tasks/BIT-1.2.md` exists and its frontmatter
       reads `id: BIT-1.2`; `.bit/tasks/BIT-1.1.md` still contains `ORIGINAL BAR ONE` and its
       title is unchanged; `.bit/tasks/` holds exactly three files.
     - **Boundary:** one existing child — the lower bound at which a missed glob collides.
       Asserting the *surviving contents* of `BIT-1.1.md`, not just its presence, is what catches
       this: the destructive case leaves the file in place with the wrong body inside.
   - [ ] Confirm fails: stdout is `bit-1.1`, `.bit/tasks/` still holds two files, and
     `BIT-1.1.md` now contains the new task with `id: bit-1.1`

2. **Implement (GREEN):**
   - [ ] Apply `normalizeID` to `parent` in `NextChildID` before the `Stat`, the glob and the
     regex, and mint the returned ID from the normalized parent.
   - [ ] Apply it to `parent`/`id`/`anchor` in `InsertAfter`, `AppendToOrder` and `Move` so the
     order list a new bar is spliced into is keyed consistently.

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` passes

## User verifies
- [ ] none — deterministic

## Commit (user)
`fix(task): stop a wrong-case parent from overwriting a bar`