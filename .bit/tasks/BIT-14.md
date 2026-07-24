---
id: BIT-14
title: Theme-native, legible TUI chrome
status: doing
---
## Why

The TUI's chrome fights the reader instead of serving them. It paints everything in a
fixed purple (`Color("99")`) that ignores whatever theme the terminal is set to, so the
tool clashes with the user's environment no matter how they've themed it. It also carries
noise the eye has to filter past — a redundant `List` heading and a duplicated empty-state
line in list view — and its pane/column titles don't clearly signal which pane has focus.

This matters because the whole point of the live-reload work (BIT-13) was to keep a human
oriented inside a running TUI while edits land underneath them. Orientation depends on the
chrome being calm and glanceable: focus you can read instantly, colors that match the rest
of the terminal, and no decoration competing for attention. Right now it isn't.

## Summary

Make the TUI's lipgloss chrome theme-native and legible without touching any copy beyond
two deletions. Replace the fixed purple with two terminal ANSI colors that follow the user's
theme: green for focus/selection, white for unselected rows. Make focus unmistakable: the
focused pane's title (list view) and the focused column's title (board) render inverted,
unfocused board columns get a framed `─| Title |─` look, selected cards render inverted, and
an opened card modal shows its title inverted. Finally, drop the two pieces of list-view
noise: the `List` heading and the duplicate `No items` line.

The glamour-rendered markdown detail pane is explicitly out of scope — it keeps its current
styling.

## Visual aid

Titled-border states, before → after:

```
FOCUSED pane / column
  before:  ─ Tasks (0) ───────      (cyan/purple text sitting in the border)
  after:   ─[ Tasks (0) ]───────    rendered inverted: dark text on a green block

UNFOCUSED board column
  before:  ─ Doing (0) ────────     (dim text in the border)
  after:   ─| Doing (0) |─────      framed inline in the border line

LIST VIEW body
  before:   List            after:   (gone)
            No items                 No items.
            No items.
```

## Decisions

- **Focus accent is terminal green (ANSI slot).** The focused-pane border and inverted titles,
  the focused column title, the selected list row, selected cards, and the modal title all use
  green. Follows the terminal theme; replaces the fixed purple `Color("99")`.
- **Everything unselected uses the terminal's normal white / default foreground.** Unselected
  list rows drop the hardcoded gray (`Color("245")`) on bar rows so they follow the terminal;
  tracks vs bars distinguish by weight (bold vs not) and indent, not by color. Unselected board
  column titles keep this same default color — their only change is gaining the framed
  `─| Title |─` look, not a recolor.
- **The glamour markdown detail pane is unchanged.** Only lipgloss chrome is in scope; the
  syntax-highlighted body keeps its current styling.
- **Selection renders inverted in board view — column titles and cards both.** A selected
  column title and a selected card both use the inverted (reverse-video) green treatment, so
  board selection is consistent. (The list view's selected row is green foreground, not inverted.)
- **Copy is unchanged except two deletions.** The `List` heading and the duplicate `No items`
  line are removed; no other wording changes.
- **"Inverted" means reverse-video.** A filled block where the accent (green) is the background
  and the title text sits dark on top — the look of the current `List` chip and the app's active tab.

## Verses

- [ ] Verse 1 — The chrome follows your terminal theme: the purple accent becomes terminal
  green and unselected rows become terminal white, so re-theming the terminal re-colors the
  tool. Touches: `tui/delegate.go` (row styles), `tui/model.go` (`titledBorder` accent).
- [ ] Verse 2 — Focus is unmistakable: the focused pane's title (list view) and focused
  column's title (board) render inverted, unfocused board columns show the framed `─| Title |─`
  look, selected cards render inverted, and an opened card modal shows its title inverted.
  Touches: `tui/model.go` (`titledBorder`, `content`), `tui/board.go` (`boardView`, `modalView`),
  `tui/delegate.go` (selected-card style).
- [ ] Verse 3 — The list view sheds its noise: the redundant `List` heading and the duplicate
  empty-state line are gone, leaving one clean empty state. Touches: `tui/model.go` (list setup),
  `tui/delegate.go` if the empty state renders there.