---
id: BIT-36.1
title: Contradiction forces serve to become a parent
status: done
approved: true
phase: 1
phase_label: serve daemon
---
## **Verse 1**

New test `TestServeCmd_DaemonIsListedInServeHelp` can't pass with a leaf `serve` — it forces the command to become a parent. Until `daemon` is a subcommand, `bp serve --help` lists nothing.

## Scope
- `cmd/serve.go` — split `newServeCmd()` into a parent (no `RunE`) that adds `newServeDaemonCmd()` as a child; `serveCmdUse = "serve"` stays; add `serveDaemonCmdUse = "daemon"`
- `cmd/serve_test.go` — add `TestServeCmd_DaemonIsListedInServeHelp`; update every invocation of `serveCmdUse` alone (e.g. `runWithContext(t, ctx, serveCmdUse)`) to pass two args (`serveCmdUse, serveDaemonCmdUse`); update `TestServeCmd_IsListedInHelp` per the scope Decision

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeCmd_DaemonIsListedInServeHelp`
     - **Behavior:** `bp serve` is a parent whose help lists `daemon` as a subcommand
     - **Setup:** `mustRun(t, "serve", "--help")`
     - **Assertions:** output contains `"daemon"`
     - **Boundary:** bare `serve` with no subcommands — the exact state before this bar; proves the parent→child relationship exists, not just that the loop still runs
   - [ ] Confirm fails: `serve` has no subcommands so `--help` does not list `daemon`

2. **Implement (GREEN):**
   - [ ] In `cmd/serve.go`, rename the existing command body to `newServeDaemonCmd()` — same `Short`, same `Args: cobra.NoArgs`, same `RunE`, keep `verbose` flag and `serveTick`; change its `Use:` to `serveDaemonCmdUse`
   - [ ] Replace `newServeCmd()` with a parent: `Use: serveCmdUse`, `Short: "Run a foreground server"` (or similar grouping short), no `RunE`, no `Args`; call `cmd.AddCommand(newServeDaemonCmd())`
   - [ ] Add `const serveDaemonCmdUse = "daemon"` above or beside `serveCmdUse`

3. **More tests (RED → GREEN):**
   - [ ] `TestServeCmd_IsListedInHelp` (rename/repurpose per scope Decision)
     - **Behavior:** `bp serve --help` lists `daemon` (this replaces the root-help check, which still holds incidentally)
     - **Setup:** `mustRun(t, "serve", "--help")`
     - **Assertions:** output contains `"daemon"`
     - **Boundary:** proves the subcommand is registered (same signal as the new test — either merge them or keep both; merging is simpler)
   - [ ] Update every call site that passes `serveCmdUse` as the sole command word:
     - `runWithContext(t, ctx, serveCmdUse)` → `runWithContext(t, ctx, serveCmdUse, serveDaemonCmdUse)`
     - `runWithContext(t, ctx, serveCmdUse, "-v")` → `runWithContext(t, ctx, serveCmdUse, serveDaemonCmdUse, "-v")`
     - The table test `args: []string{serveCmdUse, "-v"}` → `[]string{serveCmdUse, serveDaemonCmdUse, "-v"}`
     - `args: []string{serveCmdUse}` → `[]string{serveCmdUse, serveDaemonCmdUse}`
   - [ ] Confirm each now routes to the daemon subcommand (run the test suite)

## Claude verifies
- [ ] `just test ./cmd/...` passes
- [ ] `just lint` passes

## User verifies
none — deterministic

## Commit (user)
`refactor(serve): make serve a parent; daemon body moves to serve daemon`