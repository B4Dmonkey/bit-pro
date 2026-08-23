---
id: BIT-37.3
title: parent narrows the list to one track's bars, in order
status: todo
approved: true
phase: 1
phase_label: Read surface
---
## **Verse 1**

`task_list` returns everything, so it cannot answer the question the pipeline actually asks —
"what are this track's bars". A store where one track's bars are a strict subset of all tasks
contradicts the unfiltered return.

## Scope
- `cmd/serve_mcp.go` — `taskListHandler` branches on `in.Parent`.
- `cmd/serve_mcp_test.go` — the new test.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeMCPCmd_TaskListParentReturnsOnlyThatTracksBarsInOrder`
     - **Behavior:** `parent` narrows the result to one track's direct bars and preserves that
       track's explicit order — the two properties a skill reading a plan depends on.
     - **Setup:** store holding `FOO-1` with `Order: []string{"FOO-1.2", "FOO-1.1"}`, its bars
       `FOO-1.1` and `FOO-1.2`, plus an unrelated track `FOO-2` and its bar `FOO-2.1` — five
       tasks in all. Call `task_list` with `Arguments: map[string]string{"parent": "FOO-1"}`.
     - **Assertions:** `tasks` holds exactly 2 entries, `FOO-1.2` first and `FOO-1.1` second.
       Neither `FOO-1` itself nor anything under `FOO-2` appears.
     - **Boundary:** `parent` present — the other end of the optional input from the previous
       bar. The requested track's bars are 2 of the store's 5 tasks, so a filter that is a no-op
       fails; and the reversed `Order` puts the required order in conflict with ID order, so
       the order assertion cannot pass on lexical luck.
   - [ ] Confirm fails: five entries come back — the handler ignores `parent`.

2. **Implement (GREEN):**
   - [ ] `taskListHandler`: when `in.Parent != ""` call `store.Children(in.Parent)`, else
     `store.List()` as before; the mapping is unchanged. This mirrors the branch in
     `cmd/task/list.go`, which is what makes the tool's order identical to the CLI's.

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` passes

## User verifies
- [ ] Read this plan back through the real server. `just install`, then from the repo root:

  ```
  { printf '%s\n' \
    '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"hand","version":"1"}}}' \
    '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
    '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"task_list","arguments":{"parent":"BIT-37"}}}'; sleep 2; } \
    | CLAUDE_PROJECT_DIR=$PWD bp serve mcp 2>/dev/null | tail -1
  ```

  The last frame lists exactly the three bars of this plan, in plan order, each with `status`,
  `phase`, and `phase_label` as its own field and no `body`. (The `sleep` is what holds stdin
  open — without it the server exits on EOF before it answers.)

## Commit (user)
`feat(mcp): filter task_list to one track's bars`