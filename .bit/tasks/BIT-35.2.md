---
id: BIT-35.2
title: Contradiction forces real MCP server with task_read handler
status: todo
approved: true
phase: 1
phase_label: mcp skeleton
---
## **Verse 1**

The stub `newServeMCPCmd()` has no `RunE` — it closes immediately and sends nothing. `TestServeMCPCmd_TaskReadReturnsStructuredFields` can't pass until the command runs a real MCP server that handles `task_read`.

## Scope
- `cmd/serve_mcp.go` — implement `runMCPServer(ctx context.Context, root string, r io.Reader, w io.Writer) error`; update `RunE` to read `CLAUDE_PROJECT_DIR` and call `runMCPServer(cmd.Context(), root, os.Stdin, os.Stdout)`
- `cmd/serve_mcp_test.go` — new file for the in-process MCP tests
- `go.mod`, `go.sum` — new `github.com/modelcontextprotocol/go-sdk` direct dependency

## References
- `mcp-notes.md` — SDK choice ("The Go SDK is official…") and the stdout rule ("Nothing logs to stdout"). `task_read` return fields and the no-project-param decision are in Decisions.

## Needs real data
- [ ] `go get github.com/modelcontextprotocol/go-sdk/mcp@v1.7.0` then `go doc github.com/modelcontextprotocol/go-sdk/mcp StdioTransport` — confirm whether `StdioTransport` exposes `In`/`Out io.Reader/io.Writer` fields (use them in `runMCPServer`) or whether the SDK provides an in-process transport for testing. The transport shape determines how the test injects JSON-RPC frames without touching `os.Stdin`/`os.Stdout`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeMCPCmd_TaskReadReturnsStructuredFields` in `cmd/serve_mcp_test.go`
     - **Behavior:** `task_read` on an existing track returns `id`, `title`, `status`, `approved`, `phase`, `phase_label`, `parent`, and `body` as structured JSON
     - **Setup:** `dir := t.TempDir()`; `task.New(filepath.Join(dir, ".bit")).Save(&task.Task{ID: "FOO-1", Title: "a track", Status: task.StatusTodo, Body: "the body"})`; `t.Setenv("CLAUDE_PROJECT_DIR", dir)`; construct transport from r/w per Needs real data; run `runMCPServer(ctx, dir, r, w)` in a goroutine; send MCP `initialize` handshake then a `tools/call` for `task_read` with `{"id":"FOO-1"}`
     - **Assertions:** response contains `"id":"FOO-1"`, `"title":"a track"`, `"status":"todo"`, `"approved":false`, `"body":"the body"`, `"parent":""`
     - **Boundary:** root track (no dot in ID) — `parent` is empty string; proves the no-parent case; `body` non-empty proves the body round-trip
   - [ ] Confirm fails: no `RunE` means the command exits immediately; test reads EOF waiting for a response

2. **Implement (GREEN):**
   - [ ] `go get github.com/modelcontextprotocol/go-sdk/mcp@v1.7.0`; `go mod tidy`
   - [ ] Define in `cmd/serve_mcp.go`:
     ```
     type taskReadInput struct { ID string }
     type taskReadOutput struct {
         ID, Title, Status string
         Approved           bool
         Phase              int
         PhaseLabel, Parent, Body string
     }
     ```
   - [ ] Implement `runMCPServer(ctx, root, r, w)`:
     - `s := mcp.NewServer(mcp.ServerInfo{Name: "bp"}, nil)` (adjust to actual constructor shape)
     - `mcp.AddTool(s, mcp.Tool{Name: "task_read", ...}, handler)` — handler: `task.New(root).Load(input.ID)` → map all `task.Task` fields to `taskReadOutput`; derive `parent` by `strings.LastIndex(id, ".")` (`-1` → `""`)
     - `return s.Run(ctx, transport)` — transport constructed from r/w per Needs real data
   - [ ] Update `RunE` in `newServeMCPCmd()`: `root := os.Getenv("CLAUDE_PROJECT_DIR"); return runMCPServer(cmd.Context(), root, os.Stdin, os.Stdout)`
   - [ ] Confirm no path in `runMCPServer` or its callees writes to `os.Stdout` outside `s.Run` — `bitdir.Resolve()` in `PersistentPreRunE` is pure path math; `task.New(root).Load()` has no stdout I/O

3. **More tests (RED → GREEN):**
   - [ ] `TestServeMCPCmd_TaskReadReturnsParentForBar`
     - **Behavior:** `task_read` on a bar returns its track ID as `parent`
     - **Setup:** same temp dir; also save `&task.Task{ID: "FOO-1.1", Title: "a bar", Status: task.StatusTodo}` (child of FOO-1); send `tools/call` for `task_read` with `{"id":"FOO-1.1"}`
     - **Assertions:** response `"parent":"FOO-1"`
     - **Boundary:** bar ID with exactly one dot (`FOO-1.1`) — `strings.LastIndex` extracts `FOO-1`; contradicts hardcoded `parent:""` if that was the GREEN for test 1

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` passes

## User verifies
- [ ] Run `bp serve mcp` in a project terminal; type a raw MCP `initialize` frame on stdin — confirm a JSON response arrives and the connection stays open
- [ ] In Claude Code (project wired to `bp serve mcp`), confirm `task_read` appears in the tool panel; call it on a real track ID and confirm the response includes `body` as the full task body text