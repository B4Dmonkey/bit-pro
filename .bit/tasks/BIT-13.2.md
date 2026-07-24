---
id: BIT-13.2
title: A reload message rebuilds the board columns
status: done
phase: 1
phase_label: See CLI edits
---
The list rebuild from the previous step leaves the board showing stale cards. A board-mode assertion contradicts that, forcing `setTasks` to also regroup the tasks into the three columns.

**Scope:**
- `tui/model.go` — extend `setTasks` to rebuild `m.boardCols` from `groupByStatus(tasks)` (reusing `newColumnList`), then re-apply sizes with `m.layout()` so the new columns are dimensioned, and refresh the detail pane (`m.refreshDetail()`).

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestUpdate_ReloadedMsgRebuildsBoard`
     - **Behavior:** a reloadedMsg regroups tasks into the board columns, so a new `todo` task appears in the To Do column without a restart.
     - **Setup:** `m := New([]*task.Task{{ID: "BIT-1", Status: "todo"}})`; size it (`Update(WindowSizeMsg{Width: 80, Height: 24})`); then `Update(reloadedMsg{tasks: []*task.Task{{ID: "BIT-1", Status: "todo"}, {ID: "BIT-2", Status: "todo"}}})`.
     - **Assertions:** the To Do column has 2 items after the reload (`len(updated.(model).boardCols[0].Items()) == 2`).
     - **Boundary:** To Do column count grows 1 → 2 — proves the board is rebuilt from the message, not just the list.
   - [ ] Confirm fails: `setTasks` rebuilds only the list, so `boardCols[0]` still has 1 item.

2. **Implement (GREEN):**
   - [ ] In `setTasks`, after `SetItems`, rebuild the columns: `for i, cards := range groupByStatus(tasks) { m.boardCols[i] = newColumnList(cards) }`, then `m.layout()` (sizes the new columns) and `m.refreshDetail()`.

**Claude verifies:**
- [ ] `just test` passes
- [ ] `just lint` clean

**User verifies:**
- [ ] none — deterministic

**Commit (user):** `feat(tui): rebuild the board on a reload message`