---
id: BIT-40.1
title: TestInitCmd_RegistersMCPServer pulls RegisterMCP into existence
status: done
approved: true
phase: 1
phase_label: register server
---
## **Verse 1**

`TestInitCmd_RegistersMCPServer` is the outermost test — it can't pass until `claude.RegisterMCP` exists and `writeClaudeWiring` calls it and prints a confirmation line.

## Scope
- `cmd/init.go` — `writeClaudeWiring`: print before/after `RegisterMCP` call; update `pluginSyncCalls()` helper in `cmd/cmd_test.go` to include the mcp add entry so `TestInitCmd_SyncsThePlugin` stays green
- `claude/sync.go` — add `RegisterMCP(ctx context.Context, run Runner) error`
- `cmd/init_test.go` — new test `TestInitCmd_RegistersMCPServer`
- `claude/sync_test.go` — two new unit tests for `RegisterMCP`

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestInitCmd_RegistersMCPServer` in `cmd/init_test.go`
     - **Behavior:** bp init calls `claude mcp add bit -- bp serve mcp` and confirms it on stdout
     - **Setup:** recorder runner (all calls succeed), `t.Chdir(t.TempDir())`, run `init --prefix BIT`
     - **Assertions:** `calls` slice contains `[]string{"claude", "mcp", "add", "bit", "--", "bp", "serve", "mcp"}`; `out` contains `"bit MCP server registered"`
     - **Boundary:** single first-run — no prior entry; covers the "added" path; idempotency is Verse 2's bar
   - [ ] Confirm fails: `claude.RegisterMCP undefined` (or the call never appears in the recorder)

2. **Implement (GREEN):**
   - [ ] Add to `claude/sync.go`:
     ```go
     func RegisterMCP(ctx context.Context, run Runner) error {
         if err := run(ctx, "claude", "mcp", "add", "bit", "--", "bp", "serve", "mcp"); err != nil {
             return fmt.Errorf("registering bit MCP server: %w", err)
         }
         return nil
     }
     ```
   - [ ] In `writeClaudeWiring` (`cmd/init.go`), after `SyncPlugin` returns, add:
     ```go
     fmt.Fprintln(cmd.OutOrStdout(), "Registering bit MCP server...")
     if err := claude.RegisterMCP(cmd.Context(), run); err != nil {
         return err
     }
     fmt.Fprintln(cmd.OutOrStdout(), "bit MCP server registered (local scope).")
     ```
   - [ ] Update `pluginSyncCalls()` in `cmd/cmd_test.go` to append `[]string{"claude", "mcp", "add", "bit", "--", "bp", "serve", "mcp"}` so `TestInitCmd_SyncsThePlugin` keeps passing

3. **More tests (RED → GREEN):**
   - [ ] `TestRegisterMCP_CallsClaudeMCPAdd` in `claude/sync_test.go`
     - **Behavior:** `RegisterMCP` issues exactly `claude mcp add bit -- bp serve mcp`
     - **Setup:** `newRecorder(nil)`, call `RegisterMCP(t.Context(), rec.Run)`
     - **Assertions:** `rec.calls` equals `[][]string{{"claude", "mcp", "add", "bit", "--", "bp", "serve", "mcp"}}`
     - **Boundary:** happy path, single call, no args variability
   - [ ] `TestRegisterMCP_ReturnsErrorWhenClaudeFails` in `claude/sync_test.go`
     - **Behavior:** runner error surfaces wrapped as the function's return value
     - **Setup:** `newRecorder(map[int]error{0: errors.New("mcp add failed")})`, call `RegisterMCP`
     - **Assertions:** `err != nil`; `err.Error()` contains `"mcp add failed"`
     - **Boundary:** call index 0 errors — the only call, so the error is always the registration failure

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] None — idempotency and the full UX check are on the last bar (Bar 2)

## Commit (user)
`feat(init): register bit MCP server at local scope`