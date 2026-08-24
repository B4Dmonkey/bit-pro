---
id: BIT-38.3
title: task_create mints a track through the protocol
status: todo
approved: true
phase: 1
phase_label: Scope writes
---
## **Verse 1**

The first write tool. Forced by a protocol-level test: `task_create` is not registered, so calling it fails at the client. Nothing below it is new — bar 1.2 put the whole create sequence in `Store.Create`, so this handler is thin by construction.

Deliberately minimal: the schema carries `title` and `body` only. `Store.Create` already understands `parent`, `after`, and the phase tags, but no test here demands them, and verse 2's bars are the tests that do.

## Scope
- `cmd/serve_mcp.go` — `taskCreateTool` const, `taskCreateInput` / `taskCreateOutput` types, `taskCreateHandler`, and the `mcp.AddTool` registration in `runMCPServer`.
- `cmd/serve_mcp_write_test.go` — new file for the write tools' protocol tests.

## References
- `mcp-notes.md` — the "Parity map" table under Command inventory. `task_create`'s row is the authority for its params and its `{id}` return; check it before changing a param name.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeMCPCmd_TaskCreateMintsATrack`
     - **Behavior:** Claude can author a scope through the protocol — a track with a multi-line markdown body — and gets back the minted ID.
     - **Setup:** `dir := t.TempDir()`; `task.New(filepath.Join(dir, ".bit")).SaveConfig(&task.Config{Prefix: "FOO"})`; `s := mcpSession(t, dir)`; call `task_create` with `map[string]any{"title": "MCP write surface", "body": "## Why\n\nBecause `sed` owns the file format today.\n\n## Summary\n\nSix tools."}`. The body carries a fenced-inline backtick and blank lines deliberately — a multi-line body through a JSON string is the whole point of the tool.
     - **Assertions:** returned structured content is `{"id": "FOO-1"}`. Then `task.New(filepath.Join(dir, ".bit")).Load("FOO-1")` has `Title` "MCP write surface", `Status` `task.StatusTodo`, and `Body` byte-identical to the string sent.
     - **Boundary:** `parent` omitted — the lower bound of the parent param, which selects prefix minting. `body` at multiple lines rather than one, which is where the CLI needed `-d "$(cat body.md)"` and the tool needs nothing.
   - [ ] Confirm fails: the client returns an error for an unknown tool — `tool "task_create" not found` (transport-level, not `result.IsError`), because nothing registers it yet.

2. **Implement (GREEN):**
   - [ ] `const taskCreateTool = "task_create"`.
   - [ ] `type taskCreateInput struct` with `Title string` (`json:"title"`, required) and `Body string` (`json:"body,omitempty"`). Plain values, not pointers: a new task's zero values *are* the defaults, so there is nothing here to distinguish "omitted" from "empty".
   - [ ] `taskCreateOutput` — one field, `ID string`, tagged as JSON `id`.
   - [ ] `taskCreateHandler(root string)` mirroring `taskReadHandler`: `task.New(bitdir.ForRoot(root))`, `store.Create(task.CreateParams{Title: in.Title, Body: in.Body})`, wrap the error as `fmt.Errorf("creating task: %w", err)`, return `taskCreateOutput{ID: t.ID}`.
   - [ ] Register it in `runMCPServer` with a `Description` naming what a track and a bar are and that the return is the minted ID.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic

## Commit (user)
`feat(mcp): add task_create for minting a track`