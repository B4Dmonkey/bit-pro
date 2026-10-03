---
id: BIT-45.1
title: bp start/stop/status and serve daemon are unknown commands
status: todo
phase: 1
phase_label: No daemon in a v2 build
---
## **Verse 1**

Removes the daemon's CLI surface: `bp start`, `bp stop`, `bp status` and `bp serve daemon`. This is a pure removal, so it adds no new tests: the existing tests for the removed commands are deleted, and the callers that break are updated. After this bar nothing imports `daemon/` or `claude/dispatch.go`, which sets up the next bar.

## Scope
- `cmd/start.go`, `cmd/stop.go`, `cmd/status.go` and `cmd/start_test.go`, `cmd/stop_test.go`, `cmd/status_test.go`: delete all six **in this one bar**. `start_test` uses `bootoutCall` (stop_test) and `launchctlDict`, `disabledStore`, `printDisabled`, `listSubcmd` (status_test).
- `cmd/root.go`: drop the `daemon` import (:15). `NewRootCmd` becomes `newRootCmd(claude.ExecRunner)` (:125). `newRootCmd(run claude.Runner)` loses the `lc` param (:128) and the `newStartCmd`/`newStatusCmd`/`newStopCmd` wiring (:147-149).
- `cmd/serve.go`: keep only `serveCmdUse` and `newServeCmd`, with `newServeMCPCmd()` as its only child. Delete `serveDaemonCmdUse`, `claudeBinEnv`, `claudeBinFallback`, `serveTick`, `serveRunner`, `claudePath`, `newHandler` and `newServeDaemonCmd`, plus the now-unused imports (`io`, `log/slog`, `os`, `time`, `claude`, `daemon`, `db`, `orm`). Change the parent's `Short` only if lint forces it. `"Run foreground servers"` stays.
- `cmd/serve_test.go`: keep `TestServeMCPCmd_IsListedInServeHelp` (:158), converted to `TestServeCmd/lists mcp in help`, and delete the other 11 tests, including both "daemon"-in-help checks (:150, :166) and `fastTick` and any helpers only they use. The kept test runs `bp serve --help`, so its unit is the `serve` parent. It is not named `TestServeMCPCmd`, because the 34 `TestServeMCPCmd_*` tests in `cmd/serve_mcp*_test.go` will become that one top-level test when their files are converted, and a package can hold only one function of that name.
- `cmd/cmd_test.go`: drop the `daemon` import (:11), `nothingLoaded` (:24), `runWithDaemon` (:28) and `runWithContext` (:58, which lint flags as `unused` once the serve-daemon tests are gone). `runWithRunner` (:46) and `runSplit` (:128) call `newRootCmd(run)` / `newRootCmd(func…)` with one arg. Its only test function is `TestMain`, so it has nothing to convert.
- `cmd/task_test.go:18`, `cmd/root_test.go:223,243`: change to a one-arg `newRootCmd`. Delete the `serve daemon` row (`root_test.go:216`) of `TestSuppressed` (today `TestSuppressed_FullScreenCommands`, `root_test.go:208`).
- Test conversion (`.claude/rules/go-tests.md`). This bar touches `cmd/serve_test.go`, `cmd/task_test.go` and `cmd/root_test.go`, so it converts every test that survives in them to one top-level test per unit, as a pure restructure: same cases, same assertions, still green. Each flat test `TestX_Suffix` becomes `t.Run("<suffix as lowercase words>", ...)` inside `TestX`, with its body unchanged. A flat table test that is its unit's only test becomes `TestX` itself, rows unchanged. Tests already named for their unit stay as they are. The resulting names, which later bars (BIT-46, BIT-48) use:
  - `cmd/serve_test.go`: `TestServeCmd/lists mcp in help`.
  - `cmd/task_test.go`: `TestTaskCmd/subcommands are wired under root`, `TestTaskCmd/lifecycle runs through the root command`.
  - `cmd/root_test.go`:
    - `TestRootCmd/help`, `TestRootCmd/version`, `TestRootCmd/has no instructions command`, `TestRootCmd/runtime error omits usage`;
    - `TestBitDir/outside worktree uses relative dot bit`, `TestBitDir/inside claude worktree resolves to main checkout`, `TestBitDir/nested worktree resolves to outermost checkout`;
    - `TestSignalContext/cancels on termination signals`;
    - `TestExecute/behind plugin writes notice to stderr`, `TestExecute/no plugin state is silent`, `TestExecute/suppressed command writes no notice`, `TestExecute/fires the refresh`;
    - `TestSuppressed` (table, rows `tui`, `serve mcp`, `task list`, `root` once `serve daemon` is gone);
    - `TestBehind` unchanged.

## Steps
- [ ] Convert the surviving tests in `cmd/serve_test.go`, `cmd/task_test.go` and `cmd/root_test.go` as listed under Test conversion. Leave out the serve tests this bar deletes. `just test` stays green.
- [ ] Apply the Scope edits above, in this order: the root wiring and signature, then the test callers, then the file deletions and the serve trim.
- [ ] `just lint`. Delete anything `unused` now flags in `cmd/`. Expected: `runWithContext`, plus any test helper or const that only the deleted tests used. Delete it, don't suppress it.

## Claude verifies
- [ ] `just test` passes. It runs `sqlc generate` first, so never use bare `go test` here.
- [ ] `just lint` reports `0 issues`.
- [ ] `grep -rn '"github.com/B4Dmonkey/bit-pro/daemon"' --include='*.go' .` matches only files inside `daemon/`.
- [ ] `grep -rn 'claude.DirRunner\|claude.ExecDirRunner' --include='*.go' cmd/` matches nothing.

## User verifies
- [ ] Build to a temp path, never `just install`: `just db-gen-queries && go build -o /tmp/bp-v2 .`. Then `SBX=$(mktemp -d); HOME=$SBX XDG_DATA_HOME=$SBX/share /tmp/bp-v2 --help`. `start`, `stop` and `status` aren't listed, and `HOME=$SBX XDG_DATA_HOME=$SBX/share /tmp/bp-v2 serve --help` lists only `mcp`.

## Commit (user)
`feat(cmd)!: remove bp start/stop/status and serve daemon`