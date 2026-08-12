---
id: BIT-22
title: Enter moves focus into the expanded detail pane
status: done
---
## Why

The list view's detail pane widens to about 90% of the screen when you press Enter, but
pressing Enter while the task column is selected leaves keyboard focus on the list. Scrolling
the pane is gated on that focus, and expanding also reassigns left/right from moving focus to
paging between tasks — so the one key that could have moved focus no longer does. Expanding a
task to read it therefore makes it *less* readable than before: up/down change the list
selection instead of scrolling the body, and the active-pane border highlights a ten-column
strip while the pane the operator is actually looking at reads as inactive.

Bodies are the reason the pane exists. A scope body carries the Why, Decisions, and Verses; a
bar body carries its Scope and its verification checklists. These routinely run well past a
screen, which is exactly why the expanded view was built — and it's the one view where the
content can't be reached.

## Summary

Make Enter a symmetric enter/exit for the detail pane: expanding moves focus into it so the
body scrolls, and collapsing hands focus back to the task list. Then correct the help footer,
which still labels left/right as "focus" even in the expanded state where those keys page
between tasks instead.

## Visual aid

```
today — Enter pressed from the task column

  ┌ Tasks ─┐┌ Details ───────────────────────────────┐
  │ BIT-22 ││ ## Why                                 │
  │ BIT-21 ││ The list view's detail pane widens to  │
  │ BIT-20 ││ about 90% of the screen when you press │
  └────────┘└────────────────────────────────────────┘
   ^^^^^^^^                                    (cut off, and
   active border here                           unreachable)

   ↑/↓    moves the list selection      ← the viewport never sees these
   ←/→    pages between tasks           ← reassigned, so focus can't move
   Enter  collapses

   no key scrolls the 90%-wide body

after

  ┌ Tasks ─┐┌ Details ───────────────────────────────┐
  │ BIT-22 ││ about 90% of the screen when you press │
  │ BIT-21 ││ Enter, but pressing Enter while the    │
  │ BIT-20 ││ task column is selected leaves focus…  │
  └────────┘└────────────────────────────────────────┘
                        ^^^^^^^ active border here

   ↑/↓    scrolls the body
   ←/→    pages between tasks
   Enter  collapses, and focus returns to the list
```

## Decisions

- **Enter enters the pane and exits it.** Expanding focuses the details; collapsing returns
  focus to the list. One key round-trips the whole interaction, so an operator ends up back in
  the state they started from rather than stranded on a half-width pane they have to press
  left to escape.
- **Left/right keep the meaning they already have while expanded.** They page between tasks.
  Focus is Enter's job now, so the two no longer compete for the same keys, and the existing
  paging behaviour and its tests stand.
- **The help footer describes the state the operator is in.** While expanded, left/right are
  labelled for what they do there. A footer that names a binding the current state doesn't
  honour is the same confusion as the focus bug, read off the bottom of the screen.
- **No new key bindings are advertised or added.** Enter is not currently in the footer's key
  map and stays out of it; this track corrects labels that are wrong, and does not grow the
  keyboard surface.

## Verses

- [x] Verse 1 — An operator can read a task body that doesn't fit on screen: pressing Enter on
  a task moves into the widened details pane so up/down scroll it and the border shows it as
  active, and pressing Enter again returns to the task list.
  Touches: the list-mode key handling and focus state in `tui/model.go` — where to look to verify.

- [x] Verse 2 — The help footer tells the truth about the keys in the expanded state, so an
  operator who forgets how to get around reads the bottom of the screen rather than guessing.
  Touches: the list-view key map and help rendering in `tui/model.go` — where to look to verify.