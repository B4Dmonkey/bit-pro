---
id: BIT-45.3
title: Approving a track's last bar opens no Play prompt, and the TUI has no queue
status: todo
phase: 1
phase_label: No daemon in a v2 build
---
## **Verse 1**

Drops the TUI's enqueue path: the "Play? (y / n)" prompt, the `e` key, the queued (cyan) display and the `cmd/tui.go` queue wiring. Browse, reload (the poll tick and the reload after approve) and approve stay. This is a pure removal, so it adds no new tests: the existing Play-prompt, enqueue and queued tests are deleted with the code they cover. This bar must come before BIT-45.4, because `cmd/tui.go:60-80` calls `GetProjectByPath`, `EnqueueTask` and `ListQueueByProject`.

## Scope
- `cmd/tui.go`: drop the `enqueue`/`listQueue` vars and the `db.Open` block (:31-40), the `.WithEnqueue`/`.WithListQueue` calls (:45-46) and `queueFuncs` (:51-94), plus the imports `context`, `os`, `db`, `orm`. The TUI no longer opens the registry db at all.
- `tui/model.go`:
  - `reloadedMsg.queued` (:36), the fields `enqueue`/`listQueue`/`queuedIDs` (:89-91) and `playPromptOpen`/`playPromptTitle`/`playPromptTrackID`/`pendingApprovalID` (:96-99).
  - `WithEnqueue`/`WithListQueue` (:174-182), and the `listQueue` part of `reloadCmd` (:197-209), which returns `reloadedMsg{tasks, err}` again.
  - The `queuedIDs`/`applyQueued`/`pendingApprovalID` handling in the reload branch (:273-292), and `idSet`, `applyQueued`, `handlePlayPrompt`, `enqueueableBarIDs`, `enqueueSelected` and `targetBar` (:28).
  - The `playPromptOpen` guard at the top of key handling (:378-379), the `pendingApprovalID` assignment (:414), the `"e"` case (:438), and the view branches (:636, :648-649).
  - The helpers left unused (:721-753): `barChildrenOf`, `trackTitle`, `allApproved`.
  - `newColumnList(cards, nil)` (:125) and `newColumnList(cards, m.queuedIDs)` (:247) → `newColumnList(cards)`.
- `tui/board.go`: the `queued` param of `newColumnList` (:78, :84 → `delegate{board: true}`), the `pendingApprovalID` assignment (:192-194), the `"e"` case (:226-227) and `playPromptView` (:271).
- `tui/delegate.go`: `queuedColor` (:19), the `queued` field (:24) and its branch in the style resolver (:33-35).
- Tests:
  - `tui/model_test.go`: delete :38-169 (8 Play-prompt tests) and :1456-1699 (9 render/enqueue/queued tests), and the `slices` import if it's left unused.
  - `tui/board_test.go`: delete `TestUpdateBoard_EKey_EnqueuesTrackBars` (:877-905), and drop the `nil` 2nd arg of `newColumnList` at :237, :246-248, :259.
  - `tui/delegate_test.go`: delete `TestResolveStyle_Queued_IsCyan` and `TestResolveStyle_Selected_BeatsQueuedCyan` (:278-302).
  - Delete any `tui/testconst_test.go` const that `unused` flags once those tests are gone.
- Test conversion (`.claude/rules/go-tests.md`). This bar touches `tui/model_test.go`, `tui/board_test.go` and `tui/delegate_test.go`, so it converts every test that survives in them to one top-level test per unit, as a pure restructure: same cases, same assertions, still green. (`tui/testconst_test.go` has no test functions.)
  - Each flat test `TestX_Suffix` becomes `t.Run("<suffix as lowercase words>", ...)` inside `TestX`, with its body unchanged. A test that called `t.Parallel()` keeps the call in its subtest. Tests already named for their unit (`TestSameTasks`, `TestSplitWidth`, `TestSplitWidthExpanded`, `TestIsBar`, `TestVerse`, `TestDefaultColumn`, `TestFirstBarIndex`, `TestFlattenBoard`) stay as they are.
  - `TestUpdate` and `TestView` live in `tui/model_test.go`, next to `Update` and `View` in `model.go`. A package can hold only one of each, so the `TestUpdate_*` and `TestView_*` tests in `tui/board_test.go` move into them as subtests (e.g. `TestUpdate/board active column`, `TestUpdate/modal closes`, `TestView/board column counts`, `TestView/modal shows body`). Their helpers stay in `board_test.go`.
  - The resulting top-level tests:
    - `tui/model_test.go`: `TestNew` (preserves store order, empty list, list help disabled, defaults to board mode, lands on doings top bar), `TestUpdate` (every surviving `TestUpdate_*` from both files, e.g. `TestUpdate/reloaded msg rebuilds list`, `TestUpdate/focus`, `TestUpdate/space toggles approval in board mode`), `TestInit/starts polling when reload set`, `TestSelected` (tracks cursor, empty list nil), `TestLayout/expanded uses wider split`, `TestView` (every surviving `TestView_*` from both files), `TestTitledBorder` (active uses terminal green, active title inverted, inactive title framed, active title not framed), plus the unchanged unit-named tests above.
    - `tui/board_test.go`: `TestGroupByStatus` (its existing body stays; preserves order within column, empty, unapproved todos visible in list not board, unapproved todo is hidden from board, approved todo appears in board join it as subtests), `TestBoardColumns/from grouping`, plus the unchanged unit-named tests above.
    - `tui/delegate_test.go`: `TestDelegate` (every `TestDelegate_*`, e.g. `TestDelegate/selected row uses terminal green`, `TestDelegate/track vs bar distinguished by weight not color`).

## Steps
- [ ] Convert the surviving tests in `tui/model_test.go`, `tui/board_test.go` and `tui/delegate_test.go` as listed under Test conversion. Leave out the tests this bar deletes. `just test` stays green.
- [ ] Delete the `pendingApprovalID` → `playPromptOpen` path (model.go :99, :282-292, :414; board.go :192-194), `handlePlayPrompt`, `playPromptView` and the view branches.
- [ ] Then delete the rest of the Scope list (enqueue, listQueue, queued, `e`, the `cmd/tui.go` wiring) and the obsolete tests. The compiler and `just lint` drive this: every removed `With*` and field breaks its callers, and `unused` names every helper left behind. Delete them, don't suppress them.

## Claude verifies
- [ ] `just test` passes.
- [ ] `just lint` reports `0 issues`.
- [ ] `grep -rn -i 'enqueue\|queued\|playPrompt\|listQueue' --include='*.go' tui/ cmd/` matches nothing.

## User verifies
- [ ] Rebuild the temp binary (`just db-gen-queries && go build -o /tmp/bp-v2 .`). Then run the TUI on a copy of the records so approvals don't touch the real `.bit/`: `SBX=$(mktemp -d); mkdir $SBX/proj; cp -R .bit $SBX/proj/; cd $SBX/proj; HOME=$SBX XDG_DATA_HOME=$SBX/share /tmp/bp-v2 tui`. Approve the last unapproved bar of a track with space. The bar shows approved after the reload, and no "Play … ? (y / n)" prompt appears. Pressing `e` does nothing, and no row is cyan.

## Commit (user)
`feat(tui)!: drop the Play prompt, enqueue key and queued display`