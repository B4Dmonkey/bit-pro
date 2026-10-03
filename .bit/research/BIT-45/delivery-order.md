> **Superseded by the track split (banner added 2026-10-01).** The operator's order is **BIT-45 → BIT-46 → BIT-48 → BIT-49 → BIT-47 → BIT-50** (BIT-45 topic `decisions`):
> - Slice 1 is BIT-45.
> - Slices 2–3 are BIT-46 Verses 1–2. `bp init` is deleted in BIT-46 Verse 1, not folded into `add`. BIT-46 Verse 3 adds `bp remove`.
> - The global Claude wiring is BIT-48.
> - Slices 5, 6 and the sweep in 7 are BIT-49 Verses 1–3.
> - Slice 4 (anchors) now comes **last**, after feedback and migrate, not between them, and it's two tracks: BIT-47 (the commit skill and git capture on writes), then BIT-50 (merge-aware completion). Its shape changed too: no capture of head, main head and base on every write for tracks and bars. See the banner on [history-anchors](history-anchors.md).
> - Cutover is no slice and no track. The operator alone decides when v2 is ready and merges the branch.
> - Records carry no project path: the "/path" in slice 3 is out.

# Suggested delivery order for the minimum scope (Q8)

Vertical slices. Each one leaves `v2` building, with tests green and usable through `just run`/`./bin/bp` against a sandbox `XDG_DATA_HOME` ([testing](testing.md)). This is a suggestion for bit_scope, not a decision.

1. **Remove daemon/queue.** Pure deletion ([daemon-removal](daemon-removal.md)). It shrinks `db/`, `tui/`, `cmd/`, `claude/` and the tests before anything is built on them. Product call inside: the TUI "Play?" prompt.
2. **Fresh registry + resolver.**
   - `store.Dir()` → `bit`.
   - A new initial migration: `projects(id, code UNIQUE, path UNIQUE, created_at, …)`.
   - `bp add` as registration (fold in `init`'s Claude wiring, drop config.toml).
   - A `project` resolver (longest registered path) replaces `bitdir` in the CLI and MCP (with the cwd fallback).
   - Store root = `~/.local/share/bit/<code>/`, **still the v1 markdown format**.
   - Usable: register a scratch project and the whole CLI/MCP works centrally.
   - Detail: [project-resolution](project-resolution.md).
3. **JSON metadata + md content pairs.** Change `task/` (`Parse`/`Bytes`, globs, relocate moves pairs), add `created_at`/`updated_at`/`project`/path, and rewrite the frontmatter-asserting tests. Callers are unchanged if `Task.Body` stays. See [json-schema](json-schema.md).
4. **Git anchors on writes:** branch, head, main head, merge-base, session dir, behind an injectable git runner. See [history-anchors](history-anchors.md). This is independent of 5 and can swap places with it.
5. **Top-level feedback + retro.**
   - Move feedback to `~/.local/share/bit/feedback/` with a `project` field.
   - Add MCP `feedback_list`/`feedback_read` and `retro_write`/`retro_list`/`retro_read`.
   - Update the retro and learn skills to use them.
   - See [skills](skills.md).
6. **`bp migrate` + the not-migrated warning** (the resolver's error text). It depends on the final format from 3-5. Trial it on bit-pro and example, since the source is untouched. See [migrate](migrate.md).
7. **Skills/doc sweep + cutover.**
   - Delete the commit-staging lines in do/bot-dev.
   - Reword `.bit/` prose in skills, MCP descriptions, README and plugin.json.
   - Cutover checklist: [testing](testing.md).

A variant: fold 4 into 3, if every record should carry anchors from the first JSON write.

## Separate tracks, not this one
- **Rehearsal fixture rework:** `tools/example` reset/checkpoint rely on git-tracked `.bit/` ([testing](testing.md)).
- **Project/track linking** (multi-repo umbrella), and `--project`/`BIT_PROJECT` overrides.
- **Resolving the squash commit after merge** (`gh`/grep). Capture the branch at done in slice 4, and resolve later.
- **check report into the store**, and the check↔retro mismatch that already exists in v1.
- **opencode support:** MCP without `CLAUDE_PROJECT_DIR` is covered by the resolver fallback. Syncing skills and agents to opencode is separate.
- **TUI → "ui"**, and moving ID allocation into sqlite.