---
id: BIT-14.3
title: Unselected rows follow terminal default
status: done
phase: 1
phase_label: Theme
---
## **Verse 1**

Drops the hardcoded gray on unselected bar rows so they render in the terminal's default text color; tracks and bars then differ only by weight and indent, not by a fixed color. Completes the theme verse — after this nothing in the list chrome is a fixed color.

## Scope
- `tui/delegate.go` — `barStyle`: remove `.Foreground(lipgloss.Color("245"))`, leaving `barStyle = lipgloss.NewStyle()` so bar rows inherit the terminal default. `trackStyle` keeps `Bold(true)`; the two-space indent on bars already distinguishes them.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestDelegate_UnselectedBarRowFollowsTerminalDefault`
     - **Behavior:** an unselected bar row carries no hardcoded foreground, so it follows the terminal's default text color.
     - **Setup:** `l := list.New([]list.Item{item{t: &task.Task{ID: "BIT-1", Title: "Track"}}, item{t: &task.Task{ID: "BIT-1.1", Title: "Bar"}}}, delegate{}, 40, 6)`; `l.Select(0)` (so index 1 is unselected); render index 1 into a buffer.
     - **Assertions:** `strings.Contains(out, "38;5;245")` is false (gray gone); `strings.Contains(out, "BIT-1.1")` is true (row text intact).
     - **Boundary:** `isBar(id) && index != m.Index()` — the unselected-bar path, the only one that hardcoded gray.
   - [ ] Confirm fails: the unselected bar currently renders `\x1b[38;5;245m`, so the gray-absent assertion fails.

2. **Implement (GREEN):**
   - [ ] Change `barStyle` to `lipgloss.NewStyle()` (drop the foreground).

3. **More tests (RED → GREEN):**
   - [ ] `TestDelegate_TrackVsBarDistinguishedByWeightNotColor`
     - **Behavior:** an unselected track and an unselected bar differ by weight (bold) and indent, not by a foreground color — proving the "no color to distinguish rows" decision.
     - **Setup:** render an unselected track row (index 0 with `l.Select(1)`) and an unselected bar row (index 1 with `l.Select(0)`) into separate buffers.
     - **Assertions:** the track output contains the bold SGR `\x1b[1m` and the bar output does not; neither output contains a foreground color code (`38;5` absent in both).
     - **Boundary:** track vs bar, both unselected — exercises both weights of the same unselected state.
   - [ ] Confirm passes with the GREEN above (both rows colorless once the foreground is dropped; the bold difference comes from `trackStyle`).

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] Whole slice: run `bit tui`, then switch your terminal's color theme (or open it under a second theme). The focused border, the selected row, and the unselected rows all recolor to match the theme — no purple or gray stays fixed.

## Commit (user)
`feat(tui): unselected rows follow terminal default`
