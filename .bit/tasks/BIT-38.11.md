---
id: BIT-38.11
title: task_move resequences a bar
status: todo
approved: true
phase: 2
phase_label: Plan writes
---
## **Verse 2**

The last write a plan needs: resequencing a bar. Claude reaching for `mv` on a task file is the behaviour this whole phase exists to remove, and `bp task move` being invisible is why it happens — a tool is not. Forced by a protocol-level test: `task_move` is not registered. Completes verse 2.

## Scope
- `cmd/serve_mcp.go` — `taskMoveTool` const, `taskMoveInput` / `emptyOutput` types, `taskMoveHandler`, registration.
- `cmd/serve_mcp_write_test.go` — the test.

## References
- `mcp-notes.md` — "Parity map", the `task_move` row: params `bar`, `before?`, `after?` with exactly one required, and an empty return.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeMCPCmd_TaskMoveResequencesABar`
     - **Behavior:** a bar can be moved relative to a sibling through the protocol, and the track's order list — every surface that reads it included — reflects the new position.
     - **Setup:** `dir := t.TempDir()`; `seedTasks` track `FOO-1` with `Order: []string{"FOO-1.1", "FOO-1.2", "FOO-1.3"}` plus the three bars; `s := mcpSession(t, dir)`. Two subtests: (a) `{"bar": "FOO-1.3", "before": "FOO-1.1"}`; (b) `{"bar": "FOO-1.1", "after": "FOO-1.3"}`.
     - **Assertions:** each call returns structured content `{}` and `IsError` false. (a) leaves `Load("FOO-1").Order` as `["FOO-1.3", "FOO-1.1", "FOO-1.2"]`; (b) as `["FOO-1.2", "FOO-1.3", "FOO-1.1"]`. In each case a `task_list` call with `{"parent": "FOO-1"}` returns the IDs in that same order.
     - **Boundary:** the anchor pair at both of its valid states — `before` set with `after` absent, and the reverse — moving to each end of a three-element list.
   - [ ] Confirm fails: `tool "task_move" not found`
   - [ ] `TestServeMCPCmd_TaskMoveRefusesABadAnchorPair` — same setup; `callToolErr` for `{"bar": "FOO-1.3", "before": "FOO-1.1", "after": "FOO-1.2"}` and for `{"bar": "FOO-1.3"}`. Each comes back `IsError` true with the order unchanged. No new production code: bar 2.3 put the rule in the store, and this test is what proves the tool inherits it rather than needing its own copy.

2. **Implement (GREEN):**
   - [ ] `const taskMoveTool = "task_move"`.
   - [ ] `type taskMoveInput struct` — `Bar string` (`json:"bar"`, required), `Before string` (`json:"before,omitempty"`), `After string` (`json:"after,omitempty"`). Plain strings, not pointers: empty *is* absent for an anchor, and that is exactly the state the store's rule reads.
   - [ ] `type emptyOutput struct{}` — a shared return type for the tools whose result is just success. The SDK derives an object schema for it and marshals it to `{}`; verified against the in-memory transport.
   - [ ] `taskMoveHandler(root string)`: `store.Move(in.Bar, in.Before, in.After)`, wrap as `fmt.Errorf("moving bar %s: %w", in.Bar, err)`, return `emptyOutput{}`. No validation in the handler — a second copy of the one-anchor rule here is the drift the track's Decisions rule out.
   - [ ] Register it, with a `Description` stating that exactly one of `before`/`after` must be given, that the anchor must be a sibling in the same track, and that a bar's ID is stable identity so moving it keeps every existing reference to it valid.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] Whole slice: with the server registered (`claude mcp add bit -- bp serve mcp`, from bar 1.7 if it is still in place), start a session and ask for a throwaway scope track plus three bars under it with verse tags, then ask for the last bar to be moved to the front. Observe: `bp task list --parent <track>` prints the three bars in the reordered sequence with the phase column populated, and the transcript shows `mcp__bit__task_create` / `mcp__bit__task_move` with no Bash command touching `.bit/`. Clean up with `bp task delete <track> -y -f`.

## Commit (user)
`feat(mcp): add task_move for resequencing a bar`