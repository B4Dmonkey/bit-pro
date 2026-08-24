---
id: BIT-38.13
title: task_complete files a signed-off track and its bars
status: done
approved: true
phase: 4
phase_label: Close out
---
## **Verse 4**

Signing a track off is the one write left that a shell still has to do. `task complete` is one of the two commands that becomes a tool *and* stays in the CLI — one task-package implementation, two callers — so nothing moves down here; the handler is thin over `Store.Complete`. Forced by a protocol-level test: `task_complete` is not registered.

## Scope
- `cmd/serve_mcp.go` — `taskCompleteTool` const, `taskCompleteInput`, `taskCompleteHandler`, registration. Reuses `emptyOutput` from bar 2.4.
- `cmd/serve_mcp_write_test.go` — the test.

## References
- `mcp-notes.md` — "Parity map", the `task_complete` row, and the "`task complete` and `task delete` are both" Decision explaining why the CLI keeps this command.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeMCPCmd_TaskCompleteFilesATrackAndItsBars`
     - **Behavior:** a track whose bars are all done, and the bars themselves, move out of the active list into `.bit/completed/` — so a finished cycle stops cluttering everything that reads `tasks/`.
     - **Setup:** `dir := t.TempDir()`; `seedTasks` track `FOO-1` (`Status: task.StatusDone`, `Order: []string{"FOO-1.1", "FOO-1.2"}`) plus bars `FOO-1.1` and `FOO-1.2` both `task.StatusDone`; `s := mcpSession(t, dir)`; call `task_complete` with `{"id": "FOO-1"}`.
     - **Assertions:** structured content is `{}` with `IsError` false. `.bit/completed/FOO-1.md`, `.bit/completed/FOO-1.1.md`, and `.bit/completed/FOO-1.2.md` all exist; none of the three remains under `.bit/tasks/`. A `task_list` call with no parent returns an empty `tasks` array — the read surface no longer sees them.
     - **Boundary:** bar count at 2 rather than 0 — a track with children is what proves the whole tree relocates rather than just the track file.
   - [ ] Confirm fails: `tool "task_complete" not found`
   - [ ] `TestServeMCPCmd_TaskCompleteRefusesUnfinishedBars` — same setup but `FOO-1.2` left at `task.StatusDoing`. `callToolErr` for `{"id": "FOO-1"}`: `IsError` true, and all three files are still under `.bit/tasks/` with nothing under `.bit/completed/`. No new production code — `Store.Complete` already passes `force: false` to `relocateTree` and returns `*UnfinishedBarsError`. This test pins that the tool has **no** force escape hatch, which is the difference between it and `task_delete`.

2. **Implement (GREEN):**
   - [ ] `const taskCompleteTool = "task_complete"`.
   - [ ] `type taskCompleteInput struct { ID string }` tagged `json:"id"` — required.
   - [ ] `taskCompleteHandler(root string)`: `store.Complete(in.ID)`, wrap as `fmt.Errorf("completing task %s: %w", in.ID, err)`, return `emptyOutput{}`.
   - [ ] Register it, with a `Description` saying it files a signed-off track and its bars under `.bit/completed/`, that it refuses a track with an unfinished bar and there is no override, and that the ID stays reserved so older commit messages and notes referencing it remain valid.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic

## Commit (user)
`feat(mcp): add task_complete for filing a signed-off track`