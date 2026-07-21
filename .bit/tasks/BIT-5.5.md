---
id: BIT-5.5
title: Framed titled columns with card counts
status: done
phase: 2
phase_label: reads as a board
---
## Step 5 (Phase 2 — reads as a board) — Framed titled columns with card counts
**Status:** ✅ Done — verified 2026-07-19
Each column becomes a titled, bordered box whose header shows its live count (`To Do (4)`),
all three equal width, column 0 accented — reusing the list view's `titledBorder`. Forced by
a count contradiction — a hardcoded header can't show `To Do (4)` and `Done (2)` at once; the
count must come from `len(boardCols[i].Items())`.

**Scope:**
- `tui/board.go` — `boardView` wraps each column in
  `titledBorder(boardCols[i].View(), fmt.Sprintf("%s (%d)", boardColumns[i].title, len(boardCols[i].Items())), colW-2, m.height-2, i == 0)`,
  columns given equal width (`colW ≈ m.winWidth/3`, remainder absorbed so the row never
  exceeds `winWidth`), joined with `lipgloss.JoinHorizontal`. Column 0's border is accented
  via `titledBorder`'s `active` flag; Step 6 makes that follow `activeCol`.
- `tui/board_test.go` — `TestView_BoardColumnCounts`.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestView_BoardColumnCounts` (table)
     - **Behavior:** each column header names the column and its live card count, so the
       board is legible at a glance without counting cards.
     - **Setup:** `New` with 4 `todo`, 1 `doing`, 2 `done` tasks; `WindowSizeMsg{120,40}`;
       `tab` into board mode; render `View`.
     - **Assertions:** `View()` contains `"To Do (4)"`, `"Doing (1)"`, and `"Done (2)"`.
     - **Boundary:** counts driven by `len(boardCols[i].Items())` at distinct values
       (4/1/2) — a single hardcoded header can't satisfy all three.
   - [x] Confirm fails: board view has no `(N)` count text.

2. **Implement (GREEN):**
   - [x] Route each column through `titledBorder` with the `title (count)` header and equal
     width; accent column 0 via `titledBorder`'s `active` flag.

**Claude verifies:**
- [x] `just test` green (`TestView_FitsWindowHeight` behaviour holds for the board — nothing
  overflows the height budget), `just lint` clean

**User verifies:**
- [x] the three columns are titled, bordered, visibly equal in width, and each header shows
  its count; column 0's border is accented
- [x] the layout reads like the reference board and nothing overflows

**Commit (user):** `feat(tui): frame the board columns with titles and counts`