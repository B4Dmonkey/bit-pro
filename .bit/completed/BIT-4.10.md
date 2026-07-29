---
id: BIT-4.10
title: Help bar beneath both panes
status: done
phase: 4
phase_label: focus & affordances
---
## Step 10 (Phase 4 — focus & affordances) — Help bar beneath both panes
**Status:** ✅ Done — verified 2026-07-18
The keybinding hints move out of the list box — where `bubbles/list` renders its own help,
the bar you saw *inside* the pane — into a single bar under both panes, describing the
two-pane controls (`←/→` focus, `↑/↓` move·scroll, `? help`, `q quit`). Forced by turning the
list's built-in help off and reserving the row(s) the panes must give up.

**Scope:**
- `tui/model.go` — `l.SetShowHelp(false)` in `New`; a package `help.Model` + `keyMap`
  (implementing `help.KeyMap`) render the bar so `?` still toggles the full menu; `View` joins
  the two panes above `m.help.View(m.keys)` via `lipgloss.JoinVertical`; a `layout()` helper
  measures the help's actual rendered height and gives the panes the rest.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestNew_ListHelpDisabled`
     - **Behavior:** the list no longer renders its own help inside its box — the frame owns
       the help now.
     - **Setup:** `New` with one task.
     - **Assertions:** `m.ShowHelp() == false`.
     - **Boundary:** the built-in in-pane help line is off at its single source.
   - [x] `TestView_HelpBarPresentAndBounded`
     - **Behavior:** the help bar shows the two-pane controls and the whole view still fits
       the window with a row spent on it.
     - **Setup:** long body, `WindowSizeMsg{80,24}`.
     - **Assertions:** `strings.Contains(View(), "focus")` (the help text) and
       `lipgloss.Height(View()) <= 24`.
     - **Boundary:** the help row lives inside the height budget — the panes shrank to make room.
   - [x] `TestUpdate_QuestionTogglesFullHelp`
     - **Behavior:** `?` expands the full help menu and collapses it again — the `bubbles/help`
       feature a hand-rolled string would have lost.
     - **Setup:** long body, `WindowSizeMsg{80,24}`; send `?` twice.
     - **Assertions:** `help.ShowAll` false → true → false; the expanded view still
       `lipgloss.Height(View()) <= 24`.
     - **Boundary:** both toggle states, and the expanded (multi-row) help still inside the budget.
   - [x] Confirm fails: `ShowHelp()` is true / no help text below the panes / `model` has no
     `help` field.

2. **Implement (GREEN):**
   - [x] Disable the list help; render a real `help.Model` bar driven by a `keyMap`
     implementing `help.KeyMap`; intercept `?` in `Update` to toggle `help.ShowAll`; a
     `layout()` helper reserves the help's measured height (not a hardcoded one row) so both
     the collapsed and expanded bar stay bounded. A first cut used a static help string; the
     user caught that it dropped the `bubbles/help` `?` menu, so it was replaced with the
     component (bit_do case 3 — plan right, implementation wrong). `TestUpdate_WindowSizeSizesViewport`
     was revised `== 22` → `== 21` (the reserved help row moved the viewport inset down one).

**Claude verifies:**
- [x] `just build`, `just test` green, `just lint` clean

**User verifies:**
- [x] the help renders in one bar beneath both panes — not inside the list — and lists the
  focus / move / help / quit keys
- [x] `?` expands and collapses the full help menu
- [x] nothing overflows in either the collapsed or expanded state; both panes stay in their boxes

**Commit (user):** `feat(tui): show a help bar beneath the panes`