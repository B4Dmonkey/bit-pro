---
id: BIT-4
title: Browse the project's tasks in a terminal
status: done
---
# Browse the project's tasks in a terminal

## Why

The CLI can now hold the whole picture — 29 real records across three tracks, bars
parented and verse-labelled — but a human still has no way to *look* at it. `bit task
list` prints one flat line per task; to actually read a scope or walk its steps you open
markdown files by hand, which is the SaaS-web-UI problem the README set out to avoid, just
in the other direction. The README has planned a TUI from the start ("the human's window
into the same data"), and until it exists the review loop the whole tool is built around —
read the full scope, approve or disapprove, then walk the bars one at a time — happens
outside `bit` entirely.

This is also the first real second consumer of the `task` package. Extracting it was
justified on the promise of exactly this view; this scope collects on that.

## Summary

A new `bit tui` command opens a read-only terminal view over `.bit/`, built on the Bubble
Tea stack (Bubble Tea + Bubbles + Lip Gloss, and Glamour for rendering markdown bodies).
It reads through `task.Store` — never a second source of truth, never a writer. You move
through the list of tracks and bars on the left while a detail pane on the right renders
the selected task's body in place — always visible, updating as the cursor moves.

**Read-only, deliberately.** Approving or disapproving a scope stays where it already is —
in chat, by invoking the skills — and an explicit status change from the UI is a separate,
later scope (the README's "status state machine"). Keeping the TUI a pure reader means it
touches none of that: no status semantics, no file mutation. It's a window, not a control
panel.

The vocabulary — track, bar, verse — is the presentational map in
[hierarchy.md](./hierarchy.md); the code keeps saying `task` and `step`.

## Phases

- [x] Phase 1 — a human can open the project and move through it: `bit tui` launches a
  navigable list of every task, newest track first with its bars beneath (the order
  `List` already returns), and quits cleanly. Nothing is rendered richly yet — this is the
  thinnest end-to-end thing that proves the Bubble Tea stack reads real records from the
  store and paints a screen you can drive with the keyboard.
  Touches: new `cmd/tui.go` (the command), new `tui/` package (the model), `go.mod`

- [x] Phase 2 — a human can read a task without leaving the terminal: a detail pane beside
  the list renders the selected task's body — scope prose or step detail — as markdown,
  updating live as the cursor moves. This is the payload that makes the review loop possible
  in `bit`: you can read a full track top to bottom, or drop into a single bar. It's also
  where the real, backtick-heavy imported content gets to say whether the rendering holds up
  in a narrow pane.
  Touches: `tui/` (split layout + detail pane + Glamour), `cmd/tui.go`

- [x] Phase 3 — a human can see the shape at a glance: tracks read as distinct from their
  bars, and each bar shows the verse it serves, so you can navigate the tree and keep the
  thread of which slice a step is building without opening the scope. This is the indicator
  called for so that reviewing one bar at a time never loses the larger context.
  Touches: `tui/` (list rendering); possibly `task/task.go` if deriving a bar's track/verse
  earns a small accessor rather than living in the view

- [x] Phase 4 — a human can focus a pane and drive it independently, and read the frame at a
  glance: each pane's border carries a title (the list its live task count, the detail
  `Details`), the focused pane is visibly accented and takes the arrow / `j`-`k` keys —
  moving the selection when the list is focused, scrolling the body when the detail is
  focused — and `→`/`←` move focus between them. The keybinding hints live in a help bar
  beneath both panes instead of inside the list. This is what makes a long detail body
  actually readable (you can scroll it) and the two-pane layout legible as one workspace
  rather than two stacked widgets.
  Touches: `tui/model.go` (focus state, focus-routed keymap, titled borders, help bar)

> **Delivery order:** Phase 4 is being built before Phase 3. It came out of manual testing
> of the Phase 2 two-pane view — the interaction and framing needed fixing before the
> verse-column work is worth doing. The plan's steps run in build order (Phase 4 = Steps
> 7–10, Phase 3 = Steps 11–12).

## Visual aid

```
 ┌ bit ──────────────────────────────────┬ detail ──────────────────────┐
 │ BIT-3   Plans live in the project todo │ # Plans live in the project  │   ← selection
 │  BIT-3.1  reverse-sort the list v1 done │                              │
 │  BIT-3.9  the import proves...  v4 done │ Scope prose / step detail    │
 │ BIT-2   Task Management (CRUD)    todo  │ renders here as markdown,    │
 │  BIT-2.1  init wizard + create v1 done  │ updating as the cursor moves │
 │  …                                      │ …                            │
 └─────────────────────────────────────────┴──────────────────────────────┘
   Phase 1: the list, navigable.   Phase 2: the detail pane renders the selection.
   Phase 3: track/bar distinction + the verse column (left pane).
```

The detail pane sits beside the list, not on its own screen — a persistent, read-only
side-by-side preview that follows the cursor. There is no full-screen mode.

## Risks & unknowns

- **Unknown:** How much of this is testable without a terminal. A Bubble Tea program's
  `Update` is a pure `(model, msg) → model` function and any list-shaping/verse-deriving is
  pure too — those get real tests. The rendered frames do not, by choice.
  **Resolve by:** Keep logic out of `View`; the user manually tests the visuals and
  critiques from there. `teatest` golden frames stay available if manual testing proves too
  coarse.
  **De-risk before planning?** No — this is the agreed testing posture, not an open question.

- **Unknown:** Whether Glamour renders the imported bodies well — they're dense, nested code
  fences, long lines — inside a narrow pane.
  **Resolve by:** Phase 2's manual test against the real plan/scope bodies is the experiment;
  the awkward imported content is exactly the fixture worth failing against.
  **De-risk before planning?** No — the import is what makes this cheap to just try.

- **Unknown:** Where a bar's track and verse get derived for grouping — the view parsing the
  dotted ID itself, or a small accessor on `Task`/`Store`. The sort logic that already knows
  this (`compareIDs`, `idParts`) is unexported to `task`.
  **Resolve by:** Decide when planning Phase 3, the first phase that needs it — Phases 1–2
  ride on the flat order `List` already returns. Cheap either way.
  **De-risk before planning?** No — small, and isolated to one phase.

- **Known, not unknown:** the Charm stack adds several modules to `go.mod`, and `bit tui`
  needs a TTY so it never runs in the non-interactive test path. Both accepted: the TUI is
  the one deliberately interactive entrypoint, and its logic is tested without launching it.

## Out of scope

- **Any write from the UI** — status changes, approve/disapprove, editing, creating,
  deleting. Approval stays in chat via the skills; an explicit status change is the
  README's separate "status state machine" scope. The TUI stays a pure reader.
- **The kanban board.** The README's second view is its own scoping pass; this scope is the
  list only, though the board will reuse the same store reads and rendering.
- **Filtering and search.** A README open question; at 29 records the track/bar grouping
  substitutes for it, so nothing here demands it. Same for an index for fast lookup.
- **Making the TUI a source of truth.** The CLI stays primary; the TUI only ever reads what
  the CLI wrote.
- **Renaming anything in the code to the music vocabulary.**

## Context
See scope: [tui-list-scope.md](./tui-list-scope.md) — the WHY, the phase order, and the
read-only boundary live there.

## How this plan works

The normal outside-in entry point is the `bit tui` command — but that launches a Bubble
Tea program, which needs a TTY and would hang `go test`. So the entry point itself is
**not** the tested layer here; it's a thin shell.

The tested spine is one layer down: the **model**. A Bubble Tea model is constructed as a
plain struct and its `Update(msg) (model, cmd)` is a pure function — both run without a
terminal. So every step that owns real logic (mapping `Store.List()` into list items,
forwarding messages into the list, the pane width split, the track/bar/verse
row derivation) is TDD'd against the model directly. Quit isn't in that list — it's
inherited from the list's default keymap, not code we write. The steps that only wire the shell or paint pixels (`tea.NewProgram`,
Glamour rendering, the row layout) can't be unit-tested and are **User verifies** —
exactly the split the scope accepts.

The model embeds Bubbles' `list.Model`, so list navigation and scrolling are the library's
and are not re-tested. `List()` already parses every task's `Body`, so the whole project
is in memory after one call — the detail view (Phase 2) reads `task.Body` directly, with
no second I/O path.

New package `tui/`, split so the untestable parts are quarantined:
- `tui/model.go` — the model, `New`, `Update`, item type. Tested.
- `tui/model_test.go` — the tests.
- `tui/run.go` — the `tea.NewProgram(...).Run()` wrapper. Not tested (needs a TTY).
- `tui/delegate.go` — the custom row renderer (Phase 3). `Render` is visual; its pure
  helpers live in `model.go` and are tested.