# bit

A project-management CLI built for LLM-driven development workflows — git-native,
markdown-backed, and structured around the way an LLM (and its human) actually move
work forward: scope it, plan it, do it, check it.

## Why

Existing PM tools (Jira, Linear, GitHub Projects) are built for humans clicking through
a web UI. When an LLM is doing the work, that's the wrong interface: the agent wants to
read and write plain text, wants deterministic CLI commands it can call without a
browser, and wants task state to live next to the code it's changing, not behind an API
token in a SaaS product.

[Backlog.md](https://github.com/MrLesk/Backlog.md) and
[kanban-md](https://github.com/antopolskiy/kanban-md) both point at the right shape:
tasks as markdown files, tracked in git, with a CLI and a terminal UI on top. `bit`
takes that shape and wires it specifically to an LLM-first workflow — the CLI is the
primary interface (for Claude, or any agent), and the TUI is the human's window into
the same data.

## Design principles

The CLI is the source of truth and stays easy for an LLM to drive: predictable
subcommands and flags, scriptable, non-interactive by default. `bit tui` is a second,
human-facing view over the same markdown files — it never becomes a second source of
truth. Anything the TUI can show, the CLI can already answer.

## Install

```
just install
```

This builds `bit` into your Go bin dir (`$(go env GOBIN)`, falling back to `$(go env
GOPATH)/bin`) so you can run `bit` from any directory. That dir must be on your `PATH` —
check with `bit --help`; if it's not found, add it (e.g. `export
PATH="$(go env GOPATH)/bin:$PATH"`).

## Quickstart

```
cd your-project
bit init                          # prompts for a task ID prefix, creates .bit/
bit task create "Add OAuth login" -d "Why this matters…"
bit task create "Write the token test" -p PREFIX-1 --phase 1 --phase-label "Token exchange"
bit task list
bit task update PREFIX-1.1 -s doing
bit tui                           # the human view: list + board
```

`bit init` is idempotent — re-running it in an initialized project offers the existing
prefix as a default and re-seeds the agent skills without touching your tasks.

## Storage

A project is a `.bit/` directory at the repo root, created by `bit init`:

```
.bit/
├── config.toml        # prefix = "BIT"
├── tasks/             # live work
│   ├── BIT-1.md
│   └── BIT-2.md
└── archive/           # finished and deleted work; these IDs are never re-minted
    └── BIT-3.md
```

Tasks are flat — one markdown file per task, named for its ID, no per-track
subdirectories. IDs are `<PREFIX>-<N>`, where the prefix is captured once by `bit init`
and `N` is one past the highest existing task's number. There's no index: `bit task
list` globs the directory and parses each file. Frontmatter is deliberately minimal —
only what CRUD needs — since markdown+frontmatter can grow additively as the right
shape becomes clearer:

```markdown
---
id: BIT-1
title: CLI Bootstrap
status: todo
---
The task body, verbatim.
```

Frontmatter has grown additively as the model firmed up: a bar carries `phase` /
`phase_label` tying it to a scope phase, and a track that has been reordered carries an
explicit `order` list of its bar IDs. The ID is stable identity; ordering lives in that
list, so a plan can be resequenced (`task move`, `task create --after`) without renaming
anything. A track that's never been reordered has no `order` and falls back to ID order.

## Features

What `bit` does today. It's an MVP — usable end to end for the workflow it was built
for, and still moving.

**Tasks from the command line.** `bit task create/read/list/update/move/archive/delete`
covers the full lifecycle. Every command is non-interactive by default (`delete` takes
`-y`), so an agent can drive the whole thing without a prompt to answer. `bit task read
--body` prints just the markdown body for feeding straight back into a model.

**Plans nest under the scope that owns them.** A dotted ID makes a task a child:
`BIT-2.5` is the 5th step of `BIT-2`. `--parent` mints the child (and refuses if the
parent doesn't exist), `--phase`/`--phase-label` tag a step with the coarse chunk of the
scope it serves, and `bit task list --parent BIT-2` shows just that plan. The parent is
readable straight out of the ID — no index, no lookup, and `ls` shows the tree. See
[hierarchy.md](./hierarchy.md) for the vocabulary (album / track / verse / bar).

**Plans resequence without renaming.** A track carries an explicit ordered list of its
steps, so `bit task move --before/--after` and `bit task create --after` reorder a plan
mid-stream while every ID stays stable. IDs are identity; order lives in the parent.

**Archive instead of destroy.** `bit task archive` relocates a finished track and its
steps into `.bit/archive/`, so the list, board, and TUI show only live work. `bit task
delete` reuses the same primitive — a mistaken delete is a file move, not a loss.
Relocating reserves the ID forever (it's never re-minted) and drops the step from its
parent's order, so the sequence stays honest. A track only relocates once every step is
`done`, with `--force` to override.

**Status is a plain field, not a state machine.** Tasks have `todo`/`doing`/`done`, and
changing one is a direct write (`bit task update -s doing`). There's no transition graph
to fight; rollup across a plan's steps is workflow logic, deliberately left to the agent
skills rather than enforced by the CLI.

**A terminal UI for the human.** `bit tui` opens a list with a detail pane that renders
the task body as markdown, and `tab` flips to a kanban board (To Do / Doing / Done).
`enter` on a card floats a scrollable modal with its full body, so you can read a task
without leaving the board. Focus moves with `←`/`→`, `?` toggles full help, `q` quits.

**The TUI stays live.** It re-reads `.bit/tasks/` on a short timer and refreshes when
something actually changed, so edits an agent makes in another terminal appear without a
restart — and your selection, column, view mode, and open modal survive the refresh. A
burst of writes between ticks collapses into one refresh, and a failed read holds the
last good view instead of flashing an error.

**Chrome that follows your terminal theme.** Colors come from the terminal's ANSI
palette rather than fixed hex values, so `bit` matches however you've themed your
terminal. Focus is unmistakable: the focused pane or column title renders inverted.

**Agent skills ship with the binary.** `bit init` seeds a `bit_scope` → `bit_plan` →
`bit_do` → `bit_check` skill set plus a CLI contract doc into the project's `.claude/`
directory, idempotently, so the agent workflow works in any repo the moment it's
initialized. (These are Claude Code skills today; the CLI itself has no agent-specific
dependency.)

This project tracks its own scoping and planning in `.bit/tasks/` — browse it with `bit
task list`, or `bit tui`.

## Roadmap

The live tracker is `.bit/tasks/` (`bit task list`); this is the human-readable summary
of what to pick up next.

**Up next:**

- **Rename the binary `bit` → `bp`** (`BIT-15`, scoped, not started) — `bit` collides with
  other tools; `bp` (bit-pro) is unambiguous. Binary name only: the `.bit/` directory and
  `BIT-` ID prefix don't change. The Justfile, the Cobra root command, and the embedded
  skill assets move in lockstep.
- **Mark a task approved / refined** — a scope or plan that's been reviewed and is ready to
  pick up looks identical to one that was just drafted, so "what's actually ready to work
  on?" isn't answerable from the list or the board. Add that signal, and let the TUI filter
  on it. Needs a model decision first: a new frontmatter field, another status value, or a
  separate flag — and how it coexists with `todo`/`doing`/`done`.
- **Ship the skills as a Claude Code plugin** (not scoped) — the skills are embedded in the
  binary today, so editing one costs a rebuild and a re-init, and there's no versioned way to
  distribute them across repos. Package them as a plugin instead, with `bit init` wiring the
  plugin into the project. The minimum bar is porting the existing asset set over unchanged;
  `assets/` and the `//go:embed` then go away, making the plugin the only source of skills.
  Worth doing on the packaging win alone — the open question to test first is whether a
  running session picks up plugin edits automatically, but even without reload this beats the
  rebuild loop. Note the tradeoff it creates: CLI and skills version independently, so `bit`'s
  command surface becomes an API a separately-released client consumes. Overlaps with
  `BIT-15`, which currently assumes the embedded assets move with the rename.
- **Retro notes, and a `bit_retro` skill to evaluate them** (not scoped) — when a run goes
  sideways today the fix is to stop, repair the scope and plan, and carry on. That repair is
  lossy: the broken plan is gone, so a finished track that was rewritten mid-flight looks
  identical to one that went smoothly, and there's nothing left to learn from. Split it in
  two. **Capture** is a `bit` subcommand that appends an observation to the track — which bar
  it happened at, what the plan said, what the work actually required, before/after commit
  hashes — called from `bit_do` and `bit_plan` when a correction lands, so it doesn't depend
  on anyone invoking a skill mid-run. Notes hang off the track, not the bar, because
  replanning renumbers bars and would orphan them. **Evaluation** is a new `bit_retro` skill
  (the underscore-family port of `bit-retro`) that reads the track, its notes, and the plan
  against what actually changed, then routes each note to the stage that should have caught
  it — a missing exemplar back to `bit_plan`'s citations, an unasked question into
  `bit_scope`'s checklist. Needs decisions before scoping: whether notes are a markdown
  section or frontmatter, and the cause taxonomy — small and closed enough to count across
  cycles, with an explicit "not preventable" bucket so the checklists don't bloat with
  defensive questions. Aggregating notes across repos and clients wants a store above
  `.bit/`; out of scope here.

**Backlog — needs definition before scoping:**

- **Homebrew packaging** — `just install` assumes a Go toolchain, which is the wrong ask for
  anyone who just wants the tool. Ship a formula so `brew install` works. Needs decisions
  before scoping: a personal tap or homebrew-core, and whether release binaries come from
  GoReleaser (which would also cover Linux) or a from-source formula. Wants the rename to
  land first, so the formula names the final binary.
- **Jump straight to the board** — a way to open the TUI directly on the kanban board
  instead of landing on the list and tabbing over. Shape undecided (a flag on `bit tui`, a
  separate `bit board` command, or config); no details yet.
- **Search** — quickly target a task by text.
- **Broader filtering** — closer to Backlog.md; which dimensions matter is still open (see
  *Open design questions* → Filtering dimensions).
- **Viewing the archive** — a filter (or command) to surface archived tracks. Archiving
  itself is built; a restore/view command was deferred (YAGNI), so this is the read-side
  that pairs with it.
- **Vim keys for focus, and a home for paging** — `h`/`l` should move focus like `←`/`→` in
  the list view. They don't today: the arrows work, and `h`/`l` fall through to the list's
  own paging. The fix is one keymap rework — intercept `h`/`l` alongside `KeyLeft`/`KeyRight`
  in the focus handler, and rebind the list's `NextPage`/`PrevPage` to `ctrl+f`/`ctrl+b` so
  paging keeps an explicit home. (The board modal already scrolls on `h`/`j`/`k`/`l`.)
- **UI polish** — general TUI visual improvements. The one open thread with history behind
  it is the list-row delegate (`tui/delegate.go`): rows are single-line, which is denser than
  Bubbles' default but has no room for a status column, so status shows only as a `✓` on done
  rows and in full in the detail pane. Whether to bring status back as a right-aligned
  column, go back to two-line rows, or leave it is undecided — `bit task read BIT-4.12
  --body` has the fuller record.

## Open design questions

Genuinely unresolved — what future scoping needs to answer, not decisions already made.

- **Whether an index is needed.** `bit task list` globs `.bit/tasks/` and parses every file;
  there's no index. Fine at this size, and it's what keeps the storage format honest (the
  files are the state, with nothing to fall out of sync). But the TUI now re-reads the whole
  directory on a timer, so the cost is paid continuously rather than once per command — the
  question is what task count makes that felt, and whether the answer then is an index, a
  cheaper change-check (mtime/hash before parse), or a longer interval.
- **Are the board columns fixed?** They're hardcoded to To Do / Doing / Done, matching the
  three status values. A project wanting `blocked`, `review`, or a WIP limit has nowhere to
  say so. Open: whether statuses become project config (in `.bit/config.toml`, which today
  holds only the prefix) and what that does to the agent skills, which currently assume the
  three values by name.
- **Filtering dimensions.** kanban-md's filtering is the thing to match, but the frontmatter
  has no tags, no assignee, and no dates — so "filter by tag" is a data-model decision before
  it's a UI one. Open: which dimensions earn a field, and whether they're fixed keys or
  free-form.
- **How much of the workflow belongs in the CLI.** Status rollup, sign-off, and archiving
  triggers all live in the agent skills right now, which keeps the CLI to primitives and lets
  the workflow change without a release. The cost is that the rules aren't enforced and don't
  travel to anyone using the CLI without the skills. Open: whether any of it graduates into
  the binary.

## Inspiration

- [Backlog.md](https://github.com/MrLesk/Backlog.md) — task view + kanban board UI,
  milestones/docs concept, git-native markdown tasks.
- [kanban-md](https://github.com/antopolskiy/kanban-md) — markdown-backed kanban with
  filtering.
