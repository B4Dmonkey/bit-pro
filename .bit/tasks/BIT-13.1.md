---
id: BIT-13.1
title: A reload message rebuilds the list
status: done
phase: 1
phase_label: See CLI edits
---
Introduce the reload result message and the rebuild path it drives. A model built once from a snapshot can't reflect a task set it was never handed again, so a `reloadedMsg` carrying a fresh slice forces a rebuild of the list's items.

**Scope:**
- `tui/model.go` — add a `reload func() ([]*task.Task, error)` field (left nil by `New`); add `reloadedMsg struct { tasks []*task.Task; err error }`; add a `setTasks([]*task.Task)` helper that rebuilds the list items via `SetItems`; handle `reloadedMsg` in `Update` by calling `setTasks(msg.tasks)`.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestUpdate_ReloadedMsgRebuildsList`
     - **Behavior:** a reloadedMsg replaces the list's items with the incoming task set, so a task added through the CLI shows up.
     - **Setup:** `m := New([]*task.Task{{ID: "BIT-1"}})`; `updated, _ := m.Update(reloadedMsg{tasks: []*task.Task{{ID: "BIT-1"}, {ID: "BIT-2"}}})`.
     - **Assertions:** `len(updated.(model).Items()) == 2`; `items[0].(item).t.ID == "BIT-1"`, `items[1].(item).t.ID == "BIT-2"`.
     - **Boundary:** item count grows 1 → 2 — the add path; proves the list rebuilds from the message, not from the original `New` snapshot.
   - [ ] Confirm fails: no `reloadedMsg` case in `Update` (won't compile until the type exists), then `Items()` stays length 1.

2. **Implement (GREEN):**
   - [ ] Add the `reload` field and `reloadedMsg` type. Add `setTasks` (list only for now: build `[]list.Item` from the tasks, then `m.SetItems(items)`). Add `case reloadedMsg:` to `Update` calling `m.setTasks(msg.tasks)` and returning `m, nil`.

**Claude verifies:**
- [ ] `just test` passes (new test green, existing green)
- [ ] `just lint` clean

**User verifies:**
- [ ] none — deterministic

**Commit (user):** `feat(tui): rebuild the task list on a reload message`