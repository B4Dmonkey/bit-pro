---
id: BIT-13.7
title: A reload keeps the board card, column, and mode
status: todo
phase: 2
phase_label: Keep your place
---
The board rebuild resets each column's cursor, and the modal reads from it. Preserve the active column's selected card by ID across the rebuild; the active column, view mode, and open modal are model fields the reload must simply leave untouched.

**Scope:**
- `tui/model.go` — capture the active column's selected task ID before rebuilding `boardCols`; after rebuild, re-select it in that column; ensure the reload apply path leaves `activeCol`, `mode`, `modalOpen`, and `modalViewport` untouched, and re-`refreshModal()` when a modal is open.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestUpdate_ReloadPreservesBoardSelection`
     - **Behavior:** after a reload, the board keeps the active column and its selected card, so an edit elsewhere doesn't move the human.
     - **Setup:** tasks spanning columns: `{ID:"BIT-1",Status:"todo"},{ID:"BIT-2",Status:"doing"},{ID:"BIT-3",Status:"doing"}`; size; `tab` into board; `right` to the Doing column (`activeCol==1`); `down` to its second card (BIT-3); reload adding `{ID:"BIT-4",Status:"todo"}` (leaves Doing unchanged).
     - **Assertions:** `activeCol == 1`; `boardSelected().ID == "BIT-3"`; `mode == modeBoard`.
     - **Boundary:** active-column cursor preserved by ID while a *different* column changes — proves per-column restore, not a global reset.
   - [ ] Confirm fails: rebuilding `boardCols` resets the Doing cursor to its first card (BIT-2).

2. **Implement (GREEN):**
   - [ ] Capture `m.boardSelected()` ID before rebuild; after rebuilding `boardCols`, scan the active column for that ID and `Select` it; keep `activeCol`/`mode`/`modalOpen` as-is; `if m.modalOpen { m.refreshModal() }`.

**Claude verifies:**
- [ ] `just test` passes
- [ ] `just lint` clean

**User verifies:**
- [ ] In `bit tui`, scroll partway down the list, then `tab` to the board, select a card in the Doing column, and open its modal (`enter`). From another terminal edit a *different* task. Confirm your list position, board column, selected card, view mode, and the open modal all stay put while the change appears. (Whole Verse 2 slice: a reload never pulls you out of context.)

**Commit (user):** `feat(tui): keep board selection, column, and mode across a reload`