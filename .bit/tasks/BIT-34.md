---
id: BIT-34
title: The queue holds bars only
status: todo
---
## Why

The play prompt asks the operator "Play <track>? (y / n)" and, on yes, writes a queue row for
whichever task happens to be selected — the single bar just approved. The operator confirms a
track and gets one bar. Separately, pressing `e` on a track writes a row typed `track`, meaning
"work this track's bars in order" — a semantics nothing implements today and that the dispatch
design has since settled *against*: dispatch will only ever pop bars.

So the queue currently holds two kinds of row. One silently under-delivers what the operator was
asked to confirm; the other will never be honoured by any consumer. Nothing has popped the queue
yet, so both are invisible — the moment dispatch starts reading, the first becomes "I queued a
three-bar track and only one ran" and the second becomes "I queued a track and nothing happened."
Fixing the meaning of a row is cheap now and a data-migration later.

## Summary

Every enqueue path writes **bar rows only**. Answering yes at the play prompt enqueues all of the
track's approved, not-yet-done bars in track order — what the prompt already promises. Pressing `e`
on a track does the same; `e` on a bar enqueues that one bar. No path writes a `track` row again,
and re-enqueueing the same bar becomes a no-op at the storage layer rather than a duplicate row.

## Visual aid

```
today                                    after
  approve last bar → "Play BIT-7?"        approve last bar → "Play BIT-7?"
    y → queue: [BIT-7.3]        ✗           y → queue: [BIT-7.1, BIT-7.2, BIT-7.3]   ✓
        (the selected bar only)                  (approved, not-done, in track order)

  e on track BIT-7                        e on track BIT-7
    → queue: [BIT-7 (track)]    ✗           → queue: [BIT-7.1, BIT-7.2, BIT-7.3]     ✓

  e, e, e on BIT-7.2                      e, e, e on BIT-7.2
    → 3 rows                    ✗           → 1 row                                  ✓

```

## Decisions

- **Only bars are queued.** A track is a way of *selecting* bars to queue, never a kind of queue
  row. This is the whole point: one row kind means one dequeue rule in the dispatch track.
- **The popup enqueues the track's bars in track order.** `Store.List()` already sorts by the
  track's `order` list (`task/store.go:355`) and `barChildrenOf` preserves that order, so ordering
  is inherited rather than recomputed.
- **Only approved, not-yet-done bars are enqueued.** Approval gating is unchanged — it is what
  clears work to run. Skipping `done` bars honours the standing decision that the ledger is the
  source of truth and a bar already done is skipped rather than re-run; it matters because the
  prompt fires on a replanned track whose earlier bars are already finished.
- **`e` on a track is exactly the popup's yes.** Same code path, same rows. The shortcut exists for
  the operator who answered no, or never got a prompt — it should not be a second behaviour.
- **Re-enqueue is idempotent at the storage layer.** A unique index on `(project_id, target_id)`
  plus `INSERT OR IGNORE`. Enforced in the database rather than in the TUI, so no caller — the
  popup, the `e` key, or anything the dispatch track adds — can produce a double-spawn, and the
  guarantee does not depend on how fresh the TUI's reload is.
- **`target_typ` stays in the schema and is always `bar`.** Dropping a column is churn that buys
  nothing, and the dispatch track may still want to assert on the value it reads.
- **Rendering fallout is explicitly not this track's problem.** Queueing a track will colour its
  bars rather than the track row, and that is fine — no verse or bar should be spent on colour. If
  it reads badly in practice it gets cleaned up afterwards.
- **Nothing enqueueable is a silent no-op.** A track whose bars are all done, or all unapproved,
  writes no rows and shows no error — consistent with the existing behaviour for an unregistered
  project.

## Verses

- [ ] Verse 1 — Enqueueing the same bar twice stops corrupting the queue: the operator can press
  `e` repeatedly, or answer yes twice, and the queue holds one row per bar. Lands first because the
  later verses multiply one gesture into N rows, and doing that before the guarantee exists makes
  the blast radius worse than today.
  Touches: `db/migrations/` (new migration), `db/queries/queue.sql`, `db/orm/` (sqlc regenerated).

- [ ] Verse 2 — The play prompt queues the track it named: answering yes queues every approved,
  not-yet-done bar of the track, in order, so the operator gets the run they confirmed instead of a
  single bar.
  Touches: `tui/model.go` (`handlePlayPrompt`, `enqueueSelected`), `cmd/tui.go` (`queueFuncs` — the
  enqueue seam takes more than one target).

- [ ] Verse 3 — `e` on a track queues the whole track: the shortcut matches the popup, so an
  operator who declined the prompt can still queue a track in one keystroke. After this no surface
  writes a `track` row.
  Touches: `tui/model.go` (`enqueueSelected`, key dispatch), `tui/board.go`.

## References

- `automation-notes.md` — the automation phase working notes. Step 5 (BIT-33) is the queue this
  track corrects; the "Decisions" and "Open gaps" sections carry the standing rules this scope
  honours. Informs all three verses.