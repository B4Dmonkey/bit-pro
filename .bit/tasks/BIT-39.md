---
id: BIT-39
title: Dispatch — the daemon works queued bars unattended
status: todo
approved: true
---
## Why
Every part of the automation phase is built except the part that does work. BIT-28 gave us a
daemon, BIT-29 a project registry, BIT-31 a play prompt, BIT-32 counts, BIT-33 a queue — and an
operator who approves a track, answers "yes", and walks away comes back to find the bars exactly
where they left them. The queue has no consumer, which is why `clear-queue.sh` had to be written:
nothing removes a row. Until the loop dispatches, five shipped tracks add up to plumbing an
operator can watch but not use.

## Summary
Fill in the daemon's tick loop. Each tick, for every registered project, it holds if that project
already has a live Claude session; otherwise it pops the head queue row, re-reads the ledger,
spawns a background session running `/bit:do <BAR>` as `bit:bot-dev` in the track's worktree,
confirms the session came up, and deletes the row. One bar in flight per project means a track's
bars run in order. The loop body moves out of `cmd/` first, so the rest lands somewhere sane.

## Visual aid
```
tick
  └─ for each registered project
       ├─ live session for this project?   (claude agents --json, filtered by cwd)
       │      yes ─────────────────────────────────────► hold, next project
       └─ no
            └─ pop smallest queue.id for project_id
                 ├─ ledger: bar done, or not approved ──► delete row, log, next tick
                 └─ claude --bg --agent bit:bot-dev
                          -w <TRACK>-<slug>  -n <TRACK>-<slug>
                          '/bit:do <BAR>'
                      └─ session now in claude agents --json?
                            yes ────────────────────────► delete row
                            no  ────────────────────────► leave row, log, retry next tick
```

## Decisions

- **A dispatched session does not park on a folder-trust prompt.** Measured 2026-08-24 on
  2.1.241: `claude --bg -w bit-probe-39 -n bit-probe-39 '<write a file>'` in this repo went
  `state: working` → `state: done` and wrote the file, with `waitingFor` never set. Trust
  resolves against the already-trusted repo root, not the fresh
  `.claude/worktrees/bit-probe-39` path — which confirms what the 2026-08-24 config finding
  suggested and supersedes the 2026-08-21 probes, whose parking was the throwaway repo being
  untrusted, not the worktree being new. Dispatch into an enrolled project needs no trust
  handling; a project the operator has never opened is still expected to park.
- **Claude creates the worktree, not `bp`.** Dispatch passes `-w <name>` and lets Claude Code
  make it. Reverses the standing "`bp` creates the worktree" call, which invited exactly this
  recheck: there is no git-worktree code in the repo today, and `bp` creating one buys only an
  unprefixed branch name. The cost accepted is the forced `worktree-<name>` branch prefix and a
  worktree Claude locks and owns.
- **The worktree name is derived at dispatch, never stored.** `<track-id>-<kebab-cased track
  title>`, from the bar's parent track. Nothing new is recorded on a track and no skill changes,
  which is what makes this cheaper than asking bit_scope for a name at scope time.
- **One identifier for the worktree and the session.** The same derived string goes to `-w` and
  to `-n`, so a `claude agents --json` row is directly attributable to a track.
- **Bars of a track share one worktree.** The name is per-track, so bar 2 lands on bar 1's tree
  rather than restarting from `main`.
- **The loop never tears down a worktree.** After the last bar the tree holds committed work on
  `worktree-<name>`, and `claude rm` deletes the branch along with it. Reaping is the operator's,
  after they have merged. Measured 2026-08-24: `claude rm` refuses outright while the worktree
  has uncommitted changes — a real backstop, but not one to rely on, since a bot-dev session
  commits its work and so leaves the tree clean. `bp task complete` removing the worktree is now
  wrong under the Claude-owned route and is left alone here.
- **Dequeue on dispatch, after confirming the session exists.** The row is deleted once the
  spawned session appears in `claude agents --json` under its `-n` name — not on completion, and
  not by parsing the `backgrounded · <id> · <name>` line the spawn prints. Polling is the same
  surface the in-flight guard already needs, it needs no TTY (which the launchd-hosted daemon
  does not have), and it does not depend on a human-readable line staying stable.
- **A spawn that cannot be confirmed leaves its row in place.** The loop logs and retries on the
  next tick rather than dropping the bar.
- **Liveness is read from `state`, tolerating its absence.** Re-confirmed 2026-08-24: background
  rows carry `state`, interactive rows carry neither `state` nor `id`. A row counts as live only
  when its `state` is in `{working, blocked}`; a `done` row lingers in the registry and must not
  count as live.
- **A live row is matched to a project by `cwd`, not by name.** `claude agents --json` is
  machine-wide, and the per-project guard runs *before* a row is popped — so there is no track
  yet whose dispatch name could be matched against. The loop cuts each row's `cwd` back to its
  owning checkout with `bitdir.Canonical` — the BIT-27 rule that already resolves a path inside
  `.claude/worktrees/` to the main checkout's `.bit/` — and compares that to the registered
  project path. The derived `<track>-<slug>` name identifies *which* track a live row belongs
  to; it does not answer whether the project is busy. Matching on a project-code name prefix was
  rejected: it assumes codes are unique across the registry and answers the narrower question.
- **A session that ends without landing its bar is not the loop's problem.** Ownership passes to
  Claude at spawn. The loop does not stall, retry, or flag it; the slot frees and the next row
  goes. Accepted consequence: bar N+1 can start on a tree bar N left half-finished, and the
  operator finds out by reading the ledger.
- **The ledger is checked before every spawn.** A popped row whose bar is already `done`, or is
  not approved, is deleted with a log line rather than dispatched. Both states are reachable —
  approval is revoked by a replan — and leaving the row would block that project's queue head
  forever. Re-queueing is one keypress.
- **The loop body belongs in `daemon/`.** `cmd/serve.go` keeps flag parsing and wiring only;
  `writeCounts` and the tick loop move to `daemon`, beside `Start`/`Stop`/`Status`, which
  supervise the same process.
- **The `claude` binary calls go in `claude/`, not `daemon/`.** `claude/sync.go` already shells
  out to `claude` there; spawn and `agents --json` join it, which keeps `daemon` free of
  external-tool plumbing.
- **They do not share the existing `Runner` seam — `claude/` carries two runner shapes.**
  `claude.Runner` returns only an `error` and discards stdout. That is right for its three
  `plugin` callers, where pass/fail is the whole answer, and useless for `agents --json`, whose
  output *is* the answer. So a second, output-returning shape lands beside it —
  `(out string, code int, err error)`, the same shape `daemon.Runner` already uses. Widening
  `claude.Runner` instead was rejected: it churns six files (`cmd/root.go`, `cmd/add.go`,
  `cmd/init.go`, `cmd/cmd_test.go`, `claude/sync.go`, `claude/sync_test.go`) so that every
  existing caller can discard two return values it has no use for.
- **No new package.** `dispatch/` was considered and rejected — it would split "the daemon"
  across two names for one process's job.
- **Queue rows are bar rows, full stop.** `target_typ` exists but the TUI only ever writes
  `"bar"` (`tui/model.go:325,374`); a track is expanded into its approved, not-done bars at
  enqueue time. Dispatch handles bar rows only and does not implement BIT-33's dual-typed prose.
  This also means a mid-flight replan cannot surprise the loop: the bars it will work were fixed
  when the operator queued them.
- **The loop only. No new operator surface.** No `bp approve` deny rule, no `bp queue rm`, no
  TUI un-enqueue, and `clear-queue.sh` stays for now. Each is a follow-up track. The deny rule in
  particular is a real safety gap — `bit:bot-dev` forbids `bp approve` in prose and nothing
  enforces it — and it is deliberately left open rather than folded in here.
- **Permission prompts stay.** No permission mode that suppresses them. A session pausing for the
  operator is the safety mechanism that replaces the push gate.

## Verses

- [ ] Verse 1 — The loop lives where the rest can be built: `writeCounts` and the tick loop move
  out of `cmd/serve.go` into `daemon/`, beside `Start`/`Stop`/`Status`; the command keeps flag
  parsing and wiring. Nothing changes for the operator — `bp serve daemon -v` still ticks and
  `bp list` still shows refreshed counts — and every verse after this adds to a package instead
  of to a command body.
  Touches: `cmd/serve.go`, `daemon/`.

- [ ] Verse 2 — A queued bar runs unattended: with the daemon running in a terminal, an operator
  queues one approved bar, walks away, and comes back to a commit on the track's worktree branch
  and an empty queue row. Includes the guard that stops the loop dispatching a second bar while
  that project has a live session — without it the next tick stampedes the rest of the queue.
  Touches: `daemon/` (pop, spawn, confirm, dequeue), `claude/` (spawn and `agents --json`),
  `db/queries/queue.sql` (the first delete query).

- [ ] Verse 3 — A whole approved track runs bar-by-bar: several queued bars drain in FIFO order,
  one at a time, each session landing in the same per-track worktree so bar 3 builds on bar 1.
  An operator approves a three-bar track, answers "yes" at the play prompt, and the track is done
  without them touching it.
  Touches: `daemon/` (FIFO drain, ledger check), `claude/` (worktree name derivation).

- [ ] Verse 4 — It works with the terminal closed: the same cycle runs under the launchd-hosted
  daemon, so `bp start` and closing the terminal is enough. Agents inherit almost no environment,
  so the plist has to name whatever the spawned sessions need — the `claude` binary's absolute
  path above all — and the loop cannot assume a TTY.
  Touches: `daemon/plist.go` (`EnvironmentVariables`, `WorkingDirectory`), `daemon/` loop.

## References

- `automation-notes.md` — the working notes this scope is step 6 of. Its Decisions, Measured
  facts, and Open gaps sections are the source for everything above; the 2026-08-21 and
  2026-08-24 spawn-surface measurements inform all four verses.