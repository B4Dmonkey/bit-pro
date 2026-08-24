---
id: BIT-38.1
title: One session harness replaces five copies
status: done
approved: true
phase: 1
phase_label: Scope writes
---
## **Verse 1**

Five copies of the same start-server-connect-client block already sit in `cmd/serve_mcp_test.go`; this plan adds roughly a dozen more. Collapsing them into one harness first makes every later bar cheap. Pure refactor — no production code moves, no new behaviour.

## Scope
- `cmd/mcp_harness_test.go` — new. The shared harness.
- `cmd/serve_mcp_test.go` — the five existing tests rewritten onto it.

## TDD cycle

Refactor step — no RED. The existing five tests are the safety net, and they must pass unchanged in what they assert.

1. **Extract (green throughout):**
   - [ ] `func mcpSession(t *testing.T, root string) *mcp.ClientSession` — builds `mcp.NewInMemoryTransports()`, runs `runMCPServer` in a goroutine with a cancellable `t.Context()`, registers `t.Cleanup` that cancels and drains the error channel, connects a client, registers `t.Cleanup(session.Close)`. Calls `t.Helper()`.
   - [ ] `func callTool(t *testing.T, s *mcp.ClientSession, name string, args map[string]any) map[string]any` — calls the tool, `t.Fatal`s on a transport error or `result.IsError`, then round-trips `result.StructuredContent` through `json.Marshal`/`json.Unmarshal` into `map[string]any` and returns it. Calls `t.Helper()`.
   - [ ] `func callToolErr(t *testing.T, s *mcp.ClientSession, name string, args map[string]any)` — the inverse: `t.Fatal`s unless `result.IsError` is true. Later bars need it; add it here so the harness is complete in one commit.
   - [ ] `func seedTasks(t *testing.T, dir string, tasks ...*task.Task)` — `task.New(filepath.Join(dir, ".bit"))` then `Save` each, `t.Fatal` on error. Calls `t.Helper()`.
   - [ ] Rewrite all five existing tests onto the harness. Keep every assertion exactly as it is — this bar changes only how the session is obtained.
   - [ ] `TestServeMCPCmd_TaskListReturnsEveryTaskAsFields` decodes into a `Tasks []map[string]any` shape rather than a bare map. Give the harness a second decoder for that (`callToolList`, returning `[]map[string]any` from the `tasks` key) rather than making `callTool` generic.

## Claude verifies
- [ ] `just test` — the five MCP tests pass with their assertions unchanged
- [ ] `just lint`
- [ ] `git diff --stat cmd/serve_mcp.go` is empty — no production code moved in this bar

## User verifies
- [ ] none — deterministic

## Commit (user)
`refactor(mcp): extract a shared in-memory session harness for tool tests`