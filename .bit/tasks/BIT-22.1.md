---
id: BIT-22.1
title: Enter hands up/down to the body
status: todo
phase: 1
phase_label: Focus
---
## **Verse 1**

Pressing Enter starts routing up/down to the detail viewport instead of the list. This is the
entry point of the whole verse: `Update`'s Enter branch currently flips `detailExpanded` and
returns without touching `detailFocused`, and the routing tail at the bottom of `Update` keys
solely on `detailFocused` — so the widened pane can't scroll.

## Scope
- `tui/model.go` — the `msg.Code == tea.KeyEnter` branch in `Update` (line 258)
- `tui/model_test.go` — new test

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestUpdate_EnterFocusesDetailForScrolling`
     - **Behavior:** After expanding a task, up/down scroll the task's body rather than
       changing which task is selected — the pane the operator just widened is the one the
       arrows drive.
     - **Setup:** `New([]*task.Task{{ID: "BIT-2", Body: body}, {ID: "BIT-1", Body: body}})`
       where `body := strings.Repeat("line\n", 500)`. Then
       `Update(tea.WindowSizeMsg{Width: 80, Height: 24})`, `Update(tea.KeyPressMsg{Code:
       tea.KeyTab})` to reach list mode, `Update(tea.KeyPressMsg{Code: tea.KeyEnter})`, then
       `Update(tea.KeyPressMsg{Code: tea.KeyDown})`.
     - **Assertions:** `viewport.YOffset() > 0` and `Index() == 0`.
     - **Boundary:** `detailExpanded == true` — the state that had no focus path at all. The
       collapsed side of the same boolean is already pinned by `TestUpdate_Focus` and
       `TestUpdate_FocusRoutesArrows`, which press right rather than Enter and must stay green.
   - [ ] Confirm fails: both assertions — `YOffset() = 0` and `Index() = 1`, because
     `detailFocused` is still false so `Update` falls through to `m.Model.Update`, which moves
     the selection and then `refreshDetail()` resets the viewport to the top.

2. **Implement (GREEN):**
   - [ ] In the Enter branch, set `m.detailFocused = true` alongside the existing toggle. A
     bare `true` is the minimum here and is deliberately wrong for the collapse case — the
     next bar contradicts it.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic; the verse's end-to-end check is on the next bar.

## Commit (user)
`fix(tui): focus the detail pane when Enter expands it`