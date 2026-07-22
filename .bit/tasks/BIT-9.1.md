---
id: BIT-9.1
title: Quit keys exit from the detail pane
status: done
phase: 1
phase_label: Quit from anywhere
---
Quit keys exit the TUI while the detail pane is focused, instead of being swallowed by the viewport-forwarding branch. Forces the fix: the `detailFocused` branch returns every message to the viewport before the list's quit handler can see it.

**Scope:**
- `tui/model.go` — in `Update`, before the `if m.detailFocused { m.viewport, cmd = ... }` forward (lines ~133-137), intercept the quit keys and return `tea.Quit`, mirroring `updateBoard`'s `case "q", "esc", "ctrl+c": return m, tea.Quit` (`tui/board.go:78-79`). Everything else still forwards to the viewport so scrolling is unaffected.
- `tui/model_test.go` — new test.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestUpdate_QuitsFromDetail` (table-driven, mirroring `TestUpdate_BoardQuits`)
     - **Behavior:** with focus on the detail pane, each quit key returns a `tea.Quit` cmd — focus never changes what quits.
     - **Setup:** `New([]*task.Task{{ID: "BIT-1"}})`, size with `WindowSizeMsg{80,24}`, then `Update(KeyMsg{Type: tea.KeyRight})` to focus the detail pane. Table cases: `q` (`KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}`), `esc` (`KeyEsc`), `ctrl+c` (`KeyCtrlC`).
     - **Assertions:** for each key, `cmd != nil` and `cmd().(tea.QuitMsg)` type-asserts ok.
     - **Boundary:** all three quit keys, exercised from the detail-focused state (the state where they're currently swallowed) — the mirror of `TestUpdate_BoardQuits` for list mode.
   - [ ] Confirm fails: today the keys forward to `m.viewport.Update`, which returns a nil/non-quit cmd, so the `tea.QuitMsg` assertion fails (esc/q/ctrl+c never quit from the detail pane).

2. **Implement (GREEN):**
   - [ ] In `Update`, add a quit-key check on the `tea.KeyMsg` before the `detailFocused` viewport-forward: `switch msg.String()` (or `msg.Type`/runes) matching `q`, `esc`, `ctrl+c` → `return m, tea.Quit`. Keep it in the `case tea.KeyMsg` block so non-key messages are untouched. Leave the viewport forward for every other key so `ctrl+d`/`ctrl+u` still scroll.

**Claude verifies:**
- [ ] `just test` passes — including the new test and the existing `TestUpdate_CtrlDScrollsDetail`, `TestUpdate_FocusRoutesArrows`, `TestUpdate_EscQuitsFromList` (scrolling and list-mode quit unregressed).
- [ ] `just lint` clean.

**User verifies:**
- [ ] In `bit tui`, press `→` to focus the detail pane, then confirm `q`, `esc`, and `ctrl+c` each quit — and that `ctrl+d`/`ctrl+u` still scroll a long body while focused.

**Commit (user):** `fix(tui): quit keys exit from the detail pane`