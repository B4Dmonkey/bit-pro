---
id: BIT-5.3
title: Board mode renders three plain columns
status: done
phase: 1
phase_label: flip to a board
---
## Step 3 (Phase 1 — flip to a board) — Board mode renders three plain columns
**Status:** ✅ Done — verified 2026-07-19
Completes the walking skeleton: `tab` now *shows* a board — three side-by-side columns of
cards, split into equal thirds, grouped by status. Pure paint over the Step 1 toggle and
Step 2 grouping, so it is **manual-verify** with no new unit test (the same posture as list
Steps 3 and 12 — the logic underneath is already covered). Plain on purpose: no borders,
titles, counts, or accents yet (those are Phase 3); each card is just its ID + title.

**Scope:**
- `tui/board.go` — `boardView(m model) string`: for each of `m.columns`, render its cards
  (`ID + "  " + Title` per line, reusing the list's card text shape) into a column of width
  `≈ m.winWidth/3`, joined with `lipgloss.JoinHorizontal`. Height bounded to `m.height`
  (the pane height `layout()` already computes under the help bar).
- `tui/model.go` — `View` branches on `m.mode`: `modeBoard` returns `boardView(m)` (still
  joined above the existing `m.help.View(m.keys)`); `modeList` unchanged.

**Implement (GREEN):**
- [x] `boardView` laying the three grouped columns into equal thirds; `View` routes to it
  in board mode. No new test — rendering is visual.

**Claude verifies:**
- [x] `just build` succeeds
- [x] `just test` green (unchanged), `just lint` clean

**User verifies:**
- [x] `bit tui` opens on the list; `tab` switches to a three-column board; `tab` returns to
  the list
- [x] against the real records, To Do shows the `todo` tasks, Doing the one `doing` task,
  Done the `done` tasks — each in the correct column
- [x] the three columns are roughly equal thirds and nothing overflows the screen
- [x] `q`/`esc`/`ctrl+c` still quit from board mode (quit still falls through here — the
  board key gate that could break it doesn't land until Step 4)

**Commit (user):** `feat(tui): render the board view`