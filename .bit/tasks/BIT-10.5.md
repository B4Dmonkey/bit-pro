---
id: BIT-10.5
title: bit task archive wires the command
status: todo
phase: 1
phase_label: Archive
---
Wire `bit task archive <id>` (with `--force`) onto `Store.Relocate` — the CLI face of the archive verse.

**Scope:**
- `cmd/task_archive.go` (new) — `newTaskArchiveCmd()`, `Use: "archive <id>"`, `cobra.ExactArgs(1)`, `--force`/`-f` bool; `RunE` calls `task.New(bitDir).Relocate(id, force)`. Mirror `newTaskDeleteCmd`'s shape.
- `cmd/task.go` — register the subcommand alongside delete.
- `cmd/task_archive_test.go` (new).

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestTaskArchiveCmd_RelocatesTask`
     - **Behavior:** `bit task archive` moves a task into `archive/`.
     - **Setup:** `initProject(t,"BIT")`; `createTask(t,"Done thing","...")`; `mustRun(t,"task","update","BIT-1","-s","done")`; `mustRun(t,"task","archive","BIT-1")`.
     - **Assertions:** `.bit/archive/BIT-1.md` exists; `.bit/tasks/BIT-1.md` is `fs.ErrNotExist`.
     - **Boundary:** happy path, no children.
   - [ ] `TestTaskArchiveCmd_ForceArchivesUnfinished`
     - **Behavior:** `--force` archives a track with an unfinished bar.
     - **Setup:** `createTask` BIT-1; `mustRun(t,"task","create","Bar","--parent","BIT-1", ...)` (leaves `BIT-1.1` todo); `mustRun(t,"task","archive","BIT-1","--force")`.
     - **Assertions:** both `BIT-1.md` and `BIT-1.1.md` in `archive/`; none in `tasks/`.
     - **Boundary:** proves `--force` → arg wiring (the guard itself is proven at store level).
   - [ ] Confirm fails: command `archive` unknown.

2. **Implement (GREEN):**
   - [ ] Add the command + `--force`/`-f`; register under the task command.

**Claude verifies:**
- [ ] tests pass (`just test`)
- [ ] linter passes (`just lint`)
- [ ] `just build` then `bit task archive --help` shows the command and `--force`

**User verifies:**
- [ ] the `archive` command reads naturally next to `delete`

**Commit (user):** `feat(cmd): add task archive`