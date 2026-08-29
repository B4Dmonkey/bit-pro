---
id: BIT-39.18
title: The dequeue waits a tick for the session to survive
status: todo
approved: true
phase: 5
phase_label: Cleanup
---
## **Verse 5**

The dequeue moves off the spawning tick. Measured 2026-08-27 on 2.1.250: a good spawn and a doomed
one both exit 0, and the sub-second confirm poll that follows the spawn catches a dying process
while it is still registering — which is how `EX-2.1`'s row was deleted for a bar that never ran,
leaving no record anywhere that it was attempted. Losing the row is the worse half of that defect:
"dequeue on a confirmed spawn" is sound only if confirmation means the session is doing the work,
and immediately after a spawn it means the process existed for an instant.

The check itself was right; only its timing was wrong. Presence in the plain `claude agents --json`
listing is the honest signal — a session that ran and finished stays there at `state: done` /
`status: idle` across two minutes of polling, while one that died at startup is absent from it
entirely and shows up only under `--all`. So the name check stays exactly as it is and runs a tick
later, by which time a doomed process is gone.

Where it runs matters: **before** the guard, not inside it. The dispatched session is itself what
holds the project, so a confirm placed after the guard would never be reached. And it checks the
head row's derived name against the whole live listing rather than against whichever row the guard
matched — with an operator's own session also live, the guard may match that one instead, and a
confirm that depended on it would stall.

The guard's hold/release rule is untouched.

## Scope
- `daemon/loop.go` — a `confirm` step called from `Tick` between the counts write and the guard;
  `dispatch` loses its post-spawn `claude.Agents` poll, its `not visible yet` warning, and its
  `DeleteQueueRow`. `worktreeFor` takes the queued target ID rather than a loaded `*task.Task`, so
  `confirm` needs one store read instead of two — `task.ParentID` works off the ID string.
- `daemon/loop_test.go` — three existing tests encode the old timing and move with this bar:
  - `TestTick_DequeuesAConfirmedDispatch` — its runner reports the session on every call, so under
    the new order it dequeues on tick 1 before anything is spawned. Replaced by the two-tick test
    below.
  - `TestTick_KeepsTheRowWhenTheSessionCannotBeConfirmed` — asserts a `WARN` that no longer exists.
    Replaced by the never-registers test below.
  - `TestTick_DrainsATrackOneBarPerTick` — its tick script deletes the session immediately after
    each tick, so the confirm never gets a chance to see it. Needs a confirm tick between each
    spawn tick and each deletion: spawn tick → confirm tick → delete the session → repeat.

`confirm` runs only when `mayDispatch` is true — if the `claude agents --json` call failed, the
listing is empty for the wrong reason and dequeuing on it would drop a live bar's row.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTick_DequeuesOnALaterTick`
     - **Behavior:** a spawned bar's row survives the tick that spawned it and clears on a later
       tick, so a session that dies on the way up cannot take its queue row with it.
     - **Setup:** `queries, project := queuedBar(t)` — one approved bar. A `fakeSessions`-style
       runner: the `--bg` call records the spawn and registers `live[<-n value>] = dir`; every other
       call marshals `live` to JSON. Log to a `bytes.Buffer` at `slog.LevelDebug`. Call `Tick`
       twice, asserting between the calls.
     - **Assertions:** after tick 1 — exactly one `--bg` call; `ListQueueByProject` returns 1 row
       whose `TargetID` is `ACME-1.1`; a record with `msg == "dispatching"`. After tick 2 — no new
       `--bg` call; `ListQueueByProject` returns 0 rows; a record with `msg == "dispatched"` and
       `bar == "ACME-1.1"`.
     - **Boundary:** the tick count between spawn and dequeue — 1 versus 2. Tick 1 is the case the
       current code gets wrong (it deletes there); tick 2 is the case that must still eventually
       clear the row, so the row is not merely held forever.
   - [ ] Confirm fails: the row is already gone after tick 1, and there is no `dispatched` record on
     tick 2 because there is nothing left to confirm.

2. **Implement (GREEN):**
   - [ ] `daemon/loop.go`: change `worktreeFor` to take `targetID string`, deriving the parent with
     `task.ParentID(targetID)`.
   - [ ] Add `confirm(ctx, queries, log, live []claude.Agent, p orm.Project, store *task.Store)`:
     read the queue for the project, return if empty, derive the head row's worktree name, and if
     that name appears in `live`, log
     `log.Info("dispatched", "project", p.Code, "bar", head.TargetID, "worktree", name)` and
     `DeleteQueueRow(head.ID)`.
   - [ ] `Tick`: call `confirm` after the `mayDispatch` check and before the guard.
   - [ ] `dispatch`: delete the `claude.Agents` call, the `not visible yet` warning, and the
     `DeleteQueueRow` that followed them. It ends at the `dispatching` record.
   - [ ] Rewrite the three existing tests listed in Scope.

3. **More tests (RED → GREEN):**
   - [ ] `TestTick_KeepsTheRowWhenTheSpawnNeverRegisters`
     - **Behavior:** a spawn that never produces a session keeps its bar in the queue and retries,
       rather than silently dropping the bar — the loop has no state, so it cannot tell this row
       from one never dispatched, and retrying is the settled behaviour.
     - **Setup:** `queries, project := queuedBar(t)`. A recording runner that returns `"[]", 0, nil`
       for every call, including the `--bg` one — a spawn that exits 0 and registers nothing. Call
       `Tick` three times.
     - **Assertions:** `ListQueueByProject` returns 1 row after each of the three ticks, still
       `ACME-1.1`; three `--bg` calls were recorded; three `dispatching` records exist and no
       `dispatched` record does.
     - **Boundary:** three ticks with zero sessions ever appearing — N > 1, so it proves the row is
       retried rather than merely surviving one tick, and that nothing eventually gives up and
       deletes it.
   - [ ] Confirm fails: the current code warns once and leaves the row, so the row count passes but
     the `dispatching`/`dispatched` record assertions do not.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] Whole slice, in `tools/example`. `./reset.sh last`, approve `EX-2`'s bars, answer `y` at the
  play prompt, then `bp serve daemon -v` in a terminal so you can watch the ticks.

  Read the log as it runs. One bar's dispatch reads as three records across two ticks, and each one
  says something different: `dispatching` naming the bar, the worktree, and what `claude` printed;
  then `dispatched` on a later tick; then `not dispatching` naming the session that now holds the
  project. Nothing in that sequence requires you to guess what the loop is doing — which is what
  Verse 5 is for.
- [ ] Watch the row survive the spawning tick:

  ```
  sqlite3 "${XDG_DATA_HOME:-$HOME/.local/share}/bit-pro/bit.db" \
    'SELECT id, target_id FROM queue ORDER BY id'
  ```

  Run it right after the `dispatching` record appears — `EX-2.1`'s row is still there. Run it again
  after `dispatched` — it is gone, and `EX-2.2` is now the head.
- [ ] Delete the `EX-2-<slug>` session and confirm `EX-2.2` dispatches on the next tick, then
  `bp stop`.

## Commit (user)
`fix(daemon): dequeue a bar only once its session has survived a tick`