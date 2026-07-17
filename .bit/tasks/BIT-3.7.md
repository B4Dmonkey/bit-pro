---
id: BIT-3.7
title: A phase is correctable after the fact
status: done
phase: 3
phase_label: A step shows its phase
---
## Step 7 (Phase 3 — A step shows its phase) — A phase is correctable after the fact

**Status:** ✅ Done — verified 2026-07-17

`update` is a separate command with its own flag plumbing; Step 6's create path proves
nothing about it. Same red-green, own commit.

**Scope:**
- `cmd/task_update.go` — `--phase` and `--phase-label` flags

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestTaskUpdateCmd_ChangesPhase`
     - **Behavior:** a step mis-labelled during import can be fixed without recreating it
       — which the Phase 4 import will need.
     - **Setup:** create a bar with `--phase 2 --phase-label "List & read"`, then
       `task update BIT-1.1 --phase 3 --phase-label "Update"`.
     - **Assertions:** `task read BIT-1.1` shows phase 3 / `Update`; title, status, and
       body are unchanged.
     - **Boundary:** phase changes 2→3 while every other field is untouched — the
       `Flags().Changed` boundary the other update flags already rely on.
   - [x] Confirm fails: unknown flag `--phase`

2. **Implement (GREEN):**
   - [x] Add both flags and the matching `cmd.Flags().Changed(...)` guards, following the
         existing `title` / `description` / `status` pattern exactly

**Claude verifies:**
- [x] `just test`
- [x] `just lint`

**Commit (user):** `feat(update): allow correcting a step's phase`