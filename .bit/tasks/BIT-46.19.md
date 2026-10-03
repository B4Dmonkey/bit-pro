---
id: BIT-46.19
title: bp remove archives every uncompleted track and soft-deletes the project after confirmation
status: todo
approved: true
phase: 3
phase_label: bp remove soft-deletes a project
---
## **Verse 3**

The new CLI-only command. It runs from the project's folder (same resolver), lists outstanding work, asks for confirmation, archives every track still in `tasks/` (completed tracks already live in `completed/`), then sets the removed flag. It needs Verse 2's pair-aware `relocateInto`. The end-to-end command test forces it.

## Scope
- `project/store.go`: extract `func StoreFor(p Project) (*task.Store, error)` (`store.ProjectDir(p.Code)` + `task.NewProject`) from `OpenStore`, which becomes `Find` + `StoreFor`. `remove` needs both the `Project` (ID, code, path) and its store.
- new `cmd/remove.go` `newRemoveCmd()`: `Use: "remove"`, `cobra.NoArgs`, registered in `newRootCmd`. Flow:
  1. `os.Getwd` → `project.Find` → `project.StoreFor`.
  2. `tasks := store.List()`. Outstanding = every task (track or bar) whose `Status != done`. If any, print `Outstanding work:` followed by one line per task, `  <ID>\t<status>\t<title>`.
  3. Prompt `Remove <CODE> (<path>)? [y/N]: ` and read one line from `cmd.InOrStdin()` (same reader pattern as `readProjectCode`). Anything but `y`/`yes` (case-insensitive, trimmed), including EOF, prints `not removed` and returns nil with nothing changed.
  4. For each top-level task in `tasks` (no `.` in the ID), `store.Relocate(id, true)`. That moves the track and its bars to `archive/tasks/` and keeps their IDs reserved.
  5. Open the db; `SetProjectRemoved(1, p.ID)`; print `removed <CODE> <path>`.
  - Archive before the flag, so a failure part-way leaves an active project that a re-run finishes.
  - research, feedback and the store dir are left alone.
- `cmd/remove_test.go`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestRemoveCmd/archives open tracks and soft deletes` (subtest of a new `TestRemoveCmd`, `cmd/remove_test.go`)
     - **Behavior:** after confirming, the project's open work is archived (not lost), it's marked removed, and its research survives.
     - **Setup:** `dir := initProject(t, "BIT")`. Through the store: `BIT-1` (todo) with bar `BIT-1.1` (todo); `BIT-2` (done) with bar `BIT-2.1` (done), never completed; `BIT-3` with a done bar, then `Complete("BIT-3")`; `WriteResearch("BIT-1", "index", "x", Commit{})`. Run `bp remove` with stdin `"y\n"`.
     - **Assertions:** output contains `Outstanding work:`, a line for `BIT-1` and for `BIT-1.1`, and none for `BIT-2`/`BIT-2.1`; it ends with `removed BIT <canonical dir>`. `tasks/` has no `.json`. `archive/tasks/` holds `BIT-1`, `BIT-1.1`, `BIT-2`, `BIT-2.1` (`.json` and `.md`). `completed/BIT-3.json` is still there. `research/BIT-1/index.md` exists. `project.Load` shows the row with `Removed == true`.
     - **Boundary:** three track states: open, all-done-but-uncompleted (archived, not listed as outstanding), completed (untouched).
   - [ ] Confirm fails: `unknown command "remove"`.

2. **Implement (GREEN):**
   - [ ] `StoreFor`; `cmd/remove.go`; root registration.

3. **More tests (RED → GREEN):** (subtests of `TestRemoveCmd`)
   - [ ] `TestRemoveCmd/decline changes nothing` (a table inside the subtest, rows `explicit no` stdin `"n\n"`, `empty line` stdin `"\n"`, `eof` stdin `""`): output ends `not removed`; `tasks/` unchanged; row not removed. *Boundary:* explicit no, default (empty line), EOF.
   - [ ] `TestRemoveCmd/no outstanding work skips the list`: only done/completed tracks → output has no `Outstanding work:` and still prompts. *Boundary:* zero outstanding.
   - [ ] `TestRemoveCmd/then the folder says removed`: after a confirmed remove, `bp task list` in `dir` → `errors.Is(err, project.ErrRemoved)`; a second `bp remove` → the same error. *Boundary:* the post-removal state seen through the CLI.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `test ! -e ~/.local/share/bit/main.db`

## User verifies
- none here (the verse's end-to-end check is on its last bar)

## Commit (user)
`feat(bit): bp remove archives open work and soft-deletes the project`