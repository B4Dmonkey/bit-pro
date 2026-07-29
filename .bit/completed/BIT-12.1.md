---
id: BIT-12.1
title: Port the tui package to the Charm v2 stack
status: done
phase: 1
phase_label: v2 migration
---
Port the whole `tui/` package and `go.mod` from the Charm v1 stack to v2 in one atomic
commit. This is the refactor case, not a RED-first cycle: the migration adds no behavior
(scope Decision 2), so there is no new failing test to write — the guardrail is the 32
existing `tui/` tests, which pass on v1 now and must pass unchanged on v2. What forces a
single bar rather than a ladder: `go.mod` pointing at v2 breaks the build until every file
in `tui/` is ported (scope Decision 4, "one atomic cut"), so there is no committable green
intermediate to split on.

**Scope:**
- `go.mod` — bump the four Charm deps to the pinned v2 set: bubbletea v2.0.8, bubbles
  v2.1.1, lipgloss v2.0.5, glamour v2.0.1. Import paths gain `/v2`. Run `go mod tidy` so the
  indirect set and `go.sum` resolve to the v2 tree. `termenv` may fall out of direct use
  (see the background-detection note below) — let `tidy` decide.
- `tui/run.go` — `tea.NewProgram(m, tea.WithAltScreen()).Run()`.
- `tui/model.go` — the bulk. `Update` key handling, `WindowSizeMsg`, glamour renderer
  construction + light/dark style selection, viewport, help, list, and all lipgloss styling
  in `titledBorder`/`View`.
- `tui/board.go` — `updateBoard` key handling and `boardView` lipgloss joins.
- `tui/delegate.go` — the `list.ItemDelegate` implementation and its lipgloss styles.

**Consult before editing:** the official Charm v2 migration guides (bubbletea, lipgloss,
bubbles, glamour) for the exact v2 replacements — do not port from memory. Apply the **go**
skill's Charm guidance as you write, and per scope Decision 3 take the cleaner v2 idiom
where one exists (a typed API over a stringly-typed one, a built-in over a hand-rolled
loop) rather than transliterating v1 call-for-call. The guardrail on all of it is that the
existing tests stay green and observable behavior stays identical.

**v1 surfaces currently in use** (find-and-port checklist — every one must resolve to its
v2 form):
- bubbletea: `tea.NewProgram` + `tea.WithAltScreen`; `tea.Model`/`tea.Cmd`/`tea.Msg`;
  `tea.WindowSizeMsg` (`.Width`/`.Height`); `tea.KeyMsg` with `msg.Type` +
  `tea.KeyTab`/`tea.KeyRight`/`tea.KeyLeft` and `msg.String()`; `tea.Quit`.
- bubbles: `list` (`.Model`/`.New`/`.Item`/`.SetSize`/`.Update`/`.View`/`.Index`/…);
  `help` (`.Model`/`.New`/`.View`/`help.KeyMap`); `key` (`.Binding`/`.NewBinding`/
  `.WithKeys`/`.WithHelp`/`key.Matches`); `viewport` (`.New`/`.SetContent`/`.GotoTop`/
  `.Update`/`.View`/`.Width`/`.Height`).
- lipgloss: `NewStyle` + `.Border`/`.BorderTop`/`.BorderForeground`/`.Foreground`/`.Bold`/
  `.Faint`/`.Italic`/`.Width`/`.Height`/`.MaxWidth`/`.Render`; `RoundedBorder()` +
  `.TopLeft`/`.Top`/`.TopRight`; `Color("99")`/`Color("245")`; `JoinVertical`/
  `JoinHorizontal` + `Left`/`Top`; `Height`/`Width`.
- glamour: `NewTermRenderer` + `WithStandardStyle`/`WithWordWrap`; `glamour/styles`
  (`LightStyle`/`DarkStyle`).
- termenv: `HasDarkBackground()`.

**Two behavior-risk spots — port carefully, they are where "appearance identical" can
silently drift:**
1. **Key handling.** bubbletea v2 reworks key messages (the v1 `msg.Type` +
   `tea.KeyTab`/`KeyRight`/`KeyLeft` enum style changes). `model.go`'s `Update` and
   `board.go`'s `updateBoard` both branch on it — `TestUpdate_TabTogglesMode`,
   `TestUpdate_Focus`, `TestUpdate_FocusRoutesArrows`, `TestUpdate_EscQuitsFromList`,
   `TestUpdate_QuitsFromDetail`, `TestUpdate_BoardQuits`, `TestUpdate_BoardActiveColumn`,
   `TestUpdate_BoardCardSelection` are the ones that pin this. If a test constructs a
   `tea.KeyMsg` literal, it moves to the v2 message type too.
2. **Background / style detection.** `New` calls `termenv.HasDarkBackground()` to pick
   glamour's `DarkStyle` vs `LightStyle`. v2 moves terminal-background detection into
   bubbletea (delivered to the model, not queried synchronously at construction). Preserve
   the *observable* outcome — the detail body renders in the same light/dark style as
   today. If the v2 path only learns the background after `Init`, the renderer may need to
   be (re)built when that message arrives rather than in `New`; keep the rendered output
   identical either way.

**TDD cycle (refactor — green to green, no new RED):**

1. **Baseline (confirm GREEN on v1):**
   - [ ] `just test` passes on the current v1 stack before touching anything — this is the
     guardrail you are about to hold invariant.

2. **Port (keep GREEN on v2):**
   - [ ] Bump `go.mod` to the v2 pins and switch every `tui/` import to its `/v2` path.
   - [ ] Port each v1 surface in the checklist above to its v2 form, file by file, applying
     v2 idioms where cleaner (Decision 3). The build will be red until the whole package is
     ported — that is expected for an atomic cut; do not commit a half-ported tree.
   - [ ] Update any test that constructs v1 library types directly (e.g. `tea.KeyMsg`
     literals) to the v2 equivalent — this is porting the guardrail, not changing what it
     asserts. The set of behaviors asserted must not change.
   - [ ] `go mod tidy`.

**Claude verifies:**
- [ ] `go build ./...` succeeds on v2.
- [ ] `just test` — all 32 existing `tui/` tests green (same assertions as v1), whole
  suite green.
- [ ] `just lint` clean.
- [ ] `go mod tidy` leaves no diff (deps fully resolved to the v2 tree).
- [ ] `git grep -n 'charmbracelet/\(bubbletea\|bubbles\|lipgloss\|glamour\)"' -- tui/`
  returns nothing — no un-migrated v1 import path remains (every Charm import is `/v2`).

**User verifies (whole slice — the migration's acceptance bar):**
- [ ] `just install`, then run `bit tui` and compare side-by-side against the v1 build:
  the list view (list + bordered detail pane) and the kanban board look and behave exactly
  as before — same keys (`tab`, `←/→`, `↑/↓`, `?`, `q`/`esc`/`ctrl+c`), same layout, same
  focus accent. Pay special attention to the glamour-rendered detail body: it renders in
  the same light/dark style as today (the background-detection risk spot).

**Commit (user):** `refactor(tui): migrate the Charm stack to v2`