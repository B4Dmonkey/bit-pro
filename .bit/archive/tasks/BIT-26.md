---
id: BIT-26
title: List focus preserves approval color on item text
status: todo
---
## Why

BIT-23.7 established that unapproved items render in yellow so a reader can see at a glance what still needs review. That signal disappears the moment you focus an item: the selection style overrides the entire row with green, so the approval color is invisible exactly when you're looking directly at the item. The cursor indicator should keep green to show where focus is, but the text should stay on the approval palette so the signal holds even when the item is selected.

## Summary

Split the list-focus style into two concerns: the `▎` cursor glyph stays green to mark where focus is, and the row's main text uses the item's approval color — yellow for unapproved, default for approved — plus bold to maintain focus emphasis.

## Visual aid

```
before (current)              after
──────────────────────        ──────────────────────
▎ BIT-23.10  bit:do …        ▎ BIT-23.10  bit:do …   ← cursor green, text yellow (unapproved)
  BIT-21     Task IDs…         BIT-21     Task IDs…   ← unfocused, unchanged
```

## Decisions

- **Cursor keeps green.** The `▎` glyph is the sole green element on a focused row; it marks position without drowning the approval signal.
- **Bold survives on focused rows.** Bold already distinguishes focus; it stays. Color is what changes.
- **Approval color logic is unchanged.** Unapproved → yellow; approved (or non-approval states) → default. The delegate's existing branch is reused, not duplicated.
- **Board mode is unaffected.** The board delegate (`selectedBoardStyle`, reverse-video) serves a different display context and is not changed here.

## Verses

- [ ] Verse 1 — Focused list item text reflects approval state: the `▎` cursor renders in green, the row text renders in the item's approval color (yellow for unapproved, default for approved) plus bold, identical to an unfocused row's color but with bold added.
  Touches: `tui/delegate.go` — the `Render` function's `selected` branch.
