---
id: BIT-5.1
title: '`tab` toggles the view mode'
status: done
phase: 1
phase_label: flip to a board
---
## Step 1 (Phase 1 — flip to a board) — `tab` toggles the view mode
**Status:** ✅ Done — verified 2026-07-19
The walking skeleton's hinge: one key flips the model between list and board mode. Forced
by a both-directions contradiction — a handler that hardcodes `mode = modeBoard` satisfies
list→board but fails board→list, so it must actually toggle.

**Scope:**
- `tui/model.go` — new `viewMode` `iota` enum (`modeList` default, `modeBoard`); a `mode
  viewMode` field on `model`; in `Update`'s `tea.KeyMsg` branch, intercept `tea.KeyTab` to
  flip `mode` and return (placed alongside the existing `?` intercept, before the
  focus/forward logic so it fires in either mode).
- `tui/model_test.go` — `TestUpdate_TabTogglesMode`.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestUpdate_TabTogglesMode` (table)
     - **Behavior:** `tab` switches the model between list and board mode and back, so a
       human can flip views with one key and return the same way — the scope's core gesture.
     - **Setup:** `New([]*task.Task{{ID: "BIT-1"}})`; send `tea.KeyMsg{Type: tea.KeyTab}`
       zero, one, and two times.
     - **Assertions:** 0 presses → `m.mode == modeList` (default); 1 → `modeBoard`; 2 →
       `modeList`.
     - **Boundary:** both transitions — list→board and board→list; the second is what a
       hardcoded assignment can't satisfy.
   - [x] Confirm fails: `model` has no `mode` field / `modeBoard` undefined (compile error).

2. **Implement (GREEN):**
   - [x] Add the `viewMode` enum and `mode` field; intercept `tea.KeyTab` in `Update` to
     flip `mode`. `View` is unchanged — pressing `tab` is not yet visible; Steps 2–3 make
     it so. (Model logic before the paint, exactly as list Steps 1–2 preceded Step 3.)

**Claude verifies:**
- [x] `just test` green
- [x] `just lint` clean

**Commit (user):** `feat(tui): toggle between list and board view modes`