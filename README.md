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

## Vision

`bit` is a CLI-first project-management tool for tracking tasks — create, read,
update, delete — with a terminal UI (a list view and a kanban board) over the same
data:

- **CLI** — create, read, update, and delete tasks; this is the primary interface,
  built for an LLM (or any agent) to drive directly.
- **`bit tui`** — the human view: a task list and a kanban board over the same data,
  with filtering (by epic, status, tag, whatever turns out to matter) — this is the
  part directly inspired by Backlog.md's UI and kanban-md's filtering.

The CLI is the source of truth and must stay easy for an LLM to drive: predictable
subcommands and flags, scriptable, non-interactive by default, structured output where
it helps. The TUI is a second, human-facing view over the same underlying data — it
never becomes a second source of truth.

## Install

```
just install
```

This builds `bit` into your Go bin dir (`$(go env GOBIN)`, falling back to `$(go env
GOPATH)/bin`) so you can run `bit` from any directory. That dir must be on your `PATH` —
check with `bit --help`; if it's not found, add it (e.g. `export
PATH="$(go env GOPATH)/bin:$PATH"`).

## Storage

A project is a `.bit/` directory at the repo root, created by `bit init`:

```
.bit/
├── config.toml        # prefix = "BIT"
└── tasks/
    ├── BIT-1.md
    └── BIT-2.md
```

Tasks are flat — one markdown file per task, named for its ID, no per-epic
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

## Status

This project is intentionally not fully specified yet — the plan is to build it by
feel, scoping and planning one increment at a time rather than designing the whole data
model up front. The rough sequence of scopes:

1. ✅ **Bootstrap** (`BIT-1`) — a runnable, installable Cobra CLI with a
   `just`-based task runner and a `bit init` command.
2. ✅ **Simple task management (CRUD)** (`BIT-2`) — the data model: create, read,
   update, delete, and list tasks against the directory `bit init` creates,
   including scope docs (like this project's own) living as first-class tasks
   inside `.bit/` instead of loose root files.
3. **Status state machine** — decided against. CRUD gives tasks *a* status field;
   status changes stay a direct, unvalidated write (`bit task update -s`), and
   rollup across a track's bars is logic the `bit_do` skill owns, not a
   CLI-enforced transition graph.
4. ✅ **Explore UI (list)** (`BIT-4`) — a list + detail view in a terminal.
5. ✅ **Explore UI (board)** (`BIT-5`) — the kanban board view.

All of this project's own scoping and planning now lives in `.bit/tasks/` (browse
it with `bit task list`) rather than root-level markdown files — see the
`.claude/skills/bit_scope` / `bit_plan` / `bit_do` skills for how.

## Roadmap

What's scoped, committed-next, and backlogged. The live tracker is `.bit/tasks/`
(`bit task list`); this is the human-readable summary of what to pick up next.

**In progress (scoped):**

- **Install + portable init** (`BIT-7`) — bit installs locally (`just install`) and
  `bit init` seeds the `bit_*` skills + `bit-cli.md` into any project, idempotently, so
  the skills can drive bit in any repo. The unlock for using bit outside this repo.

**Committed next — not yet scoped:**

- **Spike: how to handle reordering** — the CLI is append-only today. `task create` mints
  the next dotted ID at the end of a track, and bit_do executes bars in creation order, so
  a step can't be inserted mid-plan without deleting and recreating everything after it.
  Spike the options — an `--after <id>` flag on create, a `task move` verb, or an explicit
  order field in frontmatter — and what each costs the dotted-ID scheme. De-risk before
  scoping the feature.
- **Fix the detail-pane quit bug** — small; spec'd under *Cleanup & known issues* below.
- **Archiving** — move a completed track to `.bit/archive/`, hidden from the TUI/list by
  default, to cut the noise. Viewing/restoring the archive is deliberately deferred
  (YAGNI for now) — but the storage layout should keep it cheap to add.
- **Board modal** — toggle a full-detail view for the selected card, so a `todo` item can
  be inspected without leaving the board.

**Backlog — needs definition before scoping:**

- **Search** — quickly target a task by text.
- **Broader filtering** — closer to Backlog.md; which dimensions matter is still open (see
  *Open design questions* → Filtering dimensions).
- **Approved / sign-off** — mark a track or bar reviewed-and-ready-to-pick-up, and filter
  on it. Needs a model decision first: a new frontmatter field, a status value, or a
  separate flag — and how it coexists with `todo`/`doing`/`done`.
- **Viewing the archive** — a filter to show archived tracks; paired with archiving above,
  not built yet.
- **UI polish** — general TUI visual improvements.

## Open design questions

These aren't answered yet — they're what future scoping work needs to resolve, not
decisions already made:

- ✅ **Relationships in frontmatter — resolved.** A task points at a parent via a
  dotted ID (`BIT-2.5` is the 5th step of `BIT-2`), assigned with `--parent` and
  validated (a missing parent refuses to mint). See
  [hierarchy.md](./hierarchy.md) for the vocabulary this introduced.
- **Whether an index is needed.** `bit task list` currently globs and parses every
  file. Fine at this size; an open question once the TUI wants fast filtering over a
  real backlog.
- **Deletes should eventually be soft.** `bit task delete` removes the file outright.
  The direction is for a delete to move the task into a folder instead, so it's
  recoverable — which also makes a task deleted out from under its children a survivable
  mistake rather than a lost one. What that folder is called, whether `bit task list`
  hides its contents, and whether there's a restore command are all undecided. Not
  urgent: nothing here has enough data yet to regret deleting.
- **Status/board model.** What are the kanban columns, and are they fixed or
  configurable per project?
- **Filtering dimensions.** By epic, tag, status, assignee — kanban-md's filtering was
  called out as something to match, but which dimensions matter hasn't been decided.

## Cleanup & known issues

A running list of rough edges and deferred work from the implementation — things with a
known right answer that just aren't done yet, kept separate from the still-undecided design
questions above.

- **Can't quit from the detail pane.** Once focus is on the detail pane (after `→`), every
  key routes to the viewport for scrolling — including `q` and `ctrl+c` — so nothing quits
  the TUI; you have to `←` back to the list first. The fix is to let the quit keys through
  before the focus handler forwards the message to the viewport.
- **`h`/`l` should move focus like `←`/`→`.** Vim-style editors bind `h`/`l` to left/right,
  so the focus keys ought to accept them as aliases. Deferred deliberately — focus works on
  the arrows today, and `h`/`l` currently fall through to the list's own paging. When picked
  up: intercept `h`/`l` alongside `KeyLeft`/`KeyRight` in the focus handler (which also stops
  them paging the list).
- **Paging should move to `ctrl+f`/`ctrl+b`.** The list's default keymap pages on `h`/`l`
  (also `f`/`b`, `u`/`d`, `pgup`/`pgdn`). This is the same keymap rework as the `h`/`l` item
  above: once `h`/`l` become focus aliases, list paging needs an explicit home — rebind the
  list's `NextPage`/`PrevPage` to `ctrl+f`/`ctrl+b`.
- **Revisit the list-row rendering (track/bar/verse delegate).** The custom delegate that
  renders the list (`tui/delegate.go`, added in `9442a7d` "feat(tui): render tracks, bars,
  and verses in the list") is kept but not settled — it's liked for now, but not quite what
  was originally in mind, so it may change. What it does today: one line per row (denser than
  Bubbles' default two-line title+description), plain IDs bold at the left edge for tracks,
  dotted IDs indented two columns and dimmed for bars, each phased bar's verse (`phase N —
  Label`) shown faint/italic after the title, and the selected row marked with `▎` in the
  accent colour (`99`, matching the focused-pane border). Rows are clipped with
  `MaxWidth(m.Width())` so a long title can't wrap and break the one-line layout. Known
  trade-offs to reconsider if revisited: going single-line **dropped the per-row status
  field** (`todo`/`done`) the default delegate used to show — status now lives only in the
  detail pane; the dim/faint/italic styling and the `245`/`99` colours are a first cut (same
  "flavor is deferred polish" note as the Step 9 border accent). Likely directions: bring
  status back as a right-aligned column, or restore a two-line row, or refine the styling —
  none decided. See `BIT-4.12` (`bit task read BIT-4.12 --body`) for the fuller record.

## Inspiration

- [Backlog.md](https://github.com/MrLesk/Backlog.md) — task view + kanban board UI,
  milestones/docs concept, git-native markdown tasks.
- [kanban-md](https://github.com/antopolskiy/kanban-md) — markdown-backed kanban with
  filtering.
