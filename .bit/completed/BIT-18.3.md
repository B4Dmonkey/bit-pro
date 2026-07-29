---
id: BIT-18.3
title: A completed ID is never re-minted
status: done
phase: 1
phase_label: Filed as completed
---
## **Verse 1**

`highestReserved` scans `tasks/` and `archive/` only, so the moment a track moves to
`completed/` its number is free again and the next `task create` re-mints it onto different
work. A test that completes `BIT-2` and then asks for the next ID contradicts that: it wants
`BIT-3` and gets `BIT-2`.

## Scope
- `task/store.go` — add `s.completedDir()` to the directory loop in `highestReserved`.
- `task/store_test.go` — two tests, next to the existing `_ReservesArchivedIDs` pair they
  mirror.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestStoreNextID_ReservesCompletedIDs`
     - **Behavior:** a completed track's number is never handed to a later track, so a
       commit message or a note naming `BIT-2` keeps meaning the same work forever.
     - **Setup:** `s := New(t.TempDir())`; `s.Save` `BIT-1` (`todo`) and `BIT-2` (`done`);
       `s.Complete("BIT-2", false)`.
     - **Assertions:** `s.NextID("BIT")` returns `BIT-3`.
     - **Boundary:** the highest number lives *outside* `tasks/` — the case where the active
       directory alone gives the wrong answer (`BIT-2`).
   - [ ] `TestStoreNextChildID_ReservesCompletedChildren`
     - **Behavior:** the same reservation holds for a bar, so a replanned track never reuses
       a bar number a completed bar already had.
     - **Setup:** `s.Save` `BIT-1` (`todo`) and `BIT-1.1` (`done`); `s.Complete("BIT-1.1",
       false)`.
     - **Assertions:** `s.NextChildID("BIT-1")` returns `BIT-1.2`.
     - **Boundary:** a dotted id — the child-ID regex path rather than the track path, both
       of which route through the same `highestReserved` loop.
   - [ ] Confirm fails: `NextID() = "BIT-2", want "BIT-3"` and
     `NextChildID() = "BIT-1.1", want "BIT-1.2"`.

2. **Implement (GREEN):**
   - [ ] `task/store.go`: the loop becomes
     `[]string{s.tasksDir(), s.completedDir(), s.archiveDir()}`. `highestSuffix` already
     tolerates a directory that does not exist (`filepath.Glob` returns no matches and no
     error), so no guard is needed for a project that has never completed anything.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(task): reserve IDs of completed work`