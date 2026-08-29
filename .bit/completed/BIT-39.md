---
id: BIT-39
title: Dispatch — the daemon works queued bars unattended
status: done
order:
    - BIT-39.1
    - BIT-39.2
    - BIT-39.3
    - BIT-39.4
    - BIT-39.5
    - BIT-39.6
    - BIT-39.7
    - BIT-39.8
    - BIT-39.9
    - BIT-39.15
    - BIT-39.16
    - BIT-39.17
    - BIT-39.10
    - BIT-39.11
    - BIT-39.12
    - BIT-39.18
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
  handling. Widened 2026-08-25 on 2.1.245: a `--bg` session spawned into `tools/temp` — a
  directory that is neither a git repo nor one the operator had ever opened — also went straight
  to `state: working` with `waitingFor: null`. So the residual "a project the operator has never
  opened still parks" caveat does not hold as stated; what made the 2026-08-21 scratch repo park
  is still unexplained, and it is not simply that the directory was new.
- **Claude creates the worktree, not `bp`.** Dispatch passes `-w <name>` and lets Claude Code
  make it. Reverses the standing "`bp` creates the worktree" call, which invited exactly this
  recheck: there is no git-worktree code in the repo today, and `bp` creating one buys only an
  unprefixed branch name. The cost accepted is the forced `worktree-<name>` branch prefix and a
  worktree Claude locks and owns.
- **The worktree name is derived at dispatch, never stored.** `<track-id>-<kebab-cased track
  title>`, from the bar's parent track. Nothing new is recorded on a track and no skill changes,
  which is what makes this cheaper than asking bit_scope for a name at scope time.
- **The exact kebab-casing rule is deliberately left loose.** Lowercase the title and collapse each
  run of non-alphanumerics to a single `-`, keeping the track ID's case verbatim. Long names, odd
  punctuation, and collisions are accepted risks, not blockers — if one breaks a branch name in
  practice it gets fixed then. Settled this way on purpose: blocking the plan on a naming rule
  nobody has been bitten by yet costs more than the breakage would.
- **One identifier for the worktree and the session.** The same derived string goes to `-w` and
  to `-n`, so a `claude agents --json` row is directly attributable to a track.
- **Bars of a track share one worktree.** The name is per-track, so bar 2 lands on bar 1's tree
  rather than restarting from `main`. Measured cost, 2026-08-26: sharing means bar N+1 inherits
  whatever bar N left, including wreckage. A session that dies at startup still creates the
  worktree, branches it, and `git worktree lock`s it before it fails — so every later bar of that
  track lands on a tree locked by a dead pid, and `git worktree remove` refuses until it is
  unlocked by hand. The decision stands; it is not free, and BIT-39.13 prices it.
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
- **Liveness is presence in `claude agents --json`; `state` is not read at all.** Reverses the
  2026-08-24 call that filtered on `state` in `{working, blocked}`. Measured 2026-08-25 on 2.1.245:
  a background row observed at `state: done` read `state: blocked` ten minutes later — same session,
  same name — and `claude agents --help` documents `--all` as "also include completed background
  sessions", so a row present *without* `--all` is by definition not a completed one. `state: done`
  is a transient between-turns value, which makes the old filter actively wrong: it would have read
  a mid-flight session as free and dispatched a second bar on top of it. Presence is both simpler
  and safer, and interactive rows — which carry no `state` at all — need no special case under it.
  Accepted consequence: an operator's own interactive session in a project holds that project's
  queue, so the daemon and a human cannot work the same project at once.
- **A live row is matched to a project by `cwd`, not by name.** `claude agents --json` is
  machine-wide, and the per-project guard runs *before* a row is popped — so there is no track
  yet whose dispatch name could be matched against. A row belongs to a project when its `cwd` is
  **at or beneath the registered project path**, compared by path segment so `/p/foo` never
  matches `/p/foobar`. One rule covers all three shapes a real row takes: the checkout itself, a
  dispatched session in `<repo>/.claude/worktrees/<name>`, and an interactive session started in
  a subdirectory like `<repo>/cmd`. The derived `<track>-<slug>` name identifies *which* track a
  live row belongs to; it does not answer whether the project is busy. Matching on a
  project-code name prefix was rejected: it assumes codes are unique across the registry and
  answers the narrower question.
- **The guard uses neither `bitdir.Canonical` nor `claude agents --cwd`.** This reverses the
  earlier call to canonicalize each row's `cwd`. `bitdir.Canonical` returns the literal string
  `".bit"` for any path outside `.claude/worktrees/`, so every non-worktree project would compare
  equal — one live session anywhere would mark every registered project busy and nothing would
  ever dispatch. Filtering with `claude agents --json --cwd <repo>` was rejected too: it is
  measured to return interactive rows as well as background ones, but whether it reaches into
  `.claude/worktrees/` is unmeasured, and the at-or-beneath rule settles that question instead of
  resting the guard on it.
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
- **The daemon ticks every 5 seconds, not 10.** Halves the interval inherited from BIT-28.
  Deliberately a step rather than a final value: the intent is to keep cutting it — 1s or less is
  plausible — with each cut gated on watching what a faster loop costs, since from Verse 2 every
  tick shells out to `claude agents --json` once per registered project. 5s is what BIT-39 ships
  and the next cut is a later call. It stays a hardcoded `cmd` package var under the "no new
  operator surface" Decision — no flag, no config.
- **The `claude` binary calls go in `claude/`, not `daemon/`.** `claude/sync.go` already shells
  out to `claude` there; spawn and `agents --json` join it, which keeps `daemon` free of
  external-tool plumbing.
- **A spawn's working directory comes from the calling process, not a flag.** Measured 2026-08-25
  on 2.1.245: `claude` has no cwd/`-C` flag at all — `--cwd` exists only on the `claude agents`
  subcommand, as a listing filter. `cd "$DIR" && claude --bg -n "$NAME" '<prompt>'` produced a
  registry row reading `"cwd": "<DIR>"` and wrote its file there. In Go that is `exec.Cmd.Dir`.
- **They do not share the existing `Runner` seam — `claude/` carries two runner shapes.**
  `claude.Runner` returns only an `error`, discards stdout, and runs wherever the daemon happens
  to be. That is right for its three `plugin` callers, where pass/fail is the whole answer, and
  wrong for both new callers: `agents --json`'s output *is* the answer, and a spawn has to start
  in the target project. So a second shape lands beside it — one that takes a **working
  directory** and returns `(out string, code int, err error)`. This reverses the earlier call
  that the new shape would be "the same shape `daemon.Runner` already uses": that signature
  carries no directory, so it cannot dispatch. Widening `claude.Runner` instead was rejected: it
  churns six files (`cmd/root.go`, `cmd/add.go`, `cmd/init.go`, `cmd/cmd_test.go`,
  `claude/sync.go`, `claude/sync_test.go`) so that every existing caller can discard a parameter
  and two return values it has no use for.
- **`bp start` resolves `claude` to an absolute path; the daemon never relies on a `PATH`.**
  Reverses "the daemon finds `claude` on the `PATH`, so the plist gets no `EnvironmentVariables`",
  whose accepted risk was measured and then hit. Measured 2026-08-27: `bp start` with one approved
  bar queued logged `listing live sessions … exec: "claude": executable file not found in $PATH`
  on every tick and dispatched nothing. `launchctl getenv PATH` is empty, so a LaunchAgent gets
  launchd's default `/usr/bin:/bin:/usr/sbin:/sbin`, while `claude` is `~/.local/bin/claude`.
  `bp start` is the one place that can resolve this, because it runs in the operator's shell: it
  looks the binary up there, fails loudly if it cannot find it, and the resolved path travels to
  the daemon in the plist `bp start` already writes. Chosen over giving the plist an
  `EnvironmentVariables` `PATH` because it puts the failure at `bp start`, where an operator is
  watching, rather than at tick time in a log nobody reads. Accepted cost: the path is pinned at
  enrollment, so moving or reinstalling `claude` needs another `bp start`. `WorkingDirectory`
  stays unset — nothing in the loop reads a relative path.
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
  operator is the safety mechanism that replaces the push gate. Measured 2026-08-27, and it is the
  mechanism working rather than a defect: a dispatched `bit:bot-dev` session took `EX-2.1` to
  green, marked the bar `done`, and then parked at `state: blocked` on `Bash(git add *)` — an
  `ask` rule in the operator's own settings — without committing. Dispatch's job ends at handing
  the bar to Claude; answering the prompt and carrying that session on is the operator's.

- **The plist carries the resolved path as an `EnvironmentVariables` entry, not a flag.** `bp start`
  writes `BP_CLAUDE=<absolute path>` into the plist it already writes; the daemon reads it and falls
  back to a bare `claude` when it is unset, so a foreground `bp serve daemon` in a terminal is
  unchanged. Chosen over adding a `--claude` flag to `serve daemon`, which would breach the "no new
  operator surface" Decision that already forbids a flag for the tick interval. Nothing is
  hardcoded — `exec.LookPath` reads whatever `PATH` the operator actually has — so this works on any
  machine, at the already-accepted cost of re-running `bp start` if `claude` moves.
- **The hold message logs at `INFO` and names the session holding the project.** `DEBUG` is wrong
  for it: an operator running `bp serve daemon -v` to find out why nothing dispatches is the exact
  audience, and under launchd a `DEBUG` record may not be written at all. The record carries the
  matched session's name and `cwd` alongside `project`, and states the consequence — not
  dispatching — rather than the condition, so the operator can tell which session to close.
- **A project's slot frees when its session is gone, and clearing it is the operator's job.** The
  guard already implements this and needs no change: hold while any live row's `cwd` is at or
  beneath the project path, dispatch again once none is. So the operator dispatches a bar, works
  with that session in Claude — answering prompts, reviewing, whatever it needs — and deletes it
  when they are finished, which is what lets the next bar go. Explicitly rejected reading the bar's
  ledger status as the release signal: what `done` should mean for a dispatched session is
  unexplored, and guessing it here would dispatch bar 2 on top of work that is not actually
  finished. That exploration is a future track; until then the release is manual on purpose.
- **The queue row is not deleted at spawn time.** The row and the spawn are held across ticks, and
  the dequeue decision is made on a later tick once the session's fate is legible. Reverses
  "Dequeue on dispatch, after confirming the session exists": measured 2026-08-27 on 2.1.250, a good
  spawn and a failed one both exit 0, and the sub-second confirm poll catches a doomed process while
  it is still registering — which is what deleted `EX-2.1`'s row for a bar that never ran. The row
  is held instead, and the dequeue decision is made on a later tick.
- **A dispatch is confirmed by presence in the plain listing; `done` and `idle` both confirm
  it.** Settled 2026-08-27. A session that has not failed is a good dispatch whatever it is
  currently doing, so a bar mid-turn, a bar parked on a permission prompt, and a bar already
  finished all count and all dequeue. This needs no `state` read and so leaves the "`state` is
  not read at all" Decision intact, because the plain listing already excludes failures on its
  own — see the next Decision. Rejected the earlier "dequeue once the session is gone" reading:
  measured the same day, a finished session never goes away, so that rule would never fire.
- **Presence in the plain listing is the honest success signal; `--all` is not used.** Measured
  2026-08-27 on 2.1.250: a session that ran and finished stayed in `claude agents --json` at
  `state: done` / `status: idle` across 24 polls over two minutes, while a session that died at
  startup was absent from that listing and appeared only under `claude agents --all --json` with
  `state: failed`. The existing name check is therefore the right check — only its timing was wrong.
- **An unknown `--agent` no longer fails; it silently substitutes the default.** Measured 2026-08-27
  on 2.1.250: `claude --bg --agent 'definitely-not-an-agent' …` printed `warning: no agent named
  '…' — spawning with default template`, exited 0, and spawned a session. This supersedes the
  2026-08-26 observation that such a spawn errored and produced nothing, and it means a project
  where the plugin does not resolve runs `/bit:do` under the default agent instead of `bit:bot-dev`,
  so it will not commit and will not honour the approval gate. What no longer follows is the
  conclusion this Decision originally drew — that fixing `bp add` is therefore a correctness matter
  for dispatch. BIT-41 measured that cause away; see the next Decision.
- **`bp add`'s readiness defects are out of scope; a follow-up track owns them.** Verse 5 originally
  carried them because an unresolvable `--agent` silently ran `/bit:do` as the default agent, which
  made a broken `bp add` a correctness matter for dispatch. BIT-41 measured that cause away (see
  References): it was upstream bug anthropics/claude-code#27257, a project-scope install in one
  project making the same plugin uninstallable at project scope in another, fixed in Claude Code
  2.1.248 and confirmed resolving `bit:bot-dev` in `tools/example`. Three real defects remain, and
  they are recorded here so the follow-up track does not re-derive them: `cmd/add.go:45-48`
  short-circuits on an already-enrolled project, so re-running `bp add .` to repair a project does
  nothing; `cmd/add.go:66-70` gates the wiring on `.bit/` being absent, so a project that ran
  `bp init` first never gets the wiring at all; and `claude.SyncPlugin`/`claude.RegisterMCP` take
  `claude.Runner`, which carries no working directory, so `claude plugin install --scope project`
  and `claude mcp add` resolve against wherever the operator was standing rather than the path
  `bp add` was given — with `SyncPlugin` trying `plugin update` before `install`, so a succeeding
  update swallows the install. None of the three blocks a working daemon, which is what BIT-39 is
  for.

- **Deferred with the above, and kept rather than deleted: `bp add`'s directory problem is fixed
  with the existing `DirRunner`, not by widening `Runner`.** `claude plugin install --scope project`
  and `claude mcp add` both resolve against the working directory and `claude.Runner` carries none,
  so those two call sites move to `claude.DirRunner`, which already takes one — two call sites
  instead of the six files widening `claude.Runner` would churn (`cmd/root.go`, `cmd/add.go`,
  `cmd/init.go`, `cmd/cmd_test.go`, `claude/sync.go`, `claude/sync_test.go`). One end is left open
  for that track rather than settled here: `claude.Runner` has no consumer other than those two
  functions, so moving both to `DirRunner` leaves it dead — which is in tension with the
  two-runner-shapes Decision above, and is a call for whoever plans it.
- **A stale locked worktree gets no work of its own.** Dropped 2026-08-27 (was BIT-39.13). Every
  instance observed came from a spawn that died at startup, and that cause is `bp add` leaving the
  plugin unresolvable in the project. Once a failed spawn is loud and keeps its row, a poisoned tree
  shows up as a bar that visibly refuses to start; clearing the leftover is `git worktree unlock &&
  git worktree remove`, which the fixture's `reset.sh` already does. Inheriting a broken tree is in
  any case already an accepted cost under "Bars of a track share one worktree."

- **The spawn record says `dispatching`; `dispatched` is logged only once the session is
  confirmed.** The spawn and the dequeue now land on different ticks, so one record cannot honestly
  mean both. `dispatching` carries the bar, the worktree name, and the spawn's captured output,
  which `claude.Spawn` currently discards — that output is the only place a failure diagnostic can
  reach the log, since a spawn that produced no session still exits 0 and neither `--all` nor
  parsing the `backgrounded ·` line is available. `dispatched` is logged on the later tick that
  finds the session in the plain listing, beside the dequeue. Consequence, and the point of the
  wording: a project where every spawn dies logs `dispatching` forever and never `dispatched`, with
  the diagnostic in the record.

- **The loop keeps no state between ticks, so an unconfirmed row is retried, not diagnosed.** A row
  with no matching session is either a bar never dispatched or a bar whose session died at startup,
  and nothing available to the loop tells the two apart. Rather than add a marker to the queue row,
  the loop re-spawns — which the existing "a spawn that cannot be confirmed leaves its row in place"
  Decision already prescribes. Accepted cost: a permanently broken project re-spawns every tick.
  That is the loud signal, because every one of those ticks logs `dispatching` with claude's own
  diagnostic attached.

## Verses

- [x] Verse 1 — The loop lives where the rest can be built: `writeCounts` and the tick loop move
  out of `cmd/serve.go` into `daemon/`, beside `Start`/`Stop`/`Status`; the command keeps flag
  parsing and wiring, and the tick drops to 5s. The move itself changes nothing for the operator —
  `bp serve daemon -v` still ticks and `bp list` still shows refreshed counts — the one visible
  difference is the faster cadence, and every verse after this adds to a package instead of to a
  command body.
  Touches: `cmd/serve.go`, `daemon/`.

- [x] Verse 2 — A queued bar runs unattended: with the daemon running in a terminal, an operator
  queues one approved bar, walks away, and comes back to a commit on the track's worktree branch
  and an empty queue row. Includes the guard that stops the loop dispatching a second bar while
  that project has a live session — without it the next tick stampedes the rest of the queue.
  Touches: `daemon/` (pop, spawn, confirm, dequeue), `claude/` (spawn and `agents --json`),
  `db/queries/queue.sql` (the first delete query).

- [x] Verse 3 — A track's bars run in order, one at a time: queued bars drain FIFO, never more than
  one in flight per project, each session landing in the same per-track worktree so bar 3 builds on
  bar 1. An operator approves a three-bar track and answers "yes" at the play prompt; each bar then
  runs unattended, and deleting its session is what releases the slot for the next. A fully
  unattended multi-bar drain is deliberately not this verse — see the slot Decision.
  Touches: `daemon/` (FIFO drain, ledger check), `claude/` (worktree name derivation).

- [ ] Verse 4 — It works with the terminal closed: the same cycle runs under the launchd-hosted
  daemon, so `bp start` and closing the terminal is enough. The TTY worry was already covered —
  `claude agents --json` needs none by its own documentation, and the daemon's logger already
  switches to JSON when stdout is not a terminal. The environment worry was not: the `PATH`
  assumption failed on first contact and nothing dispatched (see Decisions), so this verse now
  carries code as well as an observation — `bp start` resolves `claude` where it can still see the
  operator's shell and hands the daemon a binary it can actually execute. What remains
  unautomatable is the observation itself: that the cycle survives losing its terminal. The cycle
  observed here is **one bar** — with the terminal closed nobody is deleting sessions, so the slot
  never frees and a second bar is not expected to go.
  Touches: `cmd/start.go` (resolve, fail loudly), `daemon/plist.go` (carry the resolved path),
  `claude/` (invoke it instead of a bare `claude`).

- [ ] Verse 5 — The operator can trust what the loop tells them: the two daemon defects found while
  verifying Verses 1–3 are closed, so a held project says which session holds it and why, and a
  spawn that produced no working session is loud and keeps its row instead of silently dropping a
  bar. `bp add`'s readiness defects were the third and are now out of scope — see Decisions.
  Touches: `daemon/loop.go` (the guard's log record, the hold-and-dequeue sequence),
  `claude/dispatch.go` (the discarded spawn output).

## References

- `probe-dispatch.sh` — the bash spelling of the spawn-and-confirm surface, ten lines: `cd` into
  the project, `claude --bg -n <name> '<prompt>'`, then confirm the row in `claude agents --json`.
  The Go in Verses 2–3 is this, with `exec.Cmd.Dir` doing the `cd`.
- `BIT-41` track body, `## Decisions` — "Versioning was never what blocked BIT-39." The measurement
  that took `bp add` out of Verse 5: the agent-resolution failure was upstream bug
  anthropics/claude-code#27257, fixed in Claude Code 2.1.248 and confirmed resolving `bit:bot-dev`
  in `tools/example`. It settles the *cause* only — `bp add`'s own three defects are untouched by it.
- `automation-notes.md` — the working notes this scope is step 6 of. Its Decisions, Measured
  facts, and Open gaps sections are the source for everything above; the 2026-08-21 and
  2026-08-24 spawn-surface measurements inform all four verses.