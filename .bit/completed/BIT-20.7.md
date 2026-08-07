---
id: BIT-20.7
title: splitWidthExpanded computes the ~90% detail pane split
status: done
phase: 3
phase_label: Expanded detail pane
---
## **Verse 3**

The expanded detail pane needs its own width split (~90% detail, ~10% list), separate from the
normal 40/60 `splitWidth`. Same shape as `splitWidth` itself (already table-tested in
`tui/model_test.go`'s `TestSplitWidth`) — a pure function, tested on its own first.

## Scope
- `tui/model.go` — add `splitWidthExpanded(total int) (listW, detailW int)`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestSplitWidthExpanded/typical_terminal`
     - **Behavior:** at a normal terminal width, the detail pane takes roughly 90% of the
       space, leaving the list a narrow strip rather than disappearing entirely.
     - **Setup:** `total := 100`.
     - **Assertions:** `listW == 10`, `detailW == 89` (matches `splitWidth`'s own
       `total-listW-1` gap-for-the-divider convention).
     - **Boundary:** a representative width, not an edge case — proves the ~90% target itself.
   - [ ] Confirm fails: `splitWidthExpanded` doesn't exist yet (compile error).

2. **Implement (GREEN):**
   - [ ] `func splitWidthExpanded(total int) (listW, detailW int) { return 10, 89 }` —
     hardcoded, passes the one test above.

3. **More tests (RED → GREEN):**
   - [ ] `TestSplitWidthExpanded/scales_with_width`
     - **Behavior:** the split is a percentage, not a fixed pixel count — it scales with the
       terminal.
     - **Setup:** `total := 200`.
     - **Assertions:** `listW == 20`, `detailW == 179`.
     - **Boundary:** contradicts the hardcoded `10, 89` — forces the real `total*10/100`
       calculation.
   - [ ] `TestSplitWidthExpanded/detail_wider_than_list`
     - **Behavior:** sanity check on the shape of the split itself, mirroring
       `TestSplitWidth`'s own "detail wider than list" subtest.
     - **Setup:** `total := 120`.
     - **Assertions:** `detailW > listW`.
     - **Boundary:** confirms the *direction* of the split, independent of exact numbers.
   - [ ] `TestSplitWidthExpanded/zero_and_one_width`
     - **Behavior:** degenerate terminal widths don't return negative widths or panic.
     - **Setup:** `total` in `{0, 1}`.
     - **Assertions:** `listW >= 0`, `detailW >= 0`, `listW+detailW <= total`.
     - **Boundary:** the lower bound of the input space — mirrors `TestSplitWidth`'s own
       zero/one-width subtests.
   - [ ] Implement the real split: `listW = total * 10 / 100; detailW = max(total-listW-1, 0)`.

## Claude verifies
- [ ] `go test ./tui/... -run TestSplitWidthExpanded` passes.
- [ ] `golangci-lint run` passes.

## User verifies
- [ ] None — pure helper, not yet wired into `layout()` (the next bar).

## Commit (user)
`feat(tui): add splitWidthExpanded for the ~90% detail pane split`