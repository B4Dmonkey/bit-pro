---
id: BIT-46
title: 'v2: central registry, resolver and record format'
status: doing
---
## Why
bit keeps each project's state in a `.bit/` folder inside the repo, so knowledge is scattered across repos. That also breaks down for a client like `acme/`, where cross-cutting work happens in a folder that isn't a repo and has nowhere to keep state. On top of that, four separate code paths work out which project bp is in, and running bp from a subfolder doesn't find the project at all. Moving every project's state into one store under `~/.local/share/bit/` gives multi-repo work (linking comes later) a home, and makes the operator's daily tool resolve projects one way, everywhere. The operator uses v1 daily, so v2 is built alongside it and v1 keeps working.

## Summary
v2 is built on the `v2` branch and tested against a sandbox. v1 on `main` is untouched.
- A fresh sqlite registry (`~/.local/share/bit/main.db`) holds a `projects` table.
- bp finds the project from the longest registered path that contains the current folder, and that lookup replaces `.bit/` discovery everywhere. `bp init` is deleted.
- Each project's tasks and research live under `~/.local/share/bit/<CODE>/` as JSON metadata files next to unchanged markdown bodies, and every record carries the git state it was written in.
- `bp remove` takes a project out of the registry without losing anything: a soft delete, with its remaining work archived. `bp add` revives it.

Other v2 work is in its own tracks: the global Claude wiring (BIT-48), shared feedback and retro, `bp migrate` and the `.bit/` sweep (BIT-49), the commit skill and git capture (BIT-47), then merge-aware completion (BIT-50).

## Visual aid
```
v1 (per repo)                         v2 (central)
repo/.bit/config.toml       ──►   ~/.local/share/bit/main.db  (projects: code, path, removed, …)
repo/.bit/tasks/BIT-7.md    ──►   ~/.local/share/bit/BIT/tasks/BIT-7.{json,md}
repo/.bit/research/BIT-7/   ──►   ~/.local/share/bit/BIT/research/BIT-7/<topic>.{json,md}

cwd ── longest registered path containing it ──► project ──► store dir
```

## Decisions
- **Depends on BIT-45 (daemon removal),** which lands first, so the new db starts with only `projects`.
- **Cutover belongs to the operator.** The operator alone decides when v2 is ready and merges the branch, and no track gates it. The steps are on the cutover checklist in `v2-sketch.md`.
- **Never run `just install` on `v2`.** Dev builds run with `just run`, or as a binary built to a temp path, against a sandboxed `XDG_DATA_HOME` and a sandboxed `HOME` (BIT-45 topic `testing`, BIT-46 topic `soundness-2`). `just install` replaces the daily `bp` that every daily Claude session runs as its MCP server, and a v2 `bp` can't read `.bit/`, so this overrides the habit of installing after code changes (2026-10-01).
- **Dev sessions use `claude --plugin-dir <v2 checkout>/bit --mcp-config dev.json --strict-mcp-config`.** This tests the branch's skills against a dev MCP server.
  - `dev.json` keeps the server name `bit`, so the tools stay `mcp__bit__*`. It runs the dev `bp serve mcp` and sets `XDG_DATA_HOME` to a sandbox.
  - Headless `-p` runs also need `--allowedTools 'mcp__bit__*'`.
  - Verified with Claude Code 2.1.285 (topic `plugin-dir-testing`): the branch plugin replaces the installed `bit@bit-pro` with no duplicates, the plugin declares no MCP server, `--strict-mcp-config` drops the daily `~/.claude.json` entry, and nothing in the daily setup changes.
  - This bypasses the Claude wiring entirely, so this track doesn't need BIT-48.
- **Tests never touch the real store (Claude default, 2026-10-01).** The shared test fixture `initProject` (`cmd/cmd_test.go:88-96`, `cmd/task/helpers_test.go`, 78 call sites) creates a project with `bp init` or `SaveConfig`. It's rewritten to register the project in a registry under a `t.TempDir()` `XDG_DATA_HOME`, so `go test` never opens `~/.local/share/bit/main.db`.
- **The MCP test fixtures follow the same rule (Claude default, 2026-10-02).** The MCP tests don't use `initProject`. They write `<dir>/.bit` directly through `seedTasks` (30 uses), `seedConfig` (5) and `mcpSession(t, dir)` (36), and `store/store_test.go:23,28` and `db/open_test.go:20` hard-code today's paths (topic `claim-audit-2026-10-02`).
- **Each bar leaves the build and tests green (Claude default, 2026-10-01).** The `bitdir` and `task.Config` callers (BIT-46 topic `file-inventory`, plus `cmd/add.go:51` and the fixtures above) move to the resolver before those packages are deleted. Deleting `bp init` and rewriting `initProject` happen in one bar, because the fixture runs `bp init` today (`cmd/cmd_test.go:93`). All CLI commands switch to the central store in the same bar, or records would land in two places.
- **"Green" means `just lint` and `just test` pass (Claude default, 2026-10-02).** That's what the pre-commit hook runs (`.pre-commit-config.yaml`). Code this track leaves unused is deleted in the same bar, including `prefixFlag` and the test helper `mcpRegisterCall` (`cmd/cmd_test.go`, used only by `cmd/init_test.go`).
- **The `tools/example` rehearsal fixture stops rehearsing task state, and that's accepted (Claude default).** It resets task state through its git-tracked `.bit/` with `git reset --hard`, and its "blank" checkpoint is described as a `bp init` state in a comment (`tools/example/reset.sh:4`, corrected 2026-10-02: that line is a comment, not a call). This track deletes `bp init`. Reworking the fixture is a separate track, already listed as out of scope here and in `v2-sketch.md` (BIT-45 topic `testing`).
- **The registry is a fresh db at `~/.local/share/bit/main.db`.**
  - v1 already has a registry: `~/.local/share/bit-pro/bit.db` (`store/store.go:22`, `db/open.go:25`), with four migrations and a non-UNIQUE `code`. Verse 1 moves the data dir to `bit`, renames the file to `main.db`, and replaces all four migrations and `db/queries/projects.sql` with the new set (fact, 2026-10-02).
  - Its schema is managed with sqlc and dbmate. Migrations are embedded and applied when the db opens, so a fresh machine needs no setup command (topics `sqlite-registry`, `init-and-update`).
  - The db never lives in a project.
- **A fresh `main.db` is safe to open from several processes at once (Claude default, 2026-10-02).** Several MCP servers and the CLI can open it for the first time together. Migrations on open are serialized, so only one process applies them, and a test opens a fresh db from two processes concurrently. bit:plan picks the locking mechanism.
- **Until cutover, the schema is one initial migration, edited in place (Claude default, 2026-10-01).** Verse 3 adds the removed flag to it, and dev sandboxes are recreated after any schema change, because dbmate won't re-apply an edited migration. Nothing outside a sandbox has the db yet.
- **The db is the only place a project's path is stored.** In `projects`, both `code` and `path` are UNIQUE, and the store dir is named by the code. Records don't copy the path. Re-pointing a moved repo is out of scope for now.
- **Path uniqueness and matching are case-insensitive, and are checked in Go (Claude default, 2026-10-02).** macOS APFS is case-insensitive, `os.Getwd` returns the path in the case it was typed, and SQLite's UNIQUE is case-sensitive. bp loads the (small) project list and compares in Go, rather than relying on `COLLATE NOCASE`, whose sqlc support is unverified.
- **A project's ID prefix is its `projects.code`.** `config.toml` is dropped, and so is the track-ID minting that reads it today (`task/store.go:216-221`). The task store gets the code from the registry. `task.Store` has no code today (`task/store.go:25-31`), so its constructor changes (topic `claim-audit-2026-10-02`).
- **A code is letters and digits, starting with a letter, and `FEEDBACK` and `RETRO` are reserved (Claude default, 2026-10-01).** It's checked after uppercasing (`^[A-Z][A-Z0-9]*$`). The code names a store dir and prefixes every ID, and v1's `bp add` doesn't check it at all (`cmd/add.go:84-102`). The reserved names would collide with BIT-49's top-level `feedback/` and `retro/` folders on case-insensitive APFS. BIT-49's migrate applies the same rule.
- **`bp init` is deleted in Verse 1 (operator, 2026-10-01).** It depends on `bitdir` and `config.toml` (`cmd/init.go:10,39,75`), which Verse 1 removes. The per-project wiring helper (`writeClaudeWiring`, in `cmd/init.go` today) stays, because `bp add` still uses it, and BIT-48 replaces it with the global wiring step. `bp add` keeps today's order (wire, then register, `cmd/add.go:66-73`); BIT-48 reverses it. `scripts/install.sh`'s closing hint ("Run 'bp init'", `:22`) points to `bp add` instead (Claude default).
- **`bp add` registers the project (path and code).**
  - It checks "already registered" before "has `.bit/`" (operator, 2026-10-01). A registered project is a no-op that says so, and it doesn't re-run any wiring.
  - On a removed project's path, it clears the removed flag and revives the project (operator, 2026-10-01). Reviving restores the project, not its archived tracks: they stay in `archive/tasks/`, as after `bp task delete` (Claude default, 2026-10-01). This lands in Verse 3, with the removed flag.
  - It refuses a code that belongs to a removed project when it's asked from a different folder (operator, 2026-10-01). The code still belongs to the removed project, so the operator revives that project or picks another code. This lands in Verse 3.
  - Only an unregistered folder that still has a `.bit/` is refused, with a message to run `bp migrate` (BIT-49). That replaces today's `.bit/` check before wiring (`cmd/add.go:66`), so a fresh registration only wires a folder with no `.bit/`, as today.
  - Until BIT-48 lands, a fresh registration still runs today's per-project wiring helper.
- **The project is found by the longest registered path that contains the current folder.**
  - Matching is on whole path segments, with symlinks resolved. That covers subfolders and Claude worktrees with no special case.
  - Symlinks are resolved on both the stored path and the current folder. When a path doesn't exist, so `EvalSymlinks` fails, its deepest existing ancestor is resolved and the missing tail is joined back on (Claude default, 2026-10-02; `t.TempDir()` is `/var/...` but resolves to `/private/var/...`, and the MCP worktree test passes a path that doesn't exist, `cmd/serve_mcp_test.go:104`).
  - A registered path may sit inside another registered path, since `acme/` holds repos, and the longest match wins (Claude default, 2026-10-02).
  - The MCP server resolves on every call, never at startup, from `CLAUDE_PROJECT_DIR`, falling back to the current folder for opencode. An unregistered folder costs nothing until a tool is called. The server already builds its store on every call (`cmd/serve_mcp.go:294…489`), and only reads the env var once (`:231`), so the new part is the registry lookup (fact, 2026-10-02).
  - This replaces `bitdir` (topics `project-resolution`, `config`).
- **bitdir's main-checkout cut survives as an exported helper when `bitdir` is deleted (Claude default, 2026-10-02).** The `.claude/worktrees/` cut (`bitdir/bitdir.go:57-68`) is what BIT-49's migrate needs to find a worktree's main-checkout `.bit/`. A plain upward walk won't do, because each worktree has its own tracked `.bit/`.
- **Only commands that act on a project's work resolve it (Claude default, 2026-10-01).** `bp task`, `approve`/`unapprove`, `feedback`, `tui` and `remove` resolve. `add`, `list`, help and `--version` work from any folder, and `serve mcp` resolves per tool call. The root command's `bitdir.Resolve()` pre-run (`cmd/root.go:136`) goes. The `cmd/task` test harness never runs the root pre-run (`cmd/task/helpers_test.go:21-40`), so those commands resolve the project themselves.
- **The plugin-behind notice gets its project root from the resolver: the resolved project's path, else the current folder (Claude default, 2026-10-02).** `pluginState` (`cmd/root.go:21-28`) runs after every command and calls `bitdir.Root()` today. BIT-48 later drops the root, when the notice reads the user-scope install.
- **Resolver errors in an unregistered folder:** if a `.bit/` exists here or above, the error says "run `bp migrate`". Otherwise it says "not a bit project; run `bp add`". In a removed project's folder, it says the project was removed and points to `bp add` (operator, 2026-10-01). That error needs the removed flag, so it lands in Verse 3 (Claude default, 2026-10-02). Both the CLI and every MCP tool return them.
- **Project dirs use the uppercase code,** e.g. `~/.local/share/bit/BIT/`.
- **Each record is a JSON metadata file plus an unchanged `.md` body.** Frontmatter becomes JSON fields, and `content` holds the relative path to the `.md`. Bodies are never split into fields (topic `json-schema`).
- **Record layout (Claude default, 2026-10-02).** Neither topic `json-schema` nor the decisions settled these, and bit:plan would otherwise guess them.
  - The pair sits side by side: `<ID>.json` and `<ID>.md` for tasks, `<topic>.{json,md}` for research, `<TRACK>-NNN.{json,md}` for feedback. `content` is the `.md` file's name.
  - A task's JSON fields keep today's frontmatter names (`id`, `title`, `status`, `approved`, `phase`, `phase_label`, `order`). `parent` isn't stored, since it follows from the ID, as today.
  - Every field is written. Empty git fields are `""` and `[]`, never null. Timestamps are RFC 3339 UTC, and `updated_at` changes on every write. Rewriting a record (e.g. `research_write` on an existing topic) keeps its `created_at` and `commits`.
  - The `.md` is written first, then the `.json`. ID minting and reservation look at `*.json` only, so a stray `.md` reserves nothing.
  - `research_write` and `feedback_add` still return the `.md` path.
- **Every record carries `project`, `created_at` and `updated_at`.**
- **Every record carries git info (operator, 2026-10-01).** This covers tracks, bars, research topics, feedback notes and retro proposals. The goal is traceability: the state of the world when the record was written.
  - Tracks and bars carry a `branch` and a single `commit`. On a track, `commit` is the landing or merge commit, the anchor that matters most. On a bar, it's a best-effort hash of the bar's own commit, which may break. After a squash merge it points at the squash commit (BIT-50).
  - Research topics, feedback notes and retro proposals carry a `commits` list. Each entry is `{sha, branch, at}`. The first entry is HEAD when the record is created, and a later write appends an entry when HEAD has moved.
  - This track only adds the fields, empty. BIT-49's migrate, BIT-47 and BIT-50 fill them. Retro proposals become records in BIT-49 and carry the same `commits` field.
  - The store's write functions take the git values from their caller and write them as given, so BIT-47 and BIT-49 only pass values in (Claude default, 2026-10-02).
  - They may be empty, e.g. for `acme/`, which has no git.
- **Every store operation that moves files moves both files of a pair (Claude default, 2026-10-02).** Completing, archiving (`relocateInto`, `task/store.go:57-66`) and `bp remove` all go through it, so Verse 2 rewrites it and Verse 3's archiving works on JSON records.
- **`bp remove` takes a project out of the registry (operator, 2026-10-01).**
  - It soft-deletes the project: a removed flag on its `projects` row. There is no hard delete.
  - It archives the project's remaining work: every track that isn't completed, including tracks whose bars are all done but which were never completed.
  - It asks for confirmation first, showing a note that lists any outstanding (not-done) work.
  - It's CLI-only.
  - The project's research, feedback and retro are kept.
  - It belongs here because it's registry and store work: the `projects` row and the archive path.
- **Archived work goes through the existing archive path, `archive/tasks/` (Claude default).** That's where `bp task delete` already relocates tracks and their bars (`task/store.go` `Relocate`, `:637-650`). It keeps their IDs reserved, so a revived project never re-mints them.
- **`bp list` leaves out removed projects (Claude default).** A removed project shouldn't look active. Its row and store dir are still there for `bp add` to revive.
- **`bp remove` runs from the project's folder, like the other commands (Claude default).** It uses the same resolver, so there's no path argument to get wrong.
- **Verse 1 is tried on a fresh folder in the sandbox (fact, 2026-10-02).** Every real project has a `.bit/`, which Verse 1 refuses with "run `bp migrate`", and migrate is BIT-49's.
- **Out of scope:**
  - the global Claude wiring (BIT-48);
  - shared feedback and retro, `bp migrate` and the `.bit/` sweep (BIT-49), including the `.bit/` wording in the MCP tool descriptions (`cmd/serve_mcp.go:76,84,102,109`);
  - filling git fields and the commit skill (BIT-47), and merge-aware completion (BIT-50);
  - project and track linking, reworking the fixture, syncing skills to opencode, and distribution.
  Until BIT-49 lands, feedback stays in each project's store dir and retro and learn are broken on the branch, which is accepted.

## Verses
- [x] Verse 1 — A registered project works from the central store.
  - The operator registers a project with `bp add`.
  - The CLI and MCP find it from any subfolder or worktree, reading and writing tasks under `~/.local/share/bit/<CODE>/`, still in v1's markdown format.
  - An unregistered folder gets the "run bp migrate" or "run bp add" error. `bp init` is gone.
  Touches: `bitdir/` (replaced; its main-checkout cut kept as a helper) and `bitdir/bitdir_test.go`, `task/store.go`, `task/config.go` and its tests (`task/store_test.go:682-790`), `db/` (the new initial migration replacing all four, `db/queries/projects.sql`, `db/open.go` and `db/open_test.go:20`), `store/` (`store.go:22` and `store_test.go:23,28`), `cmd/{root,add,init,approve,feedback_add,tui,serve_mcp,list}.go` (`init.go` loses the command and keeps `writeClaudeWiring`; `root.go` loses the pre-run and changes `pluginState`), `cmd/task/*` including `cmd/task/move_test.go`, `scripts/install.sh` (the closing hint), the test fixture `initProject` in `cmd/cmd_test.go` and `cmd/task/helpers_test.go`, the MCP fixtures in `cmd/serve_mcp_test.go`, and the tests that create `.bit/`. The full list is in BIT-46 topics `file-inventory` and `claim-audit-2026-10-02` (which adds the files `file-inventory` missed and drops `cmd/serve_test.go`, deleted by BIT-45). The files it lists under BIT-48 (`claude/{sync,settings,plugin}.go`, the notice text in `cmd/root.go`) stay out of this verse.
- [ ] Verse 2 — Tasks, research topics and feedback notes are stored as JSON metadata plus markdown content. Every record carries the project, timestamps and its empty git fields: `branch` and `commit` on tracks and bars, and a `commits` list on research topics and feedback notes. The CLI, MCP and TUI behave as before, and any Verse 1 sandbox stores are thrown away.
  Touches: `task/` (store, parse, research, feedback, ID minting, and the file moves in `relocateInto`) and the tests that assert on frontmatter (7 of them; `feedback_add_test` and `serve_mcp_write_test` check raw bytes). See BIT-45 topic `json-schema` (read its superseded banner first).
- [ ] Verse 3 — The operator removes a project with `bp remove`. After confirming, with any outstanding work listed first, the project is soft-deleted in the registry and every track that isn't completed is archived. `bp list` no longer shows it, running bp in its folder says it was removed, and `bp add` revives it, or refuses that code from another folder.
  Touches: `db/` (the `projects` table in the initial migration, and its queries), the resolver (the removed-project error), `task/store.go` (the archive path), `cmd/{add,list}.go`, and a new `cmd/remove.go`.

## References
- `v2-sketch.md` (repo root): the v2 direction, the operator's decisions, the multi-repo sketch and the cutover checklist.
- `.bit/research/BIT-45/`: research notes for the whole v2 transition. Start at the `index` topic.
- `.bit/research/BIT-46/`: research specific to this track. Start at the `decisions` topic for the operator decisions of 2026-09-30 and 2026-10-01. Also see `file-inventory`, `soundness-2`, `init-and-update`, `plugin-dir-testing` and `soundness`. `review-2026-10-02` and `claim-audit-2026-10-02` (2026-10-02) verify each claim against the code, give a bar order that keeps each bar green, and win where older topics differ.