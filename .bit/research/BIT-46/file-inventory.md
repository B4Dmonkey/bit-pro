# Files that reference `.bit/`, `bitdir` or `config.toml` (v2 @ 6a1d345)

> **Updated 2026-10-01 for the track split.** The section headings now name the track and verse that own each file. "Verse N" means the current verse of that track, not the original BIT-46 numbering. `bp init` is deleted in BIT-46 Verse 1, and its wiring helper and the other wiring files belong to BIT-48 (BIT-46 topic `decisions`).
>
> **Readiness pass (2026-10-01):** added the files the grep below misses because they call `bp init` rather than naming `.bit/` (the test fixture, `scripts/install.sh`), the wiring test helpers, the complete evals and `bit/agents/bot.md:41`.

This is the checklist of files the v2 tracks have to change. Regenerate it with:
`grep -rlE '"\.bit"|bitdir\.|\.bit/|config\.toml|Config\(\)' --include='*.go' .` (Go) and `grep -rlE '\.bit/|\.bit\b' bit/ README.md update` (non-Go). Also `grep -rn '"init"\|bp init' --include='*' cmd scripts README.md`.
The `daemon/` and `task/counts*` files are omitted because BIT-45 deletes them.

## Go: resolution and config (BIT-46 Verse 1)
- `bitdir/bitdir.go`: replaced by the project resolver. `worktreeCut` goes.
- `task/store.go`: track `Create` mints its ID from `Config().Prefix` (:216-221). v2 takes the prefix from `projects.code`.
- `task/config.go`: the `config.toml` reader goes.
- `cmd/root.go`: `pluginState` needs a repo root (:27), and `PersistentPreRunE` calls `bitdir.Resolve()` (:136). The notice text (:119) belongs to BIT-48.
- `cmd/add.go`: reads `.bit/config.toml` (:51) and skips wiring when `.bit/` exists (:66). It checks "already registered" before "has `.bit/`".
- `cmd/init.go`: writes `.bit/config.toml` (:23, :39, :75). BIT-46 Verse 1 deletes the `bp init` command and keeps `writeClaudeWiring` (:51) for `bp add`.
- `cmd/approve.go`, `cmd/feedback_add.go`, `cmd/tui.go` (:24): the `bitdir.Current()` callers.
- `cmd/task/{complete,create,delete,list,move,read,update}.go`: the `bitdir.Current()` callers.
- `cmd/serve_mcp.go`: reads `CLAUDE_PROJECT_DIR` once at start (:231). Resolve per call instead, with a cwd fallback.
- `scripts/install.sh`: its closing hint says "Run 'bp init'". It points to `bp add`.

## Go tests that create `.bit/` (18, BIT-46 Verse 1)
- `cmd/{add,approve,feedback_add,init,mcp_harness,root,serve_mcp_research,serve_mcp_write,serve,task}_test.go` (`init_test.go` goes with the command)
- `cmd/task/{complete,create,delete,helpers,list,read,update}_test.go`
- `task/store_test.go`
- The shared fixture `initProject`: `cmd/cmd_test.go:89-96` runs `bp init`, and `cmd/task/helpers_test.go` runs `SaveConfig`. About 75 call sites use them. It's rewritten to register the project under a `t.TempDir()` `XDG_DATA_HOME`.

## Go: Claude wiring (BIT-48)
- `claude/sync.go`: `SyncPlugin` (project scope) and `RegisterMCP` (scope-blind `mcp get`, :22).
- `claude/settings.go`: `WriteSettings` becomes dead code.
- `claude/plugin.go`: the plugin-behind check matches only project-scope installs (:26-31).
- The `writeClaudeWiring` helper BIT-46 kept, and the notice text in `cmd/root.go` (:119).
- Test helpers `pluginSyncCalls` and `mcpRegisterCall` (`cmd/cmd_test.go:98-111`).

## Skills, agents, plugin and docs (BIT-49 Verse 3; retro, learn and feedback are BIT-49 Verse 1)
- `bit/.claude-plugin/plugin.json` (:5, description)
- `bit/agents/{bot,bot-dev,ruler}.md`
- `bit/skills/{analyze,check,complete,do,feedback,learn,plan,retro,scope}/SKILL.md`
- `bit/skills/analyze/evals/evals.json`
- `README.md`, including the command table (`bp init` out; `bp add`, `bp remove`, `bp migrate` in)
- Help text: `cmd/task/complete.go:12` (`cmd/init.go:23` goes with the command in BIT-46)

## Later edits to the same files (BIT-47, BIT-50)
- BIT-47: `bit/skills/{do,plan}/SKILL.md`, `bit/agents/bot-dev.md`, a new `bit/skills/commit/`.
- BIT-50: `bit/skills/complete/SKILL.md` and `bit/skills/complete/evals/evals.json`, `bit/skills/do/SKILL.md` (sign-off), `bit/agents/bot.md:41`.

## v1-only (left as they are)
- `update/normalize.sh`, `update/normalize_test.sh`, `update/README.md`: the v1 one-off ID normaliser. migrate refuses a store that isn't normalised; it doesn't run the normaliser.