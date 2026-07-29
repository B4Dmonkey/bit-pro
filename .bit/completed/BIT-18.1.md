---
id: BIT-18.1
title: A signed-off track files itself under completed
status: done
phase: 1
phase_label: Filed as completed
---
## **Verse 1**

`bp task complete` does not exist, so a signed-off track has nowhere to go but the bin
soft-deleted work uses. This step adds the command and the second destination, reusing the
relocate machinery `Relocate` already has (the all-bars-`done` guard, the cascade onto bars,
dropping a bar from its parent's order).

## Scope
- `task/store.go` — add `completedSubdir`, `completedDir()`, `completedPath(id)`; turn
  `relocate(id)` into `relocateInto(dir, id)` and `Relocate`'s body into
  `relocateTree(dir, id, force)`; add `Complete(id string) error`. `Relocate` keeps its
  archive destination — `delete` still uses it.
- `cmd/task_complete.go` — new; same shape as `cmd/task_archive.go`.
- `cmd/task.go` — register `newTaskCompleteCmd()`.
- `cmd/task_complete_test.go` — new.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTaskCompleteCmd_FilesTrackAndBarsUnderCompleted` in `cmd/task_complete_test.go`
     - **Behavior:** signing a track off files the whole track — the track file and every
       bar — under `.bit/completed/`, and puts nothing in the archive.
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "Separate completed work",
       "## Why\n\nFinished work needs its own home.\n")`; two bars via
       `mustRun(t, "task", "create", "<name>", "--parent", "BIT-1", "--description", "...")`;
       `-s done` on `BIT-1.1`, `BIT-1.2`, then `BIT-1`; then
       `mustRun(t, "task", "complete", "BIT-1")`.
     - **Assertions:** `.bit/completed/BIT-1.md`, `.bit/completed/BIT-1.1.md`,
       `.bit/completed/BIT-1.2.md` each stat clean; the same three under `.bit/tasks/` each
       give `fs.ErrNotExist`; `.bit/archive` itself gives `fs.ErrNotExist`.
     - **Boundary:** bar count == 2 — one above the single-child lower bound, so the cascade
       loop runs more than once. The archive-does-not-exist assertion is the discriminator:
       an implementation that reused the existing destination passes every other assertion
       and fails that one.
   - [ ] Confirm fails: `unknown command "complete" for "bp task"` — note that cobra prints
     help and exits 0 for an unknown subcommand, so the visible failure is the
     `.bit/completed/BIT-1.md` stat returning `fs.ErrNotExist`.

2. **Implement (GREEN):**
   - [ ] `task/store.go`: `completedSubdir = "completed"` with the other subdir constants;
     `completedDir()` and `completedPath(id)` mirroring `archiveDir()`/`archivePath(id)` —
     `completedPath` goes through `pathologize.Join` like the other two, since the id is
     untrusted.
   - [ ] `task/store.go`: `relocateInto(dir, id string)` does the MkdirAll + rename of
     `s.Path(id)` to `pathologize.Join(dir, id+".md")`; `relocateTree(dir, id string, force
     bool)` holds what `Relocate` does today. `Relocate(id, force)` becomes
     `relocateTree(s.archiveDir(), id, force)`.
   - [ ] `task/store.go`: `Complete(id string) error` →
     `relocateTree(s.completedDir(), id, false)`. No `force` parameter — completing is a
     sign-off on finished work, so the all-bars-`done` guard is the point of it and there is
     no override. Mark the bars `done` first.
   - [ ] `cmd/task_complete.go`: `Use: "complete <id>"`,
     `Short: "Complete a task, filing it and its bars under .bit/completed/"`,
     `Args: cobra.ExactArgs(1)`, `RunE` → `task.New(bitDir).Complete(args[0])`. No flags.
   - [ ] `cmd/task.go`: `taskCmd.AddCommand(newTaskCompleteCmd())`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(task): file a signed-off track under .bit/completed`