---
id: BIT-35.1
title: bp serve mcp skeleton registered in bp serve --help
status: doing
approved: true
phase: 1
phase_label: mcp skeleton
---
## **Verse 1**

`bp serve mcp` is not yet a subcommand of `serve` — `TestServeMCPCmd_IsListedInServeHelp` can't pass without it registered.

## Scope
- `cmd/serve_mcp.go` — new file; `newServeMCPCmd()` returns a stub command (`Use: "mcp"`, `Short: "Run the MCP server in the foreground"`, `Args: cobra.NoArgs`, no `RunE`)
- `cmd/serve.go` — add `cmd.AddCommand(newServeMCPCmd())` in `newServeCmd()` (the BIT-36 parent form)

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeMCPCmd_IsListedInServeHelp` in `cmd/serve_test.go` (alongside `TestServeCmd_DaemonIsListedInServeHelp`)
     - **Behavior:** `bp serve --help` lists `mcp` as a subcommand
     - **Setup:** `mustRun(t, "serve", "--help")`
     - **Assertions:** output contains `"mcp"`
     - **Boundary:** `serve` with no `mcp` child yet registered — the exact state before this bar; proves the subcommand relationship exists, not any server behaviour
   - [ ] Confirm fails: `bp serve --help` output does not list `mcp`

2. **Implement (GREEN):**
   - [ ] Create `cmd/serve_mcp.go`: `package cmd`; `const serveMCPCmdUse = "mcp"`; `func newServeMCPCmd() *cobra.Command` returning a `*cobra.Command` with `Use: serveMCPCmdUse`, `Short: "Run the MCP server in the foreground"`, `Args: cobra.NoArgs` — no `RunE` yet
   - [ ] In `cmd/serve.go` `newServeCmd()`, add `cmd.AddCommand(newServeMCPCmd())`

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` passes

## User verifies
none — deterministic

## Commit (user)
`feat(cmd): register bp serve mcp as a stub subcommand`