---
id: BIT-13.9
title: A failed reload holds the last good view
status: todo
phase: 3
phase_label: Quiet & safe
---
A tick can read `.bit/tasks/` mid-write and get a parse or I/O error. When the reload errors, hold the current view and keep polling instead of blanking or crashing.

**Scope:**
- `tui/model.go` — at the top of the `reloadedMsg` handler, when `msg.err != nil`, return `m, tick()` without touching the view (before the change-detection and apply).

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestUpdate_ReloadErrorHoldsView`
     - **Behavior:** a reloadedMsg carrying an error leaves the displayed task set unchanged and keeps the poll alive.
     - **Setup:** `m := New(nil).WithReload(func() ([]*task.Task, error) { return nil, nil })`; apply a good set `S` (`Update(reloadedMsg{tasks: []*task.Task{{ID:"BIT-1"},{ID:"BIT-2"}}})`); then `updated, cmd := m.Update(reloadedMsg{tasks: nil, err: errors.New("mid-write")})`.
     - **Assertions:** `len(updated.(model).Items()) == 2` (unchanged); `cmd != nil` (poll continues).
     - **Boundary:** `err != nil` with `tasks == nil` — the failure case must not rebuild to empty.
   - [ ] Confirm fails: the handler rebuilds from `msg.tasks` (nil), emptying the list.

2. **Implement (GREEN):**
   - [ ] Add `if msg.err != nil { return m, tick() }` as the first line of the `reloadedMsg` case.

**Claude verifies:**
- [ ] `just test` passes
- [ ] `just lint` clean

**User verifies:**
- [ ] With `bit tui` open and idle, watch for ~10s and confirm no flicker or redraw while nothing changes. Then make a rapid burst of CLI edits (e.g. create three tasks back-to-back) and confirm the board settles to the final state in one update, and the TUI never blanks or crashes if a read races a write. (Whole Verse 3 slice: the poll is quiet when idle and safe under churn.)

**Commit (user):** `feat(tui): hold the last good view when a reload errors`