---
id: BIT-18.6
title: Soft deletes land where .bit puts tasks
status: done
phase: 2
phase_label: Archive is the soft delete
---
## **Verse 2**

`delete` now owns `archive/` alone, and it writes flat into it. Laying it out as
`archive/tasks/` mirrors `.bit/` itself, so a future `.bit/archive/feedback/` needs no new
rule. This is the last bar of Verse 2, so it carries the verse's integration check.

**What this bar stops scanning.** Until now `highestReserved` reads the flat
`.bit/archive/*.md`; afterwards it reads `.bit/archive/tasks/` instead. The one-time move of
each project's flat archive runs at the very end of the track, so between installing this
`bp` and running that move a project's flat archive is unscanned — `bp task create` there
mints from `tasks/` and `completed/` alone. In this repo BIT-18 is already the highest number,
so nothing collides. Elsewhere, do the move before creating new work in that project.

## Scope
- `task/store.go` — `archiveDir()` becomes `archiveTasksDir()`, joining
  `archiveSubdir` and `tasksSubdir`; its three call sites follow.
- `task/store_test.go` — the destination strings in
  `TestStoreRelocate_ContainsUntrustedID`.
- `cmd/task_delete_test.go` — the two `.bit/archive/…` destination assertions.

## TDD cycle

1. **Write test (RED):** the behavior change *is* the destination, so this step edits the
   assertions that name it rather than adding a test that duplicates them.
   - [ ] `cmd/task_delete_test.go`: in `TestTaskDeleteCmd_RelocatesInsteadOfDestroying` and
     `TestTaskDeleteCmd_ForceDeletesUnfinished`, the destination becomes
     `.bit/archive/tasks/<id>.md`.
     - **Behavior:** a soft-deleted task lands where the rest of `.bit/` would put a task —
       under a `tasks/` directory — so archive can shadow the whole of `.bit/` later.
     - **Setup:** unchanged.
     - **Assertions:** `.bit/archive/tasks/BIT-1.md` (and `BIT-1.1.md` in the force case)
       stat clean; `.bit/tasks/…` gives `fs.ErrNotExist` as before.
     - **Boundary:** a track with one bar in the force case — the cascade writes both files
       into the new directory, not just the track.
   - [ ] `task/store_test.go`: `TestStoreRelocate_ContainsUntrustedID`'s `want` values become
     `.bit/archive/tasks/…`, and its `strings.HasPrefix` guard tightens to
     `.bit/archive/tasks/` — the loose `.bit/archive/` prefix would still pass and so would
     stop proving containment.
     - **Boundary:** the traversal and absolute-path ids, one directory deeper than before —
       the containment guarantee has to hold at the new depth, not just the old one.
   - [ ] Confirm fails: `os.Stat(".bit/archive/tasks/BIT-1.md")` returns `fs.ErrNotExist`,
     and `archivePath("BIT-1") = ".bit/archive/BIT-1.md", want ".bit/archive/tasks/BIT-1.md"`.
   - [ ] The remaining `s.archivePath(id)` assertions in `task/store_test.go` follow the
     implementation and need no edit — leave them alone.

2. **Implement (GREEN):**
   - [ ] `task/store.go`: rename `archiveDir()` to `archiveTasksDir()` and return
     `filepath.Join(s.root, archiveSubdir, tasksSubdir)`. The private name matters here — an
     `archiveDir` that isn't the archive directory is the kind of drift this whole track is
     about. Update `archivePath`, `Relocate`, and the `highestReserved` loop.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`
- [ ] `just install`

## User verifies
- [ ] Whole slice, in a scratch directory so nothing here is polluted:
      `mkdir /tmp/bitcheck && cd /tmp/bitcheck && git init && bp init --prefix TMP`, then
      `bp task create Keep`, `bp task create Toss`, `bp task update TMP-1 -s done`,
      `bp task complete TMP-1`, `bp task delete TMP-2 -y`. `find .bit -name '*.md'` shows
      `TMP-1` under `.bit/completed/` and `TMP-2` under `.bit/archive/tasks/` — two verbs,
      two visibly different destinations, and archive laid out like `.bit/` itself.

## Commit (user)
`feat(task): file soft-deleted tasks under .bit/archive/tasks`