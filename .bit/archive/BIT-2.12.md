---
id: BIT-2.12
title: '`bit task delete <id> --yes` removes a task'
status: done
phase: 4
phase_label: Delete
---
## Step 12 (Phase 4 — Delete) — `bit task delete <id> --yes` removes a task
**Status:** ✅ Done — verified 2026-07-16
Adds `bit task delete` with an explicit `--yes` flag that skips confirmation — the
non-interactive path first, mirroring how Step 1 built the flag path before Step 2's
interactive fallback. Without `--yes`, this step returns a plain error; Step 13
contradicts that with real prompting.

**Scope:**
- `cmd/task_delete.go` — new: `newTaskDeleteCmd()`
- `cmd/task_delete_test.go` — new
- `cmd/task.go` — `taskCmd.AddCommand(newTaskDeleteCmd())`

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestTaskDeleteCmd_RemovesFileWithYesFlag` (in `cmd/task_delete_test.go`)
     - **Behavior:** proves a task can be removed, and that `--yes` bypasses
       confirmation for scripted/LLM-driven use — the same non-interactive need `init
       --prefix` served in Phase 1.
     - **Setup:** `init --prefix BIT`; `task create "Throwaway" --description "..."`;
       run `task delete BIT-1 --yes`.
     - **Assertions:** `err` is `nil`; `os.Stat(".bit/tasks/BIT-1.md")` returns an
       `IsNotExist` error.
     - **Boundary:** `--yes` supplied — the skip-confirmation path.
   - [x] Confirm fails: `unknown command "delete" for "bit task"`.

2. **Implement (GREEN):**
   - [x] `cmd/task_delete.go`: `newTaskDeleteCmd()`, `Args: cobra.ExactArgs(1)`, a
     `--yes`/`-y` bool flag. `RunE`: if `yes`, `os.Remove(taskPath(args[0]))`; else
     `return fmt.Errorf("confirmation required; pass --yes or confirm the prompt")`
     (placeholder — Step 13 replaces this branch with real prompting).

3. **More tests (RED → GREEN):**
   - [x] `TestTaskDeleteCmd_ErrorsOnUnknownID`
     - **Behavior:** proves deleting a typo'd ID fails clearly rather than silently
       succeeding or panicking (should already pass — `os.Remove`'s `*PathError`
       propagates — asserted as a regression guard).
     - **Setup:** `init --prefix BIT` only; run `task delete BIT-99 --yes`.
     - **Assertions:** `err` is not `nil`.
     - **Boundary:** an ID with no corresponding file.
   - [x] `TestTaskDeleteCmd_RejectsPathTraversalID`
     - **Behavior:** proves the path-traversal containment Step 9 added to `taskPath`
       also protects the destructive command, not just `read` — this is the scope's
       actual "warns/confirms before removing anything" requirement extended to a
       crafted ID, not just a merely-mistyped one.
     - **Setup:** `dir := t.TempDir()`; `t.Chdir(dir)`; write the same fixture
       `os.WriteFile(filepath.Join(dir, "README.md"), []byte("# real project readme\n"),
       0o644)` Step 9 uses; `init --prefix BIT`; run `task delete "../../README" --yes`.
     - **Assertions:** `err` is not `nil` (should already pass via Step 9's
       `pathologize.Join` containment — asserted here as a regression guard specific to
       the destructive path); `os.ReadFile(filepath.Join(dir, "README.md"))` still
       succeeds and returns the unchanged fixture content afterward.
     - **Boundary:** same traversal-shaped ID as Step 9 (two levels above `tasksDir`),
       exercised against `delete` instead of `read` — proves containment isn't
       read-path-specific.

**Claude verifies:**
- [x] `go test ./...` passes
- [x] `go vet ./...` passes

**User verifies:**
- [x] none yet — the interactive confirmation (the scope's actual "warns/confirms"
  requirement) lands in Step 13; this step alone is just the scripted escape hatch

**Commit (user):** `feat(task-crud): add task delete command with --yes flag`