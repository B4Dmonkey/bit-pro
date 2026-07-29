---
id: BIT-14.8
title: Show one clean empty state
status: done
phase: 3
phase_label: Declutter
---
## **Verse 3**

Hides the list's status bar so the empty state shows a single `No items.` line instead of the status-bar `No items` duplicated above the body message. Completes the declutter verse.

## Scope
- `tui/model.go` — `New()`: call `l.SetShowStatusBar(false)`. (The status bar is what renders the `No items` / `N items` line; the body's `No items.` message stays.)

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestView_EmptyListSingleEmptyState`
     - **Behavior:** an empty list shows exactly one empty-state line, not the status-bar duplicate above it.
     - **Setup:** `New(nil)`; WindowSize 60x16; `view := View().Content`.
     - **Assertions:** `strings.Count(view, "No items") == 1`. (Currently 2: the status bar renders `No items` and the body renders `No items.`, and `No items` is a substring of `No items.`.)
     - **Boundary:** item count == 0 — the only state showing the empty message; a populated list shows `N items` in the status bar instead.
   - [ ] Confirm fails: the count is currently 2.

2. **Implement (GREEN):**
   - [ ] Add `l.SetShowStatusBar(false)` in `New()`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] Whole slice: run `bit tui` against an empty store and a populated one. No `List` heading appears, and the empty store shows a single `No items.` line — the list view reads clean.

## Commit (user)
`feat(tui): show one clean empty state`
