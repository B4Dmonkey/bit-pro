---
id: BIT-49.30
title: The README and hierarchy.md describe the v2 store, bp add, bp remove and bp migrate
status: todo
approved: true
phase: 3
phase_label: no .bit/ in sight
---
## **Verse 3**

The README still teaches `bp init`, "work lives in `.bit/` next to your code", seven skills, and a frontmatter Storage section. hierarchy.md still says an album is one `.bit/` directory. This is the last wording pass. It finishes the verse and holds its end-to-end check.

## Scope
Re-grep first: `grep -n '\.bit\|bp init\|seven' README.md hierarchy.md`. BIT-48 adds a `## Claude Code setup` section after `## Install` and leaves the command table, the Quickstart and the `bp init` text to this bar. Don't touch BIT-48's section.
- `README.md`:
  - `:3`, `:7`, `:13-16` (Why): work lives in a central store under `$XDG_DATA_HOME/bit` (default `~/.local/share/bit`), keyed by project code. Drop "next to your code" and "markdown files in git".
  - Quickstart (`:30-43`): `bp add .` (prompts for a project code) replaces `bp init`. One line says an existing v1 project runs `bp migrate` instead. Delete the `bp init` idempotency paragraph.
  - Command table (`:52-67`): remove `bp init`. Add `bp add <path>` (register a project, setting bit up for the machine on first use), `bp remove` (soft-delete a project after confirmation, archiving its uncompleted tracks; take its exact wording from BIT-46's `Short`), and `bp migrate` (copy this folder's v1 `.bit/` into the store and register it). `complete`, `delete` and `feedback add` lose their `.bit/...` paths: "files it as completed", "soft-delete into the archive", "record a note in the shared feedback store".
  - `:82`: "`bp add` sets up the `bit` Claude Code plugin, which ships these skills" with no count (today analyze, check, complete, do, feedback, learn, plan, retro, scope), plus the bot, bot-dev and ruler agents. Update the list under it to match. Leaving the count out keeps it from going stale when BIT-47 adds `bit:commit`.
  - `:103` (TUI): "re-reads the project's store on a timer".
  - Storage (`:107-135`): replace the `.bit/` tree and the frontmatter example with the v2 layout:
    ```
    $XDG_DATA_HOME/bit/
    ├── main.db                         registered projects
    ├── <CODE>/tasks|completed|archive/tasks/<ID>.json + <ID>.md
    ├── <CODE>/research/<TRACK>/<topic>.json + .md
    ├── feedback/<TRACK>-NNN.json + .md   every project, project field
    └── retro/<CODE>-<name>.json + .md    every project, project field
    ```
    Add one sentence saying the `.json` holds metadata and the `.md` holds the body unchanged, and that the `mcp__bit__*` tools and `bp` are the only way in. Drop "No index — `bp task list` globs".
  - `:137`, `:141`: this project tracks its own work in the bit store (`bp task list`, `bp tui`).
- `hierarchy.md:15`, `:23`: an album is "the project, one registered project code" and "the project's store". The tree at `:21+` uses the store layout (`<CODE>/tasks/BIT-1.json + .md`).
- `mcp-notes.md`, `automation-notes.md` and other design notes are left as history (scope).
- Text only, with no test files.

## Change checklist
- [ ] Apply the edits.
- [ ] `grep -n '\.bit' README.md hierarchy.md` shows only the `bp migrate` row and the Quickstart's v1 sentence, which name the v1 `.bit/` on purpose.
- [ ] `grep -n 'bp init\|seven' README.md` finds nothing.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `go run . --help` lists `migrate`, `add` and `remove`, and no `init`, matching the new table

## User verifies
This checks the whole verse in a dev session with no `.bit/` anywhere: BIT-46.11's sandbox, plus `--plugin-dir <v2 checkout>/bit`, never `just install`.
- [ ] In `$SB/proj` (registered with `bp add`, no `.bit/`), run `/bit:scope` on a one-line idea, then `/bit:plan`, then `/bit:do` on the first bar. The transcript never reads, globs or mentions a `.bit` path, the tools carry every write, and the do close-out doesn't tell you to stage `.bit/`.
- [ ] `grep -rn '\.bit' bit/ cmd/serve_mcp.go cmd/task/complete.go` finds nothing. In `README.md`, only the deliberate `bp migrate` mentions remain.
- [ ] Whole slice: a v2 dev session runs scope → plan → do with no `.bit/` in sight, in the skills, the agents, the tool descriptions or the docs.

## Commit (user)
`docs(bit): README and hierarchy describe the v2 store and commands`