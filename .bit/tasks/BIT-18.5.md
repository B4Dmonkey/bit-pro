---
id: BIT-18.5
title: Nothing files itself as archived any more
status: doing
phase: 2
phase_label: Archive is the soft delete
---
## **Verse 2**

With `complete` in place, `bp task archive` is a second name for a third meaning: a command
that says "file this as done" pointing at a folder that means "soft delete". It goes away, so
the only way to file something as done is to say `complete`. Removing it is also what forces
the two remaining test call sites to say which of the two things they actually meant.

## Scope
- `cmd/task_archive.go` — deleted.
- `cmd/task_archive_test.go` — deleted; its two cases are already covered by
  `cmd/task_complete_test.go` (the relocate) and `cmd/task_delete_test.go` (the force path).
- `cmd/task.go` — drop `taskCmd.AddCommand(newTaskArchiveCmd())`.
- `cmd/task_complete_test.go` — the new test.
- `cmd/feedback_add_test.go` — the two tests that drive `task archive` pick the verb they
  meant.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTaskCompleteCmd_ReplacesArchive` in `cmd/task_complete_test.go`
     - **Behavior:** `task archive` no longer does anything — it is not a hidden synonym that
       quietly files a track somewhere, so the verb and the folder can't drift apart again.
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "Done thing", "Finished work.")`;
       `mustRun(t, "task", "update", "BIT-1", "-s", "done")`; then
       `out, err := run(t, "task", "archive", "BIT-1")`.
     - **Assertions:** `.bit/tasks/BIT-1.md` still stats clean — nothing was relocated — and
       `out` does not contain `"archive"`, so the subcommand is gone from the `task` listing.
       `err` is nil: cobra prints help and exits 0 for an unknown subcommand, so the error
       value proves nothing here and asserting on it would be asserting on cobra.
     - **Boundary:** an id that exists and is `done` — precisely the input the removed
       command accepted, so the test can only pass because the command is gone rather than
       because the input was rejected.
   - [ ] Confirm fails: `os.Stat(".bit/tasks/BIT-1.md")` returns `fs.ErrNotExist`, because
     `task archive` still relocates it today.

2. **Implement (GREEN):**
   - [ ] Delete `cmd/task_archive.go` and `cmd/task_archive_test.go`; drop the `AddCommand`
     line in `cmd/task.go`.
   - [ ] `cmd/feedback_add_test.go`: `TestFeedbackAddCmd_AcceptsArchivedTrack` keeps its
     meaning — a note against a *soft-deleted* track — so its setup becomes
     `mustRun(t, "task", "delete", "BIT-1", "--yes")`.
   - [ ] `cmd/feedback_add_test.go`: `TestFeedbackAddCmd_NoteSurvivesTrackArchive` was about
     a *finished* track, so it becomes `TestFeedbackAddCmd_NoteSurvivesTrackCompletion`,
     driving `task complete` and asserting the track landed at `.bit/completed/BIT-1.md`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`
- [ ] `grep -rn "task archive\|newTaskArchiveCmd" cmd/ task/` returns nothing.

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(task)!: remove task archive in favour of task complete`