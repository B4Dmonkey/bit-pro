---
id: BIT-13.4
title: The poll loop starts, sustains itself, wired to store
status: todo
phase: 1
phase_label: See CLI edits
---
The read and rebuild halves exist but nothing drives them on a schedule, and the real command still loads once. Start the loop in `Init`, reschedule it after each reload, and hand the running TUI a store-backed reload source so an edit appears on its own.

**Scope:**
- `tui/model.go` — add `const pollInterval = time.Second` and `func tick() tea.Cmd { return tea.Tick(pollInterval, func(time.Time) tea.Msg { return tickMsg{} }) }`; make `Init` return `tick()` when `reload != nil`, else `nil`; make the `reloadedMsg` handler return `m, tick()` to schedule the next poll; add exported `func (m model) WithReload(r func() ([]*task.Task, error)) model { m.reload = r; return m }`. New import: `time`.
- `cmd/tui.go` — build `s := task.New(bitDir)`, `tasks, err := s.List()`, then `tui.Run(tui.New(tasks).WithReload(s.List))` (the method value `s.List` has exactly type `func() ([]*task.Task, error)`).

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestInit_StartsPollingWhenReloadSet`
     - **Behavior:** with a reload source attached, `Init` returns a command (the first poll); with none, it returns nil, so snapshot-only construction is unchanged.
     - **Setup:** `set := New(nil).WithReload(func() ([]*task.Task, error) { return nil, nil })`; `none := New(nil)`.
     - **Assertions:** `set.Init() != nil`; `none.Init() == nil`.
     - **Boundary:** reload present vs absent — both states of the guard.
   - [ ] `TestUpdate_ReloadedMsgReschedules`
     - **Behavior:** applying a reloadedMsg returns a command, so the poll keeps going.
     - **Setup:** `m := New(nil).WithReload(func() ([]*task.Task, error) { return nil, nil })`; `_, cmd := m.Update(reloadedMsg{tasks: nil})`.
     - **Assertions:** `cmd != nil`.
     - **Boundary:** the loop is self-sustaining — reloadedMsg always yields the next tick.
   - [ ] Confirm fails: `Init` returns nil and the reloadedMsg handler returns `nil`.

2. **Implement (GREEN):**
   - [ ] Add `pollInterval`, `tick()`, `WithReload`; make `Init` conditional on `reload`; make the `reloadedMsg` case end with `return m, tick()`.
   - [ ] Rewrite `cmd/tui.go`'s `RunE` to attach `s.List` via `WithReload`.

**Claude verifies:**
- [ ] `just test` passes
- [ ] `just lint` clean
- [ ] `just build` succeeds

**User verifies:**
- [ ] With `bit tui` open, in another terminal run `bit task create "reload smoke test"`; within ~1s the new task appears in the list, and pressing `tab` shows it in the To Do column — no quit-and-relaunch. (Whole Verse 1 slice: CLI edits land in the open TUI. Selection may jump to the top — Verse 2 fixes that.)

**Commit (user):** `feat(tui): poll the store and live-reload the open TUI`