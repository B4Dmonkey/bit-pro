---
id: BIT-13.6
title: A reload that drops the selected task clamps safely
status: done
phase: 2
phase_label: Keep your place
---
Restoring selection by ID must survive the selected task disappearing (archived or deleted between polls). When the previous ID is absent from the new set, clamp to a valid row instead of leaving an out-of-range cursor.

**Scope:**
- `tui/model.go` — in the selection-restore path, when `prevID` isn't found, clamp the index into `[0, len-1]`, and handle the empty set (no `Select`).

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestUpdate_ReloadSelectionGoneClamps`
     - **Behavior:** when the selected task is gone after a reload, the cursor lands on a valid item rather than pointing past the end.
     - **Setup:** `m := New([]*task.Task{{ID:"BIT-2"},{ID:"BIT-1"}})`; `m.Select(1)` (BIT-1); `updated, _ := m.Update(reloadedMsg{tasks: []*task.Task{{ID:"BIT-2"}}})`.
     - **Assertions:** no panic; `updated.(model).Index() == 0`; `selected().ID == "BIT-2"`.
     - **Boundary:** previous ID absent from a set that also shrank below the old index (the not-found edge) — must clamp, not overrun.
   - [ ] Confirm fails: restore leaves the index at 1, out of range for a 1-item list (panic or stale selection).

2. **Implement (GREEN):**
   - [ ] When `prevID` not found and `len(items) > 0`, `m.Select(min(prevIndex, len(items)-1))`; skip `Select` on an empty set.

**Claude verifies:**
- [ ] `just test` passes
- [ ] `just lint` clean

**User verifies:**
- [ ] none — deterministic

**Commit (user):** `fix(tui): clamp selection when the selected task disappears`