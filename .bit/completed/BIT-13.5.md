---
id: BIT-13.5
title: A reload keeps the list selection on the same task
status: done
phase: 2
phase_label: Keep your place
---
`setTasks` rebuilds via `SetItems`, which resets the cursor to the top — so a live reload yanks the human to index 0. Capture the selected task's ID before the rebuild and re-select it after, so selection tracks the task, not the row number.

**Scope:**
- `tui/model.go` — in the reload apply path, record the selected task's ID before rebuilding the list; after `SetItems`, find that ID's new index and `m.Select` it (the not-found/clamp case is the next step); run `m.refreshDetail()` after re-selecting.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestUpdate_ReloadPreservesListSelection`
     - **Behavior:** after a reload that reorders/extends the set, the same task stays selected.
     - **Setup:** `m := New([]*task.Task{{ID:"BIT-2"},{ID:"BIT-2.1"},{ID:"BIT-1"}})`; `m.Select(2)` (BIT-1); `updated, _ := m.Update(reloadedMsg{tasks: []*task.Task{{ID:"BIT-3"},{ID:"BIT-2"},{ID:"BIT-2.1"},{ID:"BIT-1"}}})`.
     - **Assertions:** `updated.(model).selected().ID == "BIT-1"` (now at index 3).
     - **Boundary:** selected task moves index 2 → 3 — selection follows the ID across a reorder, not the position.
   - [ ] Confirm fails: `SetItems` reset the cursor to 0, so `selected().ID == "BIT-3"`.

2. **Implement (GREEN):**
   - [ ] Before rebuild, capture `prevID` from `m.selected()` (guard nil). After `SetItems`, scan the new items for `prevID` and `m.Select(i)` when found; then `m.refreshDetail()`.

**Claude verifies:**
- [ ] `just test` passes
- [ ] `just lint` clean

**User verifies:**
- [ ] none — deterministic

**Commit (user):** `feat(tui): keep the list selection across a reload`