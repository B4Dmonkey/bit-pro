---
id: BIT-23.9
title: Board todo column shows only approved items
status: todo
phase: 5
phase_label: Board filter
---
## **Verse 5**

The board's todo column currently shows everything with `todo` status. That makes it answer "what exists" rather than "what is ready to work on." Filtering unapproved todos out of the board (while keeping them visible in the list view) makes the board actionable.

## Scope
- `tui/board.go` — in `groupByStatus`, skip a task from the todo column (index 0) when `t.Status == "todo" && !t.Approved`. The existing test `TestGroupByStatus` does not set `Approved`, so those tasks default to `false` and will be filtered out — update that test to set `Approved: true` on tasks that should appear in the todo column, or add a targeted new test first.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestGroupByStatus_UnapprovedTodoIsHiddenFromBoard`
     - **Behavior:** an unapproved todo task does not appear in the board's todo column
     - **Setup:** `tasks := []*task.Task{{ID: "BIT-1", Status: "todo", Approved: false}}`
     - **Assertions:** `groupByStatus(tasks)[0]` (todo column) has length 0
     - **Boundary:** `Status == "todo"` and `Approved == false` — the filtered case; proves unapproved todos are excluded
   - [ ] Confirm fails: current `groupByStatus` places any `todo` task in the first column regardless of Approved
   - [ ] `TestGroupByStatus_ApprovedTodoAppearsInBoard` (contradiction)
     - **Behavior:** an approved todo task appears in the board's todo column
     - **Setup:** `Approved: true`, `Status: "todo"`
     - **Assertions:** `groupByStatus(tasks)[0]` has length 1
     - **Boundary:** approved todo — contradicts filtering all todos; proves only unapproved are filtered

2. **Implement (GREEN):**
   - [ ] In `groupByStatus` (in `tui/board.go`), add a guard for the todo column: `if col.status == "todo" && !t.Approved { continue }` (or equivalent using the column index)
   - [ ] Update `TestGroupByStatus` (existing test) to set `Approved: true` on tasks intended for the todo column, so it stays green with the new filter
   - [ ] Update `TestView_BoardColumnCounts` and any other board tests that create `todo` tasks without setting `Approved: true`

3. **More tests (RED → GREEN):**
   - [ ] `TestGroupByStatus_UnapprovedTodosVisibleInListNotBoard`
     - **Behavior:** list view model still receives all tasks (including unapproved todos); only the board view hides them
     - **Setup:** `New([]*task.Task{{ID: "BIT-1", Status: "todo", Approved: false}})`
     - **Assertions:** `m.Items()` has length 1 (list gets everything); `m.boardCols[0].Items()` has length 0 (board filters)
     - **Boundary:** the divergence between list and board — proves unapproved tasks remain list-visible

## Claude verifies
- [ ] `just test` passes (including updated board tests)
- [ ] `just lint` clean

## User verifies
- [ ] `bp tui` → board view: unapproved tasks do not appear in the "To Do" column; switch to list view (Tab) — the same unapproved tasks are still listed there in yellow

## Commit (user)
`feat(board): filter unapproved todos from the kanban to-do column`