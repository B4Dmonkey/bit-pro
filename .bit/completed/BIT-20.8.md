---
id: BIT-20.8
title: layout() uses the expanded split when detailExpanded is set
status: done
phase: 3
phase_label: Expanded detail pane
---
## **Verse 3**

Wires `splitWidthExpanded` into `layout()`, gated on a new `detailExpanded` field — the state
the next two bars will let the user toggle and page while in. `layout()` is the single place
`listWidth`/`detailWidth` get computed (from `tea.WindowSizeMsg` and from the `?` help toggle),
so gating the split choice there covers every path that resizes the panes.

## Scope
- `tui/model.go` — add `detailExpanded bool` field to `model`; `layout()`: choose
  `splitWidthExpanded` over `splitWidth` when `m.detailExpanded`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestLayout_ExpandedUsesWiderSplit`
     - **Behavior:** setting `detailExpanded` changes which split function `layout()` uses,
       proving the gate is wired, not just the helper existing in isolation.
     - **Setup:** `m := New([]*task.Task{{ID: "BIT-1"}})`; `m.detailExpanded = true`; send
       `tea.WindowSizeMsg{Width: 100, Height: 24}`.
     - **Assertions:** resulting `listWidth == 10` (matches `splitWidthExpanded(100)`, not
       `splitWidth(100)`'s `40`).
     - **Boundary:** the "on" state of a boolean gate — the only two states this field has.
   - [ ] Confirm fails: `layout()` always calls `splitWidth`, ignoring `detailExpanded`
     (`listWidth` comes back `40`).

2. **Implement (GREEN):**
   - [ ] Add `detailExpanded bool` to the `model` struct.
   - [ ] In `layout()`, replace the unconditional `listW, detailW := splitWidth(m.winWidth)`
     with a branch: `splitWidthExpanded(m.winWidth)` when `m.detailExpanded`, else
     `splitWidth(m.winWidth)` (unchanged default path).

## Claude verifies
- [ ] `go test ./tui/...` passes — confirms the default (`detailExpanded == false`) path is
  untouched (existing `TestSplitWidth`-adjacent view tests stay green).
- [ ] `golangci-lint run` passes.

## User verifies
- [ ] None — `detailExpanded` has no way to become `true` from user input yet (the next bar).

## Commit (user)
`feat(tui): gate the detail pane's width split on detailExpanded`