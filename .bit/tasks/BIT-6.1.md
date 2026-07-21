---
id: BIT-6.1
title: create echoes an ID
status: done
phase: 1
phase_label: Learn the ID you just created
---
## Step 1 (Phase 1 — Learn the ID you just created) — create echoes an ID
**Status:** ✅ Done — verified 2026-07-20

`task create` prints nothing today, so a caller can't learn what it minted. This step makes
`create` echo an ID to stdout. A hardcoded `BIT-1` satisfies it — that's fine for Step 1;
Step 2's contradiction forces the real value.

**Scope:**
- `cmd/task_create.go` — after `s.Save(...)` succeeds, write the minted ID to `cmd.OutOrStdout()`.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestTaskCreateCmd_EchoesMintedID` in `cmd/task_create_test.go`
     - **Behavior:** creating a task reports the ID the caller can now act on, instead of forcing a list-and-diff guess.
     - **Setup:** `initProject(t, "BIT")`, then `out := mustRun(t, "task", "create", "First track", "-d", "...")`.
     - **Assertions:** `out == "BIT-1\n"`.
     - **Boundary:** the first task in an empty store — count 0 → 1, the ID-minting lower bound.
   - [x] Confirm fails: `create` prints nothing, so `out == ""`, not `"BIT-1\n"`.

2. **Implement (GREEN):**
   - [x] In `RunE`, replace the bare `return s.Save(...)` with: `if err := s.Save(&task.Task{...}); err != nil { return err }`, then `_, err := fmt.Fprintln(cmd.OutOrStdout(), "BIT-1"); return err` (hardcoded ID is deliberate — Step 2 replaces it). Add the `fmt` import; capture `cmd` in the `RunE` signature (currently `_`).

**Claude verifies:**
- [x] `just test` — the new test passes; existing create tests (which ignore stdout via `createTask`) stay green.
- [x] `just lint`

**User verifies:**
- [x] Echoing the raw ID + newline (no label, no `Created ` prefix) is the shape a skill should parse.

**Commit (user):** `feat(task): echo the minted ID from task create`