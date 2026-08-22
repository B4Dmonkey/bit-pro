---
id: BIT-34.2
title: Yes at the prompt queues the track's bars
status: todo
approved: true
phase: 2
phase_label: Play prompt
---
## **Verse 2**

Answering `y` at the play prompt queues every approved, not-yet-done bar of the track the
prompt named, instead of the single selected bar. Forced by an end-to-end test that approves
the last bar of a two-bar track and presses `y`, asserting two IDs reach the enqueue seam —
the current single-target seam cannot express that, and the prompt has to carry the track it
named rather than read the selection.

## Scope
- `tui/model.go` — `enqueue` field and `WithEnqueue` take `(targetIDs []string, targetTyp string)`;
  new `playPromptTrackID` field, set where `playPromptTitle` is already set so the prompt's
  wording and its rows come from one source; new helper expanding a track ID into its
  enqueueable bar IDs; `handlePlayPrompt` uses it instead of `enqueueSelected`;
  `enqueueSelected` adapts to the new seam and otherwise keeps today's behaviour.
- `cmd/tui.go` — `queueFuncs`' `enqueue` loops the IDs into `q.EnqueueTask`, returning on the
  first error.
- `tui/model_test.go` — the four existing `WithEnqueue` callers move to the new signature.

`target_typ` stays a seam parameter and stays in the schema, per the scope's decision. This
bar leaves `enqueueSelected` passing `targetTrack` for a track exactly as it does today —
retyping that row is Verse 3's job, so no commit in between writes a mislabelled row.

## References
- `automation-notes.md` — "Decisions": the ledger is the source of truth, so a bar already
  `done` is skipped rather than re-run. That is what the `done` filter below implements.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestUpdate_PlayPromptYes_EnqueuesTrackBars`
     - **Behavior:** the rows the operator gets match the track the prompt asked about — every
       approved, not-yet-done bar of it, in track order — driven the way the operator drives
       it, from the approval that opens the prompt through to the `y`.
     - **Setup:** `New([]*task.Task{{ID: ttid1}, {ID: ttid1_1, Approved: false, Status: task.StatusTodo},
       {ID: ttid1_2, Approved: true, Status: task.StatusTodo}})` with
       `WithApprove(func(_ string, _ bool) error { return nil })` and
       `WithEnqueue(func(ids []string, typ string) error { calls++; gotIDs, gotTyp = ids, typ; return nil })`.
       Then, mirroring `TestUpdate_BarApprovalSetsPlayPromptOpen`: `Code: tea.KeyTab` (to
       `modeList`), `Code: tea.KeyDown` (select `ttid1_1`), `Code: ' '` (approve),
       `reloadedMsg{tasks: []*task.Task{{ID: ttid1}, {ID: ttid1_1, Approved: true, Status: task.StatusTodo},
       {ID: ttid1_2, Approved: true, Status: task.StatusTodo}}}`, then `Code: 'y'`.
     - **Assertions:** `calls == 1`; `gotIDs` equals `[]string{ttid1_1, ttid1_2}` (slice
       equality, order significant); `gotTyp == targetBar`.
     - **Boundary:** bar count > 1 — the multi-row path the single-target seam cannot
       represent, and the case where today's behaviour (the one selected bar, `ttid1_1`) and
       the wanted behaviour visibly differ. The track row `ttid1` is present and must not
       appear in the slice.
   - [ ] Confirm fails: compile error — `WithEnqueue` wants `func(string, string) error`. Once
         the signature is in place it fails again on the assertion, with `gotIDs ==
         []string{ttid1_1}`: the selected bar only.

2. **Implement (GREEN):**
   - [ ] `tui/model.go`: change the `enqueue` struct field and `WithEnqueue` to
         `func(targetIDs []string, targetTyp string) error`.
   - [ ] Add `playPromptTrackID string` to the `model` struct, next to `playPromptTitle`, and
         set it to `parentID` in `handleReloaded` on the line after
         `m.playPromptTitle = trackTitle(parentID, msg.tasks)`.
   - [ ] Add `func (m model) enqueueableBarIDs(trackID string) []string` — `barChildrenOf(trackID,
         m.loaded)` (which already preserves the store's track order), keeping only bars with
         `t.Approved && t.Status != task.StatusDone`, returning their IDs.
   - [ ] In `handlePlayPrompt`, case `"y"`: set `m.playPromptOpen = false`, then
         `if ids := m.enqueueableBarIDs(m.playPromptTrackID); len(ids) > 0 && m.enqueue != nil {
         _ = m.enqueue(ids, targetBar) }`. It no longer calls `enqueueSelected`.
   - [ ] In `enqueueSelected`, change the final call to `_ = m.enqueue([]string{t.ID}, typ)` —
         the track/bar `typ` choice is unchanged here.
   - [ ] `cmd/tui.go`: `enqueue = func(targetIDs []string, targetTyp string) error` looping
         `q.EnqueueTask(ctx, orm.EnqueueTaskParams{ProjectID: project.ID, TargetID: id,
         TargetTyp: targetTyp})` and returning the first non-nil error; update the local
         `enqueue` variable's declared signature in `newTUICmd` to match.
   - [ ] `tui/model_test.go`: move `TestHandlePlayPrompt_No_SkipsEnqueue`,
         `TestUpdate_EKey_EnqueuesTrack` and `TestUpdate_EKey_EnqueuesBar` to the new
         signature — the two `e` tests assert `[]string{ttid1}` / `[]string{ttid1_1}` with
         their current `targetTrack` / `targetBar` values. Delete
         `TestHandlePlayPrompt_Yes_CallsEnqueue`; the new test above replaces it.

3. **More tests (RED → GREEN):**
   - [ ] `TestHandlePlayPrompt_Yes_SkipsUnapprovedAndDoneBars`
     - **Behavior:** the filter is both-sided — an unapproved bar is not cleared to run and a
       done bar is not re-run, so neither is queued.
     - **Setup:** `New` with track `ttid1` and bars `ttid1_1` (approved, `StatusTodo`),
       `ttid1_2` (`Approved: false`, `StatusTodo`), `ttid1_3` (approved, `StatusDone`);
       set `m.playPromptOpen = true` and `m.playPromptTrackID = ttid1` directly (the wiring is
       already proven by the test above); press `y`.
     - **Assertions:** `gotIDs` equals `[]string{ttid1_1}`.
     - **Boundary:** each of the two filter predicates exercised in its excluding state, with
       one bar passing both so the assertion is not vacuously empty.
   - [ ] `TestHandlePlayPrompt_Yes_NothingEnqueueableIsSilent`
     - **Behavior:** a track with nothing to run writes no rows and reports no error — the
       scope's "nothing enqueueable is a silent no-op".
     - **Setup:** `New` with track `ttid1` and one bar `ttid1_1` (approved, `StatusDone`);
       `m.playPromptOpen = true`, `m.playPromptTrackID = ttid1`; press `y`.
     - **Assertions:** `calls == 0`; the returned model has `playPromptOpen == false`.
     - **Boundary:** enqueueable count == 0 — the lower bound, proving the guard runs before
       the seam rather than sending an empty slice.

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` passes

## User verifies
- [ ] `just install`, then `./clear-queue.sh`. In `bp tui`, approve the last unapproved bar of
      a multi-bar track and answer `y`; then
      `sqlite3 ~/.local/share/bit-pro/bit.db 'select target_id, target_typ from queue'` — one
      `bar` row per approved, not-done bar of that track, in track order, and no row for the
      track ID itself.

## Commit (user)
`fix(tui): queue the whole track when the play prompt is accepted`