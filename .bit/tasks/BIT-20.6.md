---
id: BIT-20.6
title: Modal left/right/h/l page between tasks in board order
status: done
phase: 2
phase_label: Modal paging
---
## **Verse 2**

Wires `flattenBoard` into the modal's key handling so left/right/h/l page to the previous/next
task while the modal stays open — today those four keys only scroll the modal's viewport
(`tui/board.go`'s `updateBoard`, the `modalOpen` branch). Up/down/j/k keep scrolling exactly as
they do today; only left/right/h/l change meaning. This closes Verse 2.

Paging past the first or last task in the flattened sequence clamps at that end (mirrors the
existing column-switch clamp behavior in this same file) rather than wrapping around — the
scope's Decision covers crossing a *column* boundary mid-sequence but doesn't say what happens
at the very ends, so this bar treats "stop, don't wrap" as the safe default consistent with how
column-switching already behaves. Flag this to the user as an assumption when reviewing.

## Scope
- `tui/board.go` — `updateBoard`'s `modalOpen` branch: split the current
  `"up", "down", "left", "right", "j", "k", "h", "l"` case into a scroll case
  (`"up", "down", "j", "k"`, unchanged) and a paging case (`"left", "h"` → previous,
  `"right", "l"` → next). Add `func (m *model) pageModal(delta int)`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestUpdate_ModalPagesWithinColumn`
     - **Behavior:** the simplest paging case — moving to the next card in the same column
       the modal is already open on.
     - **Setup:** tasks `{ID: "BIT-1", Status: "doing"}, {ID: "BIT-1.1", Status: "doing"}`;
       `WindowSizeMsg`, open the modal on `BIT-1` (index 0 of Doing).
     - **Assertions:** after `right` (or `l`), `m.boardSelected().ID == "BIT-1.1"`;
       `m.modalOpen` stays `true`.
     - **Boundary:** the first, smallest paging step — one card forward, no column crossing.
   - [ ] Confirm fails: `right`/`l` currently route to `modalViewport.Update`, a no-op on
     selection.

2. **Implement (GREEN):**
   - [ ] Minimal `pageModal` that only moves the selection within the *current* column
     (`m.boardCols[m.activeCol].Select(m.boardCols[m.activeCol].Index()+delta)`) — enough to
     pass the one test above, but not yet real cross-column paging.
   - [ ] Wire `"left", "h"` → `m.pageModal(-1)`, `"right", "l"` → `m.pageModal(1)` in
     `updateBoard`'s `modalOpen` branch; keep `"up", "down", "j", "k"` on the existing
     viewport-scroll path.

3. **More tests (RED → GREEN):**
   - [ ] `TestUpdate_ModalPagesAcrossColumns`
     - **Behavior:** the scope's worked example — paging past the start of one column
       continues into the previous column, not just within it.
     - **Setup:** tasks `{ID: "BIT-2", Status: "todo"}, {ID: "BIT-1", Status: "doing"},
       {ID: "BIT-1.1", Status: "doing"}`; open the modal on `BIT-1.1` (last card in Doing).
     - **Assertions:** `left` → `boardSelected().ID == "BIT-1"`; `left` again →
       `boardSelected().ID == "BIT-2"` (now in To Do; `m.activeCol == 0`); from there, `right`
       → `"BIT-1"`; `right` again → `"BIT-1.1"`.
     - **Boundary:** contradicts the column-local hardcode — the second `left` press can't
       succeed without a real flattened, cross-column sequence.
   - [ ] `TestUpdate_ModalPagingClampsAtEnds`
     - **Behavior:** paging past either end of the whole board's sequence holds at that end
       instead of wrapping or erroring.
     - **Setup:** same three tasks as above, modal open on `BIT-2` (first in sequence).
     - **Assertions:** `left` → stays on `"BIT-2"`. Then from `BIT-1.1` (last in sequence,
       reached via two `right` presses), `right` again → stays on `"BIT-1.1"`.
     - **Boundary:** both ends of the paging sequence — `idx == 0` and `idx == len-1`.
   - [ ] `TestUpdate_ModalPagingSingleTaskNoop`
     - **Behavior:** a board with only one task doesn't panic or lose the open modal when
       paging is attempted.
     - **Setup:** one task, modal open on it.
     - **Assertions:** `left` and `right` both leave `boardSelected().ID` unchanged and
       `modalOpen == true`.
     - **Boundary:** `len(order) == 1` — the lower bound where there's nowhere to page to.
   - [ ] Implement the real `pageModal`: build `order := flattenBoard(m.boardCols)`; if empty,
     return; find the current entry via `slices.IndexFunc(order, func(e boardEntry) bool { return e.t.ID == cur.ID })` (default `0` if the modal's current task isn't found); clamp
     `idx+delta` to `[0, len(order)-1]` with `min`/`max`; set `m.activeCol = order[idx].col`,
     `m.boardCols[m.activeCol].Select(order[idx].pos)`, then `m.refreshModal()`.

## Claude verifies
- [ ] `go test ./tui/...` passes, including the pre-existing `TestUpdate_ModalScrollsLongBody`
  (confirms up/down/j/k scrolling is untouched).
- [ ] `golangci-lint run` passes.

## User verifies
- [ ] Whole slice: in `bp tui`, open the modal on a task and page through several tasks with
  `l`/`h` (or the arrows) across a column boundary — the popup stays open the whole time and
  its content updates to each task in turn, matching the board's To Do → Doing → Done order.

## Commit (user)
`feat(tui): page the Kanban modal between tasks with left/right/h/l`