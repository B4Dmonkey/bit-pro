---
id: BIT-20.3
title: firstBarIndex finds the first bar, falling back to the first row
status: todo
phase: 1
phase_label: Kanban-first default
---
## **Verse 1**

The other half of "land on the right task": within the default column, select the first *bar*
(a dotted child ID) rather than its parent track, since the track is a container an operator
isn't directly acting on. Same shape as `defaultColumn` — a small pure helper, table-tested on
its own before it's wired in.

## Scope
- `tui/board.go` — add `firstBarIndex(items []list.Item) int`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestFirstBarIndex/bar_after_its_track`
     - **Behavior:** the common shape — a track followed by its bar(s) in the same column —
       lands on the bar, not the track.
     - **Setup:** `items := []list.Item{item{t: &task.Task{ID: "BIT-1"}}, item{t: &task.Task{ID: "BIT-1.1"}}}`.
     - **Assertions:** `firstBarIndex(items) == 1`.
     - **Boundary:** the bar is the second row — proves it isn't just "always pick index 0".
   - [ ] Confirm fails: `firstBarIndex` doesn't exist yet (compile error).

2. **Implement (GREEN):**
   - [ ] `func firstBarIndex(items []list.Item) int { return 1 }` — hardcoded, passes the one
     test above.

3. **More tests (RED → GREEN):**
   - [ ] `TestFirstBarIndex/bar_before_its_track`
     - **Behavior:** the bar can appear at any position — the search isn't keyed to a fixed
       index, it's keyed to the ID shape.
     - **Setup:** `items := []list.Item{item{t: &task.Task{ID: "BIT-2.1"}}, item{t: &task.Task{ID: "BIT-2"}}}`.
     - **Assertions:** `firstBarIndex(items) == 0`.
     - **Boundary:** contradicts the hardcoded `return 1` — forces a real per-item `isBar` scan.
   - [ ] `TestFirstBarIndex/no_bars_falls_back_to_first_row`
     - **Behavior:** a column holding only tracks (no bars planned yet) still gives a sane
       default — the first row.
     - **Setup:** `items := []list.Item{item{t: &task.Task{ID: "BIT-3"}}, item{t: &task.Task{ID: "BIT-4"}}}`.
     - **Assertions:** `firstBarIndex(items) == 0`.
     - **Boundary:** zero bars present — the lower bound of "how many bars are in this column".
   - [ ] `TestFirstBarIndex/empty_column`
     - **Behavior:** an empty column doesn't panic or return an out-of-range index.
     - **Setup:** `items := []list.Item{}`.
     - **Assertions:** `firstBarIndex(items) == 0`.
     - **Boundary:** zero items — the lower bound of the input space itself.
   - [ ] Implement the real scan: iterate `items`, cast to `item`, return the index of the first
     one where `isBar(it.t.ID)`; fall through to `return 0`.

## Claude verifies
- [ ] `go test ./tui/... -run TestFirstBarIndex` passes.
- [ ] `golangci-lint run` passes.

## User verifies
- [ ] None — pure helper, not yet wired into `New()` (that's this verse's last bar).

## Commit (user)
`feat(tui): add firstBarIndex to skip past a column's track rows`