---
id: BIT-13.3
title: A tick triggers a reload from the injected source
status: done
phase: 1
phase_label: See CLI edits
---
Nothing produces a reloadedMsg yet. A `tickMsg` must run the model's reload source off the render loop and surface the result as a reloadedMsg — this wires the read half of the poll.

**Scope:**
- `tui/model.go` — add `tickMsg struct{}`; add a `reloadCmd() tea.Cmd` that calls `m.reload()` and returns `reloadedMsg{tasks, err}` (guarding a nil `reload`); handle `case tickMsg:` in `Update` by returning `m, m.reloadCmd()`.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestUpdate_TickTriggersReload`
     - **Behavior:** a tickMsg invokes the injected reload func and delivers its tasks as a reloadedMsg, so the poll actually reads the store.
     - **Setup:** `m := New(nil)`; `m.reload = func() ([]*task.Task, error) { return []*task.Task{{ID: "BIT-9"}}, nil }`; `_, cmd := m.Update(tickMsg{})`; `msg := cmd()`.
     - **Assertions:** `cmd != nil`; `rm, ok := msg.(reloadedMsg)` and `ok`; `len(rm.tasks) == 1 && rm.tasks[0].ID == "BIT-9"`; `rm.err == nil`.
     - **Boundary:** reload invoked exactly once and its result surfaced — the single happy read path.
   - [ ] Confirm fails: no `tickMsg` case (won't compile until the type exists), then `cmd` is nil.

2. **Implement (GREEN):**
   - [ ] Add `tickMsg`. Add `reloadCmd`: `if m.reload == nil { return nil }; reload := m.reload; return func() tea.Msg { tasks, err := reload(); return reloadedMsg{tasks: tasks, err: err} }` (capture the func value, not the whole model). Add `case tickMsg: return m, m.reloadCmd()`.

**Claude verifies:**
- [ ] `just test` passes
- [ ] `just lint` clean

**User verifies:**
- [ ] none — deterministic

**Commit (user):** `feat(tui): reload from the store when the poll ticks`