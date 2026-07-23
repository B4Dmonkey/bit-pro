---
id: BIT-11
title: Board modal
status: done
---
## Why

The kanban board shows only card faces — an ID, a title, a status. To actually read a
task's body (its scope prose, its plan detail) a human has to tab back to the list view,
find the row, and read it in the detail pane. That round-trip is friction: the board is
where you survey work, but you can't inspect a single card without leaving the board. The
list view already solved "read a task's body in place" with its detail pane; the board has
no equivalent, so inspection means context-switching away from the very view you were using
to decide what to inspect.

## Summary

On the kanban board, pressing Enter on the selected card opens a modal — a bordered box
floated over the board (backlog.md's detail-overlay style: no dimming, single border),
composited as a layer on top of the board render — that shows that card's full, rendered
body. Pressing q or esc closes the modal and returns to the board with the app still
running; ctrl+c remains the hard quit from anywhere. A body longer than the modal scrolls
inside it rather than overflowing, so even a long track like BIT-2 is fully readable
without leaving the board.

## Visual aid

```
board (no modal)                  modal open (Enter on selected card)
┌─ To Do (3) ─┐┌─ Doing ─┐        To Do (3)   Doing (1)   Done (5)
│ ▎BIT-11     ││ BIT-4   │        ┌─ To Do  BIT-11 — Board modal ──────┐
│  BIT-9      ││         │  Enter │  Details                           │
│  BIT-2      ││         │  ────► │  ## Why                            │
└─────────────┘└─────────┘        │  The board only shows card faces...│
   q/esc: quit app                │  (ctrl+u / ctrl+d scrolls)         │
                                  └────────────────────────────────────┘
                                     q/esc: close modal   ctrl+c: quit
```

## Decisions

- **Depends on the Charm v2 migration (BIT-12).** The overlay is drawn with lipgloss's
  cell-based compositor, which ships only in lipgloss v2, so BIT-11 builds on the v2 stack
  that BIT-12 lands. This track can't start until that migration is done.
- **Trigger is Enter on the selected board card.** Enter opens the modal for the card
  currently selected in the active column. Enter only does this on the board; the list view
  keeps its always-visible detail pane and gains nothing new.
- **The modal is a single-bordered box floated over the board** (backlog.md's detail-overlay
  style) — **no dimming** and **no double border**. The board stays visible around it,
  composited as a layer on top.
- **The overlay uses lipgloss's cell-based compositor** (`NewLayer` + `Compose`), not
  hand-rolled line-splicing — it's the library's sanctioned primitive for layered content,
  so the box renders over the board without reinventing z-ordering.
- **The modal captures input while open.** q or esc closes the modal and returns to the
  board without quitting; ctrl+c still quits the app from anywhere, including with the modal
  open. Every other key (board navigation, tab) is swallowed while the modal is open — you
  can't move the board underneath it. This is the "close the modal, *then* close the app"
  behavior: to quit you dismiss the modal first, then q/esc on the board quits as it does
  today. Existing board close behavior (no modal → q/esc/ctrl+c quit) is unchanged.
- **The modal reuses the list detail's rendering.** It shows the same glamour-rendered body
  in a scrollable viewport the list detail pane already uses — one rendering path, not a
  second one that can drift.
- **Enter on an empty column does nothing.** If the active column has no cards there's
  nothing to open, so Enter is a no-op.

## Verses

- [x] Verse 1 — Open a card's details in a modal and dismiss it: on the board, Enter on the
  selected card floats a bordered box over the board showing that card's rendered body; q or
  esc closes it back to the board (app stays open); ctrl+c still quits. Delivers the core
  inspect-without-leaving-the-board capability, and — as the walking skeleton — is what first
  exercises lipgloss's v2 compositor over the live board render. Touches: `tui/board.go`
  (updateBoard key handling, the overlay view), `tui/model.go` (model gains modal state;
  Update routes Enter/close/quit) — where to look to verify.
- [x] Verse 2 — Read a long body by scrolling inside the modal: a body taller than the modal
  scrolls within it (ctrl+u/ctrl+d) instead of overflowing the box, so long tracks are fully
  readable on the board. Touches: `tui/model.go` (the modal's viewport sizing + scroll-key
  routing, reusing the list detail's viewport) — where to look to verify.