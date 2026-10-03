---
id: BIT-48.3
title: A fresh bp add registers first, then ensures the global wiring
status: done
approved: true
phase: 1
phase_label: bp add sets bit up for the whole machine
---
## **Verse 1**

`bp add`'s fresh branch swaps the per-project `writeClaudeWiring` for the global ensure step and saves the registration before wiring. The rewritten happy-path test forces the new calls and output order; a refused code (UNIQUE clash) that still wires today forces register-first. The registered, revive, `.bit` refusal and removed-code branches from BIT-46.5/.21/.22 already assert no runner calls and stay as they are.

## Scope
- new `cmd/wiring.go`: `func ensureGlobalWiring(cmd *cobra.Command, run claude.Runner) error`. `home, err := os.UserHomeDir()`; on error return it. Print `Setting up bit in Claude Code (user scope)...` to `cmd.OutOrStdout()`, then `return claude.EnsureGlobal(cmd.Context(), run, home)`. **Seam for BIT-49:** `bp migrate` calls this after it registers a project for the first time; `newMigrateCmd(run claude.Runner)` gets `run` from `newRootCmd` as `newAddCmd(run)` does.
- `cmd/add.go` fresh branch: `CreateProject`, print `added <CODE> <path>`, then `return ensureGlobalWiring(cmd, run)`. Remove the `writeClaudeWiring` call and any import it alone needed. The other early returns are untouched.
- delete `cmd/init.go` (after BIT-46.7 it holds only `writeClaudeWiring`).
- `cmd/root.go`: delete `const claudeDir` (its only user was `writeClaudeWiring`).
- `cmd/cmd_test.go`: delete `mcpLookupCall` and `pluginSyncCalls`. `cmd/testconst_test.go`: delete `claudeBin` once unused (`updateCmd` stays, `cmd/task_test.go:53` uses it). Neither file has a test function to convert (`cmd_test.go` holds only `TestMain`).
- `cmd/add_test.go` `TestAddCmd/initialises a project without bit`: new output and calls (below); drop the `prompt`/`(` check's split on `Bringing` (split on `added` instead) and the `.claude/settings.json` assertions. Keep the "no `.bit`" check (BIT-46.11 changed it from "no `.bit/config.toml`") and the project-row assertions.
- `cmd/add_test.go`, other subtests that assert a fresh add's exact output (at least `TestAddCmd/uppercases a typed code`, both rows): the expected output becomes `"Project code: added FOO <canonical dir>\nSetting up bit in Claude Code (user scope)...\n"`, replacing whatever per-project wiring lines BIT-46 left in it. Drop the now-meaningless "no `.claude/settings.json`" checks in `TestAddCmd/rejects an empty code` and `TestAddCmd/refuses an invalid code` (wherever BIT-46 left them), since nothing writes that file any more. The other assertions in those subtests stay.
- `cmd/add_test.go` is already in the one-`TestAddCmd` shape (converted by BIT-46, which first touches it in BIT-46.3), so this bar does no conversion step.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestAddCmd/initialises a project without bit` (modified `t.Run` subtest)
     - **Behavior:** registering a new project sets bit up for the whole machine, after the project is saved.
     - **Setup:** as today (temp `HOME`, `XDG_DATA_HOME=""`, `t.Chdir(t.TempDir())`, recording runner, stdin `BIT\n`, `bp add .`); no `<HOME>/.claude.json`.
     - **Assertions:** output == `"Project code: added BIT <canonical dir>\nSetting up bit in Claude Code (user scope)...\n"`; `calls` equals `claude.GlobalWiring()` (four argvs; the literals are pinned in `TestEnsureGlobal` in `claude/global_test.go`); one project row with code `BIT`.
     - **Boundary:** fresh path × no user MCP entry — all four steps run, and `added` precedes the wiring line.
   - [ ] Confirm fails: output has the old `Bringing the bit plugin current...` lines and calls are the project-scope sync.

2. **Implement (GREEN):**
   - [ ] `ensureGlobalWiring`; the fresh-branch reorder; the deletions above.

3. **More tests (RED → GREEN):**
   - [ ] `TestAddCmd/refuses a code already taken` (BIT-46.5's subtest, extended): give the second `bp add <tmp>/b` (`FOO`) its own recording runner; assert it recorded no calls.
     - **Behavior:** a registration that fails doesn't touch Claude.
     - **Boundary:** `CreateProject` fails (UNIQUE `code`) — the wiring must come after it.
     - Confirm fails before the reorder lands (wiring ran first). If the reorder already made it pass, record that it went green with the GREEN step.

## Claude verifies
- [ ] `just lint` and `just test` pass (`unused` catches any helper left behind)
- [ ] `grep -rn "writeClaudeWiring\|claudeDir\|pluginSyncCalls\|mcpLookupCall" --include='*.go' cmd/` finds nothing (scoped to `cmd/`: the main-checkout cut BIT-46.11 moves out of `bitdir` may keep its own `claudeDir` const)
- [ ] `test ! -e ~/.local/share/bit/main.db`

## User verifies
- none here (the verse's end-to-end check is on its last bar)

## Commit (user)
`feat(bit): bp add registers first, then ensures user-scope Claude wiring`