# Operator decisions

These are operator decisions, not research findings, unless an entry is marked "Claude default". The first set came from the operator's answers to the fresh-eyes review of the v2 plan (2026-09-30). The second came from the answers to the consistency review (2026-10-01). The third is the readiness pass (2026-10-01).

# 2026-09-30

## The split of the original BIT-46
The original BIT-46 was split three ways:
- **BIT-46** keeps the registry, the resolver, the central store and the JSON + `.md` record format (the old Verses 1–2), without the Claude wiring.
- **BIT-48** takes the global Claude wiring and removing `bp init`. (2026-10-01: deleting `bp init` moved back to BIT-46 Verse 1. See below.)
- **BIT-49** takes shared feedback and retro, `bp migrate` and the `.bit/` sweep (the old Verses 3–5). It also owns the git helper, because migrate is its first user.

## Project paths
- Project paths are stored only in the db (`projects.path`).
- Records don't keep their own copy of the project path, because the db holds it.
- Re-pointing a moved repo is out of scope. It's a future consideration, not part of the MVP.

## Git fields in the record format (Verse 2)
- Records carry a branch field and a `commit` field. (2026-10-01: these apply to tracks and bars. Other kinds carry a `commits` list. See below.)
- On a track, `commit` is the track's landing or merge commit. It's the anchor that matters most.
- On a bar, `commit` is a best-effort hash of the bar's own commit, and it's allowed to break. When completion detects a squash merge, it also points the track's bars at the squash commit.
- BIT-47 fills these fields, and BIT-49's migrate sets them to HEAD at migration. (2026-10-01: completion moved to BIT-50.)

## Cutover
The operator alone decides when v2 is ready and merges the branch. No track gates cutover.

## Confidentiality (applies to BIT-49; see BIT-49 topic `decisions`)
- `feedback_list` returns only the current project's notes by default.
- Retro file names include the project code.
- learn's description gets rewritten.

# 2026-10-01

## `bp init` is deleted in Verse 1
- `cmd/init.go` depends on `bitdir` and `task.Config`/`SaveConfig` (`cmd/init.go:10,39,75`), and Verse 1 removes both. So Verse 1 deletes the `bp init` command.
- The per-project wiring helper `writeClaudeWiring` (in `cmd/init.go` today) stays, because `bp add` still calls it. BIT-48 folds the global wiring step into `add` in its place.
- Files that moved to BIT-48: `claude/sync.go`, `claude/settings.go`, `claude/plugin.go`, the `writeClaudeWiring` helper, and the plugin-behind notice text in `cmd/root.go` (:119). BIT-46 still changes `cmd/root.go` where it resolves the project (`pluginState` at :27, `bitdir.Resolve()` at :136).

## `bp add` order of checks
- `bp add` checks "already registered" before "has `.bit/`". A registered project is a no-op that says so.
- Re-running `bp add` or `bp migrate` doesn't set up the wiring. The wiring is global, user-scope, and owned by BIT-48.
- `bp add` refuses a removed project's code when it's asked from a different folder. The code still belongs to the removed project, so the operator revives that project or picks another code.

## `bp remove` (Verse 3)
- It soft-deletes the project: a removed flag on its `projects` row. There is no hard delete.
- `bp add` on a removed project's path flips the flag back and revives it.
- It archives the project's remaining work: every track that isn't completed, including tracks whose bars are all done but which were never completed.
- It asks for confirmation first, with a note listing any outstanding (not-done) work.
- In a removed project's folder, the resolver says the project was removed and points to `bp add`.
- It's CLI-only.
- The project's research, feedback and retro are kept.
- It's in BIT-46 because it's registry and store work.
- **Claude default:** archived work goes through the existing archive path `archive/tasks/` (`Relocate`, as `bp task delete` uses), so its IDs stay reserved.
- **Claude default:** `bp list` leaves out removed projects.
- **Claude default:** `bp remove` runs from the project's folder through the resolver, with no path argument.

## The fixture
- **Claude default:** the `tools/example` fixture stops rehearsing task state, and that's accepted. Reworking it is already a separate, out-of-scope track. Its "blank" checkpoint names `bp init` (`tools/example/reset.sh:4`), which Verse 1 deletes.

## Every record carries git info
- The goal is traceability: the state of the world when the record was written. This covers tracks, bars, research topics, feedback notes and retro proposals.
- Tracks and bars keep a `branch` and a single `commit` (track = landing commit, bar = best-effort).
- Research topics, feedback notes and retro proposals carry a `commits` list of `{sha, branch, at}`. The first entry is HEAD when the record is created. A later write appends an entry when HEAD has moved.
- Verse 2 defines which kind carries which field and adds them empty. BIT-49 defines what migrate writes for each kind. BIT-47 fills them on later writes, and BIT-50 records the track's landing commit at completion.

## Order
- BIT-47 was split. Merge-aware completion is now BIT-50. The order is BIT-45 → BIT-46 → BIT-48 → BIT-49 → BIT-47 → BIT-50.

# 2026-10-01 readiness pass (Claude defaults; the operator can veto these)
- **Never `just install` on `v2`** now also says why: it replaces the daily `bp` that every daily session's MCP server runs, and a v2 `bp` can't read `.bit/`. It overrides the memory `bit-pro-just-install-after-changes` for v2 work. The rule is repeated in every v2 track.
- **Tests never touch the real store.** The shared fixture `initProject` (`cmd/cmd_test.go:89-96`, `cmd/task/helpers_test.go`, about 75 call sites) runs `bp init` or `SaveConfig` today. It's rewritten to register the project under a `t.TempDir()` `XDG_DATA_HOME`.
- **Each bar leaves the build green:** the `bitdir` and `task.Config` callers move before those packages go.
- **Project code format:** after uppercasing, `^[A-Z][A-Z0-9]*$`, with `FEEDBACK` and `RETRO` reserved. The code names a store dir and prefixes every ID, v1's `bp add` doesn't check it (`cmd/add.go:84-102`), and the reserved names would collide with BIT-49's top-level `feedback/` and `retro/` on case-insensitive APFS. Known v1 prefixes (`BIT`, `EX`) pass.
- **Only project-scoped commands resolve.** `bp task`, `approve`/`unapprove`, `feedback`, `tui` and `remove` resolve. `add`, `list`, help and `--version` work anywhere, and `serve mcp` resolves per tool call. The root pre-run `bitdir.Resolve()` (`cmd/root.go:136`) goes.
- **One initial migration, edited in place until cutover.** Verse 3 adds the removed flag there, and dev sandboxes are recreated after a schema change (dbmate won't re-apply an edited migration).
- **Reviving restores the project, not its archived tracks.** They stay in `archive/tasks/`, as after `bp task delete`.
- **`scripts/install.sh`'s closing hint** ("Run 'bp init'") points to `bp add` in Verse 1.
- **Verse 2 names feedback notes** alongside tasks and research, since they get the JSON format and `commits` there too.