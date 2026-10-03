---
id: BIT-48.2
title: An existing user-scope bit MCP entry skips the mcp add step
status: done
approved: true
phase: 1
phase_label: bp add sets bit up for the whole machine
---
## **Verse 1**

`claude mcp add -s user` errors when the entry already exists, and `claude mcp get` can't see scope. So EnsureGlobal reads the top-level `mcpServers.bit` of `<home>/.claude.json` itself. A home that already has the user entry contradicts the unconditional four-call loop from BIT-48.1.

## Scope
- `claude/global.go`:
  - unexported `hasUserMCP(home string) bool`: read `filepath.Join(home, ".claude.json")`; unmarshal into `struct{ MCPServers map[string]json.RawMessage `json:"mcpServers"` }`; return whether key `bit` is present. Any read or parse error returns false, so the add runs anyway and a duplicate surfaces as a failed step (scope Decision). Read only; never write the file. `CLAUDE_CONFIG_DIR` is ignored.
  - `EnsureGlobal`: run steps 1–3 unconditionally; run step 4 only when `!hasUserMCP(home)`. The step number in the error stays `i+1` of 4.
- `claude/global_test.go`: the new cases join BIT-48.1's `TestEnsureGlobal` as rows of one table, since they share setup (a home whose `.claude.json` holds the row's contents, `newRecorder(nil)`, one `EnsureGlobal` call) and assertions (nil error, `rec.calls` equals the row's expected argvs). Write `<home>/.claude.json` with `writeFixture` from `claude/plugin_test.go` (reused, not edited).

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestEnsureGlobal/skips mcp add when a user entry exists` (the table's first row)
     - **Behavior:** re-wiring a machine that already has bit's user-scope MCP server doesn't trip `mcp add`'s "already exists" error.
     - **Setup:** `<home>/.claude.json` = `{"mcpServers": {"bit": {"type": "stdio", "command": "bp", "args": ["serve", "mcp"], "env": {}}}, "projects": {}}` (the real shape); `newRecorder(nil)`.
     - **Assertions:** nil error; `rec.calls` equals the first three argvs only.
     - **Boundary:** top-level `mcpServers.bit` present.
   - [ ] Confirm fails: four calls recorded.

2. **Implement (GREEN):**
   - [ ] `hasUserMCP` and the guard on step 4.

3. **More tests (RED → GREEN):**
   - [ ] More rows in the same `TestEnsureGlobal` table; each expects all four calls:
     - `TestEnsureGlobal/adds mcp for a local-only entry`: `{"projects": {"/p/a": {"mcpServers": {"bit": {"command": "bp"}}}}}`. *Boundary:* a local-scope entry doesn't count as user scope (the real file looks like this today).
     - `TestEnsureGlobal/adds mcp when only another user server exists`: `{"mcpServers": {"other": {}}}`. *Boundary:* `mcpServers` present, key `bit` absent.
     - `TestEnsureGlobal/adds mcp when claude json is malformed`: `{`. *Boundary:* torn read of a file every Claude session rewrites → fall back to adding.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none — deterministic

## Commit (user)
`feat(bit): skip mcp add when a user-scope bit entry exists`