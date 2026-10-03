# Claim audit + build inventory (2026-10-02, v2 @ 6a1d345, BIT-45 NOT landed)

Baseline: `go test ./...` green with sandboxed HOME/XDG_DATA_HOME. `db/orm/` is untracked (`git ls-files db/orm` empty; `.gitignore`). Probed on this Mac (Go 1.27): `t.Chdir` sets `$PWD`; `os.Getwd` returns `/var/...`, `filepath.EvalSymlinks` gives `/private/var/...`, and `EvalSymlinks` on a missing path errors (`lstat ...: no such file or directory`).

Verdicts: **CONFIRMED** = code matches; **WRONG** = false today; **STALE** = was true or was a gap that has since been fixed/superseded; **NOT RE-RUN** = empirical CLI result, consistent with code but not repeated.

## Part 1. Claim audit

### Track body
| Claim | Verdict | Evidence / correction |
|---|---|---|
| Four code paths work out the project | CONFIRMED | `bitdir.Current/Resolve` (CLI, `cmd/root.go:136`), `bitdir.ForRoot` (MCP, `cmd/serve_mcp.go:294…489`), `bitdir.Root` (`cmd/root.go:27`), `GetProjectByPath(os.Getwd())` (`cmd/tui.go:55-60`). The 4th goes with BIT-45, so 3 remain at BIT-46 start. |
| bp from a subfolder doesn't find the project | CONFIRMED | `bitdir.Canonical` returns relative `.bit` outside a worktree (`bitdir/bitdir.go:41-47`). |
| "A fresh sqlite registry (`main.db`)" | CONFIRMED as design, but v1 already has one | `store/store.go:22` (`bit-pro`), `db/open.go:25` (`bit.db`), 4 migrations. Verse 1 renames dir+file and replaces all migrations and `db/queries/projects.sql`. |
| Fixture `initProject` at `cmd/cmd_test.go:89-96` | WRONG (trivial) | Def is `:88-96`; `bp init` at `:93`. |
| "about 75 call sites" | WRONG | 78 calls (cmd 22: add 3, approve 3, feedback_add 8, root 7, task_test 1; cmd/task 56: create 14, list 11, update 9, delete 8, read 7, complete 5, move 2). |
| Fixture "creates a project with `bp init` or `SaveConfig`" | CONFIRMED | `cmd/cmd_test.go:93`, `cmd/task/helpers_test.go:59`. |
| MCP tests use the fixture | MISSING | MCP tests don't use `initProject`: they use `seedConfig` (`cmd/serve_mcp_write_test.go:56-62`, 5 uses), `seedTasks` (`cmd/mcp_harness_test.go:93-103`, 30 uses), `seedEscapedResearch` (2), `seedOrderedTrack`/`seedDoneTrack`, and `mcpSession(t, dir)` (36 sites) — all write `<dir>/.bit` directly. |
| `tools/example/reset.sh:4` names `bp init` | CONFIRMED | Comment line 4. |
| Migrations embedded, applied on open | CONFIRMED | `db/open.go:16,33` (`CreateAndMigrate`). |
| dbmate won't re-apply an edited migration | CONFIRMED (by design; not probed) | dbmate tracks applied versions in `schema_migrations`. |
| `code` and `path` both UNIQUE | Design; today `code` is NOT unique | `db/migrations/20260820232810_create_projects.sql`. |
| Track minting reads config.toml `task/store.go:216-221` | CONFIRMED | `s.Config()` :216, `s.NextID(cfg.Prefix)` :221. |
| v1 `bp add` doesn't check the code `cmd/add.go:84-102` | CONFIRMED | Only `task.NormalizeID` (`ToUpper`, `task/store.go:69-71`). |
| `cmd/init.go:10,39,75`; `writeClaudeWiring` in init.go | CONFIRMED | import :10, `SaveConfig` :39, `Config()` :75, helper :51-71. |
| `scripts/install.sh` hint | CONFIRMED | `:22`. |
| Root pre-run `bitdir.Resolve()` `cmd/root.go:136` | CONFIRMED | Only `PersistentPreRunE` in the tree. |
| `Relocate` → `archive/tasks/`, IDs stay reserved | CONFIRMED | `task/store.go:104,112,57-66`; `highestReserved` :637 scans tasks/completed/archive. |
| MCP "resolves on every call, never at startup" (as a change) | STALE framing | Env read once (`cmd/serve_mcp.go:231`), but every handler already builds its store per call (`:294,320,360,384,407,423,439,455,472,489`). New work = registry lookup per call + resolver errors as tool errors. |
| opencode needs a cwd fallback | CONFIRMED | `opencode.json` passes no env; `ForRoot("")` already = cwd-relative `.bit`. |
| Claude worktrees sit inside the repo | CONFIRMED | `.gitignore` `.claude/worktrees`; tests `cmd/root_test.go:55-94`, `cmd/serve_mcp_test.go:97-111`. |
| Verse 1 Touches list | INCOMPLETE | Missing: `db/open.go` (filename), `db/queries/projects.sql`, `store/store_test.go:23,28`, `db/open_test.go:20`, `db/queries_test.go`, `cmd/serve_mcp_test.go`, `cmd/task/move_test.go`, `task/store_test.go:739-822` (`TestStoreCreate` seeds via `SaveConfig`, rewritten not deleted) and `:678-737` (config tests, deleted), `bitdir/bitdir_test.go`. `cmd/list.go` needs no resolver, only the query changes. |
| Verse 2 cites BIT-45 `json-schema` banner | CONFIRMED | Banner present. |

### file-inventory
| Claim | Verdict | Evidence |
|---|---|---|
| `cmd/root.go` :27/:136/:119 | CONFIRMED | |
| `cmd/add.go` :51/:66 | CONFIRMED | |
| `cmd/init.go` :23/:39/:75/:51 | CONFIRMED | :23 is the Short text, not a write. |
| bitdir callers approve/feedback_add/tui:24 + cmd/task/* | CONFIRMED | approve :15,26; feedback_add :19; tui :24; cmd/task complete:15 create:42 delete:37 list:19 move:17 read:19 update:19. |
| serve_mcp "reads once; resolve per call instead" | STALE framing | See body row above. |
| "Go tests that create .bit/ (18)" | WRONG count | Includes `cmd/serve_test.go` (BIT-45 deletes it). Misses `cmd/task/move_test.go`, `cmd/serve_mcp_test.go`, `store/store_test.go`, `db/open_test.go`, `db/queries_test.go` (path/schema). |
| fixture `cmd/cmd_test.go:89-96`, about 75 | WRONG | `:88-96`, 78. |
| BIT-48 pointers sync.go:22, plugin.go:26-31, cmd_test.go:98-111 | CONFIRMED | |
| plugin.json:5, complete.go:12, bot.md:41 | CONFIRMED | |
| `update/` files | CONFIRMED | |

### init-and-update
All code pointers CONFIRMED: `cmd/init.go:25-44`, `writeClaudeWiring :51-71`, `claude/settings.go:17,19`, `claude/sync.go:12-19,21-31,33-45`, `claude/plugin.go:10-34`, `cmd/add.go:26-80,45-48,66-70`, `cmd/root.go:50-68`, `db/open.go:19-38`, `store/store.go:23`. `ExecRunner` sets no `Dir` — CONFIRMED. One STALE: "replaces the migrations (BIT-45 drops the queue)" — BIT-45 deletes the queue *queries/code* only and keeps all migrations ("No drop migrations"); the queue table disappears only with BIT-46's new initial migration. Recommendation (a) is superseded (already marked).

### plugin-dir-testing
File-level facts CONFIRMED: `bit/.claude-plugin/plugin.json` name `bit`, version 1.3.0, no `mcpServers`; no `bit/.mcp.json`; bit-pro `.claude/settings.json` enables `bit@bit-pro`; `~/.claude/settings.json` does not enable it (but does list the `bit-pro` marketplace); `~/.claude.json` local `bit` entries for bit-pro and a client project, none top-level; `bp` = `~/go/bin/bp`; cached 1.3.0 lacks `complete`, branch has it (9 skills, 3 agents). The CLI behaviour (inline replaces marketplace copy, strict-mcp drop, `-p` needs `--allowedTools`) is NOT RE-RUN.

### soundness
- `task/store.go:216-221` blocker: CONFIRMED, now in Touches.
- "`task/store.go` is not in verse 1 Touches", "`cmd/init.go` in no Touches", "Touches incomplete: cmd/task/*, approve…": STALE (all now listed).
- `bp add` semantics with `.bit/`: STALE (decided).
- `code` not UNIQUE: CONFIRMED in code; decided.
- No git exec in bp (`claude/sync.go:13`, `claude/plugin.go:78`, `daemon/daemon.go:22`): CONFIRMED (no `"git"` anywhere in Go).
- `task/feedback.go:16` feedback dir: CONFIRMED (`:15-17`).
- "~18 test files": see file-inventory row (WRONG count).
- BIT-45 `migrate` stale lines (:12 bit.db, :18 `<code>`, :26 unknown files, :33 git log): CONFIRMED still there.
- "v2-sketch delivery order puts anchors at step 4 before feedback/migrate": STALE — `v2-sketch.md:94-101` now matches the track order 45→46→48→49→47→50.
- Ordering / `task.New(root)` root-parameterised / `worktreeCut` `bitdir/bitdir.go:57`: CONFIRMED.

### soundness-2
Code pointers CONFIRMED: `claude/sync.go:13,22`, `claude/plugin.go:26-31`, `cmd/root.go:27,119`, `cmd/serve_mcp.go:229` and `cmd/tui.go:22` (quiet), `runMCPServer` `:238-274` (note says 276; trivial). Real `installed_plugins.json`: project-scope bit-pro 1.3.0 and example 0.1.0 — CONFIRMED. Real `~/.claude/settings.json` declares `bit-pro` in `extraKnownMarketplaces` — CONFIRMED. CLI scope behaviour NOT RE-RUN.

### decisions
All code pointers CONFIRMED (init.go:10,39,75; root.go:27,136,119; add.go:84-102; reset.sh:4). "Known v1 prefixes (BIT, EX) pass": CONFIRMED (`.bit/config.toml` = BIT, `tools/example/.bit/config.toml` = EX). "about 75 call sites": WRONG (78).

### index
Its "Verse 1 Touches miss cmd/task/*…" and "bp init is in no verse" bullets are STALE (resolved). Rest is summary of the above.

## Part 2. Call-site inventory (production code; BIT-45-deleted files excluded)

- `bitdir.*`: `cmd/root.go:27` (Root), `:136` (Resolve); `cmd/approve.go:15,26`; `cmd/feedback_add.go:19`; `cmd/tui.go:24`; `cmd/init.go:39,75`; `cmd/task/{complete:15,create:42,delete:37,list:19,move:17,read:19,update:19}.go`; `cmd/serve_mcp.go:294,320,360,384,407,423,439,455,472,489` (ForRoot). Tests: `bitdir/bitdir_test.go`, `cmd/root_test.go:44-94`, `cmd/serve_mcp_test.go:97`.
- `task.New`: prod `cmd/add.go:51`, `approve.go:15,26`, `feedback_add.go:19`, `init.go:39,75`, `tui.go:24`, `serve_mcp.go` ×10, `cmd/task/*` ×7 (as `taskstore.New`). Tests: `cmd/approve_test.go:15,32`, `init_test.go:34,64,103`, `mcp_harness_test.go:96`, `serve_mcp_research_test.go:313`, `serve_mcp_write_test.go:59,77,114,182,219,266,338,391,434,647`, `cmd/task/create_test.go:19,72,133,167`, `update_test.go:49,122,139,156,189,231`, `helpers_test.go:59,69`, `task/store_test.go` ×28. (`daemon/loop*.go`, `cmd/serve_test.go`, `task/counts_test.go` go with BIT-45.)
- `Config`/`SaveConfig`/`ConfigPath`/`configFileName` (there is no `LoadConfig`): `task/config.go` (all), `task/store.go:20,216`; `cmd/add.go:51`; `cmd/init.go:39,75`; tests `cmd/init_test.go:34,64,103`, `cmd/serve_mcp_write_test.go:59`, `cmd/task/helpers_test.go:59`, `task/store_test.go:678-737` (config tests) and `:783-786` (inside `TestStoreCreate`).
- `store.Dir`: `db/open.go:20` only after BIT-45 (`cmd/start.go:59` goes). Test `store/store_test.go:23,28`.
- `db.Open`: prod `cmd/add.go:32`, `cmd/list.go:19` (`cmd/tui.go:36`, `serve.go:65`, `status.go:41` go with BIT-45). Tests `cmd/add_test.go:38,116,167,208,244`, `cmd/list_test.go:75,120`, `db/open_test.go:14` (asserts `bit-pro/bit.db` at :20), `db/queries_test.go:14`.
- `orm` queries in prod after BIT-45: `ProjectExists` + `CreateProject` (`cmd/add.go:40,73`), `ListProjects` (`cmd/list.go:25`). `GetProjectByPath` has no caller once `cmd/tui.go` queueFuncs goes.
- `newRootCmd`: `cmd/root.go:125,128`; tests `cmd/cmd_test.go:31,46,61,128`, `cmd/root_test.go:223,243`, `cmd/task_test.go:18`. BIT-45 changes the signature (drops `lc`); BIT-46 only changes the body (drops `newInitCmd`, the pre-run).
- `initProject`: defs `cmd/cmd_test.go:88`, `cmd/task/helpers_test.go:53`; 78 calls (above).
- `bp init` invocations: `cmd/cmd_test.go:93`, `cmd/init_test.go` (whole file), `scripts/install.sh:22`, `tools/example/reset.sh:4` (comment), README command table (BIT-49).
- `.bit` path peeks in tests that break when the store moves: `cmd/{add,approve,feedback_add,init,mcp_harness,root,serve_mcp_research,serve_mcp_write,serve_mcp,task}_test.go`, `cmd/task/{complete,create,delete,helpers,list,move,read,update}_test.go`, `task/store_test.go` (Path tests use `New(".bit")` but are store-relative, fine). Returned-path assertions: `serve_mcp_research_test.go:89,113,132,162,308`, `serve_mcp_write_test.go:523,546,586,591,617,621,705,710`.
- `cmd/add_test.go` encodes v1 add semantics: `:17` (default code from config.toml "Project code (BIT)"), `:224` (already enrolled), and `:51-58` compares stored path to `filepath.Abs` — breaks once `bp add` stores the `EvalSymlinks` path (`/private/var`).
- `writeClaudeWiring` coverage after `init_test.go` is deleted: only `cmd/add_test.go:62-135` (asserts settings.json, `pluginSyncCalls`). init_test's MCP-registration/idempotency tests (`:222,263,301`) have no add equivalent; BIT-48 rewrites these anyway.

## Part 3. Bar order (each bar green)

**Bar check command:** `just test` (or `sqlc generate && go test ./...`), not plain `go test`. `db/orm` is gitignored; a bar that changes migrations/queries compiles against stale generated code locally and fails on a fresh clone. Run tests with `HOME`/`XDG_DATA_HOME` sandboxed. Any bar that edits the initial migration: delete dev sandboxes.

Precondition: BIT-45 merged (`newRootCmd(run)`, no queue queries, `ListProjects` without counts, `cmd/tui.go` queue code gone).

### Verse 1
1. **Fresh registry.** `store.Dir` → `<data>/bit`; `db/open.go` → `main.db`; delete the 4 migrations, add one initial migration `projects(id, code TEXT NOT NULL UNIQUE, path TEXT NOT NULL UNIQUE)` (decide `COLLATE NOCASE` here; sqlc support for it not verified); rewrite `projects.sql`: `CreateProject`, `GetProjectByPath`, `GetProjectByCode`, `ListProjects`; delete `ProjectExists` or keep. Tests: `store/store_test.go`, `db/open_test.go`, `db/queries_test.go` (UNIQUE code). No caller changes beyond `cmd/add.go` if a query is renamed.
2. **Code rule.** Exported `ValidateCode` (`^[A-Z][A-Z0-9]*$` after upper, reserved `FEEDBACK`/`RETRO`) used by `bp add`. New add_test cases. Independent.
3. **Resolver package** (new, e.g. `project/`): `Resolve(ctx, dir) (Project, error)` = load rows, `EvalSymlinks` both sides, whole-segment, case-insensitive longest match; errors `ErrNotRegistered` / needs-migrate (a `.bit/` here or above). `bp add` stores the `EvalSymlinks` path (update `cmd/add_test.go:51-58` expectations). Unit-tested, no other callers yet. Must tolerate `EvalSymlinks` failing on a non-existent dir (fall back to `filepath.Clean`), because MCP's `CLAUDE_PROJECT_DIR` worktree test passes a path that doesn't exist (`cmd/serve_mcp_test.go:104`).
4. **`task.Store` learns its code.** Add a code to `Store` (e.g. `task.NewProject(root, code)`; keep `New(root)` for tests that never mint tracks). `Create` uses the code; transitional fallback to `Config()` when empty keeps every caller green. `TestStoreCreate` gains a code-based case.
5. **Fixtures register (test-only, foldable into 6).** Both `initProject`s sandbox `HOME` (`t.Setenv`) and `XDG_DATA_HOME=""`, and also register `dir` in the registry, while still writing `.bit/config.toml`. Safe: no cmd or cmd/task test is `t.Parallel`.
6. **CLI commands use the central store — all at once.** `cmd/task/*`, `approve`, `feedback_add`, `tui` switch together, because cmd tests create via `bp task create` and then read via `bp feedback add`/`approve` (`cmd/feedback_add_test.go`, `cmd/task_test.go`); splitting them puts records in two places. Resolution goes in a `cmd/task` helper (the `cmd/task` harness `helpers_test.go:21-40` has no root pre-run), and the root pre-run is removed. Test peeks `task.New(".bit")` → a `projectStore(t)` helper (cmd/task create_test ×4, update_test ×6, helpers :69; cmd approve_test ×2, feedback_add_test path asserts). The worktree tests `cmd/root_test.go:44-94` become resolver tests (same expectations). MCP is untouched (its tests seed `<dir>/.bit` and use `ForRoot`).
7. **Delete `bp init`.** Remove `newInitCmd`, keep `writeClaudeWiring` (move it to `add.go`), delete `cmd/init_test.go`, drop the `init` call from the cmd fixture, `bp add` drops the config.toml default and refuses an unregistered folder with `.bit/` (after the registered check), and `scripts/install.sh:22` → `bp add`. Rewrite `cmd/add_test.go:17` (default-from-config test) as the refusal test.
8. **MCP resolves through the registry.** Handlers: resolve `root` (or cwd when empty) per call; resolver errors become tool errors. `runMCPServer(ctx, root, transport)` can keep its signature. Rewrite `seedTasks`/`seedConfig`/`seedEscapedResearch`/`seedOrderedTrack`/`seedDoneTrack` to register `dir` and write to the store dir; update returned-path assertions; the 36 `mcpSession` sites keep passing `dir`. Tool descriptions naming `.bit/` (`cmd/serve_mcp.go:76,84,102,109`) are BIT-49's sweep, so leave the text (or update now — `TestMCPToolDescriptions` asserts other strings only).
9. **Delete `bitdir` and `config.toml`.** `pluginState` root = resolved project path, else cwd, never an error (BIT-48 later drops the arg). Delete `task/config.go`, `configFileName`, the `Create` fallback, `task/store_test.go:678-737`, `bitdir/`.

### Verse 2 (JSON + `.md`)
1. Tasks: `Save/Load/List/listCompleted`, `Path/archivePath/completedPath`, `relocateInto` (move the pair), `trackExists`, globs → `*.json` (`task/store.go:443,467,614,626`). Every record gets `project` (needs Bar 1.4's code), `created_at`, `updated_at`, `branch`, `commit`.
2. Research topics: pair + `commits` (`task/research.go:20,83-95`).
3. Feedback notes: pair + `commits` (`task/feedback.go:19-33`).
Keep `task.Parse` (v1 frontmatter) — BIT-49's migrate reads v1 `.bit/` with it.

### Verse 3
1. `removed` column (edit the initial migration), `SetRemoved`/`ClearRemoved`, `ListProjects` excludes removed, `GetProjectByCode` and `GetProjectByPath` include removed; resolver `ErrRemoved`.
2. `bp remove` (new `cmd/remove.go`): confirm with outstanding list, soft-delete, `Relocate(id, true)` for each top-level task in `List()` (`tasks/` only, `task/store.go:466`). Needs Verse 2's pair-aware `relocateInto`.
3. `bp add` revive on a removed path; refuse a removed code from another folder; `bp list` hides removed.

## Part 4. Forced API changes
- `task.Store` needs the project code (track minting; also Verse 2's `project` field). `Store` today is `{root}` (`task/store.go:25-31`).
- `task.Config`, `(*Store).Config/SaveConfig/ConfigPath`, `configFileName` deleted.
- `store.Dir()` base changes to `bit`; add `ProjectDir(code)` (uppercase).
- `db.Open()` opens `main.db`; `orm` query set changes (see bar 1.1, 3.1).
- New resolver package with exported `Resolve`, sentinel errors, `ValidateCode`.
- `cmd/task` needs its own store getter (no root pre-run).
- `newRootCmd` body: no `init`, no `PersistentPreRunE`.
- `pluginState` source of the project root.
- `readProjectCode(cmd, existing)` loses its `existing` source (always empty).

## Part 5. What later tracks need from BIT-46
- **BIT-48:** `writeClaudeWiring(cmd, run, dir)` kept and called only on a fresh registration; `bp add`'s three branches (registered no-op / revive / fresh) cleanly separated. BIT-48 wants the row saved *before* wiring; today `add.go:66-73` wires first — either flip it in BIT-46 or leave it to BIT-48 knowingly. The resolver's "not a bit project; run `bp add`" error text (BIT-48 cites it). `pluginState` root source.
- **BIT-49 (migrate):** `GetProjectByCode` including removed rows (refuse another project's code), `GetProjectByPath` including removed ("already migrated"), `CreateProject` (insert last), `ValidateCode`, `store.ProjectDir(code)`, the store writer accepting caller-set `branch`/`commit`/`commits`/timestamps, `task.Parse` for v1 frontmatter. **Gap:** migrate "reads the main checkout's `.bit/` in a Claude worktree" — today that's `bitdir.worktreeCut` (`bitdir/bitdir.go:57-68`), which Bar 1.9 deletes, and a tracked `.bit/` (bit-pro) also exists inside each worktree, so a plain upward walk would find the worktree's copy first. Keep the segment cut as an exported helper (resolver or migrate) instead of deleting it.
- **BIT-47/50:** the empty git fields from Verse 2 and a Save path that doesn't clobber them.

## Part 6. Stall risks
- Stale `db/orm` (gitignored): bar checks must regenerate.
- Concurrent first open of `main.db`: several MCP servers + CLI can each run `CreateAndMigrate` on a fresh file; modernc's default busy handling under that race is unverified. Low odds, sandbox-only until cutover.
- `EvalSymlinks` on missing paths (MCP worktree test) — handle in the resolver.
- Verse 1 can't be dogfooded on bit-pro (it has `.bit/`, so `bp add` refuses) until BIT-49; manual checks need a fresh folder and `HOME` sandboxed (the fresh-registration path runs real `claude` wiring).
- Removed-project resolver error, revive, and removed-code refusal are Verse 3 only (column doesn't exist before).
