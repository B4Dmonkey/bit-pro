# bit-pro

A project-management CLI for LLM-driven development. Git-native, markdown-backed, and
shaped around how an agent and its human actually move work forward: scope it, plan it,
do it, check it.

The binary is `bp`. Work lives in a `.bit/` directory next to your code.

## Why

Jira, Linear, and GitHub Projects are built for a human clicking through a web UI. When an
LLM is doing the work, that's the wrong interface. The agent wants plain text, deterministic
CLI commands, and task state that lives next to the code — not behind an API token.

So: tasks are markdown files in git, the CLI is the primary interface, and `bp tui` is the
human's window onto the same files. The TUI never becomes a second source of truth.

## Install

```
./scripts/install.sh
```

Builds `bp` into your Go bin dir (`$(go env GOBIN)`, else `$(go env GOPATH)/bin`) and
registers this repo as a Claude Code plugin marketplace. That bin dir must be on your
`PATH` — check with `bp --help`.

Needs Go. The agent skills need Claude Code; the CLI itself doesn't.

## Quickstart

```
cd your-project
bp init                                    # prompts for a task ID prefix, creates .bit/
bp task create "Add OAuth login" -d "Why this matters…"
bp task create "Write the token test" -p PREFIX-1 --phase 1 --phase-label "Token exchange"
bp task list
bp task update PREFIX-1.1 -s doing
bp tui                                     # the human view: board + list
```

`bp init` is idempotent — re-run it any time. It sets the prefix, scaffolds `.bit/`, and
keeps the `bit` plugin current in `.claude/settings.json`. It never touches your tasks.

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
| `bp init` | Set the ID prefix, scaffold `.bit/`, sync the plugin |
| `bp task create <title>` | New task. `-p` parent, `--phase`/`--phase-label`, `--after` sibling, `-d` body |
| `bp task read <id>` | Full content. `--body` prints just the markdown, for feeding back to a model |
| `bp task list` | All tasks. `-p <track>` lists one plan, in step order |
| `bp task update <id>` | `-s` status, `-t` title, `-d` body, `--phase`/`--phase-label` |
| `bp task move <bar>` | `--before`/`--after` a sibling — resequence without renaming |
| `bp task complete <id>` | File a signed-off track and its bars under `.bit/completed/` |
| `bp task delete <id>` | Soft-delete into `.bit/archive/tasks/`. `-y` skips confirm, `-f` overrides the guard |
| `bp feedback add <track>` | Record a correction as a note in `.bit/feedback/` |
| `bp tui` | Terminal UI |

Notes on behavior worth knowing:

- **Status is a plain field, not a state machine.** `todo`/`doing`/`done`, set directly.
  Rollup across a plan is workflow logic, left to the skills rather than enforced here.
- **Nothing is destroyed.** `complete` and `delete` both just move files. A relocated ID is
  reserved forever and drops out of its parent's order. A track only relocates once every bar
  is `done` — absolute for `complete`, `--force`-able for `delete`.
- **Order is separate from identity.** A reordered track carries an explicit list of its bar
  IDs, so a plan can be resequenced mid-stream while every ID stays stable.

## Agent skills

`bp init` wires in the `bit` Claude Code plugin, which ships seven skills:

`bit_scope` → `bit_plan` → `bit_do` → `bit_check` is the main loop — frame the WHY, turn each
phase into TDD steps, execute one commit at a time, audit the result. Alongside it,
`bit_feedback` records a correction the moment it lands, `bit_retro` reads those notes for
patterns, and `bit_learn` turns a pattern into a skill or CLI change.

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

It re-reads `.bit/tasks/` on a timer and refreshes only when something changed, so an agent's
edits in another terminal appear without a restart — selection, column, view mode, and open
modal all survive. Colors come from your terminal's ANSI palette, so it matches your theme.

## Storage

```
.bit/
├── config.toml        # prefix = "BIT"
├── tasks/             # live work
│   ├── BIT-1.md
│   └── BIT-1.1.md
├── completed/         # signed-off work
├── feedback/          # correction notes, e.g. BIT-20-001.md
└── archive/
    └── tasks/         # soft-deleted work
```

One markdown file per task, named for its ID. No index — `bp task list` globs and parses.
Frontmatter is deliberately minimal and has grown additively:

```markdown
---
id: BIT-1
title: CLI Bootstrap
status: todo
---
The task body, verbatim.
```

Bars additionally carry `phase`/`phase_label`; a reordered track carries an `order` list.
IDs are never re-minted — the next number counts past the highest found across `tasks/`,
`completed/`, and `archive/tasks/`.

This project tracks its own work in `.bit/` — browse it with `bp task list` or `bp tui`.

## Roadmap

`.bit/tasks/` is the live tracker; this is the summary.

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
- **Filtering dimensions.** No tags, assignee, or dates in frontmatter — "filter by tag" is a
  data-model decision before it's a UI one.
- **How much workflow belongs in the CLI.** Rollup, sign-off, and archiving triggers live in
  the skills, which keeps the CLI to primitives. The cost: the rules aren't enforced and don't
  travel to anyone using the CLI without the skills.

## Inspiration

- [Backlog.md](https://github.com/MrLesk/Backlog.md) — git-native markdown tasks, board UI.
- [kanban-md](https://github.com/antopolskiy/kanban-md) — markdown-backed kanban with filtering.
