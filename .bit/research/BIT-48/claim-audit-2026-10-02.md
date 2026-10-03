# Claim audit + build map, 2026-10-02 (v2 @ 6a1d345, Claude Code 2.1.287, BIT-45/46 not landed)

**Checked:** every factual claim in the BIT-48 body, topics `decisions`, `index`, `review-2026-10-02`, and BIT-46 topics `soundness-2`, `init-and-update`, `plugin-dir-testing`, against the code and the real `~/.claude*` (read only, jq on bit keys only). `claude ... --help` only; no state-changing `claude` command run. Baseline `go build ./...` and `go test ./...` (sandboxed HOME/XDG) green.

Tally: 36 CONFIRMED, 2 WRONG, 5 STALE, 7 carried-unverified (sandbox evidence in `soundness-2` / the review's probe; re-running would change state).

## 1. Verdicts

### BIT-48 body
| Claim | Verdict | Evidence / corrected fact |
|---|---|---|
| `bp init` wrote project `.claude/settings.json`, installed the plugin at project scope, added a local `bit` MCP | CONFIRMED | `cmd/init.go:43` → `:51-71`; `claude/settings.go:19`; `claude/sync.go:38-39` (`update --scope project`, `install` fallback); `sync.go:26` `mcp add` with no `-s` = local |
| "`bp add` runs the same per-project wiring" | **WRONG** | Only when `<abs>/.bit` is absent (`cmd/add.go:66`); wiring runs **before** `CreateProject` (`:67` vs `:73`); plugin/MCP commands hit bp's cwd, not `<abs>` (`ExecRunner` sets no `Dir`, `claude/sync.go:13`); only `settings.json` lands in `<abs>` (`init.go:52`) |
| Notice only sees project-scope installs (`claude/plugin.go:26-31`) | CONFIRMED | `:28` requires non-empty `ProjectPath`; the record struct (`:16-21`) has no `scope` field yet |
| Four-command ensure order | CONFIRMED (design) | Scope defaults from `--help`: `marketplace add --scope` user, `plugin install -s` user, `plugin update -s` auto-detect, `mcp add -s` local |
| Marketplace must be user-level; project-declared install fails | carried | `soundness-2` sandbox. Real `~/.claude/settings.json` already lists `bit-pro` in `extraKnownMarketplaces` (masks it here) |
| `marketplace add` on an existing marketplace only says "already on disk" | carried | `soundness-2` |
| `SyncPlugin` updates before installing (`sync.go:34`) | CONFIRMED | |
| `claude mcp get` can't see scope | CONFIRMED | `mcp get --help`: only `-h`; it also health-checks (spawns) the server |
| `mcp add -s user` errors if the entry exists | carried | `soundness-2`; not re-run |
| `RegisterMCP` uses `mcp get` (`sync.go:22`) | CONFIRMED | func `:21`, get `:22`, add `:26` |
| BIT-46 deletes `bp init`, keeps `writeClaudeWiring` | CONFIRMED | BIT-46 body, Decision "`bp init` is deleted in Verse 1" |
| This track removes `writeClaudeWiring` and `WriteSettings` | CONFIRMED, with a catch | `WriteSettings` only caller `init.go:52`. **`claude/settings.go:12-15` also holds `pluginKey`/`marketplaceName`, used at `plugin.go:27,37,74`**: deleting the file whole breaks the build |
| Notice text `--scope project` at `cmd/root.go:119` | CONFIRMED | |
| Old local `bit` MCP entries run the same `bp serve mcp` | CONFIRMED | Real `~/.claude.json`: `projects[bit-pro]` and `projects[<client>]` `.mcpServers.bit = {stdio, bp, [serve, mcp], env {}}`; **no top-level `mcpServers` key** (step 4 will run on first add/migrate) |
| Dev runs must sandbox `HOME`; XDG doesn't cover `~/.claude*` | CONFIRMED | `store/store.go:11-19` only uses XDG/HOME for the data dir; `root.go:22` reads `os.UserHomeDir()` |
| Helpers `pluginSyncCalls`, `mcpRegisterCall` at `cmd/cmd_test.go:98-111` | **STALE** | `mcpRegisterCall` `:98-100`, `mcpLookupCall` `:102-104`, `pluginSyncCalls` `:106-112`. `mcpRegisterCall` is used only by `cmd/init_test.go:243,275,277`, which BIT-46 deletes, so by BIT-48 it is dead (golangci `standard` includes `unused`). `pluginSyncCalls` users: `cmd/add_test.go:111` (+`init_test.go:191`, gone) |
| Verse 1 Touches | **STALE (incomplete)** | Add `cmd/init.go` (BIT-46 leaves only the helper there), `cmd/add_test.go:62-134`, `claude/{sync,settings}_test.go`, `cmd/root.go:32` `claudeDir` (only user `init.go:52`), `cmd/testconst_test.go` (`prefixFlag` dead after BIT-46) |
| Verse 2 Touches `claude/plugin.go`, `cmd/root.go` | **STALE (incomplete)** | Add `claude/plugin_test.go:42-82,129-177`, `cmd/root_test.go:153,164,267` (see the HOME finding below) |
| Cutover checklist removes per-project wiring | CONFIRMED | `v2-sketch.md:114` |
| README "already describes setup" | CONFIRMED, catch | `README.md:18-43,59,82` all describe `bp init`; the table/Quickstart rewrite is BIT-49 Verse 3. `scripts/install.sh:18` already runs step 1 (`marketplace add`) |

### `decisions`
- Files owned incl. `cmd/root.go (:119)`: CONFIRMED. `SyncPlugin` updates first (`sync.go:34`): CONFIRMED. Helper range `:98-111`: **STALE** (as above).

### `index` / `review-2026-10-02`
- `settings.go` consts break `plugin.go`: CONFIRMED. User-scope MCP = top-level `mcpServers` in `$HOME/.claude.json`: carried (sandbox probe); real file consistent (local entries under `projects[...]`).
- `claude.Runner` returns only `error` (`sync.go:10`), `ExecRunner` folds output into the error (`:13-15`): CONFIRMED. `DirRunner` (`claude/dispatch.go:14`) is deleted by BIT-45: CONFIRMED.
- `HOME` temp in `cmd/add_test.go:19,64,148,191,226`: CONFIRMED. `plugin_test.go:53,69` user-scope case expects false; `:64-66` project cases; `:135`; `root_test.go:153`: CONFIRMED.
- Real `installed_plugins.json` `bit@bit-pro`: two `scope: project` entries (bit-pro 1.3.0, example 0.1.0), no user entry; marketplace clone `plugin.json` 1.3.0; `known_marketplaces.json["bit-pro"].installLocation = ~/.claude/plugins/marketplaces/bit-pro`: CONFIRMED. `CLAUDE_CONFIG_DIR` unset: CONFIRMED.
- **WRONG (by omission), review §2**: "a reader via `os.UserHomeDir()` is isolated in tests for free" holds for `add_test.go`, but **not** for `root_test.go:164` `TestExecute_NoPluginStateIsSilent` and `:267` `TestExecute_FiresTheRefresh`: they go `runSplit` (`cmd_test.go:125-139`) → `execute` → real `pluginState` with the operator's real `HOME`, and assert empty stderr. Silent today only because the temp project matches no `projectPath`. Once Verse 2 reads the user-scope entry, both fail whenever the real user install is behind the marketplace clone. Fix: `t.Setenv("HOME", t.TempDir())` in those tests (or BIT-46's new `initProject` does it).

### BIT-46 `soundness-2`
- `plugin.go:26-31`, `root.go:27`, `:119`, `sync.go:13,22`: CONFIRMED. Quiet annotations `cmd/serve_mcp.go:229`, `cmd/tui.go:22`: CONFIRMED. `runMCPServer` registers tools only (`serve_mcp.go:238+`): CONFIRMED. Real settings declares `bit-pro`: CONFIRMED. "local beats user", "plugin listed twice", "`marketplace add` writes `enabledPlugins` to user settings": carried. Note: real `~/.claude/settings.json` has `bit-pro` in `extraKnownMarketplaces` but `enabledPlugins["bit@bit-pro"]` unset (from `install.sh`'s add, no user install yet).

### BIT-46 `init-and-update`
- `init.go:25-44`, `settings.go:19`, `:17` (github source), `sync.go:21,33`, `plugin.go:10-34`, `add.go:26-80`, `:45-48`, `:66-70`, `root.go:50-68`, `db/open.go:19-38`, `store/store.go:22-23`: CONFIRMED. Its recommendation (a) "re-run add re-wires": **STALE**, superseded by the operator's "wiring only on first registration".

### BIT-46 `plugin-dir-testing`
- `bit/.claude-plugin/plugin.json` name `bit`, version 1.3.0, no `mcpServers`; no `bit/.mcp.json`; 9 skills, 3 agents; bit-pro `.claude/settings.json` enables `bit@bit-pro`; `bp` = `~/go/bin/bp`; cache 1.3.0 lacks `complete`: all CONFIRMED.

## 2. Build map

### Runner and "already exists"
`type Runner func(ctx, name, args...) error` (`claude/sync.go:10`). No stdout, no exit code; `ExecRunner` returns `"<cmd>: exit status N: <output>"`. So the ensure step **cannot** cleanly detect "already exists" from the Runner, only by substring-matching the error text (fragile). Use the review's design: read `<home>/.claude.json` top-level `mcpServers.bit` before step 4; on read/parse error, attempt the add. Steps 1-3 are exit-0 on re-run (`soundness-2`). `marketplace add` and `plugin install` both offer `--json` (one machine-readable line, same exit codes); useless while Runner discards stdout, so don't change Runner.

### Every site that changes or breaks
- `claude.WriteSettings`: def `claude/settings.go:19-85` (+`load`, `merge`, `marketplace` var `:17`); caller `cmd/init.go:52`; tests `claude/settings_test.go:12,55,102,129` (whole file).
- `claude.SyncPlugin`: def `sync.go:33-45`; caller `init.go:58`; tests `sync_test.go:36,52,72,130`; consts `projectScope`, `scopeFlag`, `bitProPlugin`, `updateSubCmd` (`sync_test.go:11-20`) partly dead.
- `claude.RegisterMCP`: def `sync.go:21-31`; caller `init.go:64`; tests `sync_test.go:85,98,114`; `mcpSubCmd`, `bitServer` consts.
- `pluginKey`/`marketplaceName` (`settings.go:12-15`): `plugin.go:27,37,74`. Move them before deleting `settings.go`.
- `writeClaudeWiring` (`cmd/init.go:51-71`): callers `init.go:43` (gone in BIT-46), `cmd/add.go:67`. Output asserted at `cmd/add_test.go:85-91`, `settings.json` at `:98-105`, calls at `:111-114`; `:204` asserts no `settings.json` (stays true). `cmd/init_test.go:162-299` (gone in BIT-46).
- `claudeDir` `cmd/root.go:32`: only user `init.go:52` (`bitdir/bitdir.go:10` has its own).
- Test helpers `cmd/cmd_test.go:98-112`; `cmd/testconst_test.go` `updateCmd` (still used if the new helper lists `marketplace update`), `prefixFlag`.
- Notice: `claude.InstalledVersion(home, projectRoot)` `plugin.go:10-34` needs a `Scope` field and `scope == "user"`; `PluginState(home, projectRoot)` `:59-71`; caller `cmd/root.go:21-28` (`bitdir.Root()` → whatever BIT-46 puts there); text `root.go:118-122`; tests `plugin_test.go:42-82` (`:53,64-66,69`), `:129-177` (`installRecordFor`, both `PluginState` tests), `root_test.go:153` (text), `:164`, `:267` (HOME).
- `cmd/root.go:30` `refreshMarketplace`, `plugin.go:73-75` `RefreshMarketplace`, `LatestVersion` `:36-57`: unchanged.

### Bar order (each green on `go build ./...` + `go test ./...`)
1. **V1** additive: `claude/global.go` `EnsureGlobal(ctx, run, home) error` + exported command list + `.claude.json` check; `claude/global_test.go` with the existing `recorder`. Nothing else touched.
2. **V1** `cmd/add.go`: fresh registration saves first, then a cmd helper (e.g. `ensureGlobalWiring(cmd, run)`) calls `EnsureGlobal` and, on error, prints the failed step and the commands. Delete `writeClaudeWiring` (and `cmd/init.go` if empty), `claudeDir`; rewrite `add_test.go:62-134`, replace `cmd_test.go:98-112` helpers; add tests for failure (project still registered), re-run (no calls), revive (no calls; needs BIT-46 V3).
3. **V1** delete `WriteSettings`, `SyncPlugin`, `RegisterMCP`, `settings_test.go`, their `sync_test.go` cases; move the two consts (keep `Runner`/`ExecRunner`). No callers after bar 2.
4. **V1** README: the four commands (near Install; leave the table to BIT-49 V3).
5. **V2** user-scope notice: `InstalledVersion(home)`/`PluginState(home)`, drop the root from `cmd/root.go:21-28`, text `--scope user`; flip plugin tests; sandbox HOME in `root_test.go:164,267`. Independent of 1-4; can go first.
Bars 2+3 can merge; 1 must precede 2.

### State BIT-46 leaves (from its body + `file-inventory`)
- `cmd/init.go`: only `writeClaudeWiring` (+ imports `claude`, `cobra`, `fmt`, `filepath`); `newInitCmd`, `readInteractivePrefix`, `initCmdUse` gone; `init_test.go` gone.
- `cmd/add.go`: registered → no-op; removed → revive (V3); `.bit/` → "run bp migrate"; code validated; then `writeClaudeWiring(cmd, run, abs)` then `CreateProject` (today's order).
- `cmd/root.go`: no `bitdir`, no `PersistentPreRunE`; `pluginState` passes resolved path or cwd (BIT-46 review default, not a body Decision); `:119` text unchanged; `claudeDir` still there. BIT-45 drops `daemon` and `newRootCmd`'s `lc`.

### Gaps
- G1 Order flip: v1 wires before `CreateProject` (`add.go:66-73`); BIT-46 doesn't change it; BIT-48's "saved before wiring" must flip it.
- G2 Exit status on wiring failure is undecided: project stays registered, but does `bp add` return non-zero? Planner needs a default (suggest: non-zero, so scripts notice).
- G3 BIT-46 doesn't say whether the `.bit/` stat guard (`add.go:66`) survives; after its refusal it's always true.
- G4 BIT-46 leaves `mcpRegisterCall`, `prefixFlag` dead → `just lint` (`unused`) red between BIT-46 and BIT-48 unless BIT-46 removes them; build/test unaffected.
- G5 `pluginState`'s root after BIT-46 is only a review default, not a body Decision.
- G6 BIT-49 reuse: `newMigrateCmd` needs `run claude.Runner` threaded from `newRootCmd` like `newAddCmd(run)` (`root.go:141`).
- G7 Root tests reading real HOME (above).
