---
id: BIT-20.9
title: Enter toggles the list view's expanded detail pane
status: done
phase: 3
phase_label: Expanded detail pane
---
## **Verse 3**

Gives the operator a way to actually flip `detailExpanded`: `Enter` toggles it in list view,
the same role `Enter` already plays in board view (opening/closing the modal). `Enter` is
currently unbound in list mode — it falls through the `KeyPressMsg` switch in `model.go`'s
`Update` and reaches the embedded `list.Model`, which doesn't act on it.

## Scope
- `tui/model.go` — `Update`'s list-mode `KeyPressMsg` handling: add an explicit `tea.KeyEnter`
  case, guarded to list mode only (board mode's `Enter` handling in `updateBoard` is untouched).

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestUpdate_EnterExpandsDetail`
     - **Behavior:** pressing `Enter` in list view expands the detail pane.
     - **Setup:** `New([]*task.Task{{ID: "BIT-1"}})` in list mode (default is now board per
       Verse 1 — press `Tab` first to reach list mode), then `tea.KeyPressMsg{Code: tea.KeyEnter}`.
     - **Assertions:** `m.detailExpanded == true`.
     - **Boundary:** the off→on transition — the only interesting edge of a boolean toggle.
   - [ ] Confirm fails: `Enter` is unhandled, falls through to `list.Model.Update`, no field
     changes.

2. **Implement (GREEN):**
   - [ ] Add a case in the list-mode `KeyPressMsg` switch: `case tea.KeyEnter: m.detailExpanded = true; return m, nil` — hardcoded one-way flip, enough to pass the test above.

3. **More tests (RED → GREEN):**
   - [ ] `TestUpdate_EnterTogglesDetailBackAndForth`
     - **Behavior:** pressing `Enter` again collapses back to the normal split — it's a toggle,
       not a one-way switch.
     - **Setup:** same as above, press `Enter` twice.
     - **Assertions:** `m.detailExpanded == false` after the second press.
     - **Boundary:** contradicts the hardcoded always-`true` — forces the real `!m.detailExpanded` flip.
   - [ ] Implement the real toggle: `m.detailExpanded = !m.detailExpanded`.
   - [ ] `TestUpdate_BoardEnterOpensModal` and the other existing board-modal `Enter` tests in
     `tui/board_test.go` stay green unmodified — confirms this case is genuinely list-mode-only
     and doesn't shadow board's `Enter` handling.

## Claude verifies
- [ ] `go test ./tui/...` passes, including the untouched board-mode `Enter` tests.
- [ ] `golangci-lint run` passes.

## User verifies
- [ ] None — the pane visually resizes (Bar 3.2 is already wired), but paging while expanded
  isn't in yet; the full flow is verified on this verse's last bar.

## Commit (user)
`feat(tui): toggle the list view's expanded detail pane with Enter`