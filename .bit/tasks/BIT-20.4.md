---
id: BIT-20.4
title: New() lands on the Doing column's topmost bar
status: todo
phase: 1
phase_label: Kanban-first default
---
## **Verse 1**

Wires `defaultColumn` and `firstBarIndex` into `New()` so the walking skeleton is real
end to end: open `bp tui`, land on the Doing column, land on its topmost bar — no `Tab`,
no hunting. This closes Verse 1.

## Scope
- `tui/model.go` — `New()`: after building `boardCols`, set `activeCol` via `defaultColumn`
  and select the right row in that column via `firstBarIndex`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestNew_LandsOnDoingsTopBar`
     - **Behavior:** the exact scenario from the scope's motivating example — a track and its
       bar both in Doing, alongside another track in To Do — lands directly on the bar.
     - **Setup:** `New([]*task.Task{ {ID: "BIT-2", Status: "todo"}, {ID: "BIT-1", Status: "doing"}, {ID: "BIT-1.1", Status: "doing"} })`.
     - **Assertions:** `m.activeCol == 1` (Doing); `m.boardSelected().ID == "BIT-1.1"`.
     - **Boundary:** the realistic multi-column, track+bar shape the whole verse exists for —
       not a single-task toy case.
   - [ ] Confirm fails: `activeCol` is `0` (Go zero value, `defaultColumn`/`firstBarIndex` not
     yet called from `New()`).

2. **Implement (GREEN):**
   - [ ] In `New()`, after the `boardCols` loop: `activeCol := defaultColumn(groupByStatus(tasks))`
     (or reuse the already-grouped `cols` from the existing loop rather than recomputing);
     set it on the returned model.
   - [ ] Call `boardCols[activeCol].Select(firstBarIndex(boardCols[activeCol].Items()))` before
     returning.

## Claude verifies
- [ ] `go test ./tui/...` passes (full package, confirming Bar 1.1's migrated tests and this
  new test are all green together).
- [ ] `golangci-lint run` passes.

## User verifies
- [ ] Whole slice: run `bp tui` against a project with a track and an in-progress bar both in
  Doing — it opens directly on the board, Doing column highlighted, the bar (not the track)
  selected. No `Tab` press needed.

## Commit (user)
`feat(tui): land on the Doing column's topmost bar by default`