---
id: BIT-32.4
title: Contradiction forces the completed bucket
status: todo
approved: true
phase: 1
phase_label: Counts in the DB
---
## **Verse 1**

`Counts()` only ever globs `tasks/`, so archived work is invisible and `Completed` is still the zero Step 2 left behind. A fixture with a track filed under `.bit/completed/` contradicts that.

## Scope
- `task/store.go` — new: enumerate the tracks under `completed/`
- `task/counts.go` — the `completed` bucket
- `task/counts_test.go` — the contradicting test

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestStoreCounts_CountsCompletedTracks` (table-driven subtests)
     - **Behavior:** a track signed off with `bp task complete` is counted as completed and disappears from the other three buckets, so the four columns sum to the project's whole history rather than double-counting.
     - **Setup:** three subtests, each with its own `.bit/` root.
       1. *a completed track and an active one* — save `ACME-1` (approved, `StatusTodo`) under `tasks/`, then write `ACME-2.md` and `ACME-2.1.md` directly into `<root>/completed/` using `(&task.Task{...}).Bytes()` and `os.WriteFile`. `ACME-2` is a completed track, `ACME-2.1` its bar.
       2. *no completed directory* — one approved `StatusTodo` track under `tasks/` and no `completed/` directory at all.
       3. *only completed work* — nothing under `tasks/`, one track under `completed/`.
     - **Assertions:** whole-struct comparison against `{Todo: 1, Completed: 1}`, `{Todo: 1}`, and `{Completed: 1}` respectively.
     - **Boundary:** `ACME-2.1` is the dotted ID inside `completed/` that must not be counted — the same exclusion `tasks/` already applies, at the second location. Subtest 2 puts the `completed/` directory count at 0 by *absence*, which must read as zero rather than as an error, since a project that has never completed a track is the common case. Subtest 3 is the inverse bound: every track archived, nothing active.
   - [ ] Confirm fails: subtest 1 reports `Completed: 0`, want `1`.

2. **Implement (GREEN):**
   - [ ] In `task/store.go`, add a method alongside `List()` that globs `filepath.Join(s.completedDir(), "*.md")` and parses each match, mirroring `List()`'s read-and-`Parse` shape. `filepath.Glob` on a directory that doesn't exist returns no matches and no error, so the absent-`completed/` case needs no special handling.
   - [ ] In `Counts()`, call it and increment `Completed` once per result whose ID is not a bar.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(task): count tracks archived under completed`