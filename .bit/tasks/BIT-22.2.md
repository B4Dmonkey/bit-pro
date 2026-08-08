---
id: BIT-22.2
title: Contradiction forces Enter to hand focus back
status: todo
phase: 1
phase_label: Focus
---
## **Verse 1**

A second Enter must hand focus back to the list, which the previous bar's hardcoded
`detailFocused = true` cannot do — collapsing currently leaves the operator driving a pane
that is no longer there. This contradiction is what forces focus to track the expanded state
rather than being set once.

## Scope
- `tui/model.go` — the `msg.Code == tea.KeyEnter` branch in `Update`
- `tui/model_test.go` — new test

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestUpdate_EnterAgainReturnsFocusToList`
     - **Behavior:** Collapsing the detail pane returns the arrows to the task list, so one key
       round-trips the interaction and the operator lands back in the state they started from.
     - **Setup:** same two 500-line tasks and window size as the previous bar; `KeyTab`, then
       `KeyEnter`, then `KeyEnter` again, then `KeyDown`.
     - **Assertions:** `Index() == 1` and `viewport.YOffset() == 0`.
     - **Boundary:** the second crossing of the same boolean — `detailExpanded` true→false.
       The previous bar pinned the first crossing, so a constant can no longer satisfy both.
   - [ ] Confirm fails: `Index() = 0` and `YOffset() > 0` — the hardcoded `detailFocused =
     true` from the previous bar keeps routing down to the viewport after the collapse.

2. **Implement (GREEN):**
   - [ ] Replace the hardcoded assignment with `m.detailFocused = m.detailExpanded`, read
     after the toggle so focus follows the state Enter just produced.

## Claude verifies
- [ ] `just test` — including the pre-existing `TestUpdate_Focus`,
      `TestUpdate_FocusRoutesArrows`, `TestUpdate_CtrlDScrollsDetail`, and the four
      `TestUpdate_Pages*`/`TestUpdate_Enter*` tests, none of which should need editing
- [ ] `just lint`

## User verifies
- [ ] Whole slice: run `bp tui`, press Tab for the list, select a long task (BIT-21), press
      Enter. Up/down scroll the body, the green active border is on `Details` and not on the
      ten-column `Tasks` strip, left/right still page between tasks, and a second Enter
      returns to the normal split with up/down moving the selection again.

## Commit (user)
`fix(tui): return focus to the list when Enter collapses the detail pane`