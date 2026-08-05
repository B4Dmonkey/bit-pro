---
id: BIT-20.2
title: defaultColumn picks the first non-empty column
status: done
phase: 1
phase_label: Kanban-first default
---
## **Verse 1**

Landing on the board is only half of Verse 1 — it also has to land on the *right* column. This
step builds the pure lookup that decides which one: the first non-empty column in
To Do → Doing → Done order. Kept as its own function (rather than inlined into `New()`) so it's
directly table-testable, matching the existing style of small tested helpers like `isBar`.

## Scope
- `tui/board.go` — add `defaultColumn(cols [3][]*task.Task) int`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestDefaultColumn/doing_is_the_default_when_populated`
     - **Behavior:** with Doing non-empty, that's the column an operator lands on — the common
       case (a task actively in progress).
     - **Setup:** `cols := [3][]*task.Task{ {&task.Task{ID: "BIT-4"}}, {&task.Task{ID: "BIT-1"}}, nil }` (To Do has one, Doing has one, Done empty).
     - **Assertions:** `defaultColumn(cols) == 1`.
     - **Boundary:** the "normal" case — the middle column populated alongside another.
   - [ ] Confirm fails: `defaultColumn` doesn't exist yet (compile error).

2. **Implement (GREEN):**
   - [ ] `func defaultColumn(cols [3][]*task.Task) int { return 1 }` — hardcoded, passes the one
     test above.

3. **More tests (RED → GREEN):**
   - [ ] `TestDefaultColumn/falls_back_to_to_do_when_doing_is_empty`
     - **Behavior:** an empty Doing column is skipped in favor of To Do, proving the search is
       real, not fixed on index 1.
     - **Setup:** `cols := [3][]*task.Task{ {&task.Task{ID: "BIT-4"}}, nil, nil }`.
     - **Assertions:** `defaultColumn(cols) == 0`.
     - **Boundary:** contradicts the hardcoded `return 1` — forces a real first-non-empty scan.
   - [ ] `TestDefaultColumn/falls_back_to_done_when_only_done_has_tasks`
     - **Behavior:** the scan reaches all the way to Done, not just past Doing.
     - **Setup:** `cols := [3][]*task.Task{ nil, nil, {&task.Task{ID: "BIT-4"}} }`.
     - **Assertions:** `defaultColumn(cols) == 2`.
     - **Boundary:** the last index in the array — the far end of the scan order.
   - [ ] `TestDefaultColumn/all_empty_defaults_to_zero`
     - **Behavior:** an empty board (e.g. a brand-new project) doesn't panic or return an
       out-of-range index.
     - **Setup:** `cols := [3][]*task.Task{}`.
     - **Assertions:** `defaultColumn(cols) == 0`.
     - **Boundary:** the zero case — no column has anything, the lower bound of the input space.
   - [ ] Implement the real loop: `for i, c := range cols { if len(c) > 0 { return i } }; return 0`.

## Claude verifies
- [ ] `go test ./tui/... -run TestDefaultColumn` passes.
- [ ] `golangci-lint run` passes.

## User verifies
- [ ] None — pure helper, not yet wired into `New()` (that's this verse's last bar).

## Commit (user)
`feat(tui): add defaultColumn to pick the board's initial column`