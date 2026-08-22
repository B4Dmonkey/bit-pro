---
id: BIT-34.3
title: e on a track queues its bars
status: done
approved: true
phase: 3
phase_label: Track shortcut
---
## **Verse 3**

`e` on a track queues that track's bars — the same rows, through the same helper, as
answering `y` at the prompt. Forced by a test that presses `e` with a track selected and
asserts the bar IDs arrive: today that call sends the track's own ID typed `track`, so
satisfying it is what removes the last surface that writes a `track` row.

## Scope
- `tui/model.go` — `enqueueSelected` expands a selected track through
  `enqueueableBarIDs`; a selected bar still enqueues itself; both now pass `targetBar`, so the
  `targetTrack` constant goes.
- `tui/board_test.go` — a board-mode case, since `board.go`'s `e` shares `enqueueSelected`.
- `tui/model_test.go` — `TestUpdate_EKey_EnqueuesTrack` is rewritten to the expansion;
  `TestUpdate_EKey_EnqueuesBar` keeps its single-ID assertion.

No change to `tui/board.go` itself — its `e` case already calls `enqueueSelected`, so the
behaviour follows from the shared helper rather than a second implementation.

## References
- `automation-notes.md` — "Decisions": the ledger is the source of truth; a bar already `done`
  is skipped. The filter here is the same `enqueueableBarIDs` the prompt uses.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestUpdate_EKey_EnqueuesTrackBars` (rewrite of `TestUpdate_EKey_EnqueuesTrack`)
     - **Behavior:** the `e` shortcut on a track produces exactly what the prompt's `y`
       produces — the track's approved, not-yet-done bars, in track order, typed `bar` — so an
       operator who declined the prompt is not getting a second behaviour.
     - **Setup:** `New([]*task.Task{{ID: ttid1, Approved: true, Status: task.StatusDoing},
       {ID: ttid1_1, Approved: true, Status: task.StatusTodo}, {ID: ttid1_2, Approved: true,
       Status: task.StatusTodo}})` with
       `WithEnqueue(func(ids []string, typ string) error { calls++; gotIDs, gotTyp = ids, typ; return nil })`;
       `Code: tea.KeyTab` to reach `modeList`, leaving the track selected at index 0; then
       `Code: 'e'`.
     - **Assertions:** `calls == 1`; `gotIDs` equals `[]string{ttid1_1, ttid1_2}`;
       `gotTyp == targetBar`.
     - **Boundary:** the selected row is a track (no dot) with bar count > 1 — the branch of
       `enqueueSelected` that currently produces a `track` row; the sibling case, a selected
       bar, is held by `TestUpdate_EKey_EnqueuesBar`.
   - [ ] Confirm fails: `gotIDs == []string{ttid1}` and `gotTyp == "track"` — the selected
         track's own ID, typed `track`.

2. **Implement (GREEN):**
   - [ ] In `enqueueSelected`, after the `t == nil` guard: if `!isBar(t.ID)`, set
         `ids = m.enqueueableBarIDs(t.ID)`; otherwise `ids = []string{t.ID}`. Return early when
         `len(ids) == 0`, then `_ = m.enqueue(ids, targetBar)`. The `typ` local goes with it.
   - [ ] Delete the now-unused `targetTrack` constant from the `const` block in `tui/model.go`.
   - [ ] `tui/model_test.go`: `TestUpdate_EKey_EnqueuesBar` asserts `[]string{ttid1_1}` and
         `targetBar` — no change beyond dropping any `targetTrack` reference.

3. **More tests (RED → GREEN):**
   - [ ] `TestUpdate_EKey_TrackWithNothingEnqueueableIsSilent`
     - **Behavior:** `e` on a track whose bars are all done or all unapproved writes no rows
       and shows no error, matching the prompt's silent no-op.
     - **Setup:** `New` with track `ttid1` and bars `ttid1_1` (approved, `StatusDone`),
       `ttid1_2` (`Approved: false`, `StatusTodo`); Tab to `modeList`; press `e` with the track
       selected.
     - **Assertions:** `calls == 0`.
     - **Boundary:** enqueueable count == 0 with bar count == 2 — proves the guard is on the
       filtered slice, not on whether the track has children.
   - [ ] `TestUpdateBoard_EKey_EnqueuesTrackBars` in `tui/board_test.go`
     - **Behavior:** the board's `e` goes through the same expansion as the list's, so the two
       views cannot drift.
     - **Setup:** `New([]*task.Task{{ID: ttid1, Approved: true, Status: task.StatusDoing},
       {ID: ttid1_1, Approved: true, Status: task.StatusTodo}, {ID: ttid1_2, Approved: true,
       Status: task.StatusTodo}})` with the same `WithEnqueue` capture. No `Tab` — `New` leaves
       the model in `modeBoard`, and with the track the only card in `Doing`, `defaultColumn`
       picks column 1 and `firstBarIndex` finds no bar there, so index 0 — the track — is
       selected. Then `Code: 'e'`.
     - **Assertions:** `gotIDs` equals `[]string{ttid1_1, ttid1_2}`; `gotTyp == targetBar`.
     - **Boundary:** the same track-selected branch reached through `updateBoard` rather than
       `handleListKey` — the second entry point into `enqueueSelected`.

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` passes
- [ ] `grep -rn targetTrack --include='*.go' .` returns nothing — the only constant that
      could type a row `track` is gone

## User verifies
- [ ] `just install`, then `./clear-queue.sh`. In `bp tui` press `e` on a track, then
      `sqlite3 ~/.local/share/bit-pro/bit.db 'select target_id, target_typ from queue'` — one
      `bar` row per approved, not-done bar, no `track` row. Press `e` on the same track twice
      more: the row count does not change.
- [ ] Whole slice: a track can be handed to the queue in one keystroke and the queue reads
      back as a list of bars — the rows the operator was asked to confirm are the rows that
      are there, whether they came from the prompt or from `e`.

## Commit (user)
`fix(tui): expand a track to its bars on the e shortcut`