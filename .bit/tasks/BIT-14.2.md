---
id: BIT-14.2
title: Selected list row follows terminal green
status: done
phase: 1
phase_label: Theme
---
## **Verse 1**

Repoints the selected list row from purple to terminal green, so the list's selection highlight follows the theme like the border now does.

## Scope
- `tui/delegate.go` — `selectedStyle`: change `Foreground(lipgloss.Color("99"))` to `Foreground(lipgloss.Color("2"))`, keeping `Bold(true)`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestDelegate_SelectedRowUsesTerminalGreen`
     - **Behavior:** the highlighted list row renders in terminal green, so selection follows the theme.
     - **Setup:** `l := list.New([]list.Item{item{t: &task.Task{ID: "BIT-1", Title: "Track"}}}, delegate{}, 40, 4)`; `var buf bytes.Buffer`; `delegate{}.Render(&buf, l, 0, l.Items()[0])` (index 0 == `l.Index()`, so the row is selected).
     - **Assertions:** `strings.Contains(buf.String(), "32m")` is true (green present) and `strings.Contains(buf.String(), "38;5;99")` is false (purple gone). (Probe reference: the bold-purple row renders `\x1b[1;38;5;99m`; bold-green renders `\x1b[1;32m` — confirm the exact SGR at RED time.)
     - **Boundary:** `index == m.Index()` — the selected path, the only one `selectedStyle` applies to.
   - [ ] Confirm fails: the selected row currently renders `\x1b[1;38;5;99m`, so the purple-absent assertion fails.

2. **Implement (GREEN):**
   - [ ] Change `selectedStyle`'s foreground to `lipgloss.Color("2")`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(tui): selected list row follows terminal green`
