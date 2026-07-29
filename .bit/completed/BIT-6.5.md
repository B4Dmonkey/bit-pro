---
id: BIT-6.5
title: a missing parent refuses to mint
status: done
phase: 4
phase_label: Refuse to orphan a bar
---
## Step 5 (Phase 4 — Refuse to orphan a bar) — a missing parent refuses to mint
**Status:** ✅ Done — verified 2026-07-20

`task create --parent BIT-99` silently mints a stray `BIT-99.1` today, so a typo'd parent
goes unnoticed. This makes child minting check the parent exists and error otherwise, so
the mistake surfaces at the call site and no orphan file is written.

**Scope:**
- `task/store.go` — in `NextChildID`, stat the parent's file before minting; return a clear error if it's absent.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestStoreNextChildID_ErrorsWhenParentMissing` in `task/store_test.go`
     - **Behavior:** minting a child ID under a non-existent parent fails instead of returning `BIT-99.1`.
     - **Setup:** `s := New(t.TempDir())` with no tasks saved; `_, err := s.NextChildID("BIT-99")`.
     - **Assertions:** `errors.Is(err, fs.ErrNotExist)` and `strings.Contains(err.Error(), "BIT-99")`.
     - **Boundary:** parent absent (zero children, zero parent) — the orphan case; mirrors the existing `NextID` / `Load` error-naming tests.
   - [x] `TestStoreNextChildID_MintsWhenParentExists` in `task/store_test.go`
     - **Behavior:** an existing parent still mints its next child, so the guard doesn't over-restrict the happy path.
     - **Setup:** `s := New(t.TempDir())`; `s.Save(&Task{ID: "BIT-1", Title: "seed", Status: "todo"})`; `got, err := s.NextChildID("BIT-1")`.
     - **Assertions:** `err == nil` and `got == "BIT-1.1"`.
     - **Boundary:** parent present, zero existing bars — the lower bound of the valid path the error case must not break.
   - [x] Confirm fails: `NextChildID` doesn't stat the parent, so the missing-parent case returns `"BIT-99.1", nil` — no error.

2. **Implement (GREEN):**
   - [x] At the top of `NextChildID`, before the glob: `if _, err := os.Stat(s.Path(parent)); err != nil { return "", fmt.Errorf("parent %s does not exist: %w", parent, err) }`. (`os` and `fmt` are already imported.)

3. **More tests (RED → GREEN) — command-level surfacing:**
   - [x] `TestTaskCreateCmd_ErrorsOnMissingParent` in `cmd/task_create_test.go`
     - **Behavior:** `create --parent <missing>` fails and writes no task file, so a typo can't produce a silent orphan.
     - **Setup:** `initProject(t, "BIT")` (no BIT-99); `_, err := run(t, "task", "create", "Orphan", "-d", "...", "--parent", "BIT-99")`.
     - **Assertions:** `err != nil`; and `os.Stat(".bit/tasks/BIT-99.1.md")` returns `fs.ErrNotExist`.
     - **Boundary:** the create path (not just the store) rejects the orphan and leaves the store unchanged.
   - [x] Passes once the `NextChildID` guard lands (create returns the error before `Save`). Existing `TestTaskCreateCmd_ParentMintsDottedID` (parent BIT-1 exists) stays green.

**Claude verifies:**
- [x] `just test`
- [x] `just lint`

**User verifies:**
- [x] The error message (`parent BIT-99 does not exist`) is clear enough to read in a skill's shell output.

**Commit (user):** `feat(task): reject task create --parent for a missing parent`