---
id: BIT-10.6
title: Delete relocates instead of destroying
status: done
phase: 2
phase_label: Non-destructive delete
---
Point `bit task delete` at `Relocate` so a deleted task survives in `archive/` instead of being destroyed. Contradicts the current `os.Remove`.

**Scope:**
- `cmd/task_delete.go` — `RunE` calls `task.New(bitDir).Relocate(id, false)` instead of `.Delete(id)`; keep the confirmation prompt.
- `task/store.go` — remove `Store.Delete` (its only caller was this command).
- `cmd/task_delete_test.go` — add the survival assertion; `task/store_test.go` — drop/migrate the `Delete` test.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestTaskDeleteCmd_RelocatesInsteadOfDestroying`
     - **Behavior:** a confirmed delete leaves the file recoverable in `archive/`.
     - **Setup:** `initProject`; `createTask`; `mustRun(t,"task","update","BIT-1","-s","done")`; `mustRun(t,"task","delete","BIT-1","--yes")`.
     - **Assertions:** `.bit/archive/BIT-1.md` exists; `.bit/tasks/BIT-1.md` gone.
     - **Boundary:** happy delete of a childless done task.
   - [ ] Confirm fails: `os.Remove` destroys the file → `archive/BIT-1.md` is `fs.ErrNotExist`.

2. **Implement (GREEN):**
   - [ ] Swap the call; delete `Store.Delete` and its store test.

Keep the existing delete suite green: `TestTaskDeleteCmd_ErrorsOnUnknownID` and `_ContainsPathTraversalID` expect an error wrapping `fs.ErrNotExist` that names the id. `os.Rename` on a missing source returns a `*LinkError` wrapping `ENOENT` (so `errors.Is(err, fs.ErrNotExist)` holds); ensure `relocate` wraps it as `"...%s...: %w"` with the id so `strings.Contains(err.Error(), "BIT-99")` still passes. Traversal: `../../README` → source `tasks/README.md` (absent) fails before the real README is touched.

**Claude verifies:**
- [ ] tests pass, including the existing delete suite (`just test`)
- [ ] linter passes (`just lint`)

**Commit (user):** `feat(cmd): delete relocates to archive instead of removing`