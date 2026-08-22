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

Everything under "Measured facts" was checked on 2026-08-22 — the first half read out of this
repo, the second half looked up from the SDK and Claude Code docs.

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
      *Built on:* `github.com/modelcontextprotocol/go-sdk/mcp` — `NewServer`, the generic
      `AddTool`, `StdioTransport`. See Measured facts for why not hand-rolled.
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

- [ ] `task_create`, `task_update`, `task_move`, `task_complete`, `task_delete`, `feedback_add`.
      *Done when:* a full scope → plan cycle runs start to finish without Bash touching
      `.bit/` once.
- [ ] `task_update` preserves the approval-revocation rule exactly (see Decisions).
      *Done when:* a body edit through the tool revokes approval and a `doing → done` move
      does not.

### 4. Registration

- [ ] `bp init` registers the server at **local scope** — the nested
      `projects."<abs path>".mcpServers` entry in `~/.claude.json`.
      *Done when:* `bp init` in a fresh repo, then a new session, and the tools are listed.
- [ ] It registers by shelling out to `claude mcp add bit -- bp serve mcp`, not by editing
      `~/.claude.json` directly.
      *Done when:* `bp init` adds the entry and leaves the rest of the file untouched.
Getting the server in front of a **dispatched** session is the daemon's job, not this phase's —
see "Scope boundary" under Decisions, and the dispatch step in `automation-notes.md`.

### 5. Skills migration

- [ ] All seven skills under `bit/skills/` and both agents under `bit/agents/` call tools
      instead of shelling out.
      *Done when:* no skill or agent file contains a `bp task` or `bp feedback` bash block.
- [ ] The domain half of `assets/bit-cli.md` lands on the MCP surface; `bp instructions` retires
      with the other Claude-only commands in step 7 (see Decisions).
      *Done when:* no skill runs `bp instructions` and no skill has lost the domain it taught.
- [ ] This is skill-creator work, not code — and it ships through the plugin, so remember the
      GitHub coupling under "Relationship to the automation phase".

### 6. Close the Bash path

- [ ] Deny rules in `.claude/settings.json` for shell writes against `.bit/`.
      *Done when:* an attempt to `sed` a task file is refused rather than silently succeeding.
- [ ] **Strictly after step 5.** Denying the shell path while a skill still depends on it
      breaks every cycle.

### 7. CLI removal — optional, and last

- [ ] Delete the Claude-only commands from the CLI (see the inventory table).
      *Done when:* `bp --help` lists only commands the operator actually runs.
- [ ] **This step may never run, and the phase is a success either way.** The goal is Claude on the
      MCP; deleting the commands afterwards is tidying, not the point. Steps 1–6 stand alone, so
      revisit this only once a real cycle has run on tools alone and the commands have gone unused
      long enough to be obviously dead.

---

## Decisions

**Scope boundary** — settled 2026-08-22.
- **This phase is exactly one thing: how Claude reaches `bp`, moving from Bash to a typed
  surface.** Not orchestration, not sessions, not worktrees.
- **Worktrees are the daemon's.** The MCP server needs no worktree machinery of its own — it
  resolves `.bit/` through the same BIT-27 path cut the CLI already uses, so a worktree session
  lands on the canonical store for free. Everything else about worktrees — creating them, spawning
  into them, and passing the server config to a session started in one — belongs to
  `automation-notes.md`.

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
- **Eight tools, not a redesign.** `task_create`, `task_read`, `task_list`, `task_update`,
  `task_move`, `task_complete`, `task_delete`, `feedback_add`.
- **Rollup stays skill logic.** The CLI does not cascade a bar's status up to its track and
  neither does the MCP. bit_do owns that rule today and keeps owning it.
- **Intent-shaped tools were rejected for now.** `scope_write` / `plan_add_bar` / `rollup`
  would migrate skill logic into Go — a much larger change, against YAGNI, and the right cut
  lines are not known yet. Revisit only if the noun-shaped tools prove to invite the same
  read-modify-write dances the CLI does.
- **No `bit_` prefix on tool names.** Claude Code already namespaces them as
  `mcp__bit__task_read`, so the nouns stay bare.
- **Approval-revocation semantics are preserved verbatim.** A change to title, description,
  phase, or phase-label revokes approval; a status write of `todo` revokes it; a forward status
  move keeps it. This is load-bearing for the automation phase — an approved bar has to stay
  approved for the whole run.

**`task complete` and `task delete` are both** — settled 2026-08-22.
- **One implementation, two callers.** The logic lives in the task package where it already does;
  the operator keeps `bp task complete` and `bp task delete`, and Claude gets `task_complete` and
  `task_delete`. This is the only pair that stays in the CLI *and* becomes tools, so step 7 does
  not touch it.
- **The confirmation prompt does not cross over.** `-y/--yes` exists because a terminal has to ask
  before an irreversible move; a tool call is already surfaced to the operator by Claude Code's own
  permission prompt, which is the same guarantee by a different mechanism.
- **`--force` does cross over**, because "a track with unfinished bars needs an override" is a
  domain rule, not a terminal affordance.

**`bp instructions` retires; the MCP carries the domain** — settled 2026-08-22.
- **Schemas absorb the how**, so the shell-technique half of `assets/bit-cli.md` simply stops
  existing. The domain half — track vs. bar, approval as an axis separate from status,
  rollup-is-skill-logic, IDs reserved rather than freed on delete — moves onto the MCP surface, and
  `bp instructions` is deleted with the other Claude-only commands.
- **Where exactly on the surface is deferred to the point the tools are written.** The cheapest
  answer is that the tool descriptions carry it per-tool and nothing separate ships; the fallback is
  a single `get_instructions` tool. An MCP *resource* is the wrong shape — a resource has to be
  pulled in deliberately, and this whole phase exists because a thing Claude has to be told about is
  a thing Claude does not use.

**Ordering against the automation phase does not constrain** — settled 2026-08-22.
- **Either phase can go first, and parallel streams are fine.** Neither has a hard dependency on
  the other: no step here waits on dispatch, and nothing in the automation phase waits on a tool.
- **The two streams collide in `cmd/serve.go` and nowhere else.** Automation turns `serve` into a
  parent; this phase hangs `mcp` off that parent. Whichever lands second inherits it.
- **"Dispatched sessions need the typed surface most" is a preference, not a constraint.** If
  dispatch ships on Bash, unsupervised sessions reach for `mv` until step 6 closes that path.

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

**Where the server is registered** — settled 2026-08-22.
- **Local scope: the nested `projects."<abs path>".mcpServers` entry in `~/.claude.json`.** Per
  project, so a repo with no `.bit/` never sees the tools, and nothing is written into the
  project — `bp init` keeps the direction it took at `TestInitCmd_WritesNoSkills`.
- **Project `.mcp.json` was rejected** because it reverses that direction, and **user scope was
  rejected** because it loads in every project on the machine, taxing unrelated sessions with six
  tools that can only error.
- **Declaring the server in the plugin manifest was rejected** on two counts, both recorded under
  Measured facts: the inline `mcpServers` key is silently stripped during manifest parsing
  (`anthropics/claude-code` #16143, open), and a plugin-declared server registers under the scoped
  name `plugin:<plugin>:<server>` rather than a bare `bit`, which changes every tool name step 6's
  deny rules have to spell. Neither was ever verified first-hand, and neither needs to be — they
  are why the route is closed, not open questions.
- **`bp init` shells out to `claude mcp add bit -- bp serve mcp`** (local is the default scope)
  rather than editing `~/.claude.json` itself. That file holds all of Claude Code's per-project
  state, so `bp` must not read-modify-write it. Cost: `bp init` now requires `claude` on `PATH`.
- **Getting the server in front of a dispatched session is not this phase's problem.** Local scope
  covers the sessions this phase is about — an operator working in the checkout. Whether a session
  the daemon spawns elsewhere also needs the server, and how, is a property of dispatch; it is
  owned in `automation-notes.md` and changes nothing here.

**How the server finds `.bit/`** — settled 2026-08-22, once the lookups came back.
- **Read `CLAUDE_PROJECT_DIR`, not cwd.** It is set in the server's own environment to the stable
  project root, and one server process exists per session, so the resolution happens once and is
  right for that session.
- **Worktrees need no new machinery.** `CLAUDE_PROJECT_DIR` is the session's launch directory, so
  a dispatched worktree session gets the worktree path and BIT-27's existing cut at
  `.claude/worktrees/` reaches the same canonical `.bit/` the CLI does.
- **So no project/root param on every tool, and no `roots/list`** — the latter is deprecated at
  the `2026-07-28` revision anyway.

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
| **Both** | `task complete`, `task delete` | become tools **and** stay CLI — one task-package implementation, two callers |

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
| `task complete <id>` | `task_complete` | `id` | `{}` |
| `task delete <id> [-y] [-f/--force]` | `task_delete` | `id`, `force?` | `{}` |
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
- **`task_delete` drops `-y` and keeps `--force`.** The prompt is a terminal affordance the
  permission prompt already covers; the unfinished-bars override is a domain rule. See Decisions.
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
- **Dispatched sessions are the strongest argument for landing this early — but it stays an
  argument, not a constraint.** A `bit:bot-dev` session spawned by the daemon into a worktree has
  no operator watching it choose `mv` over `bp task move`. Ordering is settled as unconstrained
  (see Decisions); this is the cost of dispatching on Bash in the meantime.
- **Registering the server for a dispatched session is the daemon's line item**, recorded in
  `automation-notes.md`, not here — see "Scope boundary" under Decisions.
- **The server itself needs nothing for worktrees.** It runs the same BIT-27 path cut the CLI
  does — see "How the server finds `.bit/`" under Decisions.

---

## Measured facts

### Read out of the repo on 2026-08-22

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

### Looked up on 2026-08-22

Sources under Docs. These are the answers to what used to be the Unverified section.

- **The Go SDK is official, stable, and beats hand-rolling.**
  `github.com/modelcontextprotocol/go-sdk` is maintained in collaboration with Google; **v1.0.0
  shipped an explicit no-breaking-API-changes guarantee** and v1.7.0 is current. It needs Go
  1.25+ and `go.mod` is on 1.26.5. The whole server is three calls — `mcp.NewServer`,
  `mcp.AddTool`, `s.Run(ctx, &mcp.StdioTransport{})` — and the generic `mcp.AddTool[In, Out]`
  **derives the JSON Schema from the Go structs**, with handlers shaped `func(ctx,
  *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error)`. That is what settles it:
  hand-rolling six tools means hand-writing six schemas and keeping them in sync with the
  structs, which is the same drift this phase exists to kill. The parity map's params and
  returns become two structs per tool.
- **v1.7.0 formally deprecates `roots`, sampling, and logging** at the `2026-07-28` protocol
  revision, keeps the Go types for older peers, and preserves backward compatibility on every
  endpoint. A plain stdio tool server is unaffected — but it does rule `roots/list` out as the
  project-resolution mechanism.
- **Three scopes, and two of them share one file.** Local (the default for `claude mcp add`)
  loads in the current project only and lives in `~/.claude.json` under that project's path;
  project loads in the current project only and lives in `.mcp.json` at the project root; user
  loads in **all** your projects and lives at the top level of `~/.claude.json`. Precedence is
  local, then project, then user, then plugin-provided, then claude.ai connectors — matched by
  name across the three scopes, and the whole entry from the winning source is used with no
  field merging.
- **`--mcp-config` takes JSON files or inline JSON strings, space-separated**, and its servers
  run with the working directory Claude Code started in. Noted because dispatch may want it; the
  decision is `automation-notes.md`'s.
- **A plugin manifest can declare MCP servers, and the inline form is currently broken.**
  `plugin.json` documents `mcpServers` as either an inline object or a path string
  (`"mcpServers": "./.mcp.json"`), with `${CLAUDE_PLUGIN_ROOT}`, `${CLAUDE_PLUGIN_DATA}`, and
  `${CLAUDE_PROJECT_DIR}` substituted in `command`, `args`, and `env`. Plugin servers connect at
  session startup once the plugin is enabled. But `anthropics/claude-code` issue **#16143** —
  open, reported against v2.0.76 — has the inline `mcpServers` key stripped during manifest
  parsing and silently ignored; the stated workaround is a separate `.mcp.json` in the plugin
  root.
- **A plugin-declared server registers under the scoped name
  `plugin:<plugin-name>:<server-name>`**, not a bare name.
- **`CLAUDE_PROJECT_DIR` is set in the spawned stdio server's environment**, to the stable
  project root — the same value hooks get — and it does not change when working directories are
  added or removed mid-session. Plugin-provided configs substitute `${CLAUDE_PROJECT_DIR}`
  directly; a project `.mcp.json` or a `~/.claude.json` entry needs a default
  (`${CLAUDE_PROJECT_DIR:-.}`) because the variable is set in the *server's* environment, not
  Claude Code's own.
- **One stdio server process per session, launched at session startup.** Stdio servers are local
  processes, are not shared across sessions, and are **not** reconnected automatically if they
  die (only HTTP/SSE are). So "resolves the project once, at startup" is once *per session* —
  the right granularity, since a dispatched worktree session gets its own server rooted at its
  own worktree.
- **cwd is unreliable, but only outside the CLI.** The CLI launches stdio servers from the
  project directory; the **desktop app** launches them with `cwd=$HOME` (issue #75266, open).
  The daemon spawns the CLI, so dispatch is unaffected — but it is a second reason to read the
  env var rather than `os.Getwd`.
- **Project-scoped `.mcp.json` servers load without the approval prompt in `claude -p`, Agent
  SDK, and cloud sessions**, because those sessions cannot show the prompt. Interactive sessions
  approve once, and `claude mcp reset-project-choices` resets that choice. So a tracked
  `.mcp.json` needs zero interaction in exactly the sessions the daemon dispatches.
- **Claude Code namespaces every MCP tool as `mcp__<server>__<tool>`**, normalizing invalid
  characters to underscores. A server named `bit` yields `mcp__bit__task_read`. Provenance is
  visible in transcripts, and these are the names step 6's deny rules have to spell.

---

## Open gaps

Everything here needs the operator's review. The two under "Assumptions" would change the
plan, not just fill it in.

- **Version skew between plugin skills and the installed binary.** Stated as a risk under
  "Relationship to the automation phase"; no mitigation designed. A version handshake, a
  capability check at session start, and "just live with it" are all on the table.

### Assumptions — both confirmed 2026-08-22

1. **"Remove commands the operator will never run" means the Claude-only ones get deleted from
   the CLI** once the MCP owns them — *not* "hide operator commands from Claude". Confirmed as the
   reading the inventory table is built on. Confirmed with a caveat: deletion is a last step and may
   not happen at all, so the table records where each command *would* go, not a commitment. See
   step 7.
2. **Parity means equivalent capability, not equivalent flags.** Confirmed: the bar is that Claude
   can do everything it can do today. Dropping `--body`, the tab columns, and the `$( )` idiom is
   achieving parity, not losing it, because those exist to get data through a shell and a tool call
   has no shell.

---

## Position

One thing: change how Claude reaches `bp`, from Bash to a typed surface. The operator keeps a CLI
and a TUI, Claude gets tools, and both run over one task package rather than two implementations.
Deleting the commands nothing drives any more is optional tidying at the end, not the goal.

Sequenced additive-first on purpose: the shell path stays open until a real cycle has run
without it. The removal steps are last because they are the only irreversible ones.

## Docs

- <https://modelcontextprotocol.io/>
- <https://code.claude.com/docs/en/mcp>
- <https://code.claude.com/docs/en/settings>
- <https://code.claude.com/docs/en/plugins-reference>
- <https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp>
- <https://github.com/modelcontextprotocol/go-sdk/releases>
- <https://github.com/anthropics/claude-code/issues/16143> — inline `mcpServers` dropped
- <https://github.com/anthropics/claude-code/issues/75266> — desktop app stdio cwd is `$HOME`
