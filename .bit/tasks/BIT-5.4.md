---
id: BIT-5.4
title: Columns become list components
status: done
phase: 2
phase_label: reads as a board
---
## Step 4 (Phase 2 — reads as a board) — Columns become list components
**Status:** ✅ Done — verified 2026-07-19
The reshape's core: each column is the same `list` the list view uses, not plain text.
Forced by a state contradiction — a test asserting each board column exposes its grouped
tasks as list items (`.Items()`) can't be satisfied by the plain `[3][]*task.Task` field; it
requires real `list.Model`s. This is the load-bearing "columns are the same component" bet
made testable.

**Scope:**
- `tui/model.go` — replace the `columns [3][]*task.Task` field with `boardCols [3]list.Model`.
  In `New`, after `groupByStatus(tasks)`, build each column's list from the shared
  `delegate{}` over that column's slice, mirroring the main list's setup
  (`SetFilteringEnabled(false)`, `SetShowHelp(false)`, plus `SetShowTitle(false)` /
  `SetShowStatusBar(false)` since the title lives in the border frame). Size each column list
  in `layout()` alongside the main list (a third-width inner box).
- `tui/board.go` — `boardView` renders each `boardCols[i].View()` into an equal-width column
  (still plain — no border yet; the frame is Step 5), joined with `lipgloss.JoinHorizontal`.
- `tui/board_test.go` — `TestBoardColumns_FromGrouping`.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestBoardColumns_FromGrouping` (table)
     - **Behavior:** each board column is a `list.Model` populated by the status grouping —
       so the board is genuinely built from the list component, and every downstream feature
       (selection, scrolling, card styling) rides on it rather than a parallel code path.
     - **Setup:** `New` with 1 `todo`, 1 `doing`, 2 `done` tasks —
       `{ID:"BIT-4",Status:"todo"}`, `{ID:"BIT-2.1",Status:"doing"}`,
       `{ID:"BIT-4.1",Status:"done"}`, `{ID:"BIT-4.2",Status:"done"}`.
     - **Assertions:** `len(m.boardCols[0].Items()) == 1`, `[1] == 1`, `[2] == 2`; the first
       item of column 2, asserted as `item`, wraps `BIT-4.1`.
     - **Boundary:** a column with >1 card (column 2) vs. columns with 1 (0/1) — proves
       membership and count flow from grouping into three distinct lists, which one shared
       list or a plain slice can't hold.
   - [x] Confirm fails: `m.boardCols` undefined / `columns` field removed (compile error).

2. **Implement (GREEN):**
   - [x] Add `boardCols [3]list.Model`; build them in `New` from `groupByStatus`; size them
     in `layout()`; `boardView` renders each column's `.View()`. Remove the `columns` field.

**Claude verifies:**
- [x] `just test` green (`TestGroupByStatus`/`_Empty` still pass — `groupByStatus` is
  unchanged, only its consumer moves)
- [x] `just lint` clean

**User verifies:**
- [x] `tab` still shows three columns, now rendered with the list's card styling (indent,
  the `▎` marker on each column's selected card) — the same look as the list pane, in thirds;
  nothing overflows. (Whether *inactive* columns should also show a `▎` is a visual call —
  note it if it reads wrong; suppressing it is a later polish, not this step.)

**Commit (user):** `feat(tui): build board columns from the list component`