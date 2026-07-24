---
id: BIT-14.1
title: Focus accent follows terminal green
status: done
phase: 1
phase_label: Theme
---
## **Verse 1**

Repoints the focused-pane accent from the fixed 256-palette purple to the terminal's ANSI green, so the most prominent chrome follows the theme. Walking skeleton for the theme verse: the smallest end-to-end recolor a user can see.

## Scope
- `tui/model.go` — `titledBorder`: change the active-branch `accent := lipgloss.Color("99")` to `lipgloss.Color("2")` (ANSI slot 2 = terminal green; `Color("2")` returns `ansi.BasicColor(2)`, which follows the theme, whereas `Color("99")` is a fixed 256 index).

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTitledBorder_ActiveUsesTerminalGreen`
     - **Behavior:** a focused pane/column draws its border in terminal ANSI green, following the theme rather than a fixed purple.
     - **Setup:** `got := titledBorder("body", "Tasks (0)", 20, 3, true)`.
     - **Assertions:** `strings.Contains(got, "\x1b[32m")` is true (green SGR present) and `strings.Contains(got, "38;5;99")` is false (256-purple gone).
     - **Boundary:** `active == true` — the focused state, the only one that draws an accent; the inactive branch draws no color at all.
   - [ ] Confirm fails: the border currently renders `\x1b[38;5;99m`, so the purple-absent assertion fails.

2. **Implement (GREEN):**
   - [ ] Change the accent color in `titledBorder`'s active branch to `lipgloss.Color("2")`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic (the theme verse's observe-it check lands on its last bar).

## Commit (user)
`feat(tui): focus accent follows terminal green`
