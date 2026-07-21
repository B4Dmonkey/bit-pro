---
id: BIT-5.7
title: '`↑`/`↓` select a card in the active column'
status: done
phase: 3
phase_label: drive the board
---
## Step 7 (Phase 3 — drive the board) — `↑`/`↓` select a card in the active column
**Status:** ✅ Done — verified 2026-07-19
The other axis: `↑`/`↓` move the selected card inside the active column by forwarding to
that column's `list.Model` — clamping, scrolling, and per-column independence all come from
the list itself. Forced by a clamp-plus-independence contradiction — switching columns must
preserve each column's own selection, which only per-column list cursors (not one shared
index) can hold. This is where reusing the list component pays off: no `colCursor` arithmetic.

**Scope:**
- `tui/board.go` — `updateBoard` on `KeyUp`/`KeyDown` forwards to the active column's list:
  `m.boardCols[m.activeCol], cmd = m.boardCols[m.activeCol].Update(msg)`; returns `m, cmd`.
  No new field — the `list.Model` owns its cursor.
- `tui/board_test.go` — `TestUpdate_BoardCardSelection`.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestUpdate_BoardCardSelection` (table)
     - **Behavior:** `↑`/`↓` move the selection within the active column only, clamped to
       its cards, and switching columns preserves each column's own selection — so
       navigating one column never disturbs another.
     - **Setup:** `New` with a `doing` task and three `done` tasks (column 1 has 1 card,
       column 2 has 3); `tab` into board mode. Case A: on column 0 (empty) `↓` stays 0.
       Case B: `→→` to column 2, `↓` → `boardCols[2].Index() == 1`, `↓↓` clamps at 2.
       Case C: on column 2 scroll to card 1, `←←` to column 0 then `→→` back → still card 1.
     - **Assertions:** A `boardCols[0].Index() == 0`; B `boardCols[2].Index()` goes 0→1→2
       and clamps at 2; C column 2's index survives a round-trip through other columns.
     - **Boundary:** empty column (len 0 — the list keeps index 0, no negative index), last
       card (the list clamps at `len-1`), and the per-column independence a single cursor
       can't hold.
   - [x] Confirm fails: `↑`/`↓` don't move `boardCols[activeCol].Index()` (updateBoard
     currently ignores them).

2. **Implement (GREEN):**
   - [x] Forward `↑`/`↓` to the active column's list in `updateBoard`.

**Claude verifies:**
- [x] `just test` green
- [x] `just lint` clean

**User verifies:**
- [x] in board mode `↑`/`↓` move the card marker within the active column and clamp at top
  and bottom; moving to another column and back keeps each column's own selection

**Commit (user):** `feat(tui): select a card within the active board column`