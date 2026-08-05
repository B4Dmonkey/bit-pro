---
id: BIT-20.5
title: flattenBoard produces the modal's cross-column paging order
status: done
phase: 2
phase_label: Modal paging
---
## **Verse 2**

Paging through the modal needs a single, board-order sequence of tasks to move through — the
scope's Decision is "To Do, then Doing, then Done, top-to-bottom within each column." This step
builds that sequence as a pure function, carrying each entry's column and position so the next
bar can jump straight to the right card without a second search.

## Scope
- `tui/board.go` — add `type boardEntry struct { col, pos int; t *task.Task }` and
  `flattenBoard(cols [3]list.Model) []boardEntry`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestFlattenBoard/single_column`
     - **Behavior:** a column's own items come back in their own order, each tagged with its
       column and position.
     - **Setup:** `cols[1] = newColumnList([]*task.Task{{ID: "BIT-1"}, {ID: "BIT-1.1"}})`, the
       other two columns left as their zero-value `list.Model{}`.
     - **Assertions:** `flattenBoard(cols)` == `[]boardEntry{{col:1,pos:0,t:BIT-1}, {col:1,pos:1,t:BIT-1.1}}` (compare IDs/col/pos, not pointer identity).
     - **Boundary:** one populated column among three — proves per-column ordering, not yet
       cross-column concatenation.
   - [ ] Confirm fails: `flattenBoard`/`boardEntry` don't exist yet (compile error).

2. **Implement (GREEN):**
   - [ ] Minimal version that only reads `cols[1]` (hardcoded column index), builds `boardEntry`
     for each of its items — enough to pass the one test above.

3. **More tests (RED → GREEN):**
   - [ ] `TestFlattenBoard/all_three_columns_concatenate_in_order`
     - **Behavior:** the real scope Decision — the whole board's paging order is To Do, then
       Doing, then Done, not just one column.
     - **Setup:** `cols[0]` has `BIT-2`; `cols[1]` has `BIT-1`, `BIT-1.1`; `cols[2]` has `BIT-3`.
     - **Assertions:** `flattenBoard(cols)` IDs in order: `["BIT-2", "BIT-1", "BIT-1.1", "BIT-3"]`,
       with `col` `0,1,1,2` and `pos` `0,0,1,0` respectively.
     - **Boundary:** contradicts the column-1-only hardcode — forces a real loop over all three
       columns in array order.
   - [ ] `TestFlattenBoard/empty_columns_are_skipped_not_padded`
     - **Behavior:** an empty column contributes nothing — it doesn't leave a gap or a
       zero-value entry in the sequence.
     - **Setup:** `cols[0]` empty, `cols[1]` has `BIT-1`, `cols[2]` empty.
     - **Assertions:** `flattenBoard(cols)` has exactly one entry, `BIT-1`, `col:1, pos:0`.
     - **Boundary:** the lower bound of "how many items a column contributes" — zero.
   - [ ] `TestFlattenBoard/all_empty_returns_empty_not_nil_panic`
     - **Behavior:** a brand-new board with nothing in any column doesn't panic.
     - **Setup:** all three columns empty.
     - **Assertions:** `len(flattenBoard(cols)) == 0`.
     - **Boundary:** the zero case across the whole board.
   - [ ] Implement the real loop: `for col := range cols { for pos, it := range cols[col].Items() { if bi, ok := it.(item); ok { append(entries, boardEntry{col, pos, bi.t}) } } }`.

## Claude verifies
- [ ] `go test ./tui/... -run TestFlattenBoard` passes.
- [ ] `golangci-lint run` passes.

## User verifies
- [ ] None — pure helper, not yet wired into modal key handling (the next bar).

## Commit (user)
`feat(tui): add flattenBoard for the modal's cross-column paging order`