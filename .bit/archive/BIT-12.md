---
id: BIT-12
title: Migrate the TUI to the Charm v2 stack
status: done
---
## Why

The TUI runs on the Charm v1 stack (bubbletea v1.3.10, bubbles v1.0.0, lipgloss v1
pre-release, glamour v1.0.0). The board modal needs lipgloss's cell-based compositor to
float a detail box over the board — and that compositor ships only in lipgloss v2. v1 is
now superseded across the whole Charm line, so staying on it blocks the compositor and any
future TUI polish that leans on v2-only APIs. Moving the stack to v2 unblocks the modal and
keeps the TUI on the maintained line.

## Summary

Move every bubbletea, bubbles, lipgloss, and glamour call in `tui/` to the v2 APIs
(bubbletea v2.0.8, bubbles v2.1.1, lipgloss v2.0.5, glamour v2.0.1). The TUI's behavior and
appearance stay identical — this is enabling work, not a feature. Success is "the list view
and the board still look and behave exactly as before, tests green," on the new stack. The
port adopts v2 idioms rather than translating v1 call-for-call, so the code reads as native
v2 — but what a user sees is unchanged. bubbletea's Program/Model is the root of the TUI, so
the migration lands as one coherent cut rather than a half-v1/half-v2 intermediate that
can't compile.

## Decisions

- **Target the released v2 stack, pinned together.** bubbletea v2.0.8, bubbles v2.1.1,
  lipgloss v2.0.5, glamour v2.0.1 — the coherent v2 set, adopted as a unit so the TUI never
  straddles two major versions of interdependent libraries.
- **Observable behavior and appearance are unchanged.** No new keys, no visual rework, no
  new panes — the acceptance bar is "identical to today, side-by-side, on v2." Feature work
  (the board modal, any visual polish) still rides its own track.
- **Port to idiomatic v2, not a mechanical 1:1 translation.** Where a v2 API offers a
  cleaner idiom than the direct v1 equivalent — a typed API for a stringly-typed one, a
  built-in for a hand-rolled loop — take it. The goal is code that reads as native v2, not a
  line-for-line transliteration of the v1 version; the guardrail is that observable behavior
  stays identical and tests stay green.
- **One atomic cut, not a v1/v2 straddle.** bubbletea's Model interface is the TUI root, so
  the whole `tui/` package moves to v2 in one landing rather than shipping a mixed stack
  that can't build or that runs two lipgloss majors side by side.

## Verses

- [x] Verse 1 — The TUI runs entirely on the Charm v2 stack, unchanged: launch `bit tui`,
  and the list view (list + detail pane) and the kanban board look and behave exactly as
  they do today — same keys, same layout, same rendering — but every underlying call is the
  v2 API. Delivers the migrated substrate the board modal builds on; the human check is
  side-by-side identical behavior, especially glamour's rendered detail body.
  Touches: all of `tui/` (`model.go`, `board.go`, `delegate.go`, `run.go`) and `go.mod` —
  where to look to verify.