# Automation phase — working notes

**Last synced 2026-08-21. Route: daemon.**

A long-running `bp serve daemon` process (still spelled `bp serve` in the code — see the pending
rename) watches registered projects and dispatches queued bars. This
replaces the earlier chaining design (a `Stop` hook re-invoking `bp dispatch` after every bar).
BIT-25 is shelved in `.bit/completed/` at `status: todo` — it is the chaining version and should
not be planned as written.

Everything under "Measured facts" was run and observed on this machine, Claude Code 2.1.231.
Where a doc and a measurement disagreed, the measurement won.

---

## Todo

Ordered. Each line is done when its "done when" clause is true.

### 1. Daemon lifecycle — **BIT-28** — done

- [x] **Verse 1** — `bp serve` runs the stub loop attached to the terminal.
      *Done when:* it ticks in the terminal and exits cleanly on Ctrl-C.
- [x] **Verse 2** — `bp start` / `bp stop` / `bp status` over a LaunchAgent; plist written on first
      start; logs in `~/.local/share/bit-pro/`.
      *Done when:* start → status says running → stop → status says stopped, and the daemon is still
      down after a reboot.

Shipped in BIT-28, filed under `.bit/completed/`. The daemon body is still a stub loop — step 6 fills
it in.

### 2. Project registry — **BIT-29** — done

- [x] SQLite at `~/.local/share/bit-pro/bit.db`, `projects` table, `bp add <path>`, `bp list`.
      *Done when:* a project enrolled from two different cwds is one row.
- [x] **Amend the schema before it ships:** `id INTEGER PRIMARY KEY`, `path` demoted to `UNIQUE`.
      The queue needs a stable FK target that survives a project moving on disk.

Shipped in BIT-29, filed under `.bit/completed/` — read it there for the dbmate-at-runtime and sqlc
decisions the dev workflow now rests on. `projects.id` exists, so step 5's FK target is in place;
the table has no count columns yet — step 4 adds them.

### 3. TUI — the play prompt — **BIT-31** — done

- [x] Approving the final unapproved bar on a track opens a modal asking whether to play the
      track. Popup only — "yes" does nothing yet.
      *Done when:* the modal appears on that approval and on no other, and both answers dismiss it.
- [x] Settle the trigger rule (see Open gaps): bars approve in any order, so the condition is
      "track has ≥1 bar and none unapproved", not "the last one in the list".

Shipped in BIT-31, filed under `.bit/completed/`. The trigger is a **stateless post-reload state
check** — "does this track have ≥1 bar and no unapproved bar right now?" — so unapprove-then-reapprove
fires it again, and a zero-bar track never does. The overlay renders in both `modeBoard` and
`modeList`. "Yes" was inert here; step 5 wired it.

### 4. Project counts — **BIT-32** — done

- [x] `projects` gains `backlog` (unapproved tracks), `todo` (approved tracks), `completed`
      (completed tracks). The loop refreshes them each iteration.
- [x] `bp list` and `bp status` render them.
      *Done when:* counts match what the TUI shows for the same project.
- [x] Efficiency of the refresh is a planning-time question — do not pre-optimise it here.

Shipped in BIT-32, filed under `.bit/completed/`. Four buckets, not three: **backlog** (unapproved),
**todo** (approved, not done), **done** (status `done`), **completed** (archived via
`bp task complete`). `bp list` renders all four, `bp status` three. The daemon loop is the sole
writer and `bp list` reads the cache, so **counts lag by design** — see the two count entries under
Open gaps, both still open.

### 5. Queue — **BIT-33** — done

- [x] `queue` table, FK to `projects.id`. **Rows are track rows or bar rows** — a track row means
      "work this track's bars in order", a bar row means "work exactly this bar".
- [x] Popup "yes" enqueues the track.
- [x] TUI shortcut to enqueue the selected track *or* bar directly, for the case where the operator
      answered "no" at the popup, or never got one.
- [x] Queued rows render **cyan** in the TUI.
      *Done when:* enqueue from the popup and enqueue from the shortcut produce the same row shape,
      and the cyan clears when the row leaves the queue.

Shipped in BIT-33 — **still at `.bit/tasks/BIT-33.md`, not yet filed under `.bit/completed/`.** Read
it there for the full decision list; the parts step 6 inherits:

- **The pop contract.** `queue` is `id INTEGER PRIMARY KEY`, `project_id` (FK `projects(id)`),
  `target_id TEXT` (BIT-N or BIT-N.M), `target_typ TEXT` (`track` | `bar`) — the shipped column names
  are `target_*`, not the `subject_*` in BIT-33's decision prose. FIFO within a project: the loop pops
  the smallest `id` for a `project_id`. Track and bar rows are sequenced together; there is no
  separate track queue. Only `EnqueueTask` and `ListQueueByProject` exist — **there is no delete
  query**, so step 6 writes the first one.
- **Nothing dequeues yet.** No `bp queue` commands were built, deliberately. Dispatch is where a real
  dequeue surface earns its keep.
- **`clear-queue.sh` is a throwaway.** Repo root, beside `install.sh`, clears every row against the
  runtime db (`~/.local/share/bit-pro/bit.db`, *not* the Justfile's `db/bit.db`). **Delete it, don't
  migrate it,** once step 6 owns dequeuing.
- **Enqueue on an unregistered project is a silent no-op**, and the TUI never checks whether the
  daemon is running. Both are Open gaps below.

### 6. Dispatch

- [ ] The loop pops the head of a project's queue, spawns a background Claude session prompted
      `/bit_do <BAR>` as `bit:bot-dev` in the track's worktree, and waits for it before dispatching
      that project's next bar. Cleanup on completion.
      *Done when:* a three-bar approved track runs bar 1 → 2 → 3 unattended, in order.
- [ ] Expect gaps here. Revisit this list before planning it.

### Before step 6 can do anything — carried over, still true

- [ ] **`bit_do` commits and pushes.** This is the automation. Today the skill forbids it outright
      (`bit/skills/do/SKILL.md:77`, `:99`) and all 43 BIT-2x bars say `## Commit (user)`. Mostly
      skill-creator work, plus dropping the push deny rule. **Nothing runs unattended until this
      lands.**
- [x] **Fix approval revocation** — **done, BIT-30.** A forward status move
      (`todo → doing → done`) now keeps approval; only a send-back to `todo` revokes it, along with
      title/description/phase edits (`cmd/task_update.go:45-53`). `bp task update <BAR> -s doing` no
      longer unapproves the bar it just started, so `bit_do`'s approval gate can resume it.
- [ ] **Worktree name recorded at scope time.** bit_scope asks for and stores it; bit_plan and
      bit_do carry it down. Dispatch cannot run before the name exists.
- [ ] **`bit:bot-dev` agent definition** — denied `approve`, written for an operator who is not
      continuously present.

### Pending rename — `bp serve` → `bp serve daemon` — needs its own scope track

Not a blocker for step 6; listed here so it isn't lost. Decided under "Daemon hosting"; the code has
not moved yet.

- [ ] `serve` becomes a parent command; the daemon body moves to `bp serve daemon`; the plist's
      `ProgramArguments` follows.
      *Done when:* `bp serve` prints help, `bp serve daemon` runs the loop, and a freshly enrolled
      machine's plist points at the two-word form.
- [ ] **`bp start` has to rewrite a stale plist.** `enrollDaemon` writes the plist *only when the
      file is absent* (`cmd/start.go:60-63`), so an already-enrolled machine keeps a plist pointing
      at bare `bp serve` — which, as a parent with no `RunE`, prints help and **exits 0** (verified
      against `bp task`). `KeepAlive {SuccessfulExit: false}` therefore does not restart it: the job
      loads, dies immediately, and nothing anywhere says so. The fix is `bp start` comparing the
      plist on disk against what it would write and, on a difference, doing **bootout → rewrite →
      bootstrap** — the "plist contents changed" case under "Two staleness cases", not `kickstart`.
      *Done when:* upgrading a binary whose plist is stale leaves a running daemon on the new form.

---

## Decisions

**Loop shape**
- **One bar in flight per project.** Projects advance in parallel; each project is serial.
- **A parked bar holds its slot.** A bar waiting on a permission prompt or on `## User verifies`
  items stalls that project's queue until the operator acts. Other projects keep moving.
- **Completion is detected by polling `claude agents`.** Commit history is a candidate second
  signal; which signals and how they combine is deferred to planning.
- **The ledger is the source of truth.** The loop re-reads `.bit/` rather than trusting a snapshot,
  so a bar already `done` is skipped rather than re-run.
- **Fresh session per bar.** `bit_do` never rolls into the next bar in its own session — fresh
  context per bar is the anti-drift mechanism, and the reason a ten-bar track doesn't end up with
  bar 8 reasoning out of bar 1's conversation.
- **Permission prompts are retained deliberately.** The session pausing for the operator is the
  safety mechanism that replaces the push gate. No permission mode that suppresses it.

**Daemon hosting** — settled 2026-08-19, see "Daemons on macOS" for the mechanics.
- **launchd hosts the daemon**, as a per-user LaunchAgent in `gui/$UID`. `bp` does not fork itself;
  the self-fork `Setsid` route is dropped. Label `com.github.b4dmonkey.bit-pro`.
- **`bp start`/`stop`/`status` stay the operator surface** and become `launchctl` wrappers. `bp start`
  writes the plist on first run — `bp init` does not, being per-project.
- **A stop is durable, a crash and a reboot are not.** `bp stop` keeps it down across a reboot; an
  unstopped daemon comes back on its own. `KeepAlive {SuccessfulExit: false}` restarts on crash only.
- **macOS only for now.** First platform boundary in the repo; Linux would be a systemd user unit.

**Approval**
- Approval gates everything: approve a track → planning may proceed; approve every bar → work may
  proceed. Approval happens in the TUI.
- **Claude never approves.** `bp approve` ships auto-denied for Claude.
- **Editing an approved record revokes its approval** — bars are approved in a batch and then run,
  so without revocation a replan could slip unreviewed work into an approved queue.

**Worktrees and naming**
- **Named per track, never per bar** — per-bar naming restarts each bar from `main`.
- **`bp` creates the worktree**, because `-w` forces a `worktree-` branch prefix and a pre-created
  worktree keeps its exact branch name.
- **One identifier everywhere** — worktree name, branch name, and `-n` session name are the same
  string, so a `claude agents` row is directly recognisable. Default `<track-id>-<short-name>`;
  an explicit operator-supplied name passes through verbatim, no truncation.
- **`bp` removes the worktree on `bp task complete`.**

**State**
- YAML frontmatter for task records; SQLite for global state (registry, queue). Structured state is
  machine-owned, the body is human-authored markdown. `bp task read --body` is the skill contract
  either way.

**Deferred, not dropped**
- **`blocked_by: [ID...]`** on any record, readiness derived (`done` for every listed ID), one
  direction only, cycles rejected across declared *and* implicit within-track edges. Earns its keep
  on cross-track edges. Interim brake is withholding approval.
- **Approving a track fires planning.** Makes the pipeline one rule: approval dispatches the next
  skill for that record type.
- **Provenance** — actor + timestamp on transitions.
- **Rate-limit exhaustion mid-track.** Weekly quota is the scarce resource; park and resume on
  reset. Nothing designed.

---

## Daemons on macOS

macOS has no supported "daemonize yourself" path. A Go process can re-exec itself with
`syscall.SysProcAttr{Setsid: true}` and it will survive the terminal, but nothing supervises the
result: no restart if it dies mid-track, no start at login, and its TCC/permission identity is
inherited from whichever terminal happened to spawn it. **launchd is the supervisor.** The program
you write is a plain foreground process; launchd owns the pid, liveness, restart policy, and log
redirection.

**Agent, not daemon.** `~/Library/LaunchAgents/` (the `gui/$UID` domain) is the fit — the loop
spawns Claude sessions as the operator and needs their environment and keychain. A LaunchDaemon
(`/Library/LaunchDaemons/`, root, runs pre-login) is the wrong domain and buys nothing here.

**What this does to `bp start` / `stop` / `status`.** They stay the operator-facing surface and
become `launchctl` wrappers, which is *less* Go code than the self-fork route — no pid file, no
liveness probe, no signal handling beyond `SIGTERM`:

| command | launchd equivalent |
| --- | --- |
| `bp start` | `launchctl enable gui/$UID/<label>`, then `bootstrap gui/$UID <plist path>` |
| `bp stop` | `launchctl bootout gui/$UID/<label>`, then `disable gui/$UID/<label>` |
| `bp status` | `launchctl list <label>` — dict with `"PID" = N;` when running, non-zero exit when not loaded; plus `print-disabled gui/$UID` for the `stopped` state |
| restart in place | `launchctl kickstart -k gui/$UID/<label>` |

`bootstrap` takes a **path to the plist**; `bootout` and `kickstart` take a **`<domain>/<label>`**;
`list` takes a **bare label**. `launchctl load` / `unload` are the deprecated forms that appear in
older docs and blog posts.

**Three locations, and the plist is a pointer file.** The binary is wherever it is installed on
`PATH`, the plist is at `~/Library/LaunchAgents/<label>.plist`, and logs and `bit.db` are in
`~/.local/share/bit-pro/`. The plist ties them together — `ProgramArguments[0]` is the absolute path
to the binary (resolved via `os.Executable()` by whichever `bp` writes the file) and
`StandardOutPath` points into the state dir. It has to be in `~/Library/LaunchAgents/` because
launchd auto-loads that directory at login; that is the entire mechanism for the daemon surviving a
reboot.

**`bootout` is session-scoped; `disable` is durable.** launchd re-walks
`~/Library/LaunchAgents/` at login and honours a persistent disabled store, so `bootout` alone lets a
stopped daemon resurrect itself at the next login. `disable` is what makes a stop stick, and
`launchctl print-disabled gui/$UID` reads the store back. Hence `bp stop` = `bootout` + `disable` and
`bp start` = `enable` + `bootstrap`, and `stopped` is a third status distinct from `not running`.
The ordering within `bp stop` is settled — **`bootout` first, then `disable`** (`daemon/stop.go`,
BIT-28.12). `disable` does not kill a running job, so disabling first would leave the daemon alive
while marked disabled, and `bp status` — which reads the disabled store first — would report
`stopped` about a live process. A `bootout` that exits non-zero is fine (a job that is not loaded is
already booted out) and must not short-circuit the `disable`; a `disable` that fails is a stop that
will not survive a reboot, and is reported as an error rather than as `stopped`.

**Two staleness cases, and they need different commands.** This is the one thing to get right:

- **The binary was replaced at the same path.** The plist is still correct. `kickstart -k` kills the
  running process and restarts it against the new bytes.
- **The plist's contents changed** — new `EnvironmentVariables`, a moved binary, a different log
  path. Editing the file is **not** enough: launchd holds the loaded job definition in memory and
  will not notice. This needs `bootout` then `bootstrap`.

**`--fg` is gone; the daemon body is its own command.** It was never a fork marker under this route —
there is no fork — and a flag that swaps `bp start`'s whole job is the wrong shape, on top of making
the plist point `bp start` at `bp start`. So the body is its own command: the plist's
`ProgramArguments` points at it, and running it by hand stays the way to watch the loop live rather
than tailing a log. `dispatch` was rejected as a name — it means the shelved chaining design's
one-shot run — and `srv` as an abbreviation Cobra convention spells out.

**The name is `bp serve daemon`, not `bp serve`** — settled 2026-08-22, once the MCP phase turned
out to need a second foreground server. `serve` becomes a parent with two children, `daemon` and
`mcp` (`mcp-notes.md`), so `--help` says there are exactly two servers and neither one is *the*
server. Grouped by shape, not by audience: both are plain foreground processes that assume nothing
about launchd. `bp start`/`stop`/`status` stay daemon-only and never gain an MCP mode, and nobody
types `bp serve mcp` by hand — Claude Code's config does.

**The code today is still `bp serve`** (`cmd/serve.go:18`, `daemon/plist.go:25`). The rename is real
work and needs its own scope track — see the Todo.

**Plist gotchas.** Agents inherit almost no `PATH` or environment, so everything is an absolute path
and anything the spawned Claude sessions need has to be declared in the plist (`EnvironmentVariables`,
`WorkingDirectory`). `KeepAlive` restarts on crash but `bootout` still unloads regardless, so it does
not fight `bp stop`. `StandardOutPath`/`StandardErrorPath` can point into
`~/.local/share/bit-pro/`, so the state dir decision is unaffected. Unsigned binaries under
`/Library/LaunchDaemons` hit Gatekeeper/TCC prompts — another reason to stay in the agent domain.
`man launchd.plist` is the authoritative key list.

**Who writes the plist.** Not `bp init` — that is per-project (`claude.WriteSettings`,
`cmd/init.go:48`) and the agent is machine-global. It belongs to a first-run path inside `bp start`
or an explicit `bp install`.

**Portability.** launchd is macOS-only; the Linux equivalent is a systemd user unit
(`~/.config/systemd/user/`, `systemctl --user`). The repo has no `GOOS`-specific code or build tags
today, and `Setsid` would not compile on Windows either — so either route eventually needs build
tags. Not a reason to prefer the self-fork route; it is unsupervised on Linux too.

---

## Measured facts

- **Worktrees are imposed, not opted into.** A `--bg` session that *edits files* auto-creates
  `<repo>/.claude/worktrees/<slug>` on branch `worktree-<slug>`, locked, with `cwd` set to it. A
  no-edit session does not. The trigger is editing.
- **`-w <name>` controls the path but not the branch prefix.** `worktree-` is forced.
- **A pre-created worktree is not re-isolated** — dispatch with that cwd and no `-w` and the branch
  keeps its exact name. Only route to full branch-name control.
- **Bars can share one worktree.** Two sessions with the same `-w` report the same `cwd`.
- **`Stop` fires on completion; `SessionEnd` does not.** A completed `--bg` session sits at
  `state: done`, `status: idle`, still in the registry.
- **Sessions park on permission prompts within ~3s**, under `auto`, `acceptEdits`, and
  `dontAsk --allowedTools` alike. One block point was *before any tool call at all*, which smells
  like a folder-trust prompt.
- **Registry rows** carry `pid, id, cwd, kind, startedAt, sessionId, name, status, waitingFor,
  state`. A finished session lingers as `state: done`, so **"is a bar live?" must filter on
  `state ∈ {working, blocked}`, not on presence.** This is what the loop polls.
- **Teardown.** `claude stop <id>` retains the worktree; `claude rm <id>` removes it *and* deletes
  its branch. Plain git needs `worktree remove -f -f` (Claude locks its worktrees) and the branch
  survives.
- **Pushing needs no worktree knowledge.** All worktrees share one `.git`. `git worktree list
  --porcelain` lists the main checkout first and paths may contain spaces — strip the `worktree `
  prefix, never split on whitespace.
- **Subscription auth.** `claude -p … --output-format json` works; `--bare` fails without
  `ANTHROPIC_AUTH_TOKEN`. Cold `-p` startup was ~21k tokens in this repo.
- **`bp init` already writes `.claude/settings.json`** via `claude.WriteSettings`
  (`cmd/init.go:48`) — the existing seam for anything installed at enrollment. The plugin ships
  skills and one agent; **there is no MCP server**, so `bp` reaches a model only as Bash.
- **BIT-27** — `bp` cuts the path at `.claude/worktrees/` to find the canonical `.bit/`, so any
  invocation from inside a worktree writes to the main checkout. Dispatch passes nothing.
- **`launchctl list <label>` is a usable status source.** It prints a plist dict including
  `"PID" = N;` when the job is running and exits non-zero when the label is not loaded. Bare
  `launchctl list` prints a `PID  Status  Label` table where a dash means loaded-but-not-running.
- **The repo has no platform-specific code.** No `syscall`, `GOOS`, or `_darwin`/`_linux` files
  anywhere under `*.go`. Whichever daemon route wins introduces the first build-tag boundary.

### Re-measured 2026-08-21 on Claude Code 2.1.239 (step-6 spawn surface)

Probed in a throwaway git repo. Every 2.1.231 claim above about `-w` held; these are the additions.

- **`--bg` and `-p` conflict** — refused outright: `--print` never starts the interactive session
  `claude agents` attaches to. Dispatch spawns with the **positional prompt**: `claude --bg '<task>'`.
- **`-w` and `-n` are separate flags.** `-w, --worktree [name]` creates the worktree; `-n, --name
  <name>` sets the session display name. Passing the same string to both is what makes one
  identifier; it is not one flag doing both jobs.
- **`-w <name>` sets the path verbatim, and the branch prefix is still forced.** `-w bit-99-probe` →
  worktree at `.claude/worktrees/bit-99-probe`, branch `refs/heads/worktree-bit-99-probe`.
- **A Claude-created worktree is locked; a `bp`-created one is not.** The lock reason is
  `claude session <name> (pid N start …)`. A pre-created worktree kept `refs/heads/bit-98-pre`
  exactly and carried **no lock**, so plain `git worktree remove` works on it — the `-f -f` in the
  fact above is only needed for worktrees Claude made.
- **Teardown follows ownership.** `claude stop` on the Claude-created session said *"worktree retained
  … run `claude rm` to remove worktree and job state"*; on the pre-created one it just said `stopped`.
  Claude only reaps what it created, so a `bp`-created worktree is `bp`'s to remove.
- **Spawn prints the landed confirmation on stdout:** `backgrounded · 7a3cb43c · bit-99-probe` — short
  id and session name. This is the signal a delete-on-dispatch dequeue would key on.
- **`claude agents` needs a TTY** — it refuses when stdout is not one and points at
  **`claude agents --json`**. A launchd-hosted daemon has no TTY, so the JSON form is the *only*
  poll surface available to the daemon.
- **`claude agents --json` is machine-wide, not per-project.** It returned an unrelated client
  project's interactive session alongside the probes, so the loop must filter by `cwd` (or `name`).
  **Background rows carry `id` and `state`; interactive rows carry neither** — filtering on `state`
  must tolerate the field being absent rather than treating it as not-running.
- **Both fresh background sessions parked within seconds** at `state: blocked`,
  `waitingFor: "permission prompt"`, in an untrusted scratch repo — the folder-trust prompt the
  2.1.231 note suspected. A dispatched session in a project the operator has never opened parks
  immediately and forever.
- **There is no `--max-turns` flag.** It is absent from `claude --help` on 2.1.239. The only cost
  bound is `--max-budget-usd`, documented "only works with `--print`" — and `--bg` refuses `--print`.
  **A backgrounded session cannot be capped from the command line**, so the "one bounded session per
  bar, capped with `--max-turns`" line under Position does not describe anything the CLI offers.
  An unattended bar runs until it finishes, parks, or is stopped by hand.

---

## Open gaps

- **Counts vs. a track whose approval and status disagree.** BIT-32's four count buckets — backlog
  (unapproved), todo (approved and not done), done, completed — partition cleanly only while approval
  and status agree. An unapproved track that is already `doing` or `done` matches more than one
  definition, and the first-match chain BIT-32 settled on (approval checked before status) files it
  under **backlog**. That placement is incidental, not designed. The state shouldn't arise, and
  BIT-32 deliberately has no bar for it — a backfill or a separate track is the fix if it turns up in
  practice. Note that `tui/board.go:64` already cuts this differently: it hides unapproved tracks from
  **To Do** only, leaving an unapproved `doing` track visible in **Doing**. So the TUI and the counts
  can disagree about the same track, which matters because "counts match what the TUI shows" is
  step 4's own done-when clause.
- **A registered project the loop cannot read.** BIT-32 settled that a project with no `.bit/`
  directory, or with an unparseable task file, is skipped for that tick and keeps whatever counts it
  already had. Undecided: whether the operator ever finds out. A path that was `bp add`ed and then
  moved or deleted goes on reporting stale counts indefinitely with no surface saying so — `bp list`
  can't distinguish "0 todo" from "unreadable since Tuesday". Candidates are a warn-level log line per
  skipped project, a marker in the rendered row, or a `bp doctor`-style check; none chosen.
- **Count freshness.** Counts cached in SQLite and refreshed per loop iteration go stale the moment
  `.bit/` is edited by hand between ticks, so `bp list` can lie. Recompute on read, or accept lag?
- **TUI ↔ registry.** The TUI is project-local; the queue is global. Enqueuing needs the project's registry row. Current behavior (BIT-33 decision): silently do nothing if the project has no `bp add` row. The operator gets no feedback — no error, no offer to register. A future pass should surface something (inline error, daemon log, `bp doctor` hint); nothing chosen yet.
- **Daemon not running.** Answering "yes" writes queue rows that nothing dispatches. The TUI should
  probably say so.
- **Track rows that get replanned mid-flight.** A track row expands to bars as it goes; a replan
  adds unapproved bars underneath it. Presumably the row parks. Not decided.
- **Ctrl+C at an interactive prompt.** `signal.NotifyContext` cancels the root context but does not
  stop the command, so Ctrl+C at `bp add`'s code prompt lets the read complete, takes the offered
  default, and only fails at the insert with `registering <path>: context canceled`. Nothing is
  written, so it is not a correctness bug, but the exit is wrong: the user meant abort and got a
  context error. Every interactive prompt (`bp init` too) has the same shape. Undecided whether the
  fix is checking `ctx.Err()` after the read, reading through a context-aware reader, or letting the
  second signal kill the process.
- **Abort.** `abort-run.md` predates the queue — it must also dequeue the track's rows, and delete
  the remote branch now that `bit_do` pushes.

---

## Escape hatch

`abort-run.md` (repo root, untracked, deliberately not gitignored) is a prompt for undoing a bad
run: stop the track's sessions, show what is about to be discarded, reap the worktree and branch,
reset every bar to `todo` and unapproved, verify. It is only as good as the `.bit/` commit taken
before approving. Needs the two updates listed under Open gaps.

---

## Position

Explicitly high-touch. Automate only what has earned trust; scoping and planning probably never.
The target is the tail end of implementation. A correct pipeline first — smooth beats fast. The
permission pause is what makes "high-touch" real rather than aspirational.

Cost model: not "one shot per bar" but **one bounded session per bar**, capped with `--max-turns`.
A bar whose check fails once is already multi-turn.

**Reference points:** *beads* (`gastownhall/beads`) for the data model — bit-pro plays this role;
*gastown* (`gastownhall/gastown`) for the orchestration, minus the merge queue and escalation
channel. Nearly every complicated part of Gas Town is a consequence of concurrency, not of
automation. At one bar per project, what remains is a queue, a worktree, and a gate.

## Docs

- <https://code.claude.com/docs/en/headless>
- <https://code.claude.com/docs/en/hooks>
- <https://code.claude.com/docs/en/agent-sdk/permissions>
