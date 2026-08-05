---
id: BIT-20.1
title: New() defaults to board mode, and the test suite's mode assumptions are migrated
status: done
phase: 1
phase_label: Kanban-first default
---
## **Verse 1**

`New()` currently zero-values `mode` to `modeList`. Flipping the default to `modeBoard` is a
one-line change, but nearly every existing test in `tui/model_test.go` and `tui/board_test.go`
was written assuming the *old* default — list-mode tests exercise list keys straight after
`New()`, and board-mode tests press `Tab` once to reach the board. Flipping the default silently
inverts what that `Tab` press does everywhere it appears, so this step isn't just the one-line
default — it's the one-line default *plus* the mechanical migration required to keep the suite
honest about which mode it's actually testing.

## Scope
- `tui/model.go` — `New()`: set `mode: modeBoard` in the returned literal.
- `tui/model_test.go` — add a `Tab` press to reach list mode wherever a test now needs it, and
  fix the mode-toggle expectations.
- `tui/board_test.go` — remove the now-redundant `Tab` press wherever a test was using it only
  to reach board mode (already the default).

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestNew_DefaultsToBoardMode`
     - **Behavior:** `bp tui` lands on the Kanban board without a manual `Tab`.
     - **Setup:** `New([]*task.Task{{ID: "BIT-1"}})`, no key presses.
     - **Assertions:** `m.mode == modeBoard`.
     - **Boundary:** the zero-interaction state — proves the *default*, not a post-toggle state.
   - [ ] Confirm fails: `mode` is `modeList` (Go zero value), assertion mismatch.

2. **Implement (GREEN):**
   - [ ] In `New()`, add `mode: modeBoard` to the returned `model{...}` literal.
   - [ ] Update `TestUpdate_TabTogglesMode` (`tui/model_test.go`): the table's meaning inverts —
     `{"default is board", 0, modeBoard}`, `{"one tab to list", 1, modeList}`,
     `{"two tabs back to board", 2, modeBoard}`.
   - [ ] Add a `Tab` press to reach list mode in every test that exercises list-mode-specific
     behavior straight after `New()`/`WindowSizeMsg` (they all currently rely on list being the
     default): `TestUpdate_ForwardsNavigationToList`, `TestUpdate_CtrlDScrollsDetail`,
     `TestUpdate_NavigationResetsDetailScroll`, `TestUpdate_Focus` (every subtest),
     `TestUpdate_RightDoesNotPageList`, `TestUpdate_FocusRoutesArrows`,
     `TestUpdate_QuitsFromDetail`, `TestView_PaneTitles`, `TestView_ListHidesTitleHeading`,
     `TestView_ListHidesItemCount`, `TestView_EmptyListSingleEmptyState`,
     `TestView_HelpBarPresentAndBounded`. Insert the `Tab` press right after the
     `WindowSizeMsg` (or right after `New()` where there's no window sizing), before any
     list-mode key or view assertion.
   - [ ] Remove the now-redundant `Tab` press (board is already default) from every test in
     `tui/board_test.go` that presses it only to reach board mode:
     `TestView_BoardColumnCounts`, `TestUpdate_BoardActiveColumn`,
     `TestUpdate_BoardCardSelection`, `TestUpdate_BoardEnterOpensModal`,
     `TestView_ModalShowsBody`, `TestUpdate_ModalCloses`, `TestUpdate_ModalCapturesInput`
     (only the setup `Tab` before `Enter` — leave the "tab swallowed" test *case*'s own `KeyTab`
     press alone, that one is verifying tab is swallowed while the modal is open),
     `TestUpdate_ModalScrollsLongBody`, `TestUpdate_BoardEnterEmptyColumnNoop`,
     `TestUpdate_BoardQuits`. Also remove it from `tui/model_test.go`'s
     `TestUpdate_ReloadPreservesBoardSelection` and `TestView_ModalTitleInverted`.
   - [ ] `TestView_BoardHelp` (`tui/board_test.go`) is the one case that goes both ways in the
     same table: invert its `toBoard` press so it presses `Tab` when the case wants **list**
     mode (`!tt.toBoard`) and presses nothing when it wants board (already default).

## Claude verifies
- [ ] `go test ./tui/...` passes — every test above included, none accidentally left on the
  wrong mode.
- [ ] `golangci-lint run` (or the project's configured linter) passes.

## User verifies
- [ ] None — this step only flips a default and repairs test setup; no new user-observable
  behavior beyond what Verse 1's last bar demonstrates end to end.

## Commit (user)
`feat(tui): default to Kanban board view on startup`