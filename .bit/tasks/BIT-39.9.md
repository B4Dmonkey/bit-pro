---
id: BIT-39.9
title: A whole track drains one bar per tick
status: done
approved: true
phase: 3
phase_label: Track runs bar-by-bar
---
## **Verse 3**

A whole track drains bar-by-bar across ticks, and a busy project does not hold up a free one.
**This bar is expected to need no production change** — FIFO falls out of `ListQueueByProject`'s
`ORDER BY id`, one-at-a-time out of the Verse 2 guard, and the shared worktree out of per-track
name derivation. What it adds is the multi-tick test that pins those three as guaranteed rather
than incidental. If it goes red, that is a Verse 2 defect, not new work — fix it there.

## Scope
- `daemon/loop_test.go` — the multi-tick drain test. No production file should need to change; if
  one does, say which in the commit.

## TDD cycle

1. **Write test (RED, expected green):**
   - [ ] `TestTick_DrainsATrackOneBarPerTick` in `daemon/loop_test.go`
     - **Behavior:** consecutive ticks work a track's queued bars in enqueue order, one per tick,
       every one of them in that track's single worktree — so an operator who approves a three-bar
       track and answers "yes" gets the track built in order without touching it.
     - **Setup:** `dirA := t.TempDir()`; save track `ACME-1` ("a track", approved) and bars
       `ACME-1.1`, `ACME-1.2`, `ACME-1.3` (all `todo`, approved). A second project `dirB` with
       track `ZULU-1` ("other track", approved) and bar `ZULU-1.1` (`todo`, approved), enqueued
       too. Register both, enqueue `ACME-1.1`, `.2`, `.3` in that order, then `ZULU-1.1`.
       A programmable fake runner: it holds a set of "live" cwds, answers `agents --json` from that
       set plus the names it has been told to confirm, records `--bg` calls, and lets the test
       clear the live set between ticks — the stand-in for a session finishing.
     - **Assertions:** tick 1 records exactly two `--bg` calls, one per project — `ACME-1.1` in
       `dirA` and `ZULU-1.1` in `dirB` — proving a project is guarded independently, not globally.
       With `dirB`'s session left live and `dirA`'s cleared, tick 2 records exactly one call, for
       `ACME-1.2`. Tick 3, same, for `ACME-1.3`. All three `ACME` calls carry
       `-w ACME-1-a-track` — the identical string, so bar 3 lands on bar 1's tree. After tick 3,
       `ListQueueByProject(dirA's id)` is empty and `dirB`'s still holds nothing but what was never
       confirmed.
     - **Boundary:** queued rows for one project == 3 — above the 1 every Verse 2 test used, which
       is the smallest count that can distinguish FIFO from LIFO from arbitrary. Registered
       projects == 2, likewise the smallest count that can show the guard is per-project.
   - [ ] Confirm result: expected to pass on the first run. Note in the commit body that it did.
     A red here means Verse 2 is wrong — most likely the guard skipping all projects instead of
     one, or the row popped by position rather than by `id`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] Whole slice, in `tools/example`. Reset first so the track is clean: `./reset.sh last` returns
  to the `ck/ex2` checkpoint — `EX-1` and `EX-2` written and unapproved, no worktrees, no run
  branches, no leftover sessions. (Stamp your own `./checkpoint.sh <label>` once the bars are
  approved and `reset.sh last` will come back approved, which saves re-approving on every re-run.)

  In `bp tui`, approve all three `EX-2` bars. Approving the last one opens the play prompt — answer
  `y`, and three queue rows appear. `just install`, then run `bp serve daemon -v` with **no Claude
  session open in `tools/example`** and leave it.

  Each bar's session comes up, commits, and exits before the next one starts. While it runs,
  `claude agents --json` never shows two rows under `tools/example` at once — that is BIT-39.8's
  guard doing its job on a real queue rather than a fake runner. When the queue is empty,
  `./check.sh EX-2` shows all three bars `done` and
  `git log --oneline worktree-EX-2-shout-dispatch-drain-workload` shows three commits in bar order:
  the shouted greeting, then the contradiction that forces real upper-casing, then the trim. The
  track built itself.

## Commit (user)
`test(daemon): pin the one-bar-per-tick drain across a whole track`