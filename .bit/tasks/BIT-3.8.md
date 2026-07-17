---
id: BIT-3.8
title: The phase is visible while scanning
status: done
phase: 3
phase_label: A step shows its phase
---
## Step 8 (Phase 3 — A step shows its phase) — The phase is visible while scanning

**Status:** ✅ Done — verified 2026-07-17

`read` shows one task; `list` is what you scan. A phase you have to open a task to see
isn't the indicator the scope asked for. This changes the list's output format, so the
Step 1/2/5 assertions move with it.

**Scope:**
- `cmd/task_list.go` — render the phase number per row
- `cmd/task_list_test.go` — existing `want` strings pick up the new column

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestTaskListCmd_ShowsPhaseOnBars`
     - **Behavior:** scanning the list tells you which slice each step belongs to, so
       bars stay legible once a track has thirteen of them.
     - **Setup:** BIT-1 (a track), plus two bars with `--phase 1` and `--phase 2`.
     - **Assertions:** the bars' rows carry their phase; BIT-1's row does not.
     - **Boundary:** a track (phase 0) and a bar (phase non-zero) in one listing — both
       states of the same field, side by side.
   - [x] Confirm fails: no phase in list output

2. **Implement (GREEN):**
   - [x] In `newTaskListCmd`'s loop, add the phase to the `Fprintf`, empty when 0
   - [x] Update the `want` strings in `TestTaskListCmd_ShowsNewestFirst`,
         `TestTaskListCmd_OrdersNumericallyNotLexically`,
         `TestTaskListCmd_GroupsBarsUnderTheirTrack`

**Claude verifies:**
- [x] `just test`
- [x] `just lint`

**User verifies:**
- [x] The column layout survives a real listing — 13 bars under BIT-2 is the case that
      matters, and Step 9 is what produces it

**Commit (user):** `feat(list): show each step's phase`