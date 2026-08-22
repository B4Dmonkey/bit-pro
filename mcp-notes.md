# MCP phase — working notes

**Last synced 2026-08-22. Route: `bp serve mcp`, a stdio MCP server in the same binary.**

`bp` is a CLI designed to be driven by a model, but it is reached through Bash — one of a
thousand things that can be typed into a shell. The consequences are visible: Claude reaches
for `mv`, `cat`, and `sed` against `.bit/` instead of the command that owns the file format,
and the command contract has to spend half its length teaching shell technique rather than
the domain. An MCP server makes the tool surface typed and enumerable, and makes the commands
Claude should never run *absent* rather than merely denied.

This phase runs **during or after** the automation phase (`automation-notes.md`); the two are
mostly orthogonal, but see "Relationship to the automation phase".

Everything under "Measured facts" was read out of this repo on 2026-08-22. Everything under
"Unverified" has **not** been checked and must not be planned against as written.

---

## Todo

Ordered. Additive first — nothing is removed until the replacement is proven in a real cycle.
Each line is done when its "done when" clause is true.

### 1. Server skeleton

- [ ] `bp serve mcp` runs a stdio MCP server exposing exactly one read-only tool, `task_read`.
      *Done when:* Claude Code lists the tool in a project wired to it, and a `task_read` on a
      real track returns its body.
      *Depends on:* the `serve` parent existing — see the pending rename in `automation-notes.md`.
      Landing this under a bare `bp mcp` first and moving it later is fine; the two-word form is
      where it ends up.
- [ ] The server is a plain foreground process speaking stdio on stdin/stdout, so it can be driven
      by hand.
      *Done when:* running it in a terminal and typing an `initialize` frame gets a response.

**Nothing logs to stdout.** stdio transport owns stdout; any stray `fmt.Println` in a shared
code path corrupts the protocol stream. This is the one mechanical hazard of the phase.

### 2. Read surface

- [ ] `task_read` and `task_list` return structured JSON — no header line, no `--body` flag,
      no tab columns.
      *Done when:* a skill's read step needs no Bash and no tab counting.

### 3. Write surface

- [ ] `task_create`, `task_update`, `task_move`, `feedback_add`.
      *Done when:* a full scope → plan cycle runs start to finish without Bash touching
      `.bit/` once.
- [ ] `task_update` preserves the approval-revocation rule exactly (see Decisions).
      *Done when:* a body edit through the tool revokes approval and a `doing → done` move
      does not.

### 4. Registration

- [ ] A project enrolled with `bp init` gets the server with no manual step.
      *Done when:* `bp init` in a fresh repo, then a new session, and the tools are listed.
- [ ] Settle where the registration lives — see the first Open gap. This step cannot be
      planned before that decision.

### 5. Skills migration

- [ ] All seven skills under `bit/skills/` and both agents under `bit/agents/` call tools
      instead of shelling out.
      *Done when:* no skill or agent file contains a `bp task` or `bp feedback` bash block.
- [ ] Decide the fate of `bp instructions` / `assets/bit-cli.md` (see Open gaps). Roughly half
      the file is shell technique that schemas absorb; the domain half has to land somewhere.
- [ ] This is skill-creator work, not code — and it ships through the plugin, so remember the
      GitHub coupling under "Relationship to the automation phase".

### 6. Close the Bash path

- [ ] Deny rules in `.claude/settings.json` for shell writes against `.bit/`.
      *Done when:* an attempt to `sed` a task file is refused rather than silently succeeding.
- [ ] **Strictly after step 5.** Denying the shell path while a skill still depends on it
      breaks every cycle.

### 7. CLI removal

- [ ] Delete the Claude-only commands from the CLI (see the inventory table).
      *Done when:* `bp --help` lists only commands the operator actually runs.
- [ ] Last, and only once a real cycle has run on tools alone. The commands are the fallback
      until then.

---

## Decisions

**What the MCP is for**
- **Typed and enumerable beats documented.** The tools appear in the tool list with schemas.
  The reason Claude reaches for `mv` is that `bp task move` is invisible until something tells
  it the command exists; a tool is not.
- **Absence is stronger than denial.** `approve` simply is not a tool. "Claude never approves"
  stops being a deny rule that has to be maintained per project and becomes a property of the
  surface.
- **The MCP makes the right path typed; the deny rules make the wrong path impossible.** Both
  are needed — Bash does not disappear because a better tool exists. Step 6 is not optional.

**Parity means the domain, not the flags**
- **The CLI's shape is half shell accident.** `$( )` to capture a minted ID, `-d "$(cat
  body.md)"` because multi-line markdown through a shell arg is treacherous, `read --body | sed
  | update -d` round-trips, "count tabs rather than assuming the phase label is the fourth
  field". None of it survives structured params and structured returns, and dropping it is
  parity rather than a gap.
- **`status` becomes a schema enum.** The gotcha that `-s doen` succeeds and silently breaks
  rollup forever is a consequence of a stringly-typed CLI flag, not a domain decision. The
  schema fixes it for free without introducing the state machine the CLI deliberately lacks.
- **Structured returns, not text.** `task_read` returns fields; `task_list` returns an array of
  objects. The header line, the `--body` flag, and the five-tab-column format are all
  serialisation for a terminal and have no reason to reach a model.

**Tool shape — mirror the nouns, fix the accidents**
- **Six tools, not a redesign.** `task_create`, `task_read`, `task_list`, `task_update`,
  `task_move`, `feedback_add`.
- **Rollup stays skill logic.** The CLI does not cascade a bar's status up to its track and
  neither does the MCP. bit_do owns that rule today and keeps owning it.
- **Intent-shaped tools were rejected for now.** `scope_write` / `plan_add_bar` / `rollup`
  would migrate skill logic into Go — a much larger change, against YAGNI, and the right cut
  lines are not known yet. Revisit only if the noun-shaped tools prove to invite the same
  read-modify-write dances the CLI does.
- **Approval-revocation semantics are preserved verbatim.** A change to title, description,
  phase, or phase-label revokes approval; a status write of `todo` revokes it; a forward status
  move keeps it. This is load-bearing for the automation phase — an approved bar has to stay
  approved for the whole run.

**Same binary, not a second one**
- `bp serve mcp` lives in the `bp` binary, like the daemon. One install, one version, no separate
  distribution story. It also means the MCP and the CLI share the task package, so there is no
  second implementation of the file format to drift.

**Two servers under one verb** — settled 2026-08-22, the naming half of the same question.
- **`bp serve` is a parent with two children: `daemon` and `mcp`.** They are grouped because they
  are the same *shape* — a plain foreground process that assumes nothing about launchd — not because
  they share an audience. `--help` then says there are exactly two servers and neither one is *the*
  server, which is the thing a single `bp serve` plus a single `bp mcp` could never say.
- **The daemon's rename is the automation phase's work**, not this one. See "Pending rename" in
  `automation-notes.md`, including the stale-plist trap that comes with it.
- **Nobody types `bp serve mcp`.** Claude Code's config does, per session. So it is not a service in
  the operator's sense: `bp start`/`stop`/`status` stay daemon-only and never gain an MCP mode, and
  there is no plist, no label, and no liveness question here.

**MCP does not dispatch; the daemon does**
- **An MCP tool only fires while a session is already alive and chooses to call it.** Dispatch has to
  happen when no session exists — that is the entire point of the automation phase — so the MCP
  server structurally cannot be the dispatcher. It cannot spawn the thing that calls it.
- **The split is by what each owns.** The daemon owns *time and the queue*: pop the head, spawn the
  session, one bar in flight per project, poll `claude agents --json` for completion. The MCP server
  owns *the ledger, per session*: typed reads and writes for whichever session happens to be alive,
  dispatched or interactive.
- **They meet at exactly one point** — the session the daemon spawns is a client of the MCP server.
  That is also the ordering argument: see "Relationship to the automation phase".

---

## Command inventory

The full surface as of 2026-08-22, split by who actually runs it.

| | commands | fate |
| --- | --- | --- |
| **Operator-only** | `tui`, `approve`, `unapprove`, `init`, `add`, `list`, `start`, `stop`, `status`, `serve daemon` | stay CLI; **never** become MCP tools |
| **Claude-only** | `task read`, `task list`, `task create`, `task update`, `task move`, `feedback add`, `instructions` | become tools, then delete from the CLI (step 7) |
| **Contested** | `task complete`, `task delete` | undecided — see Open gaps |

`completion` and `help` are Cobra's and are nobody's decision. `serve mcp` is in neither column — it
is the surface the tools arrive *through*, and its only caller is Claude Code's own config.

### Parity map

Flags on the left are what exists today; params on the right are the proposed schema.

| CLI | tool | params | returns |
| --- | --- | --- | --- |
| `task create <title> -d -p/--parent --after --phase --phase-label` | `task_create` | `title`, `body?`, `parent?`, `after?`, `phase?`, `phase_label?` | `{id}` |
| `task read <id> [--body]` | `task_read` | `id` | `{id, title, status, approved, phase, phase_label, parent, body}` |
| `task list [-p/--parent]` | `task_list` | `parent?` | array of the above, minus `body` |
| `task update <id> -t -d -s --phase --phase-label` | `task_update` | `id`, `title?`, `body?`, `status?` (enum `todo\|doing\|done`), `phase?`, `phase_label?` | `{id, approved}` — so revocation is visible |
| `task move <bar> --before\|--after` | `task_move` | `bar`, `before?`, `after?` (exactly one) | `{}` |
| `feedback add <track> -d` | `feedback_add` | `track`, `body` | `{path}` |

Notes on the map:

- **`--body` disappears** because a structured return has no header to suppress. The
  byte-for-byte round-trip guarantee the CLI provides still matters and still has to hold —
  it is what makes read → edit → write-back safe.
- **`task_update` returning `approved`** is new. Today revocation is silent: the skill has to
  know the rule and infer that it fired. Returning the flag makes it observable, which matters
  most in exactly the place it is most dangerous — a replan touching approved bars.
- **`create --after` is a real flag that the contract never documents.** It exists at
  `cmd/task_create.go:31` and `assets/bit-cli.md` mentions `--after` only for `task move`. So
  a bar can be inserted mid-track at create time and no skill knows it. Carrying it into the
  schema fixes the drift by construction, since the schema *is* the documentation.
- **Whole-body writes stay coarse.** Toggling one verse checkbox still means rewriting the
  body. It is much less painful without the shell (`body` is just a JSON string, no temp file,
  no quoting), so a targeted edit tool is deferred rather than designed.

---

## Relationship to the automation phase

- **Version skew is a new failure mode.** The skills ship via the plugin **from GitHub**
  (`claude.WriteSettings` wires `bit@bit-pro` from `B4Dmonkey/bit-pro`), while `bp serve mcp` is the
  locally installed binary. A skill that calls a tool the installed binary does not have yet
  fails in a way the Bash route never could — Bash at least produced a legible `unknown
  command`. Nothing about this is solved; it is the sharpest new risk in the phase.
- **The daemon is the MCP server's most important client.** The two phases are not competing for the
  same job — the daemon spawns sessions, those sessions talk MCP. See "MCP does not dispatch; the
  daemon does" for the split.
- **Dispatched sessions are the strongest argument for landing this early.** A `bit:bot-dev`
  session spawned by the daemon into a worktree has no operator watching it choose `mv` over
  `bp task move`. That is the session that most needs a typed surface — which argues for
  landing this before automation step 6 rather than after.
- **Worktrees.** BIT-27 has `bp` cut the path at `.claude/worktrees/` to find the canonical
  `.bit/`. Whatever the MCP does about project resolution has to reach the same answer, or a
  dispatched session writes to the wrong ledger. See the second Open gap.
- **A tracked registration file follows into a worktree; an untracked one does not.** A
  worktree is a checkout, so this decides whether dispatched sessions see the tools at all.

---

## Measured facts

Read out of the repo on 2026-08-22.

- **The surface is 15 top-level commands and 7 `task` subcommands.** Top level: `add`,
  `approve`, `unapprove`, `completion`, `feedback`, `help`, `init`, `instructions`, `list`,
  `serve`, `start`, `status`, `stop`, `task`, `tui`. Under `task`: `complete`, `create`,
  `delete`, `list`, `move`, `read`, `update`.
- **Skill usage, counted across `bit/skills/` and `bit/agents/`:** `task list` 16, `task read`
  11, `task update` 9, `instructions` 7, `task complete` 3, `task create` 2, `feedback add` 2,
  `approve` 2, `tui` 1. The read/write pair plus `instructions` is nearly the whole traffic.
- **`bp instructions` is 158 lines of embedded markdown** (`assets/bit-cli.md`, embedded via
  `assets/assets.go`). Sections that schemas absorb: shell ID capture, "Writing a body from the
  shell", the tab-column warning, the `status` spelling gotcha. Sections that are domain and
  survive: track vs. bar, approval as a separate axis from status, rollup-is-skill-logic, and
  IDs being reserved rather than freed on delete.
- **`bp init` writes only `.claude/settings.json`.** It writes no skills and no
  `.claude/bit-cli.md` — asserted deliberately by `TestInitCmd_WritesNoSkills`
  (`cmd/init_test.go:145`). The settings file gets `extraKnownMarketplaces` and
  `enabledPlugins` only. So the contract reaches Claude *only* by a skill telling it to run
  `bp instructions`, which is why that command has 7 call sites.
- **The plugin manifest is metadata only.** `bit/.claude-plugin/plugin.json` carries `$schema`,
  `name`, `displayName`, `description` — nothing else.
- **`go.mod` has no MCP dependency.** Go 1.26.5, Cobra 1.10.2, bubbletea/lipgloss, dbmate,
  modernc sqlite, pathologize, yaml.v3. An SDK would be the first new direct dependency since
  the daemon work.
- **`task delete` has `-y/--yes` and `-f/--force`; nothing else Claude-facing has a prompt.**
  A confirmation flag is a tell that the command was built for a human.
- **`approve`/`unapprove` take any task ID and print nothing on success.** No separate track
  and bar commands.
- **`create --after` exists in code and not in the contract.** `cmd/task_create.go:31` versus
  `assets/bit-cli.md`, which covers `--after` only under `task move`.

---

## Unverified — check before planning

Load-bearing for effort estimates. None of this has been looked up.

- **The Go MCP SDK.** `github.com/modelcontextprotocol/go-sdk` is the presumed import path;
  its API shape, stability, and whether it is the right choice over a hand-rolled JSON-RPC
  stdio loop are all unchecked. A hand-rolled server is not obviously wrong for six tools.
- **Whether a plugin manifest can declare MCP servers.** If `plugin.json` supports an
  `mcpServers` key, registration could ride the plugin that already ships the skills — which
  would tie the two together and mostly answer the first Open gap. Unconfirmed.
- **How a stdio server's working directory is set**, and whether it is re-launched per session
  or shared. This decides the project-resolution question outright.
- **Whether tool names are namespaced by the server** when Claude sees them, and therefore
  whether `task_read` needs a `bit_` prefix to be unambiguous.

---

## Open gaps

Everything here needs the operator's review. The two under "Assumptions" would change the
plan, not just fill it in.

- **Where the server is registered.** Candidates: a `.mcp.json` at the project root written by
  `bp init`; the plugin manifest, if it supports servers; user-level settings. Tracked
  `.mcp.json` follows into a worktree and into a clone, which suits dispatched sessions — but
  it reverses the direction `bp init` deliberately took when it stopped writing content files
  into projects (`TestInitCmd_WritesNoSkills`). Nothing chosen. **Step 4 cannot be planned
  until this is.**
- **How the server resolves `.bit/`.** A CLI call carries cwd implicitly and BIT-27 cuts the
  path at `.claude/worktrees/` to reach the canonical ledger. A stdio server launched once per
  session resolves the project *once*, at startup — fine if it is launched per project, wrong
  if a dispatched worktree session inherits a server rooted somewhere else. The alternative is
  an explicit project/root param on every tool, which is uglier and unambiguous. Undecided.
- **What happens to `bp instructions`.** Schemas absorb the how; the domain half has to land
  somewhere. Candidates: an MCP resource, a tool that returns the doc, or folded into the
  skills themselves. Whichever wins, `assets/bit-cli.md` shrinks to roughly its domain
  sections. Undecided.
- **`task complete` and `task delete` — operator or Claude?** `complete` is the track sign-off,
  and `automation-notes.md` says the operator's explicit sign-off is what flips a track to
  `done` — which reads operator-side, except bit_do runs it today (3 call sites). `delete` has
  a confirmation prompt, which reads human, except dropping a bar mid-plan is a planning
  action. Both could be operator-only, both could be tools, or `complete` could be operator
  and `delete` a tool. Undecided.
- **Ordering against the automation phase.** Narrowed, not settled. Additive-first says ship the
  MCP, migrate the skills, then remove. Against that, dispatched sessions are the ones that most
  need the typed surface, which argues for steps 1–5 landing before automation step 6 — the removal
  steps (6 and 7) are unaffected either way, since they only close a path nothing depends on by
  then. So the live question is narrow: **does dispatch wait on the read/write tools and the skills
  migration, or ship on Bash and get the tools underneath it afterwards?** Operator's call.
- **Version skew between plugin skills and the installed binary.** Stated as a risk under
  "Relationship to the automation phase"; no mitigation designed. A version handshake, a
  capability check at session start, and "just live with it" are all on the table.

### Assumptions that need confirming

These came out of the brainstorm and were read *into* the request. Either one being wrong
changes the plan rather than a detail of it.

1. **"Remove commands the operator will never run" means the Claude-only ones get deleted from
   the CLI** once the MCP owns them — *not* "hide operator commands from Claude". This is the
   reading the inventory table is built on, and flipping it inverts the whole removal list.
2. **Parity means equivalent capability, not equivalent flags.** Dropping `--body`, the tab
   columns, and the `$( )` idiom counts as achieving parity rather than losing it. If parity
   means flag-for-flag, the parity map above is wrong.

---

## Position

The CLI is not being replaced; it is being split by audience. The operator keeps a CLI and a
TUI, Claude gets a typed tool surface, and the commands that only ever existed to be driven by
a model stop existing. Two surfaces over one task package, not two implementations.

Sequenced additive-first on purpose: the shell path stays open until a real cycle has run
without it. The removal steps are last because they are the only irreversible ones.

## Docs

- <https://modelcontextprotocol.io/>
- <https://code.claude.com/docs/en/mcp>
- <https://code.claude.com/docs/en/settings>
