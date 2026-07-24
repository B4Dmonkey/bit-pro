---
id: BIT-14.6
title: Selected board cards render inverted
status: doing
phase: 2
phase_label: Focus
---
## **Verse 2**

Makes a selected board card render as a reverse-video green block, matching the focused column title, while the main list's selected row stays green foreground. The two selection styles diverge by context, which the shared delegate can't express today — so this gives the delegate a board variant.

## Scope
- `tui/delegate.go` — add a `board bool` field to `delegate`; in `Render`, when a row is selected use reverse-video green if `board` is set, otherwise the plain green foreground (`selectedStyle`) as before.
- `tui/board.go` — `newColumnList`: construct the delegate as `delegate{board: true}` so board columns get the inverted selection.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestView_SelectedBoardCardInverted`
     - **Behavior:** the selected card in a board column renders as a reverse-video green block; the main list's selected row does not.
     - **Setup:** board case — `New([]*task.Task{{ID: "BIT-1", Status: "doing", Title: "Card"}})`; WindowSize 120x40; Tab to board; Right to the Doing column (card auto-selected); `board := View().Content`. List case — a fresh `New([]*task.Task{{ID: "BIT-1", Title: "Row"}})` sized, default list mode; `listView := View().Content`.
     - **Assertions:** `strings.Contains(board, "\x1b[7;32m")` is true (card is a reverse-green block); the list-view selected row contains green (`32m`) but `strings.Contains(listView, "\x1b[7;32m")` around the row is false — the list row is not inverted. (Note the modal/title also uses `7;32m`; assert on the card row specifically, e.g. that the `BIT-1`/`Card` segment carries the reverse SGR.)
     - **Boundary:** `board == true` vs `board == false` on the selected path — both states of the new field.
   - [ ] Confirm fails: after Verse 1 the board's selected card is plain green (no reverse), so the card-inverted assertion fails.

2. **Implement (GREEN):**
   - [ ] Add `board bool` to `delegate`; branch the selected-row style to reverse-video green when `board`; set `delegate{board: true}` in `newColumnList`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] Whole slice: run `bit tui`, press Tab. The focused column's title is a green reverse-video block and unfocused columns read `| Title |`; moving the card cursor shows the selected card as a green block; Enter shows the modal title as a green block; back in list view (Left/Right), the focused pane title is a block while the other reads `| Details |`. Focus is unmistakable everywhere.

## Commit (user)
`feat(tui): selected board cards render inverted`
