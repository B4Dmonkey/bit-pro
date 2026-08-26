---
id: BIT-39
title: Dispatch — the daemon works queued bars unattended
status: doing
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
- **The daemon finds `claude` on the `PATH`; nothing names its absolute path.** So the plist gets no
  `EnvironmentVariables`, and no `WorkingDirectory` either — nothing in the loop reads a relative
  path. The accepted risk is measured and specific: `claude` is `~/.local/bin/claude` on this
  machine and a LaunchAgent does not inherit the operator's login `PATH`. Verse 4 is where this
  either holds or fails with `claude: executable file not found in $PATH`; that failure would buy
  the Decision, whereas guessing at an absolute path now does not.
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

- [x] Verse 3 — A whole approved track runs bar-by-bar: several queued bars drain in FIFO order,
  one at a time, each session landing in the same per-track worktree so bar 3 builds on bar 1.
  An operator approves a three-bar track, answers "yes" at the play prompt, and the track is done
  without them touching it.
  Touches: `daemon/` (FIFO drain, ledger check), `claude/` (worktree name derivation).

- [ ] Verse 4 — It works with the terminal closed: the same cycle runs under the launchd-hosted
  daemon, so `bp start` and closing the terminal is enough. No code is expected. The TTY worry is
  already covered — `claude agents --json` needs none by its own documentation, and the daemon's
  logger already switches to JSON when stdout is not a terminal — and the environment worry was
  decided away above. What is left is the observation that the unattended cycle survives losing its
  terminal, which no test can make.
  Touches: nothing expected; `daemon/plist.go` only if the `PATH` assumption fails.

## References

- `probe-dispatch.sh` — the bash spelling of the spawn-and-confirm surface, ten lines: `cd` into
  the project, `claude --bg -n <name> '<prompt>'`, then confirm the row in `claude agents --json`.
  The Go in Verses 2–3 is this, with `exec.Cmd.Dir` doing the `cd`.
- `automation-notes.md` — the working notes this scope is step 6 of. Its Decisions, Measured
  facts, and Open gaps sections are the source for everything above; the 2026-08-21 and
  2026-08-24 spawn-surface measurements inform all four verses.