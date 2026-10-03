---
id: BIT-48
title: 'v2: global Claude wiring'
status: todo
approved: true
---
## Why
v1 sets bit up in Claude Code one project at a time. `bp init` wrote a project `.claude/settings.json`, installed the plugin at project scope and added a local `bit` MCP entry. `bp add` runs the same per-project wiring, but only for a folder with no `.bit/` (`cmd/add.go:66`), before it saves the registration (`:67` vs `:73`), and from bp's own folder rather than the target path, because the `claude` runner sets no working directory (`claude/sync.go:13`; corrected 2026-10-02). So each new project needs its own setup, each release needs a refresh in every project, and the "plugin behind" notice only sees project-scope installs (`claude/plugin.go:26-31`). Once BIT-46 resolves every project from one central registry, there's no per-project state left for that wiring to serve, so bit can be set up once per machine.

## Summary
One idempotent "ensure global wiring" step installs the marketplace, the plugin and the `bit` MCP server at user scope. `bp add` runs it when it registers a new project, and so does `bp migrate` (BIT-49). Re-running either one on a registered project doesn't run it again. The per-project wiring helper and settings writer are removed. (`bp init` is already gone, deleted in BIT-46.) The plugin-behind notice checks the user-scope install and tells the operator to run a global update.

## Decisions
- **Depends on BIT-46.** `bp add` registers the project in BIT-46, and this track adds the wiring step to it.
- **Claude wiring is global, not per project.** bit's MCP server, skills and agents load in every Claude session, including non-bit projects. The operator accepts that: in a non-bit project, the broad triggers only hit the resolver's "not a bit project" error.
- **The ensure step runs these in order:** `claude plugin marketplace add B4Dmonkey/bit-pro`, then `claude plugin marketplace update bit-pro`, then `claude plugin install bit@bit-pro --scope user`, then registers the `bit` MCP server at user scope.
  - The marketplace has to be added at user level. Installing from a marketplace declared by a project fails ("Plugin bit not found").
  - **The update step is a Claude default (2026-10-01).** On a machine that already has the marketplace, `marketplace add` only says "already on disk" and doesn't refresh it, so `install` could take a stale cached plugin. v1's `SyncPlugin` already updates before installing (`claude/sync.go:34`). That makes four `claude` commands.
  - `claude mcp get` can't see scope, and `mcp add -s user` errors if the entry already exists. So the step checks for a user-scope entry itself, instead of using `RegisterMCP`'s `mcp get` check (`claude/sync.go:22`).
  - Evidence: BIT-46 research topic `soundness-2`.
- **The MCP registration is `claude mcp add -s user bit -- bp serve mcp`, with `bp` by name (Claude default, 2026-10-02).** It matches v1's local entry, which runs the same command.
- **The user-scope check reads the top-level `mcpServers.bit` key of `<home>/.claude.json` (Claude default, 2026-10-02).** That's where Claude Code keeps user-scope servers (checked in a sandbox copy). `claude.Runner` returns only an error, with no output or exit code (`claude/sync.go:10`), so "already exists" can't be read from the command itself without matching error text. If the file can't be read or parsed (every running session rewrites it), the step runs the add anyway, and a duplicate shows up as a failed step. The real file has no top-level `mcpServers` today, so the first `add` or `migrate` registers the server. `CLAUDE_CONFIG_DIR` is ignored.
- **The four command lines live in one function in `claude/`, and a thin helper in `cmd/` calls it for both `bp add` and BIT-49's `bp migrate` (Claude default, 2026-10-02).** The failure message, the tests and the README then quote the same commands. `migrate` gets the `claude.Runner` from `newRootCmd`, the same way `newAddCmd` does.
- **The wiring runs only when a project is first registered (operator, 2026-10-01).** Re-running `bp add` or `bp migrate` on a registered project doesn't set up the wiring again. The wiring is global and user-scope, and this track owns it.
- **Reviving a removed project with `bp add` doesn't run the wiring (Claude default, 2026-10-01).** It was registered before, and the wiring is global.
- **Repairing lost global wiring is out of scope (operator, 2026-10-01).** The global wiring is the user-scope plugin and MCP registration this track sets up. If the operator removes it, they re-run the documented `claude` commands. There's no bp command for it.
- **The README documents the four `claude` commands (Claude default).** That repair path needs somewhere to live, and the README already describes setup.
- **A fresh registration is saved before the wiring runs (Claude default).** If a wiring step fails, `bp add` reports which step failed and prints the `claude` commands to run by hand. The project stays registered. A re-run doesn't re-wire, so printing the commands is the only useful recovery, and it matches the out-of-scope decision above. This track reverses today's order, wire then register (`cmd/add.go:66-73`), which BIT-46 keeps.
- **When a wiring step fails, `bp add` exits non-zero after printing the commands (Claude default, 2026-10-02).** The project stays registered, but the operator still has work to do, and a zero exit would hide it.
- **`bp init` is deleted in BIT-46, not here (operator, 2026-10-01).** This track removes what BIT-46 kept: the per-project wiring helper `writeClaudeWiring` and `WriteSettings`. Nothing needs re-running to refresh a project. After a release, the only refresh is one global plugin update.
- **The settings constants move before `claude/settings.go` is deleted (fact, 2026-10-02).** `plugin.go:27,37,74` use the constants at `settings.go:12-15`.
- **The plugin-behind notice checks the user-scope install,** and its text drops `--scope project` (`cmd/root.go:119`). It reads the `scope: "user"` entry and no longer takes a project root (Claude default, 2026-10-02).
- **Tests that reach the notice sandbox `HOME` (Claude default, 2026-10-02).** `cmd/root_test.go:164` and `:267` reach the operator's real `HOME` through `pluginState` and expect no warning. Once the notice reads the user-scope install, they'd fail whenever the real install is behind.
- **Old local `bit` MCP entries run the same `bp serve mcp`,** so they don't conflict with the new user-scope entry before cleanup.
- **Dev runs sandbox `HOME` as well as `XDG_DATA_HOME`.** The wiring writes to `~/.claude*`, which the XDG sandbox doesn't cover (BIT-45 topic `testing`). As in BIT-46, dev builds never use `just install` on `v2`, because it replaces the daily `bp` (Claude default, 2026-10-01).
- **Each bar leaves the build and tests green (Claude default, 2026-10-01).** The test helpers `pluginSyncCalls`, `mcpLookupCall` and `mcpRegisterCall` (`cmd/cmd_test.go:98-112`) assert on the per-project calls and change with them. `mcpRegisterCall` is used only by `cmd/init_test.go`, so BIT-46 deletes it with `bp init`.
- **"Green" means `just lint` and `just test` pass (Claude default, 2026-10-02).** That's what the pre-commit hook runs (`.pre-commit-config.yaml`).
- **Cutover belongs to the operator.** The operator alone decides when v2 is ready and merges the branch, and no track gates it. Removing each project's old per-project wiring is a manual step on the cutover checklist in `v2-sketch.md`.

## Verses
- [ ] Verse 1 — `bp add` sets bit up for the whole machine. Registering a new project ensures the user-scope marketplace, plugin and MCP server. Re-running `bp add` on a registered project changes nothing. If a wiring step fails, the operator is told which `claude` commands to run, and the same commands are in the README.
  Touches: `claude/{sync,settings}.go` and their tests, `cmd/add.go` and `cmd/add_test.go` (`:62-134`), the `writeClaudeWiring` helper BIT-46 kept in `cmd/init.go`, `claudeDir` (`cmd/root.go:32`), the helpers in `cmd/cmd_test.go`, `cmd/testconst_test.go`, and `README.md`. See BIT-46 topics `init-and-update` and `soundness-2`, and BIT-48 topics `review-2026-10-02` and `claim-audit-2026-10-02`.
- [ ] Verse 2 — The operator gets told when the global plugin is behind, with the right update command.
  Touches: `claude/plugin.go` and `claude/plugin_test.go` (`:42-82`, `:129-177`), `cmd/root.go`, `cmd/root_test.go` (`:153`, `:164`, `:267`). It doesn't depend on Verse 1 and can be built first.

## References
- `v2-sketch.md` (repo root): the v2 direction and the cutover checklist.
- `.bit/research/BIT-48/`: start at `index`. `decisions` holds the operator decisions of 2026-09-30 and 2026-10-01. `review-2026-10-02` and `claim-audit-2026-10-02` (2026-10-02) verify each claim, list the call sites and give a green bar order. They win where older topics differ.
- `.bit/research/BIT-46/`: `soundness-2` (the global wiring evidence), `init-and-update` (what init and add do today; its "re-run `bp add` re-wires" advice was overruled by the operator), `plugin-dir-testing`.
- `.bit/research/BIT-45/`: `testing`.