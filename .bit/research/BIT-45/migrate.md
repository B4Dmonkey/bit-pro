# `bp migrate` and the not-migrated warning (Q5)

**Checked:** conversion plus copy, idempotency, verification, rollback, completed/ and archive/, feedback and retro going top-level, and the warning. This rewrites the first-pass note to match the operator decisions:
- a fresh db
- JSON metadata + markdown content pairs
- feedback and retro in top-level dirs
- hard cutover
- `.bit/` removed from bit-pro afterwards

## Target layout (from the decisions)
```
~/.local/share/bit/bit.db
~/.local/share/bit/<code>/{tasks,completed,archive/tasks}/<ID>.{json,md}
~/.local/share/bit/<code>/research/<TRACK>/<topic>.{json,md}
~/.local/share/bit/feedback/<TRACK>-NNN.{json,md}      # project field
~/.local/share/bit/retro/<name>.{json,md}              # project field
```
**Open:** the `<code>` dir casing. IDs are uppercase (`task.NormalizeID`). The operator's example is `bit/bit`. APFS is case-insensitive, so choose once, e.g. `strings.ToLower(code)`, and derive the dir only in the resolver.

## Source (per project, from bit-pro's data)
- `config.toml` holds `prefix` → the `projects.code` row. The file itself is dropped.
- `tasks/`, `completed/`, `archive/tasks/`: yaml frontmatter + body → JSON (the same keys) + the `.md` body, byte-identical.
- `feedback/*.md` has no frontmatter. Track and seq are parsed from the filename.
- `research/<TRACK>/*.md` has no frontmatter. Track and topic come from the path.
- `retro/*-proposals.md`, if present. bit-pro has none.
- **Unknown files** (anything else under `.bit/`): don't drop them silently. Copy them raw into `<code>/unmigrated/` and list them, or refuse. Decide which.
- No v1 registry rows carry over, since the db is fresh. `migrate` therefore runs **per project from its own dir**. It can't enumerate v1 projects (and a client project was never registered in v1 anyway).

## Steps
1. Resolve the source. Use the repo root's `.bit/`. If cwd is a Claude worktree, use the main checkout's `.bit/`, because v1 kept state there (see memory: worktree bp writes to the main checkout).
2. Get the code from `prefix`, normalised. If the project is **already registered with a store dir**, stop with "already migrated" (idempotent no-op). If the code is taken by a different path, refuse.
3. Build the whole tree in a temp dir **inside** `~/.local/share/bit/`, so the final rename stays on one filesystem.
   - Timestamps: `git log --follow` per file when `.bit/` is git-tracked, else mtime.
   - Git anchors: empty, plus `migrated: true`.
   - Top-level feedback and retro files go into their own staging dirs.
4. **Verify before commit:**
   - Per-kind counts match.
   - Every source ID has a JSON with equal metadata.
   - Every `.md` body is byte-equal to the source body.
   - `task.Store.List()` on the new root succeeds.
5. Commit: rename staging → `<code>/`, move the feedback/retro files in, and insert the `projects` row in one db transaction. Insert the row last, so a crash leaves no registered-but-empty project.
   - Feedback and retro files for this project that already exist in the top-level dirs mean a partial prior run. Compare them and skip if identical.
6. Print next steps; bp itself runs no git.
   - bit-pro, where `.bit/` is tracked: `git rm -r .bit`.
   - Other projects, where `.bit/` is untracked: `rm -rf .bit` or rename it.

## Rollback
- **Migrate never modifies the source.** Rollback is: delete `<code>/`, delete this project's top-level feedback/retro files (by `project` field), and delete the db row.
- Could this be `bp migrate --undo`? YAGNI until needed.
- Before cutover, v1 keeps using the untouched `.bit/`. That makes a trial migrate on live projects safe (see [testing](testing.md)).

## completed/ and archive/
Migrate them as-is into the same dir names. They matter for **ID reservation** (`store.go:637-650` scans all three dirs) and `trackExists`. Dropping `archive/` would let old IDs be re-minted. See [json-schema](json-schema.md).

## Not-migrated warning
- **Condition:** resolution finds no registered project for `dir`, **and** a `.bit/` exists at `dir` or an ancestor.
  - Resolution has no walk-up, so check ancestors only on this failure path.
  - This is the only state where it fires. A registered project with a leftover `.bit/` gets a softer "you can remove .bit/" hint, or nothing.
- **CLI:** `cmd/root.go execute()` already prints a post-command stderr notice (plugin behind) and skips `bit.quiet` commands (`tui`, `serve mcp`). Since resolution fails anyway, the simplest option is that **the resolver's error text is the warning**: "found .bit/ but this project isn't registered — run `bp migrate`". That needs no separate notice machinery.
- **MCP:** stderr is invisible to the agent, so:
  - (a) each tool returns the same error text (resolution is per call), and
  - (b) optionally, `mcp.ServerOptions.Instructions`, computed at start from `CLAUDE_PROJECT_DIR`/cwd.

  (a) alone is enough for the minimum.
- Only v2 warns. v1 on `main` is untouched (operator decision).
