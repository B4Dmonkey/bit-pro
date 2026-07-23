---
id: BIT-4.2
title: Update forwards messages to the embedded list
status: done
phase: 1
phase_label: open & navigate
---
## Step 2 (Phase 1 — open & navigate) — Update forwards messages to the embedded list
**Status:** ✅ Done — verified 2026-07-18
A Bubble Tea program drives `model.Update`, but the promoted `list.Model.Update` returns
`list.Model`, not `tea.Model` — so the model isn't a `tea.Model` and can't be handed to
`tea.NewProgram` (Step 3). Our `Update` exists to adapt that signature and route messages
into the list; a navigation test forces it into being. **Quit is inherited, not owned:**
Bubbles' default list keymap binds `q`/`esc` → Quit and `ctrl+c` → ForceQuit
(`bubbles@v1.0.0 list/keys.go`; `list.go` returns `tea.Quit`), so delegating messages to
the list gives clean quit for free — we write no quit logic. (Esc's meaning narrows only
inside detail mode in Step 5, by intercepting it there before delegating.)

**Scope:**
- `tui/model.go` — `Update(msg tea.Msg) (tea.Model, tea.Cmd)` that delegates to
  `m.Model.Update(msg)`, persists the returned `list.Model` back into the embedded field,
  and returns `m` with the list's cmd.

**TDD cycle:**

1. **Write test (RED):**
   - [x] `TestUpdate_ForwardsNavigationToList`
     - **Behavior:** a navigation key routed through our `Update` actually moves the list
       cursor — proving `Update` delegates to the embedded list *and* persists the updated
       model, rather than dropping the `list.Model` the library returns (the classic
       "cursor frozen at 0" bug).
     - **Setup:** `New` with 3 tasks (cursor at index 0); send `tea.KeyMsg{Type: tea.KeyDown}`.
     - **Assertions:** the returned model's `Index() == 1`.
     - **Boundary:** cursor 0 → 1, the first move — the minimal proof that delegation plus
       state persistence works.
   - [x] Confirm fails: won't compile — without our `Update`, `m.Update` yields the promoted
     `list.Model` (a concrete type), so reading it back as `model` has no valid assertion.

2. **Implement (GREEN):**
   - [x] `Update` assigns `m.Model, cmd = m.Model.Update(msg)` and returns `m, cmd`. This one
     delegation carries navigation *and* the inherited quit keys.
   - [x] `Init() tea.Cmd { return nil }` — forced forward from Step 3: `Update` returning
     `tea.Model` makes the compiler demand the full interface, and `list.Model` promotes
     `View` but not `Init`, so this is the minimum to make the test compile.

**Claude verifies:**
- [ ] `just test` green
- [ ] `just lint` clean

**Commit (user):** `feat(tui): forward update to the embedded list`