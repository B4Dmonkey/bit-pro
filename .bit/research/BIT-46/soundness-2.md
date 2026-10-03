# BIT-46 soundness pass 2 (v2 @ 6a1d345, Claude Code 2.1.286)

> **Note (2026-10-01):** this pass predates the split. "Verse 1" here is the original BIT-46 Verse 1. Its wiring parts are now BIT-48 Verse 1 (the ensure step) and BIT-48 Verse 2 (the notice). The "Blocking" section below is resolved by BIT-49. See BIT-46 topic `decisions`.

**Checked:** the revised body, above all the new "Claude wiring is global" decision, against the code and against the `claude` CLI. The CLI was probed in a throwaway `HOME=<scratchpad>/home`, so the operator's real `~/.claude*` was only read, never written.

## Global wiring: CLI facts (sandbox-verified)
- **A marketplace declared only in project `extraKnownMarketplaces` is not visible to the CLI.** In a sandbox project whose `.claude/settings.json` declares `bit-pro` (and is marked trusted in the sandbox `.claude.json`), `claude plugin marketplace list` printed "No marketplaces configured". `claude plugin install bit@bit-pro --scope user` failed with `Plugin "bit" not found in marketplace "bit-pro"`, and so did v1's own `marketplace update bit-pro` + `install --scope project` sequence. After `claude plugin marketplace add B4Dmonkey/bit-pro` (its `--scope` defaults to user), `install --scope user` succeeded.
  - `marketplace add` itself writes `extraKnownMarketplaces.bit-pro` and `enabledPlugins["bit@bit-pro"]` into **user** `settings.json`. So `claude/settings.go` `WriteSettings` becomes dead code, not a "user scope" rewrite.
  - Re-running is safe: a second `marketplace add` exits 0 ("already on disk"), a second `install --scope user` exits 0 ("already installed"), and `update --scope user` exits 0.
  - This machine hides the problem: the real `~/.claude/settings.json` already declares `bit-pro` in `extraKnownMarketplaces`, so `install --scope user` would work here without `marketplace add`, but not on a fresh machine.
  - **Verdict:** the ensure step must be `plugin marketplace add B4Dmonkey/bit-pro` then `plugin install bit@bit-pro --scope user`. The Decision names only `claude plugin install --scope user`.
- **`claude mcp get bit` can't tell user from local.** It takes no `--scope` (`claude mcp get --help`). In the sandbox it exited 0 from a project that had a local `bit` entry and 1 from a bare folder. `claude/sync.go:22` `RegisterMCP` does get-then-add from bp's cwd, and `ExecRunner` sets no `Dir` (`claude/sync.go:13`). Run from bit-pro or a client project, which today hold local `bit` entries in `~/.claude.json`, it would skip the add, and **no user-scope entry would ever be created**.
  - `claude mcp add -s user bit ...` exits **1** when a user entry already exists.
  - **Verdict:** "ensure once" needs its own user-scope check (e.g. read `mcpServers.bit` at the top level of `~/.claude.json`, or treat "already exists in user config" as success), not `mcp get`.
- **Precedence: local beats user.** With both scopes set, `claude mcp get bit` reported `Scope: Local config`. When the endpoints differ, `claude mcp list` prints a `[Conflicting scopes]` warning. When they're identical (`bp serve mcp` in both), it lists one server and warns about nothing. So the v1 local entries shadow the user entry harmlessly until cleanup. That's true only while both run the same `bp` on PATH, which they do after cutover's `just install`.
- **The plugin at both scopes is listed twice.** After a user-scope and a project-scope install, `claude plugin list` shows `bit@bit-pro` twice (Scope: user and Scope: project). Real `installed_plugins.json` has project-scope entries for bit-pro (1.3.0) and example (**0.1.0**). Which copy a session loads was **not verified**. Cutover step 3 removes the project installs, so this only matters between steps 1 and 3.

## "Plugin behind" notice: per-project, and it must change (should-fix)
- `claude/plugin.go:26-31` `InstalledVersion` matches only entries with a non-empty `projectPath` equal to `bitdir.Root()` (`cmd/root.go:27`). A user-scope entry has **no `projectPath`** (sandbox `installed_plugins.json`: `scope: user`, no projectPath), so once the project installs are gone, the notice never fires.
- The text hardcodes `--scope project` (`cmd/root.go:119`).
- **Fix:** read the `scope: "user"` entry, drop the project-root argument, and say `--scope user`. The Decision's "the existing notice points at" the global update is false as written. `claude/plugin.go` and `cmd/root.go` are already in Verse 1's Touches, so no Touches change is needed. `serve mcp` and `tui` are quiet-annotated (`cmd/serve_mcp.go:229`, `cmd/tui.go:22`), so the notice only shows on CLI commands.

## Unregistered folders once MCP and plugin are global (should-fix: decide)
- Every Claude session on the machine then starts `bp serve mcp`, and bit's skills and agents load everywhere.
- Today startup touches no files: `runMCPServer` only registers tools (`cmd/serve_mcp.go:238-276`) and resolves `bitdir.ForRoot(root)` per call. If v2 keeps resolution per call, the server connects cleanly everywhere and each call fails with the resolver's error. Resolving at startup would instead show a failed MCP server in every non-bit project.
- Noise: 10 `mcp__bit__*` tools plus 9 skills and 3 agents in every session. Broad triggers like "retro", "make a plan" and "check the work" can fire in unrelated projects and then fail on the MCP call.
- Also unspecified: the error for an unregistered folder **without** `.bit/` (the body only specifies the `.bit/` → "run bp migrate" case), and the fact that `db.Open` creates `main.db` from any folder.

## Resolved by BIT-49 (was "Blocking"): cutover never creates the global wiring
**Resolved:** BIT-49's `bp migrate` ensures the global wiring on a first migration, and `bp add` checks "already registered" before "has `.bit/`" (operator, 2026-10-01). Cutover step 2 ("the first migrate sets up the global wiring") reflects this. The original finding follows.
- Cutover (v2-sketch "Cutover") runs `bp migrate` in each project, then step 3 removes every per-project plugin install and local `bit` MCP entry. Only `bp add` ensures the global wiring, and:
  - migrate "registers the project last", and no Decision says it ensures wiring;
  - `bp add` refuses a folder that still has `.bit/`. migrate never touches the source, so every migrated project keeps `.bit/` (bit-pro until `git rm`, the others indefinitely, since `.bit/` is untracked there). So `bp add` can't be run on any of them.
- Result: after step 3 there's no `bit` plugin or MCP anywhere. **Fix:** migrate also ensures global wiring, or add checks "registered" before "has `.bit/`" (earlier note `init-and-update`), or cutover gets an explicit wiring step.

## Should-fix: dev testing of `bp add` mutates the daily setup
The XDG sandbox (Decision "Hard cutover") doesn't sandbox Claude. A dev `bp add` runs the real `claude` (`claude/sync.go:12`) against `~/.claude*`. It would install the GitHub-main (v1) plugin at user scope and a user `bit` MCP pointing at the v1 `bp` on PATH, into every session on the machine. That contradicts "Nothing in the daily setup changes." Unit tests stub the Runner, but Verse 1's manual check needs `HOME=<sandbox>` (verified above: plugin and mcp commands honour it) or a no-wiring flag.

## Other claims re-checked
- `task/store.go:216` `s.Config()` for track IDs: **confirmed**. It's in Verse 1 Touches now.
- `tools/example/reset.sh:4` mentions `bp init`: **confirmed** (a comment describing the "blank" checkpoint). The Risk stands.
- bitdir callers: `cmd/{approve,feedback_add,init,tui,root}.go` and `cmd/task/{create,list,update,delete,complete,move,read}.go`, plus `cmd/serve_mcp.go:294-489` (`ForRoot`): **all covered** by Verse 1 Touches.
- Stale evidence pointer: the Decision cites topic `init-and-update` for global wiring, but that note analyses per-project wiring and recommends re-wiring per project (option a). This note is the evidence for user scope.
- nit: v2-sketch's cutover step 3 lists removing the local `bit` MCP entries but not verifying that the user-scope entry exists first.