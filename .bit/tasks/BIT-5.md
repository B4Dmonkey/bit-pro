---
id: BIT-5
title: See the project as a kanban board
status: done
---
# See the project as a kanban board

## Why

`bit tui` can now show the whole backlog as a navigable list with a detail pane, but a
list only answers "what is there" — it can't answer "where is the work." A human glancing
at the project wants to see, at once, what's waiting, what's moving, and what's finished.
That shape — cards grouped into status columns — is the second view the README planned for
the TUI from the start ("a task list and a kanban board over the same data"), and it's the
view that makes the status field legible instead of just present. Building it now, before
the status state machine, lets us *see how the columns actually sit* against real records
and let that inform what the state machine should later formalize.

## Summary

`bit tui` still opens on the list. Pressing `tab` switches the whole screen to a **board**:
every task laid out as a card in one of three equal-width columns — **To Do**, **Doing**,
**Done** — grouped by its `status` field. Pressing `tab` again switches back to the list.
The board reads from the same in-memory tasks the list already holds — no reload, no second
store read. Each column is built from the same list component the list view uses, framed in
a titled, bordered box. Navigation mirrors the list's feel: `←`/`→` moves the active column,
`↑`/`↓` moves the selection within it, the active column is shown by its accented border,
and a help bar sits beneath.

**Read-only, deliberately — same as the list.** The board groups tasks by the status they
already have; it never *changes* a status. Moving a card between columns is a write, and
writes are the README's separate "status state machine" scope. This board is a second way
to look, not a control panel.

The vocabulary stays: "column" and "card" are plain kanban terms; the code keeps saying
`task` and `status`. Nothing here renames anything.

## Phases

- [x] Phase 1 — a human can flip to a board grouped by status: `bit tui` opens the list as
  today, `tab` switches to a board that lays every task into three side-by-side columns —
  To Do / Doing / Done — by its status field, and `tab` switches back. This is the walking
  skeleton: it proves the view-mode toggle and the status→column grouping over the same
  in-memory tasks. Rendering can be plain (each card an ID + title, columns split in
  equal thirds) — the point is that the toggle works and the cards land in the right column.
  (Delivered with a plain render; Phase 2 reframes the columns as real list components.)
  Touches: `tui/` (a view-mode on the model; board grouping + layout — likely a new
  `tui/board.go`), `tui/model.go` (route `tab`)

- [x] Phase 2 — the board reads as a board: the three columns become titled, bordered,
  equal-width boxes — each header showing its card count (`To Do (4)`) — built from the
  same list component and framing the list view already uses, with the active column shown
  by its accented border (column 0 by default, mirroring how the list view opens with the
  list pane focus-accented). This is the difference between three columns of plain text and
  a board a human can take in at a glance — and it establishes the border as the language
  for "active", which the next phase moves.
  Touches: `tui/board.go` (columns as list components + layout), reusing `titledBorder` and
  the list `delegate`.

- [x] Phase 3 — a human can drive the board with the keyboard: in board mode `←`/`→` move
  the active column (its accented border follows), `↑`/`↓` move the selected card within
  the active column (each column keeping its own selection via its list's cursor and `▎`
  accent), and the help bar advertises the board's controls (`←/→ column · ↑/↓ card ·
  tab list · ? help · q quit`) rather than the list's. This turns a legible board into one
  you can actually navigate. Board mode owns its keys here — so quit stays wired through the
  new key gate.
  Touches: `tui/model.go` (mode-routed keymap + help; active-column state), `tui/board.go`
  (per-column selection; board keymap).

## Visual aid

```
 list mode  ── tab ──▶  board mode  ── tab ──▶  list mode
```

```
 ┌ To Do (4) ──────────┬ Doing (1) ──────────┬ Done (37) ──────────┐
 │▎BIT-4  Browse the…  │ BIT-2.x  some doing… │ BIT-4.1  Model maps…│
 │ BIT-3  Plans live…  │                      │ BIT-4.2  Update fwd…│
 │ BIT-2  Task Mgmt…   │                      │ BIT-4.3  Wire bit…  │
 │                     │                      │ …                   │
 └─────────────────────┴──────────────────────┴─────────────────────┘
   ←/→ column · ↑/↓ card · tab list · ? help · q quit
```

Three equal-width columns fill the screen — no detail pane in board mode (that's the list
view's layout). `▎` marks the selected card in the active column, reusing the list's accent.

## Risks & unknowns

- **Unknown:** The status→column mapping. There are three columns but `status` is a freeform
  field the state machine hasn't formalized yet. Today's data is exactly `todo` / `doing` /
  `done`, so it maps 1:1 — but where would a future `backlog` go, and what happens to a task
  whose status matches no column?
  **Resolve by:** For this scope, hardcode the three columns keyed to `todo`/`doing`/`done`
  (covers all current data); defer `backlog` and any catch-all to the status state machine
  scope. Manual testing will show whether an unmapped task quietly vanishes.
  **De-risk before planning?** No — current data maps cleanly and the mapping is a one-liner
  to revise later. But the *assumption* (fixed three columns, those three statuses) is
  load-bearing and named here so a reader can challenge it.

- **Unknown:** Which tasks become cards — tracks and bars alike? The list groups
  hierarchically (a track, then its bars). A flat status board splits a track from its bars:
  `BIT-4` (a track, `todo`) lands in To Do while its `done` bars land in Done. Is that the
  intent, or should the board show only leaf bars, or only tracks?
  **Resolve by:** Start with every task as a card (simplest — the in-memory slice as-is) and
  judge from manual testing whether the track/bar split reads wrong. Cheap to change the
  filter later.
  **De-risk before planning?** No — cheapest to try flat and critique visually, which is the
  stated workflow.

- **Unknown:** Does `bubbles/list` sit well as a ~third-width column, and do three list
  instances share the existing `delegate` while splitting keyboard focus cleanly (only the
  active column responds to `↑`/`↓`)? The list view drives a single list; three side-by-side
  is new, and it's the load-bearing assumption behind "columns are the same list component".
  **Resolve by:** Phase 2 renders the three framed list columns — manual testing shows
  immediately whether a narrow list reads right and whether focus routing behaves.
  **De-risk before planning?** No — it's the natural first slice of Phase 2 and cheap to
  see; but naming it here lets the reuse-the-list-three-up decision be challenged before the
  plan commits to it.

- **Known, not unknown — testing posture (same as the list scope):** the toggle state, the
  status→column grouping, the active-column index, and each column's selection (its list's
  cursor) are all pure `(model, msg) → model` logic and get real tests. The rendered frames
  do not, by choice — the user manually tests the visuals and critiques from there.

## Out of scope

- **Moving a card between columns** (or any status change from the board). That's a write —
  the README's separate "status state machine" scope. The board stays a pure reader.
- **A detail pane / open-a-card in board mode.** The board is the three columns; reading a
  task's full body stays the list view's job (`tab` back to it). No Enter-to-open here.
- **Filtering and search** (the reference's Filters bar). A README open question, excluded
  from the list scope for the same reason; nothing here demands it at this size.
- **Configurable columns.** Fixed three columns for now; whether columns become
  per-project configurable is a README open question, not this scope.
- **Renaming anything in the code to the music vocabulary.**

## Context
See scope: [tui-board-scope.md](./tui-board-scope.md) — the WHY, the phase order, the
read-only boundary, and the two load-bearing assumptions (fixed `todo`/`doing`/`done`
columns; every task a card) live there.

## How this plan works

The board is a **second view mode over the same in-memory model** the list already built.
The entry point is unchanged — `bit tui` launches the same Bubble Tea program — so the
tested spine is still `model.Update`, a pure `(msg) → (model, cmd)` that runs without a TTY.

Phase 1's toggle and grouping already landed (Steps 1–3). The rest follows the scope's
**reshaped order**: **Phase 2 makes the board *read* like a board** (Steps 4–5), then
**Phase 3 makes it *drive* like one** (Steps 6–8).

The load-bearing decision from the reshape: **each column is the same `bubbles/list`
component the list view uses**, not plain text. That changes what's tested and what's
delegated:

- The three columns become `[3]list.Model` (`boardCols`), each built from the shared
  `delegate{}` over that column's grouped tasks. This **replaces** the plain
  `columns [3][]*task.Task` field — one source of truth, and `.Items()` gives the header
  counts.
- **Card selection is the list's own cursor, not hand-rolled arithmetic.** `↑`/`↓` forward
  to `boardCols[activeCol].Update`, so clamping, scrolling, and per-column independence come
  from `list.Model` for free — the payoff for reusing the component (no `colCursor [3]int`).
- **Active = an accented border**, via the list view's existing `titledBorder(..., active)`.
  Phase 2 frames all three and accents column 0; Phase 3 makes `activeCol` move.

What gets real tests: the active-column index, each column's list membership and cursor
(`.Index()`), and the header/count strings — all pure `(model, msg) → model` state. The
rendered *frames* (equal-width boxes, border accent, card marker) stay manual-verify — the
same split the list plan and scope accept.

Two facts still shape the steps:

- **`New` maps tasks into list items but keeps no raw slice.** `groupByStatus` (Step 2,
  done) stays the pure grouping helper; `New` feeds each column's slice into a `list.New`.
- **Quit is inherited, not owned.** `q`/`esc`/`ctrl+c` quit today only because unhandled
  keys fall through to `list.Update`. Board mode's key gate (Step 6) returns early, so it
  must carry the quit keys through or quit silently dies in board mode. A tested concern.

`tui/board.go` holds the board's logic (`groupByStatus`, the column builders, `updateBoard`,
`boardView`, the board keymap); `tui/board_test.go` its tests. `tui/model.go` owns the
`mode` field, the `tab` toggle, the board key gate, and the board state (`boardCols`,
`activeCol`). No new dependency — `list`, `help`, `key`, `lipgloss` are all already required.

Idiomatic Go throughout: the view mode is an `iota` enum, grouping is a pure function tested
directly, all tests are table-driven, and there are no code comments (per project convention).