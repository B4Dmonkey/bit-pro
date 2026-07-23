---
id: BIT-11.2
title: Enter on an empty column is a no-op
status: todo
phase: 1
phase_label: Open & dismiss
---
Enter on an empty column does nothing. The unconditional `modalOpen = true` from the previous bar can't satisfy this — it forces a guard on whether the active column actually has a selected card.

**Scope:**
- `tui/board.go` — add a `boardSelected()` helper (the active column's selected `*task.Task`, or `nil` when the column is empty), and gate the Enter handler on it: open the modal only when `boardSelected() != nil`.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestUpdate_BoardEnterEmptyColumnNoop`
     - **Behavior:** Enter while the active column has no cards leaves the modal closed.
     - **Setup:** `New` with tasks that populate only the Doing column (e.g. one `doing` task), so the default active column (To Do, index 0) is empty; `WindowSizeMsg{80,24}`; `Tab` to board; `Update(tea.KeyPressMsg{Code: tea.KeyEnter})`.
     - **Assertions:** `mdl.(model).modalOpen` == `false`.
     - **Boundary:** active column card count == 0 — the lower bound; proves Enter is a no-op with nothing to open, contradicting the always-open behavior from the prior bar.
   - [ ] Confirm fails: modal opens anyway (`modalOpen` == true), because Enter is currently unconditional.

2. **Implement (GREEN):**
   - [ ] Add `boardSelected()` returning the active column's selected task or `nil` (an empty `list.Model` yields no `SelectedItem`).
   - [ ] In `updateBoard`, only set `m.modalOpen = true` when `m.boardSelected() != nil`.

**Claude verifies:**
- [ ] tests pass (`just test`)
- [ ] linter passes (`just lint`)

**User verifies:**
- [ ] none — deterministic.

**Commit (user):** `feat(tui): ignore board Enter on an empty column`