---
id: BIT-11.4
title: q/esc close the modal, app stays open
status: todo
phase: 1
phase_label: Open & dismiss
---
q or esc closes the modal and returns to the board with the app still running. Contradicts today's board behavior, where q/esc return `tea.Quit` — with the modal open they must close it and swallow, not quit. Forces a modal-open input branch ahead of the normal board handling.

**Scope:**
- `tui/board.go` — at the top of `updateBoard` (or a `updateModal` helper it delegates to when `m.modalOpen`), intercept q and esc: set `m.modalOpen = false` and return `m, nil` (no command). This branch runs before the existing `case "q","esc","ctrl+c": return m, tea.Quit`.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestUpdate_ModalCloses` (table-driven: `q`, `esc`)
     - **Behavior:** with the modal open, q and esc each close it without quitting the app.
     - **Setup:** `New` with one `todo` task + body; `WindowSizeMsg{80,24}`; `Tab`; `Enter` to open. Then `Update` with the subcase key: `tea.KeyPressMsg{Code: 'q', Text: "q"}` / `tea.KeyPressMsg{Code: tea.KeyEsc}`.
     - **Assertions:** `mdl.(model).modalOpen` == `false` **and** the returned `cmd` is `nil` (not a quit command).
     - **Boundary:** the close keys while `modalOpen == true` — the state the app is only ever in on the board; proves q/esc are diverted from their quit meaning while the modal owns input.
   - [ ] Confirm fails: q/esc return `tea.Quit` (cmd non-nil) and leave `modalOpen` true.

2. **Implement (GREEN):**
   - [ ] Add the modal-open branch closing on q/esc and swallowing.

**Claude verifies:**
- [ ] tests pass (`just test`)
- [ ] linter passes (`just lint`)

**User verifies:**
- [ ] none — deterministic (the whole-slice check lands on the next bar).

**Commit (user):** `feat(tui): close the board modal on q/esc`