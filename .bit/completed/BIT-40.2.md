---
id: BIT-40.2
title: Second init run still registers — no skip branch
status: done
approved: true
phase: 2
phase_label: idempotency
---
## **Verse 2**

Bar 1's implementation always calls `claude mcp add` with no skip-if-already-registered branch. This test contradicts any future attempt to add one: if a skip branch is added, only one mcp add call appears after two init runs, and this test catches it.

## Scope
- `cmd/init_test.go` — new test `TestInitCmd_MCPRegistrationIsIdempotent`

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestInitCmd_MCPRegistrationIsIdempotent` in `cmd/init_test.go`
     - **Behavior:** a second `bp init` run still calls `claude mcp add` and prints the status line — no skip branch exists
     - **Setup:** recorder runner (all calls succeed), `t.Chdir(t.TempDir())`, run `init --prefix BIT` twice; collect calls and output from each run separately (use `runWithRunner` for the second run with a fresh recorder to isolate per-run assertions)
     - **Assertions:** both recorders contain `[]string{"claude", "mcp", "add", "bit", "--", "bp", "serve", "mcp"}`; output from each run contains `"bit MCP server registered"`
     - **Boundary:** second run over an already-configured project — the `already initialized` case from `TestInitCmd_IsIdempotent`'s scope, now extended to the MCP registration path
   - [ ] Confirm fails: test cannot be written yet because `RegisterMCP` doesn't exist — this bar is written and run after Bar 1's GREEN is committed

2. **Implement (GREEN):**
   - [ ] No code change — Bar 1's `RegisterMCP` call has no early-exit or skip branch, so both runs call mcp add and print the line. The test passes as written.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] In a real project directory: `bp init --prefix BIT` → observe output ends with `"bit MCP server registered (local scope)."` → run `bp init --prefix BIT` again → observe the same line appears, no error

## Commit (user)
`test(init): prove MCP registration is idempotent`