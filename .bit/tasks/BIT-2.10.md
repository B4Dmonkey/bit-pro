---
id: BIT-2.10
title: '`bit task update <id> --title` changes one field'
status: done
phase: 3
phase_label: Update
---
## Step 10 (Phase 3 — Update) — `bit task update <id> --title` changes one field
**Status:** ✅ Done — verified 2026-07-16
Adds `bit task update`, built directly on Step 8's shared `Task`/`loadTask`/`.save()`.
Only the `--title` flag exists yet — nothing has forced `--description` or `--status`
into existence.

**Scope:**
- `cmd/task_update.go` — new: `newTaskUpdateCmd()`
- `cmd/task_update_test.go` — new
- `cmd/task.go` — `taskCmd.AddCommand(newTaskUpdateCmd())`

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestTaskUpdateCmd_ChangesTitleOnly` (in `cmd/task_update_test.go`)
     - **Behavior:** proves a task can be corrected in place without recreating it, and
       that an update to one field leaves the others untouched.
     - **Setup:** `init --prefix BIT`; `task create "Old title" --description "Body
       text."`; fresh `rootCmd` with `SetArgs([]string{"task", "update", "BIT-1",
       "--title", "New title"})`.
     - **Assertions:** `err` is `nil`; `loadTask("BIT-1")` gives `Title: "New title"`,
       `Status: "todo"` (unchanged), `Body` still contains `"Body text."` (unchanged),
       `ID: "BIT-1"` (unchanged).
     - **Boundary:** exactly one field flag supplied — proves partial update doesn't
       clobber the fields it wasn't asked to touch.
   - [x] Confirm fails: `unknown command "update" for "bit task"`.

2. **Implement (GREEN):**
   - [x] `cmd/task_update.go`: `newTaskUpdateCmd()`, `Args: cobra.ExactArgs(1)`, a
     `--title`/`-t` string flag. `RunE`: `loadTask(args[0])`; if
     `cmd.Flags().Changed("title")`, set `t.Title` to the flag value; `t.save()`.

**Claude verifies:**
- [x] `go test ./...` passes
- [x] `go vet ./...` passes

**User verifies:**
- [x] none — mechanical field update, matches the scope's Phase 3 description exactly

**Commit (user):** `feat(task-crud): add task update command (title)`