---
id: BIT-39.13
title: A failed session leaves a locked worktree behind
status: todo
---
## Collected, not planned

Found 2026-08-26 alongside BIT-39.12, during the track's own verification in `tools/example`.
**This bar is not planned and not approved** — the Method below is a sketch, not settled detail.
It needs a `/bit:plan` pass before anyone runs it.

It carries no verse. It prices a cost the **"Bars of a track share one worktree"** Decision has
always had and that nothing had measured until now.

## The problem

A session that dies at startup still leaves a worktree behind, created, branched, and locked.
Proven with a single run that did nothing but print `--agent 'bit:bot-dev' not found` — the
worktree below did not exist before it and did after:

```
.claude/worktrees/EX-2-shout-dispatch-drain-workload  a64586d [worktree-EX-2-...] locked
```

The lock names its creator, and that process is already gone:

```
locked claude session EX-2-shout-dispatch-drain-workload (pid 25988 start Wed Aug 26 11:20:25 2026)
$ ps -p 25988
(no such process)
```

So the worktree is created and locked **before** the agent is resolved, and the lock is not
released when startup fails. Two consequences:

- **`git worktree remove` refuses** until someone runs `git worktree unlock` by hand. The fixture's
  `reset.sh` already unlocks before removing, which is why it recovers and a bare `remove` does
  not.
- **The wreckage is inherited.** The name is per-track, so every later bar of that track lands on
  the same locked tree. One failed bar poisons the rest of the track.

Worth being precise about the blame here: leaving a stale lock is Claude Code's behaviour, not
`bp`'s, and this track's **"The loop never tears down a worktree"** Decision deliberately keeps
the daemon out of worktree teardown. So the fix is probably not "the loop cleans up" — that
reverses a standing decision. It is more likely detection and a legible log line.

## Scope
- `daemon/loop.go` — `worktreeFor` and the dispatch path around it.
- `daemon/loop_test.go`.

## Method (sketch — not settled)
- [ ] Open question for planning, and the one that decides the bar's size: does the loop detect a
      stale locked worktree before spawning, or does it only report the failure after the fact?
      Detection means git knowledge in `daemon/`, which the track has so far avoided entirely.
- [ ] Cheapest option to price first: do nothing structural, and let BIT-39.12's spawn-failure
      reporting carry it. If a failed spawn is loud and keeps its row, a poisoned worktree shows up
      as a bar that visibly refuses to start rather than as silence.
- [ ] If detection does land, `git worktree list --porcelain` already reports `locked` and its
      reason string, so no new mechanism is needed to see it.
- [ ] Do not add teardown to the loop without revisiting **"The loop never tears down a
      worktree"** — that is a scope change, not a bar.

## Claude verifies
- [ ] `just test`, `just lint`.

## User verifies
- [ ] Leave a locked worktree from a failed run in place, dispatch the track's next bar, and
      confirm the log says why it will not start rather than going quiet.

## Commit (user)
`fix(daemon): a stale locked worktree is reported, not ignored`