---
id: BIT-11.1
title: Enter flags the modal open
status: done
phase: 1
phase_label: Open & dismiss
---
Enter on the board flips modal-open state on. Forces the `modalOpen` field into existence; nothing renders it yet.

**Scope:**
- `tui/model.go` — add `modalOpen bool` to the `model` struct.
- `tui/board.go` — in `updateBoard`, handle Enter (`msg.Code == tea.KeyEnter`, a verified v2 constant) by setting `m.modalOpen = true` and returning `m, nil`. Unconditional for now — the empty-column guard is the next bar.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestUpdate_BoardEnterOpensModal` (table-driven)
     - **Behavior:** Enter in board mode marks the modal open; without it the modal stays closed.
     - **Setup:** `New` with one `todo` task carrying a `Body`; `Update(tea.WindowSizeMsg{Width: 80, Height: 24})`; `Update(tea.KeyPressMsg{Code: tea.KeyTab})` to reach board mode. Two subcases: (a) no further key; (b) `Update(tea.KeyPressMsg{Code: tea.KeyEnter})`.
     - **Assertions:** `mdl.(model).modalOpen` == `false` for (a), `true` for (b).
     - **Boundary:** the open flag in both states — `false` is the zero-value default, `true` is the single transition Enter causes.
   - [ ] Confirm fails: `mdl.(model).modalOpen` undefined — `model` has no such field.

2. **Implement (GREEN):**
   - [ ] Add `modalOpen bool` to `model`.
   - [ ] In `updateBoard`, set `m.modalOpen = true` on `tea.KeyEnter`.

**Claude verifies:**
- [ ] tests pass (`just test`)
- [ ] linter passes (`just lint`)

**User verifies:**
- [ ] none — deterministic.

**Commit (user):** `feat(tui): flag modal open on board Enter`