---
id: BIT-11.5
title: Modal captures input while open
status: done
phase: 1
phase_label: Open & dismiss
---
The modal captures input while open: ctrl+c still quits from anywhere, but board navigation and tab are swallowed so the board can't move underneath it. Contradicts the close-only branch from the prior bar — that handled q/esc but let other keys fall through to `updateBoard` and move the board; this forces the branch to swallow everything except ctrl+c.

**Scope:**
- `tui/board.go` — in the modal-open branch, return `m, tea.Quit` for ctrl+c, and return `m, nil` (swallow) for every other key, so nav (`left`/`right`/`up`/`down`) and `tab` never reach the board handling while the modal is open.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestUpdate_ModalCapturesInput` (table-driven)
     - **Behavior:** with the modal open, ctrl+c quits, while board-nav and tab keys change nothing and issue no command.
     - **Setup:** `New` with tasks across columns (so nav would otherwise move `activeCol`) + bodies; `WindowSizeMsg{80,24}`; `Tab`; `Enter` to open. Subcases: ctrl+c (`{Code: 'c', Mod: tea.ModCtrl}`) → expect quit; `{Code: tea.KeyRight}` and `{Code: tea.KeyTab}` → expect swallowed.
     - **Assertions:** ctrl+c: `cmd` non-nil and `cmd().(tea.QuitMsg)` ok. Right: `mdl.(model).activeCol` unchanged (0) and `modalOpen` still true. Tab: `mdl.(model).mode` still `modeBoard` and `modalOpen` still true.
     - **Boundary:** each captured key while `modalOpen == true` — ctrl+c at the "always quits" extreme, nav/tab at the "fully swallowed" extreme; proves the modal owns input except for the hard quit.
   - [ ] Confirm fails: right advances `activeCol` / tab flips `mode` (keys fall through to the board), and/or ctrl+c isn't handled in the modal branch.

2. **Implement (GREEN):**
   - [ ] In the modal-open branch: ctrl+c → `tea.Quit`; any other unmatched key → swallow (`m, nil`).

**Claude verifies:**
- [ ] tests pass (`just test`)
- [ ] linter passes (`just lint`)

**User verifies:**
- [ ] Whole slice: on the board, Enter opens a card's details without leaving the board; while open, arrows/tab do nothing and q/esc return you to the board with the app still running; ctrl+c quits with the modal open. Re-open, and q/esc on the board (modal closed) still quit as before.

**Commit (user):** `feat(tui): capture input while the board modal is open`