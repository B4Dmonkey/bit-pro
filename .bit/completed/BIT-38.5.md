---
id: BIT-38.5
title: task_update rewrites a body and reports revoked approval
status: done
approved: true
phase: 1
phase_label: Scope writes
---
## **Verse 1**

Closes the scope-authoring loop: a track's body can be rewritten in place, and the result says whether approval survived. Revocation is silent in the CLI, so returning the flag is the new behaviour this bar adds — that is what makes a replan touching approved bars observable. Forced by a protocol-level test: `task_update` is not registered.

Deliberately minimal: this bar's schema carries `id` and `body` only, and the handler patches only the body. That is the least code its test can demand, and bar 1.6 is the test that contradicts it.

## Scope
- `cmd/serve_mcp.go` — `taskUpdateTool` const, `taskUpdateInput` / `taskUpdateOutput` types, `taskUpdateHandler`, registration.
- `cmd/serve_mcp_write_test.go` — the test.

## References
- `mcp-notes.md` — "Parity map", the `task_update` row and the note beneath it explaining why the return carries `approved`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeMCPCmd_TaskUpdateRewritesBodyAndReportsRevocation`
     - **Behavior:** an approved track whose body is rewritten through the tool comes back with `approved: false`, and the new body is on disk — the revocation the CLI performs silently is now visible in the tool result.
     - **Setup:** `dir := t.TempDir()`; `seedTasks(t, dir, &task.Task{ID: "FOO-1", Title: "MCP write surface", Status: task.StatusTodo, Approved: true, Body: "## Why\n\nold reason"})`; `s := mcpSession(t, dir)`; call `task_update` with `map[string]any{"id": "FOO-1", "body": "## Why\n\nnew reason\n\n## Decisions\n\n- settled"}`.
     - **Assertions:** structured content is `{"id": "FOO-1", "approved": false}`. `Load("FOO-1")` has `Body` byte-identical to the string sent and `Approved` false.
     - **Boundary:** `approved` at the state where the rule fires (true → false). `body` is the only field sent — the set-field count at 1, its lower non-empty bound.
   - [ ] Confirm fails: `tool "task_update" not found`

2. **Implement (GREEN):**
   - [ ] `const taskUpdateTool = "task_update"`.
   - [ ] `type taskUpdateInput struct` — `ID string` (`json:"id"`) and `Body string` (`json:"body,omitempty"`). No other fields yet; no pointers yet. Nothing in this bar's test distinguishes an omitted field from an empty one, so nothing here should try to.
   - [ ] `taskUpdateOutput` — `ID string` tagged `id`, `Approved bool` tagged `approved`.
   - [ ] `taskUpdateHandler(root string)`: load the store, call `store.Update(in.ID, task.Patch{Body: &in.Body})`, wrap the error as `fmt.Errorf("updating task %s: %w", in.ID, err)`, return `taskUpdateOutput{ID: t.ID, Approved: t.Approved}`.
   - [ ] Register it, with a `Description` stating that the returned `approved` reflects whether the edit revoked approval.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic

## Commit (user)
`feat(mcp): add task_update returning whether approval survived`