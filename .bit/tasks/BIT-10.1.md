---
id: BIT-10.1
title: A task relocates out of the active list
status: done
phase: 1
phase_label: Archive
---
Introduce the `archive/` sibling folder and a single-file move, and prove a relocated file drops out of `List()`. Walking skeleton: one childless task, no cascade yet.

**Scope:**
- `task/store.go` — add `archiveSubdir` const, `archiveDir()`, `archivePath(id)` (via `pathologize.Join`, mirroring `Path`); unexported `relocate(id)` (single-file `os.Rename` after `os.MkdirAll(archiveDir)`, errors wrapped with the id); exported `Relocate(id string, force bool) error` calling `relocate` for the one id (`force` unused until bar 1.4 — an unused param is legal Go).
- `task/store_test.go` — new tests (`package task`, so unexported helpers are reachable).

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestStoreRelocate_MovesFileOutOfList`
     - **Behavior:** a relocated task leaves both the tasks dir and the `List()` result, and lands in `archive/`.
     - **Setup:** `s := New(t.TempDir())`; `s.Save(&Task{ID:"BIT-1", Status:"done"})`; `s.Relocate("BIT-1", false)`.
     - **Assertions:** no task with ID `BIT-1` in `List()`; `os.Stat(<root>/archive/BIT-1.md)` succeeds; `os.Stat(<root>/tasks/BIT-1.md)` is `fs.ErrNotExist`.
     - **Boundary:** a single task, zero children — the smallest unit the move applies to.
   - [ ] `TestStoreRelocate_ContainsUntrustedID` (table, mirrors `TestStorePath_ContainsUntrustedID`)
     - **Behavior:** the archive destination cannot escape `archive/` via a traversal id.
     - **Setup:** ids `../../README`, `/etc/passwd`, `a:b*c`.
     - **Assertions:** `archivePath(id)` stays prefixed with `<root>/archive/`.
     - **Boundary:** hostile id — the traversal case (fileflow-pathologize).
   - [ ] Confirm fails: `Relocate`/`relocate` undefined (compile error).

2. **Implement (GREEN):**
   - [ ] Add `archiveDir()`/`archivePath()`, `relocate(id)` (MkdirAll + os.Rename, wrapped), and `Relocate(id, force)` moving just that id.

**Claude verifies:**
- [ ] tests pass (`just test`)
- [ ] linter passes (`just lint`)

**Commit (user):** `feat(task): relocate a task file into .bit/archive`