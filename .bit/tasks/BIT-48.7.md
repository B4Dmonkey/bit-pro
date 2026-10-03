---
id: BIT-48.7
title: The plugin-behind notice reads the user-scope install and names the user-scope update
status: done
approved: true
phase: 2
phase_label: global plugin-behind notice
---
## **Verse 2**

With the plugin installed once at user scope, the notice must read the `scope: "user"` record, not a project's, and tell the operator to update at user scope. A user-scope record that today's `ProjectPath` filter ignores forces the change; the project root BIT-46.10's `pluginRoot` computed has nothing left to feed. Independent of Verse 1; can be built first.

## Scope
- `claude/plugin.go`:
  - `InstalledVersion(home string) (string, bool)`: add `Scope string `json:"scope"`` to the record struct, drop `ProjectPath` and the `projectRoot` parameter, return the first `bit@bit-pro` entry with `Scope == "user"`.
  - `PluginState(home string)`: drop `projectRoot`.
  - `LatestVersion`, `RefreshMarketplace` unchanged.
- `cmd/root.go`:
  - `pluginState` calls `claude.PluginState(home)`.
  - delete `pluginRoot()` (BIT-46.10) and the imports only it used (`project`, possibly `context`).
  - `notice`: `... — run: claude plugin update bit@bit-pro --scope user`.
- Test conversion (`.claude/rules/go-tests.md`). No earlier bar touches `claude/plugin_test.go`, so this bar converts it first, as a pure restructure: same cases, same assertions, still green. Each flat test `TestX_Suffix` becomes `t.Run("<suffix as lowercase words>", ...)` inside `TestX`, body unchanged:
  - `TestPluginState_ReportsThisProject` → `TestPluginState/reports this project`; `TestPluginState_SilentWhenEitherReadFails` → `TestPluginState/silent when either read fails` (its table stays inside the subtest).
  - `TestStart_DoesNotWaitForTheChild` → `TestStart/does not wait for the child`; `TestStart_MissingBinaryIsSilent` → `TestStart/missing binary is silent`.
  - `TestInstalledVersion` and `TestLatestVersion` are already unit-named and stay.
  - `cmd/root_test.go` was converted in BIT-45.1, so it needs no step here.
- `claude/plugin_test.go`: rewrite `TestInstalledVersion` (below); `installRecordFor(projectRoot)` becomes a user-scope record with no `projectPath`; `TestPluginState/reports this project` → `TestPluginState/reports the user install`, and both `TestPluginState` subtests drop `projectRoot`.
- `cmd/root_test.go`: delete `TestPluginRoot` (removal); update the expected notice in `TestExecute/behind plugin writes notice to stderr` (`:153`). Leave BIT-46.10's `HOME`/`XDG_DATA_HOME` sandboxing in place: those tests now read `installed_plugins.json` under `HOME`, so they still need it.
- Not needed here: `TestExecute/no plugin state is silent` (`:164`) and `TestExecute/fires the refresh` (`:267`) already get a temp `HOME` from BIT-46.7's `initProject`.

## TDD cycle

0. **Convert (pure restructure):**
   - [ ] Convert `claude/plugin_test.go` as listed under Test conversion. `just test` stays green.

1. **Write test (RED):**
   - [ ] `TestInstalledVersion` (table-driven, rewritten; one row per case below)
     - **Behavior:** the notice compares against the plugin install every Claude session loads, the user-scope one.
     - **Setup / Assertions** (`InstalledVersion(home)`):
       - `TestInstalledVersion/user only`: `{"bit@bit-pro": [{"scope": "user", "version": "0.1.0"}]}` → `"0.1.0", true`.
       - `TestInstalledVersion/user beside a project install`: `[{"scope": "project", "projectPath": "/p/a", "version": "0.1.0"}, {"scope": "user", "version": "0.2.0"}]` → `"0.2.0", true`.
       - `TestInstalledVersion/project installs only`: two `scope: project` entries → `"", false`.
       - `TestInstalledVersion/another plugin's user install only` (`go@go-skills`) → `"", false`.
       - `TestInstalledVersion/file absent`, `TestInstalledVersion/file malformed` (`{`), `TestInstalledVersion/no plugins recorded` (`{"plugins": {}}`) → `"", false`.
     - **Boundary:** scope `user` vs `project` (the transition state with both, which the real machine has until cutover cleanup), and the plugin key.
   - [ ] Confirm fails: first as a compile error on the dropped argument; with the parameter removed and the body still filtering on `ProjectPath`, the user rows return `"", false` — the real red.

2. **Implement (GREEN):**
   - [ ] The `Scope` filter; drop the parameter through `PluginState` and `pluginState`; delete `pluginRoot` and `TestPluginRoot`.

3. **More tests (RED → GREEN):**
   - [ ] `TestExecute/behind plugin writes notice to stderr` (modified subtest): expected stderr `bp: bit plugin 0.1.0 → 0.2.0 available — run: claude plugin update bit@bit-pro --scope user\n`. *Boundary:* the command names user scope explicitly (`plugin update -s` auto-detects, and could pick a leftover project install). Confirm fails on the old `--scope project` text.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `grep -rn "pluginRoot\|projectRoot\|--scope project" --include='*.go' .` finds nothing
- [ ] `test ! -e ~/.local/share/bit/main.db`

## User verifies
In a sandbox (never `just install`):
- [ ] Before exporting `HOME`: `SB=$(mktemp -d); M=$SB/home/.claude/plugins; mkdir -p $M/marketplaces/bit-pro/bit/.claude-plugin $SB/data $SB/bin; printf '#!/bin/sh\nexit 0\n' > $SB/bin/claude; chmod +x $SB/bin/claude; just db-gen-queries && go build -o $SB/bin/bp .`, then `echo '{"name":"bit","version":"0.2.0"}' > $M/marketplaces/bit-pro/bit/.claude-plugin/plugin.json`, `echo '{"plugins":{"bit@bit-pro":[{"scope":"user","version":"0.1.0"}]}}' > $M/installed_plugins.json`, `export HOME=$SB/home XDG_DATA_HOME=$SB/data PATH=$SB/bin:$PATH`.
- [ ] `cd $SB && bp list` prints `bp: bit plugin 0.1.0 → 0.2.0 available — run: claude plugin update bit@bit-pro --scope user` on stderr.
- [ ] `echo '{"plugins":{"bit@bit-pro":[{"scope":"project","projectPath":"/p/a","version":"0.1.0"}]}}' > $M/installed_plugins.json` → `bp list` prints no notice.
- [ ] Whole slice: the operator is told when the global plugin is behind, with a command that updates the global install.

## Commit (user)
`feat(bit): plugin-behind notice reads the user-scope install`