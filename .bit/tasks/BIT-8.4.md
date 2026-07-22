---
id: BIT-8.4
title: List and board preserve order everywhere
status: doing
phase: 1
phase_label: Resequence
---
Lock the holistic invariant with a guard test: the TUI list and the kanban board render in the exact order `List()` hands them, never re-sorting. No production change is expected — `New` builds items in slice order and `groupByStatus` buckets while preserving order — so this bar proves (and freezes) that every surface inherits ordering from the one source. If it forces a production change, that's a real finding the scope depends on.

**Scope:**
- `tui/board_test.go` (new or existing) — assert `groupByStatus` preserves incoming order within each status column.
- `tui/model_test.go` (new or existing) — assert `New(tasks)` yields list items in the same order as the input slice.
- No expected change to `tui/board.go` / `tui/model.go`. If a test fails, fix the consumer to preserve order rather than changing the test.

**TDD cycle:**

1. **Write test (RED → likely GREEN immediately):**
   - [ ] `TestGroupByStatus_PreservesOrderWithinColumn` (table-driven, tui pkg)
     - **Behavior:** the board does not impose its own ordering — cards appear in a column in the same relative order `List()` produced.
     - **Setup:** build a slice of `*task.Task` all with status `todo` in a deliberately non-ID order, e.g. IDs `[BIT-1.2, BIT-1.1, BIT-1.3]` (mimicking a reordered `List()` result).
     - **Assertions:** `groupByStatus(tasks)[0]` (the To Do column) has IDs in exactly `[BIT-1.2, BIT-1.1, BIT-1.3]`.
     - **Boundary:** input order deliberately differs from ID order — the case that would expose a hidden re-sort.
   - [ ] `TestNew_ListItemsFollowInputOrder`
     - **Behavior:** the flat TUI list is a pure consumer of `List()` order.
     - **Setup:** `New` with the same non-ID-ordered slice.
     - **Assertions:** the model's list items, read back in index order, match the input slice IDs.
     - **Boundary:** same non-ID input order.
   - [ ] Confirm result: expected to pass without production changes (documents current behavior); if red, the consumer is re-sorting and must be fixed.

2. **Implement (GREEN):**
   - [ ] Only if a guard test is red: make the offending consumer preserve incoming order. Otherwise no production change.

**Claude verifies:**
- [ ] `just test` passes.
- [ ] `just lint` clean.

**User verifies:**
- [ ] After `just install`, open `bit tui`, `bit task move` a bar, reopen: the moved bar sits in its new spot in both the list pane and the board column. (Manual, cross-surface confirmation of the scope's core claim.)

**Commit (user):** `test(tui): guard that list and board preserve List order`