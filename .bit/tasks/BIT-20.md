---
id: BIT-20
title: 'TUI: Kanban-first default view + task navigation'
status: doing
---
## Why
`bp tui` opens in list view and lands on whatever the list's natural top item is, so every
session starts with a manual detour — Tab into the board, find the Doing column, find the
actual in-progress bar (not its parent track) — before you're looking at what you're
actually working on. Reviewing task detail has the same friction from a different angle:
the board's modal only shows one task at a time, so paging through several open tasks
means repeatedly closing and reopening it; the list view's detail pane is a fixed 60% split,
too narrow for comfortably reading a longer body. None of this blocks work, but it's a small
tax paid every single time the TUI opens or a task gets reviewed.

## Summary
Make Kanban the default landing view, focused on the Doing column's top bar. Let the Kanban
modal page through tasks with left/right/h/l instead of only scrolling. Let the list view's
detail pane expand to ~90% width for focused reading, with the same prev/next-task paging
while expanded.

## Decisions
- **Kanban is the default view on startup**, replacing list view as `bp tui`'s zero-value
  default. List view remains reachable via the existing Tab toggle.
- **Default column and selection.** On open, focus Doing whenever it has any tasks — an
  operator opening `bp tui` wants to land on what they're actively working on, regardless of
  whether other columns are also populated. Only when Doing is empty does it fall back to the
  first non-empty of To Do, then Done. Within the chosen column, default to the first *bar*
  (a dotted child ID) in list order, skipping over track-level rows; if the column contains
  no bars, default to its first row (a track).
- **Keybindings are context-sensitive, not a global rebind.** In Kanban, left/right/h/l mean
  "switch column" while the modal is closed (unchanged from today) and "previous/next task"
  only while the modal is open. In list view, left/right/h/l mean "switch focus between the
  list and details panes" in the normal split layout (unchanged from today) and
  "previous/next item" only while the details pane is expanded. Enter is the toggle in both
  cases: it opens/closes the Kanban modal, and it expands/collapses the list view's detail
  pane.
- **Task order for paging is the board's existing render order**: To Do, then Doing, then
  Done, top-to-bottom within each column. Paging past the end/start of one column continues
  into the next/previous column rather than stopping.
- **Vertical scroll is untouched.** Up/down (and j/k) keep scrolling the Kanban modal's
  content exactly as they do today; only left/right/h/l change meaning when the modal opens.

## Verses
- [x] Verse 1 — Operator opens `bp tui` and lands directly on the task they're already
  working on: Kanban view, Doing column, topmost bar selected — no Tab, no hunting.
  Touches: `cmd/tui.go`, `tui/model.go` (initial view mode), `tui/board.go` (initial column
  and selection).
- [ ] Verse 2 — With the Kanban modal open, an operator pages through every task's detail
  with left/right or h/l without ever closing the popup, reviewing a whole column (or the
  whole board) in one sitting.
  Touches: `tui/board.go` (modal key handling, `refreshModal`).
- [ ] Verse 3 — In list view, an operator expands the details pane to ~90% width with Enter
  for a comfortable read, pages to the next/previous item with left/right/h/l while it's
  expanded, and collapses back to the normal split with Enter again.
  Touches: `tui/model.go` (layout width state, key dispatch by layout mode).