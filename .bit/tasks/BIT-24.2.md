---
id: BIT-24.2
title: 'Board cards contradict: the marker is list-only'
status: todo
phase: 1
phase_label: In-progress marker
---
## **Verse 1**

A board card at `doing` must render no `→` — the board already answers "what's in progress"
with its `Doing` column. Forced by contradiction: the previous bar's branch fires on status
alone, so the same task rendered through `delegate{board: true}` shows the marker too.

## Scope
- `tui/delegate.go` — the `doing` branch added in the previous bar, gated on `d.board`
- `tui/delegate_test.go` — new test alongside `TestDelegate_SelectedBoardCardInverted`

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestDelegate_BoardCardHasNoInProgressMarker`
     - **Behavior:** the in-progress marker is a list-view affordance only; board rendering
       is unchanged by this verse.
     - **Setup:** `l := list.New([]list.Item{item{t: &task.Task{ID: "BIT-1", Title: "Card", Status: "doing"}}}, delegate{board: true}, 40, 4)`;
       render index 0 with `delegate{board: true}.Render(&buf, l, 0, l.Items()[0])`.
     - **Assertions:** `buf.String()` does not contain `"→"`; does contain `"BIT-1"` (the
       card still renders — the gate suppresses the mark, not the row).
     - **Boundary:** `delegate.board == true`; the `false` side of the same flag is covered
       by the previous bar's test, which builds a bare `delegate{}`.
   - [ ] Confirm fails: the unconditional `doing` branch renders `→` on the board card.

2. **Implement (GREEN):**
   - [ ] Gate the branch on the existing flag: `else if t.Status == "doing" && !d.board`.
     The `done` branch stays ungated — `✓` on board cards is existing behaviour this verse
     does not touch.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] Whole slice: `just install`, then `bp tui`. Run `bp task list` first to note a task at
      `doing` (the bar being worked qualifies). On the board, that card shows no `→`; press
      `tab` to the list view and its row shows `→` while `todo` rows stay blank.

## Commit (user)
`feat(tui): keep the in-progress marker out of the board view`