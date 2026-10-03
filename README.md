# bit-pro

A project-management CLI for LLM-driven development. Markdown-backed, and shaped around how an agent and its human actually move work forward: scope it, plan it,
do it, check it.

The binary is `bp`. Work lives in a central store under `$XDG_DATA_HOME/bit` (default
`~/.local/share/bit`), keyed by project code.

## Why

Jira, Linear, and GitHub Projects are built for a human clicking through a web UI. When an
LLM is doing the work, that's the wrong interface. The agent wants plain text, deterministic
CLI commands, and task state on the local disk — not behind an API token.

So: task bodies are plain markdown in a local store, the CLI and the `mcp__bit__*` tools are
the primary interface, and `bp tui` is the human's window onto the same store. The TUI never becomes a second source of truth.

## Install

```
./scripts/install.sh
```

Builds `bp` into your Go bin dir (`$(go env GOBIN)`, else `$(go env GOPATH)/bin`) and
registers this repo as a Claude Code plugin marketplace. That bin dir must be on your
`PATH` — check with `bp --help`.

Needs Go. The agent skills need Claude Code; the CLI itself doesn't.

## Claude Code setup

`bp add` runs these commands once, the first time it registers a project, at user scope,
so bit loads in every Claude session. If a step fails, or the wiring is removed later,
run them by hand.

```
claude plugin marketplace add B4Dmonkey/bit-pro
claude plugin marketplace update bit-pro
claude plugin install bit@bit-pro --scope user
claude mcp add -s user bit -- bp serve mcp
```

## Quickstart

```
cd your-project
bp add .                                   # prompts for a project code, registers the folder
bp task create "Add OAuth login" -d "Why this matters…"
bp task create "Write the token test" -p PREFIX-1 --phase 1 --phase-label "Token exchange"
bp task list
bp task update PREFIX-1.1 -s doing
bp tui                                     # the human view: board + list
```

A project that already has a v1 `.bit/` directory runs `bp migrate` instead of `bp add`.

## Tracks and bars

A **track** is a top-level task — one scope, one deliverable. Its ID has no dot: `BIT-7`.
A **bar** is one step of that track's plan, minted with `--parent`: `BIT-7.3`. The parent is
readable straight out of the ID — no index, no lookup, and `ls` shows the tree.
See [hierarchy.md](./hierarchy.md) for the full vocabulary.

## Commands

Every command is non-interactive by default, so an agent can drive the whole lifecycle
without a prompt to answer.

| Command | What it does |
|---|---|
| `bp add <path>` | Register a project under a code you type, setting bit up in Claude Code on first use |
| `bp remove` | Archive this project's open work and remove it from the registry, after confirmation |
| `bp list` | List the registered projects |
| `bp migrate` | Copy this folder's v1 `.bit/` into the store and register it. Checks the files first, verifies the copy, and prints the cleanup step without running it |
| `bp approve <id>` / `bp unapprove <id>` | Approve a task, or revoke its approval |
| `bp task create <title>` | New task. `-p` parent, `--phase`/`--phase-label`, `--after` sibling, `-d` body |
| `bp task read <id>` | Full content. `--body` prints just the markdown, for feeding back to a model |
| `bp task list` | All tasks. `-p <track>` lists one plan, in step order |
| `bp task update <id>` | `-s` status, `-t` title, `-d` body, `--phase`/`--phase-label` |
| `bp task move <bar>` | `--before`/`--after` a sibling — resequence without renaming |
| `bp task complete <id>` | File a track and its bars as completed, with no landing check (manual override) |
| `bp task delete <id>` | Soft-delete into the archive. `-y` skips confirm, `-f` overrides the guard |
| `bp feedback add <track>` | Record a correction as a note in the shared feedback store |
| `bp tui` | Terminal UI |

Notes on behavior worth knowing:

- **Status is a plain field, not a state machine.** `todo`/`doing`/`done`, set directly.
  Rollup across a plan is workflow logic, left to the skills rather than enforced here.
- **Nothing is destroyed.** `complete` and `delete` both just move records. A relocated ID is
  reserved forever and drops out of its parent's order. A track only relocates once every bar
  is `done` — absolute for `complete`, `--force`-able for `delete`.
- **Order is separate from identity.** A reordered track carries an explicit list of its bar
  IDs, so a plan can be resequenced mid-stream while every ID stays stable.

## Agent skills

`bp add` sets up the `bit` Claude Code plugin, which ships these skills:

- `bit:analyze` — deep code research before a track is scoped, kept as research notes
- `bit:scope` — frame the WHY and the delivery order as a track of verses
- `bit:plan` — turn each verse into bars, one TDD step and one commit each
- `bit:do` — execute one bar, run its checks, roll the track up
- `bit:commit` — commit a bar's files after asking, and record the hash on the bar
- `bit:check` — audit the finished work against the plan
- `bit:complete` — once a track's work has landed, record its landing commit and file it as completed
- `bit:feedback` — record a correction the moment it lands
- `bit:retro` — read the feedback notes for patterns and write proposals
- `bit:learn` — turn a proposal into a skill or CLI change

It also ships three agents: `bot` (a general session agent that knows the task tools),
`bot-dev` (runs one bar through `bit:do` and lands it), and `ruler` (takes new work through
analyze, scope and plan to an approved plan).

Skills release independently of the binary — edit one, `/reload-plugins`, done. No rebuild.

## TUI

Opens on the kanban board, focused on the top of the Doing column.

| Key | Board | List |
|---|---|---|
| `←`/`→`, `h`/`l` | change column | move focus (page tasks when expanded) |
| `↑`/`↓`, `j`/`k` | move card | move selection / scroll |
| `enter` | open a scrollable modal | expand the detail pane and focus it |
| `tab` | switch view | switch view |
| `?` / `q` | help / quit | help / quit |

It re-reads the project's store on a timer and refreshes only when something changed, so an agent's
edits in another terminal appear without a restart — selection, column, view mode, and open
modal all survive. Colors come from your terminal's ANSI palette, so it matches your theme.

## Storage

```
$XDG_DATA_HOME/bit/
├── main.db                                         registered projects
├── <CODE>/tasks|completed|archive/tasks/<ID>.json + <ID>.md
├── <CODE>/research/<TRACK>/<topic>.json + .md
├── feedback/<TRACK>-NNN.json + .md                 every project, project field
└── retro/<CODE>-<name>.json + .md                  every project, project field
```

Each record is a pair: the `.json` holds its metadata and the `.md` holds its body unchanged.
The `mcp__bit__*` tools and `bp` are the only way in. Bars additionally carry
`phase`/`phase_label`; a reordered track carries an `order` list. IDs are never re-minted —
the next number counts past the highest found across `tasks/`, `completed/`, and
`archive/tasks/`.

This project tracks its own work in the bit store — browse it with `bp task list` or `bp tui`.

## Roadmap

The bit store is the live tracker; this is the summary.

**Up next:**

- **Mark a task approved / refined** — a reviewed scope looks identical to a just-drafted one,
  so "what's ready to work on?" isn't answerable from the board. Needs a model decision first:
  new frontmatter field, another status, or a flag — and how it coexists with `todo`/`doing`/`done`.

**Backlog — needs definition before scoping:**

- **Homebrew packaging** — `scripts/install.sh` assumes a Go toolchain. Open: personal tap vs.
  homebrew-core, and GoReleaser binaries vs. a from-source formula.
- **Search** — quickly target a task by text.
- **Broader filtering** — closer to kanban-md; which dimensions matter is still open.
- **Viewing completed and archived work** — a read-side filter or command for tracks that have
  left `tasks/`.
- **UI polish** — general visual improvements. The `✓` on done rows stays; single-line rows are
  denser than Bubbles' default and that's the point. Open only: whether the other statuses need
  more than the detail pane gives them today.

## Open design questions

- **Whether an index is needed.** No index today, which keeps the files honest. But the TUI
  re-reads the directory on a timer, so the cost is now continuous — at what task count does
  that get felt, and is the answer an index, an mtime check, or a longer interval?
- **Are the board columns fixed?** Hardcoded to To Do / Doing / Done. A project wanting
  `blocked` or `review` has nowhere to say so. Open: whether statuses become project config,
  and what that does to the skills, which assume the three by name.
- **Filtering dimensions.** No tags, assignee, or dates in a task record — "filter by tag" is a
  data-model decision before it's a UI one.
- **How much workflow belongs in the CLI.** Rollup, sign-off, and archiving triggers live in
  the skills, which keeps the CLI to primitives. The cost: the rules aren't enforced and don't
  travel to anyone using the CLI without the skills.

## Inspiration

- [Backlog.md](https://github.com/MrLesk/Backlog.md) — git-native markdown tasks, board UI.
- [kanban-md](https://github.com/antopolskiy/kanban-md) — markdown-backed kanban with filtering.
