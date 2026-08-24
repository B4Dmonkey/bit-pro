---
id: BIT-38.14
title: task_delete archives a task and reserves its ID
status: todo
approved: true
phase: 4
phase_label: Close out
---
## **Verse 4**

The other write that has to survive without a terminal. `-y/--yes` does not cross over — a tool call is already surfaced by Claude Code's own permission prompt, which is the same guarantee by a different mechanism — so the tool's shape is `{id}` and nothing else yet. Forced by a protocol-level test: `task_delete` is not registered.

## Scope
- `cmd/serve_mcp.go` — `taskDeleteTool` const, `taskDeleteInput`, `taskDeleteHandler`, registration.
- `cmd/serve_mcp_write_test.go` — the test.

## References
- `mcp-notes.md` — "Parity map", the `task_delete` row and the note on why `-y` is dropped and `--force` kept.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeMCPCmd_TaskDeleteRelocatesAndReservesTheID`
     - **Behavior:** a bar dropped mid-plan leaves the active list, comes out of its track's order, and its ID is never handed to a different task — so a note or commit message citing it stays meaningful.
     - **Setup:** `dir := t.TempDir()`; `SaveConfig(&task.Config{Prefix: "FOO"})`; `seedTasks` track `FOO-1` with `Order: []string{"FOO-1.1", "FOO-1.2"}` plus both bars; `s := mcpSession(t, dir)`; call `task_delete` with `{"id": "FOO-1.2"}`, then `task_create` with `{"title": "A replacement step", "parent": "FOO-1"}`.
     - **Assertions:** the delete returns `{}` with `IsError` false. `.bit/archive/tasks/FOO-1.2.md` exists and `.bit/tasks/FOO-1.2.md` does not. `Load("FOO-1").Order` is `["FOO-1.1"]`. The follow-up create returns `{"id": "FOO-1.3"}` — **not** `FOO-1.2`, because `NextChildID` counts the archive.
     - **Boundary:** deleting a bar rather than a track — the branch where `relocateTree` also calls `removeFromOrder`, which the track case never reaches. The sibling count going 2 → 1 is what makes the order assertion meaningful.
   - [ ] Confirm fails: `tool "task_delete" not found`

2. **Implement (GREEN):**
   - [ ] `const taskDeleteTool = "task_delete"`.
   - [ ] `type taskDeleteInput struct { ID string }` tagged `json:"id"` — required. No `force` field yet; nothing in this bar's test demands one.
   - [ ] `taskDeleteHandler(root string)`: `store.Relocate(in.ID, false)`, wrap as `fmt.Errorf("deleting task %s: %w", in.ID, err)`, return `emptyOutput{}`.
   - [ ] Register it, with a `Description` saying the task is relocated to `.bit/archive/tasks/` rather than destroyed, that its ID stays reserved, and that a bar is also removed from its track's order.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic

## Commit (user)
`feat(mcp): add task_delete for archiving a task`