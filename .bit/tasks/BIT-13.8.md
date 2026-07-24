---
id: BIT-13.8
title: A reload only rebuilds when the listing changed
status: done
phase: 3
phase_label: Quiet & safe
---
Every tick currently rebuilds, redrawing the screen even when nothing changed. Compare the incoming set against the last applied one and skip the rebuild when they match, so an idle TUI stays quiet and a burst of writes between ticks collapses into a single refresh.

**Scope:**
- `tui/model.go` — add a `loaded []*task.Task` field seeded by `New` and updated on each apply; add `func sameTasks(a, b []*task.Task) bool` comparing the view-relevant fields (ID, Status, Title, Body, Phase, PhaseLabel) in order; in the `reloadedMsg` handler, when `sameTasks(m.loaded, msg.tasks)`, skip the rebuild and just reschedule.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestSameTasks` (table-driven)
     - **Behavior:** `sameTasks` is true only when the two sets render identically.
     - **Setup / Assertions:** identical single task → true; different length → false; same length different ID → false; same ID different Status → false; same ID different Body → false; two tasks reordered → false.
     - **Boundary:** each discriminating field (length, ID, Status, Body) flipped one at a time, plus order — proves the comparison covers what the view shows.
   - [ ] Confirm fails: `sameTasks` doesn't exist (won't compile).

2. **Implement (GREEN):**
   - [ ] Add `loaded` (set in `New` to the initial tasks, and set at the end of the apply path to `msg.tasks`). Add `sameTasks` (length check, then per-index field compare). Guard the `reloadedMsg` rebuild: `if sameTasks(m.loaded, msg.tasks) { return m, tick() }`, else apply, set `m.loaded = msg.tasks`, and `return m, tick()`.

**Claude verifies:**
- [ ] `just test` passes (including `TestSameTasks`)
- [ ] `just lint` clean

**User verifies:**
- [ ] none — deterministic (the no-flicker observation lives on the last bar)

**Commit (user):** `feat(tui): skip the rebuild when the task set is unchanged`