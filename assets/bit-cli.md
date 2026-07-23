# Driving bit through `bit`

The bit_scope, bit_plan, bit_do, and bit_check skills all track their own work inside
`.bit/` by driving the local `bit` CLI. This is the shared command contract they
rely on. Read it once at the start of a session; the individual skills tell you *when* to
run each command, this file tells you *how*.

The single rule that makes the rest safe: **every write goes through `bit`. Never
edit `.bit/tasks/*.md` by hand.** The CLI owns the file format (YAML frontmatter + body);
hand-edits drift from what the tool expects and defeat the point of the tool tracking its
own development.

## The two kinds of task

- A **track** is a top-level task — one whole scope. Its ID has no dot: `BIT-7`.
  Its **body** holds the scope prose (Why / Summary / Verses / Risks).
- A **bar** is a child of a track — one plan step. Its ID is dotted: `BIT-7.3`.
  Its **body** holds that step's detail; its `--phase`/`--phase-label` tag the scope
  phase it serves.

A track and its bars *are* the scope-and-plan pairing. There is no `foo-scope.md` /
`foo-plan.md` filename pairing anymore — the parent link is the structural relationship.

## Commands

Capture minted IDs with `$( )`. `task create` prints the new ID on its own line, and
command substitution strips the trailing newline — so `ID=$(...)` holds exactly the ID.

```bash
# Create a track (a scope). Prints the new track ID.
TRACK=$(bit task create "<scope title>" -d "<scope body>")

# Create a bar (a step) under a track. Prints the new dotted bar ID.
# --phase / --phase-label carry the scope phase the step serves.
BAR=$(bit task create "<step name>" --parent "$TRACK" \
        --phase 1 --phase-label "<phase label>" -d "<step body>")

# Read a body only — no header line. Round-trips byte-for-byte, so this is the
# read side of a read → edit → write-back refine.
bit task read "$ID" --body

# Read the one-line summary: <ID>\t<status>\t<title>[\tphase N — label]
bit task read "$ID" | head -1

# Rewrite a body wholesale. -d "$(...)" is a proven byte-safe whole-body write
# (backticks, $, code fences, and --- lines all survive).
bit task update "$ID" -d "<new body>"

# Set status. Status is a plain field you set directly — there is no state machine.
# Use exactly one of todo | doing | done.
bit task update "$ID" -s todo|doing|done

# You can change body and status (and title/phase) in one call:
bit task update "$TRACK" -d "<new body>" -s done

# List one track's bars only, in step order: <ID>\t<status>\t<title>\tphase N — label
bit task list --parent "$TRACK"

# List everything. Tracks are the rows whose ID has no dot; bars have a dotted ID.
bit task list

# Resequence a bar within its track. Rewrites the track's order list so every surface
# that reads it (task list --parent, the TUI list, the board, bit_do's next-step resume)
# reflects the new position. Pass exactly one of --before / --after; the anchor is a sibling.
bit task move "$BAR" --before "$SIBLING"
bit task move "$BAR" --after  "$SIBLING"
```

A track owns an explicit ordered list of its bars, and the CLI is the only thing that
writes it: `create` **appends** to the list, `move` **rewrites** it, `delete` removes from
it. That's why order never depends on filesystem glob order or on hand-editing — the same
"every write goes through `bit`" rule that protects the file format also protects the
ordering.

## Writing a body from the shell

A body is multi-line markdown. The reliable way to author or substantially edit one is to
build the text in a file, then `-d "$(cat body.md)"`. This is the path for a scope body, a
step body, or any multi-line rewrite.

For a small, surgical change to a stored body — toggling one phase checkbox, fixing a line —
read it out, stream-edit, and write it back:

```bash
bit task read "$ID" --body | sed 's/- \[ \] Verse 1/- [x] Verse 1/' > body.md
bit task update "$ID" -d "$(cat body.md)"
```

Don't try to hand-edit a body held in a shell variable — it's multi-line text, so route it
through a file or a stream editor. Either way it round-trips byte-for-byte (minus a trailing
newline, which `$( )` strips and which never matters for markdown).

## Gotchas

- **Status is stored verbatim — spelling matters.** There's no validation: `task update -s
  doen` succeeds and stores `doen`. Rollup logic keys on the exact words (`all bars "done"
  → track "done"`), so a typo silently breaks it — a bar that reads `doen` will never count
  as done and the track will never roll up. Always pass exactly `todo`, `doing`, or `done`.
- **Deleting or archiving a task reserves its ID — it isn't freed.** `task delete` and
  `task archive` both *relocate* the file into `.bit/archive/` instead of destroying it, and
  `NextID`/`NextChildID` count `archive/` when choosing the next number. So a removed ID is
  never re-minted onto a different task, and older commit messages or notes that reference it
  stay valid. Dropping a bar mid-plan with `task delete`/`task archive` is safe for this
  reason — the file is recoverable on disk and its ID stays put.

## Rollup is skill logic, run through the CLI

The CLI does **not** cascade status from a bar up to its track — setting a bar `done`
leaves the track untouched. Rolling a track up is the skill's job, done with ordinary
commands: read the bars with `task list --parent`, decide the track's status, then
`task update <track> -s …`. bit_do owns that logic; see its skill for the exact rule.
