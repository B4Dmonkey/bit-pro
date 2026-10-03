# v1 and a v2 dev build side by side

**Checked:** Q6 — binary name, install path, plugin source, risks.

## Facts
- `scripts/install.sh` builds to `$(go env GOBIN)` or `$GOPATH/bin` — here `/Users/appstack/go/bin/bp` — always named **`bp`**, then `claude plugin marketplace add B4Dmonkey/bit-pro`.
- The MCP server is registered as `claude mcp add bit -- bp serve mcp` (`claude/sync.go:26`): command `bp` resolved from **PATH at server launch**. Registered for bit-pro and a client project (in `~/.claude.json`).
- launchd plist (`daemon/plist.go`) pins `os.Executable()` of whatever ran `bp start`, label `com.github.b4dmonkey.bit-pro`, log in `~/.local/share/bit-pro/daemon.log`. Env only sets `BP_CLAUDE`.
- Plugin: marketplace source is `{"source":"github","repo":"B4Dmonkey/bit-pro"}` with no ref → default branch (`main`). `bp init`/`SyncPlugin` update from there. Skills on the `v2` branch reach no project until merged to main (memory: unpushed skill edits never reach a target project). Currently `v2` == `main` at `6a1d345`.
- Tool names are `mcp__bit__*` because the server is named `bit`; skills hardcode those names.

## Risks
1. **`just install` on the v2 branch overwrites the daily `bp`.** Every Claude session's MCP server and every CLI call switches to v2 instantly. Needs a separate target (e.g. `just install-dev` → `bp2` or `bit`, or a different GOBIN).
2. **Shared data dir.** Both binaries using `store.Dir()` → same `bit-pro/bit.db`; v2 schema migrations run on open and could alter tables v1's queries use. Mitigation: v2 uses `~/.local/share/bit/` (the operator's name), so dirs differ naturally; tests already sandbox via `XDG_DATA_HOME`.
3. **Daemon label collision.** v2 `start` with the same launchd label replaces v1's daemon (plist rewritten to the v2 exe). Needs a distinct label for dev, or don't run v2's daemon until cutover.
4. **MCP name.** If v2 registers its server under a new name, tool names change (`mcp__bit2__*`) and v1 skills won't call them. If it keeps `bit`, only one of v1/v2 can serve a given project — consistent with the sketch's "single cutover per project".
5. **Plugin skills.** v2 skill edits can't ship via GitHub main without hitting all v1 projects. Options: a ref'd marketplace entry for v2, a separate plugin name, or local `--plugin-dir` during dev. Unknown: whether Claude Code's github marketplace source supports a `ref` for this — verify before relying on it.
6. **Version string:** install stamps `cmd.version` from `plugin.json`; the plugin-behind notice compares plugin versions, not binary — a v2 dev build shows the same version as v1 unless changed.
7. **Merge-back:** v1 fixes land on main while v2 lives on a branch; `.bit/` is tracked in bit-pro, so BIT tasks edited on both branches will conflict. After migrating bit-pro itself, that goes away (data leaves git).
