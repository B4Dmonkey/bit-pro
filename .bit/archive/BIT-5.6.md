---
id: BIT-5.6
title: '`←`/`→` move the active column'
status: done
phase: 3
phase_label: drive the board
---
## Step 6 (Phase 3 — drive the board) — `←`/`→` move the active column
**Status:** ✅ Done — verified 2026-07-19
The board's steering, and the point where board keys get their own gate. `←`/`→` move the
active column (clamped 0..2) *in board mode only* — in list mode they still move focus.
Forced by a both-directions-plus-clamp contradiction against Step 5's hardcoded `i == 0`
accent — the accented border must follow a real, movable index. The gate returns early, so
it must also carry quit through, or `q`/`esc`/`ctrl+c` die in board mode (they no longer
fall through to the list).

**Scope:**
- `tui/model.go` — an `activeCol int` field (default 0); in `Update`'s `tea.KeyMsg` branch,
  after the shared `?`/`tab` intercepts, add `if m.mode == modeBoard { return m.updateBoard(msg) }`
  before the existing list-mode focus/forward logic.
- `tui/board.go` — `updateBoard(msg tea.KeyMsg) (tea.Model, tea.Cmd)`: `KeyLeft`/`KeyRight`
  adjust `activeCol` clamped to `[0, 2]`; the quit keys (`q`, `esc`, `ctrl+c`) return
  `tea.Quit`; anything else is a no-op (the board is read-only). `boardView` accents
  `i == activeCol` instead of `i == 0`.
- `tui/board_test.go` — `TestUpdate_BoardActiveColumn`, `TestUpdate_BoardQuits`.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestUpdate_BoardActiveColumn` (table)
     - **Behavior:** in board mode `→` advances the active column and `←` retreats it,
       clamping at both ends, so arrow steering can never point off the board.
     - **Setup:** `New` with a task in each status; `tab` into board mode; drive sequences
       of `KeyRight`/`KeyLeft`.
     - **Assertions:** default `activeCol == 0`; `→` → 1; `→→` → 2; `→→→` stays 2 (clamp);
       from 2, `←` → 1; `←←←` clamps at 0.
     - **Boundary:** both clamps — `→` at the last column and `←` at the first; the two ends
       a hardcoded increment/decrement would run past.
   - [x] `TestUpdate_BoardQuits` (table)
     - **Behavior:** the quit keys still quit from board mode even though the board gate now
       intercepts keys before the list's inherited quit can see them.
     - **Setup:** `New` with one task; `tab` into board mode; send `q`, then (fresh) `esc`,
       then `ctrl+c`.
     - **Assertions:** each returns a non-nil `cmd` whose `cmd()` is a `tea.QuitMsg`.
     - **Boundary:** the board branch that returns early — proving it preserves the escape
       hatch rather than trapping the user.
   - [x] Confirm fails: `model` has no `activeCol` / `updateBoard` undefined (compile error).

2. **Implement (GREEN):**
   - [x] Add `activeCol`; add the `modeBoard` gate in `Update`; implement `updateBoard`
     with clamped `←`/`→` and quit passthrough; accent `activeCol` in `boardView`.

**Claude verifies:**
- [x] `just test` green (list-mode focus tests — `TestUpdate_Focus`, `TestUpdate_RightDoesNotPageList`
  — still pass, since the gate only fires in board mode)
- [x] `just lint` clean

**User verifies:**
- [x] in board mode `←`/`→` move the accented border between columns and clamp at the ends
- [x] `q`/`esc`/`ctrl+c` still quit from board mode

**Commit (user):** `feat(tui): move the active board column`