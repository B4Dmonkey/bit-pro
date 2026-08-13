---
id: BIT-24
title: 'List view: in-progress marker and done/total counter'
status: doing
---
## Why

The list view tells you almost nothing about momentum. A row is either checked or blank,
so a bar someone is actively working on looks identical to one nobody has started — the
single most useful fact in a working session is the one the view hides. The pane header
compounds it: `Tasks (10)` is a raw inventory count that never moves as work lands, so
there is no glanceable sense of how far along anything is. The board view already answers
both questions by construction (a `Doing` column, three visibly-sized columns), which is
why this gap only bites in the list view — and the list view is where you sit when you
want the tree structure, not the swimlanes.

## Summary

Two changes to the list view only. First, an in-progress marker: a row whose status is
`doing` renders `→` in the same mark column that already renders `✓` for `done`. Second,
the list pane's header becomes a progress fraction — `Tasks (5/10)` instead of `Tasks (10)`
— counting done rows over all rows. The board view keeps its current rendering unchanged.

## Visual aid

```
        before                            after
┌ Tasks (7) ──────────────┐   ┌ Tasks (3/7) ────────────┐
│ ✓ BIT-7   Add board view │   │ ✓ BIT-7   Add board view │
│ ✓   BIT-7.1  step one    │   │ ✓   BIT-7.1  step one    │
│ ✓   BIT-7.2  step two    │   │ ✓   BIT-7.2  step two    │
│   BIT-8   List polish    │   │ → BIT-8   List polish    │
│     BIT-8.1  step one    │   │ ✓   BIT-8.1  step one    │
│     BIT-8.2  step two    │   │ →   BIT-8.2  step two    │
│     BIT-8.3  step three  │   │     BIT-8.3  step three  │
└──────────────────────────┘   └──────────────────────────┘
   momentum invisible             what's moving, and how far along
```

## Decisions

- **In-progress marker is `→`, in the existing mark column.** Reads as "moving" and sits
  quietly beside `✓` rather than competing with it; `▶` was rejected for reading like an
  expand/collapse caret next to the detail pane's Enter-to-expand behaviour.
- **The marker is driven by status alone, so tracks get it too.** A track sitting at
  `doing` shows `→` exactly as a bar does. No special-casing by row kind — the mark column
  already treats tracks and bars alike for `✓`.
- **The counter counts every list row — tracks and bars alike.** `Tasks (<rows with status
  done>/<all rows>)`. Chosen over bars-only or tracks-only: it is the literal reading of the
  request and needs no new notion of which rows "really" count. It does mean a fully-done
  track contributes to both sides alongside its own bars; that is accepted, because the
  number is a glanceable progress signal, not an accounting figure.
- **Board view is untouched.** Both changes are gated to the list view. The board already
  answers "what's in progress" with its `Doing` column and "how many" with its column
  contents, so applying either change there would be redundant at best and misleading at
  worst.
- **No new status values, no data-model change.** Everything renders from the `status`
  field the CLI already writes (`todo` / `doing` / `done`). This is presentation only.

## Verses

- [ ] Verse 1 — See at a glance which work is in progress: a list row whose status is
  `doing` renders `→` where a done row renders `✓`, so an active bar is distinguishable
  from an untouched one without opening anything.
  Touches: the list row renderer (`tui/delegate.go`) — the mark column, gated off the
  existing board flag so the board view is unaffected.

- [ ] Verse 2 — Read overall progress off the pane header: the list pane's title shows
  done-over-total instead of a bare total, so how far along the work is is visible without
  counting rows.
  Touches: the list pane title (`tui/model.go`, the list-mode branch of `content`).