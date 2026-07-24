---
id: BIT-13
title: Live-reload the TUI
status: doing
---
## Why

When an agent — or a second terminal — changes a task while `bit tui` is open, the change
doesn't show up: the TUI reads every task once at startup and never looks again. The only
way to see new state today is to quit and relaunch, which throws away where you were and
breaks your train of thought at exactly the moment you're trying to follow work as it
happens. The whole point of the human view is to watch the source of truth move; a frozen
snapshot defeats it.

## Summary

Have the running TUI re-read `.bit/tasks/` on a short recurring timer (bubbletea's
`tea.Tick`) and refresh itself when the task set has changed, so an edit made through the
CLI shows up in the list and board with no restart. The refresh reuses the existing load
path (`Store.List()`) and preserves where the human was — selection, column, view mode, open
modal — so an incoming change updates the view without yanking them out of context. Because
each tick only refreshes on a real change, a burst of writes between ticks collapses into a
single refresh.

## Visual aid

```
before:  bit tui ──(List once)──> model ──> render     [frozen snapshot]

after:   tea.Tick ──interval──> Store.List() ──changed?──> reload
                                                              │
        model (keep selection / col / mode / modal) ──> render   [live]
              └───────────────── reschedule tick ─────────────┘
```

## Decisions

- **Poll with `tea.Tick`, no filesystem-watch dependency.** The model schedules a short
  recurring `tea.Tick`; each tick re-runs `Store.List()` off the render loop and refreshes
  only when the listing changed, then reschedules. The whole mechanism stays inside the
  native bubbletea Cmd loop — no new dependency, no raw goroutine to start or stop. Chosen
  over fsnotify, whose only edge (instant latency, no idle wakeups) doesn't matter for a
  human view where sub-second latency is fine.
- **Watch `.bit/tasks/` only.** Create / update / delete and archive all surface there
  (archiving moves a file out of `tasks/`, which reads as a delete); `config.toml` and
  `.bit/archive/` don't affect the live view, so they're out of scope.
- **Reload means re-running `Store.List()`.** The refresh reuses the existing glob-and-parse
  load path rather than a new incremental one — the simplest correct thing. Incremental
  diffing is a later optimization only if it's ever needed (YAGNI).
- **A refresh preserves the human's place.** Selection, active column, view mode, and an open
  modal survive a reload; a live change never resets the view to the top. This is the
  acceptance bar for the "maintain understanding" goal.
- **A failed reload keeps the last good view.** If a `List()` lands mid-write and errors, the
  TUI holds its current state rather than crashing or blanking.

## Verses

- [ ] Verse 1 — See CLI edits without restarting: with the TUI open, a task created, changed,
  or removed through the CLI appears in the list and board on its own — no quit-and-relaunch.
  The thinnest end-to-end path: a tick reloads and rebuilds every interval; selection may
  reset until Verse 2.
  Touches: `cmd/tui.go` (give the model a reload source — the store / bit dir — not just a
  one-shot task slice), `tui/model.go` (`Init` starts the tick; `Update` handles the tick →
  reload Cmd + reschedule, and the loaded tasks → rebuild list + board).
- [ ] Verse 2 — The refresh keeps your place: after a live reload the selected card, active
  column, view mode, and open modal stay put, so an agent's edit updates the view without
  pulling you out of context.
  Touches: `tui/model.go` (reconcile the new task set against the current selection by ID;
  preserve column / mode / modal / scroll).
- [ ] Verse 3 — The poll stays quiet and safe: a tick refreshes only when the listing
  actually changed, so nothing rebuilds (and nothing flickers) while the task set is
  unchanged, and a burst of writes between ticks collapses into one refresh; a tick that
  reads a file mid-write holds the last good view instead of blanking or crashing.
  Touches: the tick / reload handler in `tui/model.go`.