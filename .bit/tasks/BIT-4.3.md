---
id: BIT-4.3
title: Wire `bit tui` to launch the model
status: done
phase: 1
phase_label: open & navigate
---
## Step 3 (Phase 1 — open & navigate) — Wire `bit tui` to launch the model
**Status:** ✅ Done — verified 2026-07-18
The walking skeleton: a real `bit tui` that reads the store and paints a drivable screen.
Pure shell — no unit test, because launching the program needs a TTY (the scope's accepted
"known, not unknown"). This is where the deps land and where you first see it run.

**Scope:**
- `go.mod` — `go get` `bubbletea`, `bubbles`, `lipgloss` (latest).
- `tui/run.go` — new: `Run(m model) error` wrapping `tea.NewProgram(m, tea.WithAltScreen()).Run()`.
- `tui/model.go` — extend `Update` (from Step 2) to size the embedded list on
  `tea.WindowSizeMsg` (without a size the list renders empty); `View() string` delegates to
  `list.View()`. (`Init() tea.Cmd` already landed in Step 2, forced by the `tea.Model`
  signature.) Call `SetFilteringEnabled(false)` in `New` — filtering is out of scope, and
  with it off `enter` stays unbound so Step 5 can own it.
- `cmd/tui.go` — new: `newTUICmd()`; `RunE` = `task.New(bitDir).List()` → `tui.New(tasks)`
  → `tui.Run(m)`. `Args: cobra.NoArgs`.
- `cmd/root.go` — register `newTUICmd()` in `NewRootCmd`.

**Implement (GREEN):**
- [x] Add the three deps; wire `run.go`, `Init`/`View`/window-size in `model.go`, the
  command, and its registration. (bubbletea/bubbles are direct imports; lipgloss stays
  `// indirect` until Step 8 uses it — YAGNI. `Init` already landed in Step 2. Window
  size is handled by intercepting `tea.WindowSizeMsg` and calling `m.SetSize` — the
  list's own `Update` doesn't size itself.)

No cobra-layer test: driving `bit tui` through the test harness would block on a TTY. The
model logic it depends on is already covered by Steps 1–2.

**Claude verifies:**
- [x] `just build` succeeds
- [x] `just test` green (unchanged) and `just lint` clean
- [x] `go build ./...` — deps resolve

**User verifies:**
- [x] `bit tui` in this project lists tasks, newest track first with bars beneath
- [x] arrows / j-k move the selection; the list scrolls past a screenful
- [x] `q` and `ctrl+c` both quit cleanly, terminal restored (alt-screen)

**Commit (user):** `feat(tui): add read-only bit tui command`