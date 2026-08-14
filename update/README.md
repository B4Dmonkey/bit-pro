# Uppercasing task IDs

A one-time migration that rewrites every task ID in a `.bit/` project to uppercase, so
`BIT-20` is the only spelling that exists on disk.

This is not a product feature. There is no `bp doctor`, nothing detects an un-migrated
project, and nothing repairs one on startup. The operator names the projects and runs the
script; that is the whole mechanism.

## Why it matters

Task IDs used to be compared case-sensitively, so a lowercase argument silently corrupted
state — `bp task complete bit-20` filed a track while leaving its bars behind, and
`bp task create --parent bit-1` overwrote an existing bar with a different task's contents.
Both exited 0 with no output. Once a project is fully uppercase, those wrong-case paths are
unreachable in it.

Run this on every `.bit/` project on the machine. A project that is skipped stays exposed.

## Running it

```
usage: normalize.sh <project-dir>...
```

The arguments are the directories that *contain* `.bit/` — not the `.bit/` paths themselves.
Pass as many as you like; each is migrated independently.

```bash
bash update/normalize.sh ~/Developer/bit-pro ~/Developer/evus ~/Developer/marketplace
```

A successful run prints nothing and exits 0 — the confirmation is the `git status` diff, not
the script's output. A directory with no `.bit/` inside it is an error, not a silent skip: the
script reports it and exits non-zero, so a mistyped path can't look like a successful run.

Requires macOS / BSD `sed` (the in-place edits use `sed -i ''`).

## Before you run it

Make sure each project's git tree is clean.

A case-only rename is invisible to git on a case-insensitive filesystem — a plain `mv` leaves
`git status` completely empty, so the change can't be reviewed or committed. The script uses
`git mv --force` in any project that is a git working tree, which does register the rename.
Starting clean is what makes that diff readable afterwards.

## What it rewrites

Five carriers of an ID:

1. **Task filenames** in `.bit/tasks/`, `.bit/completed/`, and `.bit/archive/tasks/`
2. **`id:` frontmatter** in every one of those files
3. **`order:` entries** — the child list inside a track's frontmatter
4. **Feedback note filenames** in `.bit/feedback/` (`<TRACK>-NNN.md`)
5. **The `prefix`** in `.bit/config.toml`

All three task directories matter, not just `tasks/` — a relocated ID is still reserved, and
the next-ID scan reads `completed/` and `archive/tasks/` too.

Two things it deliberately leaves alone:

- **Body prose** — a step body citing `BIT-11.4` keeps its original case.
- **Git commit messages** — history is not rewritten.

Nothing parses either one, so a stale citation there is cosmetic.

Running it twice is safe: an already-uppercase project comes back untouched.

## After

- `bp task list` in each project shows every ID uppercase, with bars still nested under
  their tracks.
- `git status` shows the renames, but in two halves — `git mv` stages the rename while the
  in-file edits stay unstaged:

  ```
  RM .bit/tasks/bit-1.md -> .bit/tasks/BIT-1.md
   M .bit/config.toml
  ```

  Stage everything before committing, or you will commit the renames and leave the `id:`,
  `order:`, and `prefix` rewrites behind — a half-migrated project:

  ```bash
  git add -A
  git commit -m "chore: uppercase task IDs"
  ```

## Testing the script

```bash
bash update/normalize_test.sh
```

It builds throwaway projects in a temp directory and asserts each carrier, plus idempotency,
the error cases, and git visibility. Exits 0 when everything passes.
