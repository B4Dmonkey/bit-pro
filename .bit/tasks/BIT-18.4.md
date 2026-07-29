---
id: BIT-18.4
title: A note still attaches to a completed track
status: todo
phase: 1
phase_label: Filed as completed
---
## **Verse 1**

`trackExists` looks in `tasks/` and `archive/`, so the moment a track is completed
`feedback add` refuses it — and a completed track is exactly what a retro reads. A test
demanding that note be accepted contradicts the guard and forces it to treat `completed/`
as a real home for a track. This is the last bar of Verse 1, so it carries the verse's
integration check.

## Scope
- `task/feedback.go` — add `s.completedPath(track)` to the path list in `trackExists`.
- `cmd/feedback_add_test.go` — one new test, next to
  `TestFeedbackAddCmd_AcceptsArchivedTrack` which it mirrors.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestFeedbackAddCmd_AcceptsCompletedTrack`
     - **Behavior:** a track's notes stay writable after it is completed, because a note
       about finished work is the note most worth having.
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "Ship the bit plugin", "## Why\n\nThe
       skills only exist in this repo.\n")`; `mustRun(t, "task", "update", "BIT-1", "-s",
       "done")`; `mustRun(t, "task", "complete", "BIT-1")`; then
       `mustRun(t, "feedback", "add", "BIT-1", "-d", firstNote)`.
     - **Assertions:** stdout is `".bit/feedback/BIT-1-001.md\n"`, and the file's contents
       equal `firstNote` byte for byte.
     - **Boundary:** track location — `completed/` rather than `tasks/` or `archive/`, the
       third and last of the directories the guard now accepts.
   - [ ] Confirm fails: `bp feedback add BIT-1 ... returned error: track BIT-1 does not
     exist`.

2. **Implement (GREEN):**
   - [ ] `task/feedback.go`: the slice becomes
     `[]string{s.Path(track), s.completedPath(track), s.archivePath(track)}`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`
- [ ] `just install` — the checks below drive the built `bp`.

## User verifies
- [ ] Whole slice: run `bp task complete BIT-17` in this repo, then `bp task list`. BIT-17
      and its seven bars are gone from the list, `ls .bit/completed/` shows all eight files,
      and `ls .bit/archive/ | wc -l` is unchanged — finished work now files itself somewhere
      other than the soft-delete bin.

## Commit (user)
`feat(feedback): accept a note against a completed track`