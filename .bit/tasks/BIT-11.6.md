---
id: BIT-11.6
title: Long bodies scroll inside the modal
status: done
phase: 2
phase_label: Scroll in modal
---
A body taller than the modal scrolls inside it instead of overflowing the box. Forces the modal-open branch to route ctrl+u/ctrl+d to `modalViewport` (carving them out of the swallow-everything rule) and proves the viewport bound set up when the modal opened actually clips.

**Scope:**
- `tui/board.go` — in the modal-open branch, before the swallow, forward ctrl+u/ctrl+d to the modal viewport: `m.modalViewport, cmd = m.modalViewport.Update(msg); return m, cmd`. (`viewport.New()` already binds ctrl+u/ctrl+d to half-page scroll via its default keymap — verified — so no keymap wiring is needed.)

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestUpdate_ModalScrollsLongBody`
     - **Behavior:** ctrl+d scrolls the modal's content, and a long body is clipped to the modal rather than overflowing the window.
     - **Setup:** `New` with one `todo` task whose `Body` is `strings.Repeat("line\n", 500)`; `WindowSizeMsg{80,24}`; `Tab`; `Enter` to open; then `Update(tea.KeyPressMsg{Code: 'd', Mod: tea.ModCtrl})`.
     - **Assertions:** `mdl.(model).modalViewport.YOffset() > 0` after ctrl+d (was `0` immediately after open); and `lipgloss.Height(mdl.(model).View().Content) <= 24` with the 500-line body (clipped, not overflowing).
     - **Boundary:** body far taller than the modal — the overflow case; YOffset moves off its `0` floor and total height stays within the window.
   - [ ] Confirm fails: ctrl+d is swallowed (YOffset stays 0) because the modal branch forwards nothing to the viewport.

2. **Implement (GREEN):**
   - [ ] Route ctrl+u/ctrl+d to `modalViewport.Update` in the modal-open branch.

**Claude verifies:**
- [ ] tests pass (`just test`)
- [ ] linter passes (`just lint`)

**User verifies:**
- [ ] In `bit tui`, open the modal on a long track (e.g. BIT-2): the body sits inside the box and scrolls with ctrl+u/ctrl+d — it never spills past the modal border or the screen.

**Commit (user):** `feat(tui): scroll long bodies inside the board modal`