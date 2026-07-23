---
id: BIT-11.3
title: Modal floats the card body over the board
status: done
phase: 1
phase_label: Open & dismiss
---
The open modal renders the selected card's body in a bordered box floated over the board. Forces the compositor, the modal viewport, and the shared render path into existence — a flag with no view can't put the body on screen.

**Scope:**
- `tui/model.go` — add `modalViewport viewport.Model` to `model` (built in `New` with `viewport.New()`, like `m.viewport`). Add a `refreshModal()` (pointer receiver): render `boardSelected().Body` through `newRenderer(m.style, innerWidth)` (same helper the detail pane uses — one glamour path), `SetContent`, `GotoTop`, and size `modalViewport` to the modal's inner dims. Call it when the modal opens.
- `tui/board.go` — in `boardView`/a new `modalView(m)`, when `m.modalOpen` build the box with the existing `titledBorder(m.modalViewport.View(), title, innerW, innerH, true)` (title = selected card's `ID + " — " + Title`), then float it over the board string with `lipgloss.NewCompositor(base, modal).Render()` where `base := lipgloss.NewLayer(boardStr)` and `modal := lipgloss.NewLayer(boxStr).X(cx).Y(cy).Z(1)`, centered. Note: `Layer.Draw` does not recurse into children, so `NewCompositor(...).Render()` (which flattens + draws in z-order) is the correct primitive, not `Canvas.Compose(rawLayer)`.
- Modal dimensions (assumption — tune to taste): `modalW = min(m.winWidth-8, 80)`, `modalH = m.winHeight-6`; `cx = (winWidth-modalW)/2`, `cy = (winHeight-modalH)/2`; inner dims follow `titledBorder`'s `+2/+1` geometry. Fold a minimal modal help hint (`q/esc: close · ctrl+c: quit`) into the composed output.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestView_ModalShowsBody` (table-driven: modal closed vs open)
     - **Behavior:** with the modal open, `View().Content` contains text from the selected card's body that never appears on a card face; with it closed, that text is absent.
     - **Setup:** `New` with one `todo` task whose `Body` contains a distinctive short token on its own line (e.g. `MODALBODYTOKEN`) — a plain alphanumeric word so glamour renders it verbatim; `WindowSizeMsg{80,24}`; `Tab` to board. Subcases: (a) no Enter; (b) `Update(tea.KeyPressMsg{Code: tea.KeyEnter})`.
     - **Assertions:** `strings.Contains(mdl.(model).View().Content, "MODALBODYTOKEN")` is `false` for (a), `true` for (b).
     - **Boundary:** body text present only in the modal layer — proves the compositor draws the floated box over the board (card faces show ID/title, never body).
   - [ ] Confirm fails: token absent when open — no modal is composed into the view yet.

2. **Implement (GREEN):**
   - [ ] Add `modalViewport` + build it in `New`.
   - [ ] Add `refreshModal()`; call it where the modal opens (`updateBoard` Enter handler).
   - [ ] Compose the modal box over the board in the board view when `m.modalOpen`.

**Claude verifies:**
- [ ] tests pass (`just test`)
- [ ] linter passes (`just lint`)

**User verifies:**
- [ ] In `bit tui`, `tab` to the board, select a card, press Enter: a single-bordered box floats centered over the board (board still visible around it, no dimming), showing that card's rendered body.

**Commit (user):** `feat(tui): float a card-details modal over the board`