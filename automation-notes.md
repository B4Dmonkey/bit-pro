# Automation phase — working notes

**Last synced 2026-08-29. Route: daemon. Dispatch is shipped.**

A long-running `bp serve daemon` process watches registered projects and dispatches queued bars.
This replaced the earlier chaining design (a `Stop` hook re-invoking `bp dispatch` after every bar);
BIT-25 sits in `.bit/completed/` at `status: todo` — it is the chaining version and should not be
planned as written.

Every "Measured facts" entry was run and observed on this machine. Where a doc and a measurement
disagreed, the measurement won. Version is noted per entry, because several reversed each other.

---

## Checklist

| # | Step | Track | State |
|---|---|---|---|
| 1 | Daemon lifecycle — `bp serve`, `bp start/stop/status` over a LaunchAgent | BIT-28 | done |
| 2 | Project registry — SQLite, `projects` table, `bp add`, `bp list` | BIT-29 | done |
| 3 | TUI play prompt — modal on approving the last unapproved bar | BIT-31 | done |
| 4 | Project counts — four buckets, refreshed each tick | BIT-32 | done |
| 5 | Queue — `queue` table, enqueue from popup and shortcut, cyan rows | BIT-33 | done |
| 6 | **Dispatch** — the loop pops, spawns, confirms, dequeues | **BIT-39** | **done** |

All six are filed in `.bit/completed/`; read the tracks there for the full decision lists. What
step 6 inherited and what it changed:

- **Pop contract.** `queue` is `id INTEGER PRIMARY KEY`, `project_id` (FK), `target_id`,
  `target_typ` (`track` | `bar`). FIFO within a project — smallest `id` wins. In practice the TUI
  only ever writes `bar` rows; a track is expanded to its approved, not-done bars at enqueue time.
- **`clear-queue.sh` can go.** It was a throwaway for the era when nothing dequeued. Dispatch owns
  dequeuing now.
- **Four count buckets** — backlog (unapproved), todo (approved, not done), done, completed. The
  daemon is the sole writer and `bp list` reads the cache, so **counts lag by design**.
- **The `--mcp-config` idea was cut** on 2026-08-24 and has not come back. A dispatched session runs
  `/bit:do` out of the **plugin** and reaches `bp` through Bash. The MCP surface changes what the
  skills call internally, not who starts the session. Revisit only after `mcp-notes.md` steps 4–5.

---

## Next steps

Written 2026-08-29, straight after BIT-39 signed off. Not researched — these are the notes to pick
up from.

### 1. How do we tell a session is finished?

This is the open question. Right now the loop knows a session **exists**; it has no notion of one
being **done with its bar**. Presence in `claude agents --json` is the liveness test, and a finished
session never leaves that listing — it sits at `state: done` / `status: idle` indefinitely. So the
project's slot is held by a session that has nothing left to do.

Candidate signals, none evaluated: the bar's own ledger status flipping to `done`; a commit landing
on the worktree branch; `state`/`status` combinations (unreliable — see the transience measurement);
a `Stop` hook writing a marker. BIT-39 deliberately punted on this — "what `done` should mean for a
dispatched session is unexplored, and guessing it would dispatch bar 2 on top of unfinished work."

### 2. Dispatched agents can't be deleted, and that blocks the queue

Observed while running BIT-39, not yet diagnosed. The slot only frees when the session is gone, so a
session that refuses to be removed wedges that project's queue permanently. Worth pinning down which
command is actually failing and why, because the two available ones differ:

- `claude stop <id>` — keeps the conversation and the worktree. Unmeasured whether it removes the
  row from the plain `--json` listing, which is what would actually free the slot.
- `claude rm <id>` — deletes the session **and its worktree, and the branch with it**. That destroys
  the shared per-track tree, so bar N+1 would restart from `main` and bar N's commit would be gone.

If neither frees the slot without destroying work, the release mechanism itself needs redesigning —
which is really the same question as (1).

### 3. A dispatched session cannot commit unattended on this machine

Measured 2026-08-29 (see Measured facts). The operator's `ask` rules on `Bash(git add *)` and
`Bash(git commit *)` mean a `bit:bot-dev` session takes its bar to `done` and then parks. Not a
defect — it is the "permission prompts stay" Decision working — but it means the unattended cycle
stops one step short of a commit, and Verse 4's "one commit with the terminal closed" cannot happen
as written.

The only route that clears it without weakening the operator's own settings is passing
`--setting-sources project,local --allowedTools "Bash(git add *)" "Bash(git commit *)"` at spawn,
which costs the session all user-level config (model, effort, marketplaces) unless each is
re-supplied. That is a real capability worth its own scope, not a patch.

### 4. Small, already-known

- **Enforce "Claude never approves."** An unattended session can type `bp approve` and clear its own
  gate. `bit:bot-dev` forbids it in prose; nothing enforces it. A `Bash(bp approve:*)` deny rule in
  `.claude/settings.json` gets the property — `bp init` owns that file (`cmd/init.go:48`), but
  `claude.merge` assumes an object-shaped section and `permissions.deny` is an array, so it needs a
  sibling helper.
- **`bp add`'s three readiness defects**, deferred out of BIT-39 Verse 5 and recorded in BIT-39's
  Decisions so a follow-up doesn't re-derive them.
- **`abort-run.md`** predates the queue — it must also dequeue the track's rows and delete the remote
  branch now that `bit_do` pushes.
- **BIT-39.9 and BIT-39.10's `User verifies` text is stale** — both describe behaviour the track's
  own Decisions later reversed. Fix before either is used as a regression script.

---

## Decisions that still bind

**Loop shape**
- **One bar in flight per project.** Projects advance in parallel; each project is serial.
- **A parked bar holds its slot.** A bar waiting on a permission prompt or on `## User verifies`
  stalls that project's queue until the operator acts. Other projects keep moving.
- **The ledger is the source of truth.** The loop re-reads `.bit/` rather than trusting a snapshot,
  so a bar already `done`, or not approved, is dropped rather than run.
- **Fresh session per bar.** `bit_do` never rolls into the next bar in its own session — fresh
  context per bar is the anti-drift mechanism.
- **Permission prompts are retained deliberately.** The session pausing for the operator is the
  safety mechanism that replaces the push gate. No permission *mode* that suppresses it.
- **The loop keeps no state between ticks.** An unconfirmed row is retried, not diagnosed. Accepted
  cost: a permanently broken project re-spawns every tick, with claude's diagnostic in the log.
- **Dequeue waits a tick.** The row survives the spawning tick and clears on a later one, once the
  session is confirmed present. Reverses the original "dequeue on dispatch" — a doomed spawn exits 0
  and the sub-second confirm caught it mid-registration, deleting the row for a bar that never ran.

**Worktrees and naming**
- **Named per track, never per bar** — per-bar naming restarts each bar from `main`, so bars of a
  track share one tree and bar N+1 inherits whatever bar N left, wreckage included.
- **Claude creates the worktree, not `bp`.** Reverses the earlier call. Dispatch passes `-w <name>`
  and accepts the forced `worktree-` branch prefix; `bp` creating one buys only an unprefixed branch
  name and costs new git-worktree code in Go.
- **The name is derived at dispatch, never stored.** `<track-id>-<kebab-cased track title>`. No skill
  change, no new record field. The casing rule is deliberately loose.
- **One identifier everywhere** — the same string goes to `-w` and `-n`, so a `claude agents` row is
  directly attributable to a track.
- **The loop never tears down a worktree.** Reaping is the operator's, after they have merged.

**Approval**
- Approval gates everything: approve a track → planning may proceed; approve every bar → work may
  proceed. Approval happens in the TUI.
- **Claude never approves.** Not yet enforced — see Next steps.
- **Editing an approved record revokes its approval.** A forward status move (`todo → doing → done`)
  keeps it; only a send-back to `todo`, or a title/description/phase edit, revokes.

**Daemon hosting** — settled 2026-08-19.
- **launchd hosts the daemon** as a per-user LaunchAgent in `gui/$UID`, label
  `com.github.b4dmonkey.bit-pro`. No self-fork.
- **A stop is durable, a crash and a reboot are not.** `bp stop` keeps it down across a reboot;
  `KeepAlive {SuccessfulExit: false}` restarts on crash only.
- **`bp start` resolves `claude` and pins it in the plist** as `BP_CLAUDE`. A LaunchAgent inherits no
  login `PATH`, and this puts the failure at `bp start` where someone is watching.
- **macOS only for now.** Linux would be a systemd user unit; the repo still has no `GOOS` code.

**State**
- YAML frontmatter for task records; SQLite for global state (registry, queue). Structured state is
  machine-owned, the body is human-authored markdown.

**Deferred, not dropped**
- **`blocked_by: [ID...]`** with derived readiness, one direction, cycles rejected. Earns its keep on
  cross-track edges. Interim brake is withholding approval.
- **Approving a track fires planning** — would make the pipeline one rule.
- **Provenance** — actor + timestamp on transitions.
- **Rate-limit exhaustion mid-track.** Nothing designed.

---

## launchd mechanics

The parts that would bite again if forgotten. `man launchd.plist` is the authoritative key list.

| command | launchd equivalent |
| --- | --- |
| `bp start` | `launchctl enable gui/$UID/<label>`, then `bootstrap gui/$UID <plist path>` |
| `bp stop` | `launchctl bootout gui/$UID/<label>`, then `disable gui/$UID/<label>` |
| `bp status` | `launchctl list <label>`, plus `print-disabled gui/$UID` for the `stopped` state |
| restart in place | `launchctl kickstart -k gui/$UID/<label>` |

`bootstrap` takes a **plist path**; `bootout` and `kickstart` take **`<domain>/<label>`**; `list`
takes a **bare label**. `load`/`unload` are the deprecated forms in older docs.

- **`bootout` is session-scoped; `disable` is durable.** launchd re-walks `~/Library/LaunchAgents/`
  at login, so `bootout` alone lets a stopped daemon resurrect. Order is **bootout then disable** —
  `disable` doesn't kill a running job, so disabling first leaves a live process marked disabled and
  `bp status` reports `stopped` about it. A non-zero `bootout` is fine and must not short-circuit
  the `disable`.
- **Two staleness cases need different commands.** Binary replaced at the same path → `kickstart -k`.
  Plist *contents* changed → `bootout` then `bootstrap`; launchd holds the job definition in memory
  and editing the file is not enough.
- **Agents inherit almost no environment.** Everything is an absolute path, and anything the spawned
  sessions need must be declared in the plist. This is what `BP_CLAUDE` exists for.
- **Three locations.** Binary on `PATH`, plist at `~/Library/LaunchAgents/<label>.plist`, logs and
  `bit.db` in `~/.local/share/bit-pro/`. The plist ties them together, and living in
  `~/Library/LaunchAgents/` is the entire mechanism for surviving a reboot.

---

## Measured facts

**Spawn surface** — 2.1.239 through 2.1.251.

- **`--bg` and `-p` conflict**; dispatch spawns with the positional prompt.
- **`-w` and `-n` are separate flags.** Passing the same string to both is what makes one identifier.
- **`-w <name>` sets the path verbatim; the `worktree-` branch prefix is forced.**
- **Bars can share one worktree** — two sessions with the same `-w` report the same `cwd`.
- **`claude` has no cwd/`-C` flag.** A spawned session takes its working directory from the calling
  process; in Go that is `exec.Cmd.Dir`. `--cwd` exists only on `claude agents`, as a filter.
- **There is no `--max-turns`.** `--max-budget-usd` needs `--print`, which `--bg` refuses. **An
  unattended bar cannot be capped from the command line** — it runs until it finishes, parks, or is
  stopped by hand.
- **Spawn prints `backgrounded · <id> · <name>` on stdout.** Do not key on it: it needs a TTY the
  launchd daemon does not have, and it is human-readable text.
- **An unknown `--agent` no longer errors** (2.1.250) — it warns, exits 0, and spawns with the
  default template. So a project where the plugin doesn't resolve silently runs `/bit:do` unagented.
- **No folder-trust park in a never-opened, non-git directory** (2.1.245). This widened the earlier
  trust finding and leaves the 2026-08-21 scratch-repo parking **unexplained** — "the directory is
  new to the operator" is not the cause.

**Liveness and teardown.**

- **`claude agents --json` needs no TTY** and is the only poll surface available to the daemon. It is
  **machine-wide**, so rows must be filtered by `cwd` — at or beneath the project path, compared by
  segment. Background rows carry `id` and `state`; interactive rows carry neither.
- **`--json` without `--all` excludes completed background sessions**, per its own `--help`.
- **`state` is transient — never filter on it.** A row read `state: done` and ten minutes later the
  same row read `state: blocked`. `done` means "finished this turn". **Presence in the plain listing
  is the whole liveness test.** This killed the earlier `state ∈ {working, blocked}` filter, which
  would have dispatched a second bar on top of a mid-flight session.
- **A finished session never leaves the listing** (2.1.250) — it stayed at `state: done` /
  `status: idle` across 24 polls over two minutes. A session that *died at startup* was absent from
  the plain listing and appeared only under `--all` at `state: failed`. So plain presence is an
  honest success signal, and the slot is never freed by the session simply finishing.
- **Teardown follows ownership.** `claude stop <id>` retains the worktree; `claude rm <id>` removes
  it **and deletes its branch**. Claude locks the worktrees it creates, so plain git needs
  `worktree remove -f`; a `bp`-created worktree carries no lock.

**Permissions** — measured 2026-08-29 on 2.1.251, throwaway git repo, `git add` as the target, with
`git status --short` as a passing control.

- **An `ask` rule beats an `allow` from every source.** Docs confirm: rules evaluate deny → ask →
  allow, first match wins, and specificity does not reorder them. Denied under each of
  `--allowedTools "Bash(git add *)"`, `--allowedTools "Bash(git add:*)"`,
  `--settings '{"permissions":{"allow":[...]}}'`, and `--permission-mode dontAsk`.
- **`--setting-sources project,local` plus `--allowedTools` is the only combination that clears it.**
  Dropping the `user` source removes the rule rather than losing to it; the flag then supplies the
  grant. Verified it stays narrow — an unlisted command is still denied, because dropping the user
  source removes its allow rules too.
- **The cost** is the session's whole user-level config: `.env` deny rules, `model`, `effortLevel`,
  and user-level `extraKnownMarketplaces` (`go-skills` lives only there).

**Repo and config.**

- **BIT-27** — `bp` cuts the path at `.claude/worktrees/` to find the canonical `.bit/`, so any
  invocation from inside a worktree writes to the **main checkout**. `.bit/` changes therefore never
  join a worktree commit.
- **`.claude/settings.json` is tracked** and `.gitignore` excludes only `.claude/worktrees`, so a
  worktree checkout carries it and a dispatched session has `bit@bit-pro` enabled.
- **`bp init` writes `.claude/settings.json`** via `claude.WriteSettings` (`cmd/init.go:48`) — the
  seam for anything installed at enrollment. It writes `extraKnownMarketplaces` and `enabledPlugins`
  only, and `claude.merge` assumes an object-shaped section.
- **Pushing needs no worktree knowledge.** All worktrees share one `.git`. `git worktree list
  --porcelain` lists the main checkout first and paths may contain spaces — strip the `worktree `
  prefix, never split on whitespace.

---

## Open gaps

- **Counts vs. a track whose approval and status disagree.** The four buckets partition cleanly only
  while approval and status agree; an unapproved track that is already `doing` files under
  **backlog** by first-match, which is incidental rather than designed. `tui/board.go:64` cuts it
  differently — it hides unapproved tracks from **To Do** only — so the TUI and the counts can
  disagree about the same track.
- **A registered project the loop cannot read** is skipped for that tick and keeps its old counts.
  Undecided whether the operator ever finds out; a moved or deleted path reports stale counts
  indefinitely with nothing saying so.
- **Count freshness.** Counts go stale the moment `.bit/` is edited by hand between ticks, so
  `bp list` can lie. Recompute on read, or accept lag?
- **TUI ↔ registry.** Enqueuing on a project with no `bp add` row is a silent no-op — no error, no
  offer to register. Nothing chosen.
- **Daemon not running.** Answering "yes" at the play prompt writes queue rows that nothing
  dispatches. The TUI should probably say so.
- **Ctrl+C at an interactive prompt.** `signal.NotifyContext` cancels the root context but does not
  stop the command, so Ctrl+C at `bp add`'s prompt takes the offered default and fails at the insert
  with `registering <path>: context canceled`. Nothing is written, so not a correctness bug, but the
  exit is wrong. Same shape in `bp init`.

---

## Escape hatch

`abort-run.md` (repo root, untracked, deliberately not gitignored) is a prompt for undoing a bad run:
stop the track's sessions, show what is about to be discarded, reap the worktree and branch, reset
every bar to `todo` and unapproved, verify. Only as good as the `.bit/` commit taken before
approving. Needs the two updates listed under Next steps.

---

## Position

Explicitly high-touch. Automate only what has earned trust; scoping and planning probably never. The
target is the tail end of implementation. A correct pipeline first — smooth beats fast. The
permission pause is what makes "high-touch" real rather than aspirational.

Cost model: not "one shot per bar" but **one session per bar** — a bar whose check fails once is
already multi-turn, and it cannot be bounded from the command line.

**Reference points:** *beads* (`gastownhall/beads`) for the data model — bit-pro plays this role;
*gastown* (`gastownhall/gastown`) for the orchestration, minus the merge queue and escalation
channel. Nearly every complicated part of Gas Town is a consequence of concurrency, not of
automation. At one bar per project, what remains is a queue, a worktree, and a gate.

## Docs

- <https://code.claude.com/docs/en/headless>
- <https://code.claude.com/docs/en/hooks>
- <https://code.claude.com/docs/en/permissions>
- <https://code.claude.com/docs/en/agent-sdk/permissions>
