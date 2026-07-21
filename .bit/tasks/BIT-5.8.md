---
id: BIT-5.8
title: Help bar keyed to board mode
status: done
phase: 3
phase_label: drive the board
---
## Step 8 (Phase 3 — drive the board) — Help bar keyed to board mode
**Status:** ✅ Done — verified 2026-07-19
The help bar describes the controls of the mode you're in: board mode shows `←/→ column ·
↑/↓ card · tab list · ? help · q quit`, list mode keeps its focus/move hints. Forced by a
mode contradiction — one static keymap can't describe both views, so `View` must pick the
keymap by mode.

**Scope:**
- `tui/board.go` — a `boardKeyMap` (implementing `help.KeyMap`) whose `ShortHelp`/`FullHelp`
  name the board bindings (column, card, tab→list, help, quit).
- `tui/model.go` — `View` passes the board keymap to `m.help.View(...)` in board mode and
  the existing `m.keys` in list mode; `layout()` measures the active mode's help height so
  the reserved row(s) stay correct in both.
- `tui/board_test.go` — `TestView_BoardHelp`.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestView_BoardHelp` (table)
     - **Behavior:** the help bar matches the current mode — board mode advertises column
       and card navigation and `tab` back to the list; list mode still advertises focus —
       so the hints are never wrong for the view on screen.
     - **Setup:** `New` with a few tasks, `WindowSizeMsg{120,40}`. Case A: stay in list
       mode. Case B: `tab` into board mode.
     - **Assertions:** A — `View()` contains `"focus"` (list hints) and not `"column"`; B —
       `View()` contains `"column"` and `"card"` (board hints).
     - **Boundary:** the two modes — each proves the *other* mode's hints are absent, which
       a single shared keymap can't do.
   - [x] Confirm fails: board mode shows the list's `"focus"` help, not `"column"`.

2. **Implement (GREEN):**
   - [x] Add `boardKeyMap`; `View` selects the keymap by mode; `layout()` measures the
     active mode's help height.

**Claude verifies:**
- [x] `just test` green (`TestView_HelpBarPresentAndBounded`, `TestUpdate_QuestionTogglesFullHelp`
  still pass in list mode), `just lint` clean

**User verifies:**
- [x] in board mode the help bar shows the column/card/tab-list controls; in list mode it
  shows the focus/move controls; `?` expands the full menu in both modes and nothing
  overflows

**Commit (user):** `feat(tui): show board controls in the help bar`