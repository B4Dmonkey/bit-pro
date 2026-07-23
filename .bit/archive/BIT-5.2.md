---
id: BIT-5.2
title: Group tasks into three status columns
status: done
phase: 1
phase_label: flip to a board
---
## Step 2 (Phase 1 — flip to a board) — Group tasks into three status columns
**Status:** ✅ Done — verified 2026-07-19
The board's data shape: every task bucketed by its `status` into To Do / Doing / Done.
Forced by a todo-vs-done contradiction — putting a `todo` task in column 0 could be a
hardcoded "everything in column 0"; a `done` task landing in column 2 forces a real
status→index mapping.

**Scope:**
- `tui/board.go` — new: `boardColumns` (the fixed `{title, status}` triples for To
  Do/`todo`, Doing/`doing`, Done/`done`) and `groupByStatus(tasks []*task.Task)
  [3][]*task.Task` mapping each task to its column by status, dropping any task whose status
  matches no column.
- `tui/model.go` — a `columns [3][]*task.Task` field; `New` calls `groupByStatus` once and
  stores the result (no reload — the same in-memory slice the items came from).
- `tui/board_test.go` — new: `TestGroupByStatus`.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestGroupByStatus` (table)
     - **Behavior:** each task lands in the column its status names, order within a column
       preserved, and an unmapped status is dropped rather than crashing or leaking into a
       column — so the board can only ever show a task where it belongs.
     - **Setup:** a realistic mix — `{ID:"BIT-4",Status:"todo"}`, `{ID:"BIT-2.1",Status:"doing"}`,
       `{ID:"BIT-4.1",Status:"done"}`, `{ID:"BIT-4.2",Status:"done"}`, and one
       `{ID:"BIT-9",Status:"backlog"}` (a status with no column — the scope's named
       deferral).
     - **Assertions:** column 0 IDs `["BIT-4"]`; column 1 `["BIT-2.1"]`; column 2
       `["BIT-4.1","BIT-4.2"]` in that order; `BIT-9` appears in no column.
     - **Boundary:** a status outside the fixed set (`backlog`) — the drop case that a
       hardcoded three-way split would leak; plus a column with >1 card (order observable)
       and a column with 0 cards is exercised by the empty case below.
   - [x] `TestGroupByStatus_Empty`
     - **Behavior:** no tasks yields three empty columns, not a panic — a fresh project.
     - **Setup:** `groupByStatus(nil)`.
     - **Assertions:** all three columns have length 0.
     - **Boundary:** count == 0 — the lower bound.
   - [x] Confirm fails: `groupByStatus` undefined.

2. **Implement (GREEN):**
   - [x] `boardColumns` slice + `groupByStatus` matching each task's `Status` against the
     three column statuses (append to that column, skip on no match); `New` stores
     `groupByStatus(tasks)` in `m.columns`.

**Claude verifies:**
- [x] `just test` green
- [x] `just lint` clean

**Commit (user):** `feat(tui): group tasks into status columns`