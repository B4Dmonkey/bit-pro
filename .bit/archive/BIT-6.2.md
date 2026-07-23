---
id: BIT-6.2
title: contradiction forces the real minted ID
status: done
phase: 1
phase_label: Learn the ID you just created
---
## Step 2 (Phase 1 — Learn the ID you just created) — contradiction forces the real minted ID
**Status:** ✅ Done — verified 2026-07-20

Step 1's hardcoded `BIT-1` can't report a second track (`BIT-2`) or a child (`BIT-1.1`).
Contradicting inputs force `create` to print the ID it actually computed.

**Scope:**
- `cmd/task_create.go` — echo the real `id` variable, not a constant.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestTaskCreateCmd_EchoesSecondTrackID` in `cmd/task_create_test.go`
     - **Behavior:** the echoed ID tracks the real minted value across successive tracks.
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "First", "...")`; then `out := mustRun(t, "task", "create", "Second", "-d", "...")`.
     - **Assertions:** `out == "BIT-2\n"`.
     - **Boundary:** count 1 → 2 — the first increment past the hardcoded floor, which is exactly what a constant can't satisfy.
   - [x] `TestTaskCreateCmd_EchoesChildID` in `cmd/task_create_test.go`
     - **Behavior:** a `--parent` bar echoes its dotted child ID so a caller can hang further work off it.
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "Track", "...")`; then `out := mustRun(t, "task", "create", "A bar", "-d", "...", "--parent", "BIT-1")`.
     - **Assertions:** `out == "BIT-1.1\n"`.
     - **Boundary:** the dotted-ID path (parent set) vs. the flat-ID path — the other branch of `create`'s ID minting.
   - [x] Confirm fails: both print `BIT-1\n` from the Step 1 hardcode.

2. **Implement (GREEN):**
   - [x] Replace the hardcoded `"BIT-1"` with the `id` variable already computed above the `Save` call: `fmt.Fprintln(cmd.OutOrStdout(), id)`.

**Claude verifies:**
- [x] `just test`
- [x] `just lint`

**User verifies:**
- [x] none — deterministic.

**Commit (user):** `feat(task): echo the real minted ID for tracks and bars`