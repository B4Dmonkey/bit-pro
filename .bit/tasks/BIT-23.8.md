---
id: BIT-23.8
title: Space key toggles approval on the focused item
status: done
phase: 4
phase_label: TUI approval display
---
## **Verse 4**

Yellow signals "needs a look," but the user still has to drop to the terminal to act on it. Space is unbound in both views and reads as "stamp it" — toggling approval without leaving the board closes the review loop in one place.

## Scope
- `tui/model.go` — add `approve func(id string, approved bool) error` field; add `WithApprove(f func(id string, approved bool) error) model` builder; handle `" "` (space) in both `Update` paths (list mode and board mode): call `m.approve(id, !t.Approved)` if the callback is set, then return `m.reloadCmd()` to pick up the new state from the store; add `"space"` to `keyMap.ShortHelp` if space is recognized as a meaningful key
- `cmd/tui.go` — wire `s.SetApproved` via `.WithApprove(func(id string, approved bool) error { return s.SetApproved(id, approved) })`

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestUpdate_SpaceTogglesApprovalInListMode`
     - **Behavior:** pressing space in list mode calls the approve callback with the selected task's ID and inverted approval
     - **Setup:** `var called []struct{ id string; approved bool }`; `m := New([]*task.Task{{ID: "BIT-1", Status: "todo", Approved: false}}).WithApprove(func(id string, a bool) error { called = append(called, struct{id string; approved bool}{id, a}); return nil })` ; send `tea.KeyPressMsg{Code: ' '}`
     - **Assertions:** `len(called) == 1`; `called[0].id == "BIT-1"`; `called[0].approved == true`
     - **Boundary:** `Approved == false` → toggle sends `true` — the unapproved-to-approved direction; proves inversion
   - [ ] Confirm fails: space is not handled in `Update`; model returns unchanged

2. **Implement (GREEN):**
   - [ ] Add `approve func(id string, approved bool) error` field and `WithApprove` builder to `model`
   - [ ] In `Update` for `tea.KeyPressMsg` in list mode: handle `msg.Code == ' '`; if `m.approve != nil` and a task is selected, call `m.approve(t.ID, !t.Approved)` and return `m, m.reloadCmd()`
   - [ ] Same handler in board mode inside `updateBoard` for the non-modal path
   - [ ] In `cmd/tui.go`, chain `.WithApprove(s.SetApproved)` (or a closure wrapping `s.SetApproved`)

3. **More tests (RED → GREEN):**
   - [ ] `TestUpdate_SpaceTogglesApprovalInBoardMode`
     - **Behavior:** pressing space in board mode calls the approve callback for the focused card
     - **Setup:** same callback capture; start in board mode (default); press space
     - **Assertions:** `called[0].id` matches the active board card's ID
     - **Boundary:** board mode — proves both modes are handled
   - [ ] `TestUpdate_SpaceOnApprovedItemSendsUnapproved` (contradiction)
     - **Behavior:** pressing space on an already-approved item sends `approved=false`
     - **Setup:** `Approved: true`; press space
     - **Assertions:** `called[0].approved == false`
     - **Boundary:** `Approved == true` → toggle sends `false`; contradicts always sending true; proves inversion is real
   - [ ] `TestUpdate_SpaceWithNoCallbackIsNoop`
     - **Behavior:** pressing space when no approve callback is wired causes no panic and no state change
     - **Setup:** `New(tasks)` without `WithApprove`; press space
     - **Assertions:** no panic; returned model is equivalent to input model
     - **Boundary:** nil callback — proves the nil guard

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` clean

## User verifies
- [ ] Whole verse: `bp tui`, navigate to an unapproved item (shown in yellow), press space — item turns white (approved); press space again — item returns to yellow (unapproved); the toggle is durable: after quitting and re-opening the TUI the state reflects what was toggled

## Commit (user)
`feat(tui): space key toggles approval on the focused item`