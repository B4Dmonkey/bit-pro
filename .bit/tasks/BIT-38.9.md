---
id: BIT-38.9
title: after places a new bar mid-track
status: todo
approved: true
phase: 2
phase_label: Plan writes
---
## **Verse 2**

`--after` exists at `cmd/task/create.go:30` and no skill knows it, because the command contract only ever documents `--after` for `task move`. Putting it in the schema fixes that drift by construction, since the schema *is* the documentation. Forced by a test placing a bar mid-track, which the current handler cannot do — it appends.

## Scope
- `cmd/serve_mcp.go` — `taskCreateInput` gains `after`; the handler forwards it.
- `cmd/serve_mcp_write_test.go` — the test.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeMCPCmd_TaskCreateAfterPlacesABarMidTrack`
     - **Behavior:** a bar can be inserted at a chosen position in the track's order at create time, not only appended — so a step missed during planning lands where it belongs without a follow-up move.
     - **Setup:** `dir := t.TempDir()`; `SaveConfig(&task.Config{Prefix: "FOO"})`; `seedTasks` a track `FOO-1` with `Order: []string{"FOO-1.1", "FOO-1.2", "FOO-1.3"}` plus the three bars `FOO-1.1`/`FOO-1.2`/`FOO-1.3`; `s := mcpSession(t, dir)`; call `task_create` with `{"title": "Contradiction forces the pointer patch", "parent": "FOO-1", "after": "FOO-1.1", "phase": 1, "phase_label": "Scope writes"}`.
     - **Assertions:** the call returns `{"id": "FOO-1.4"}`. `Load("FOO-1").Order` is exactly `["FOO-1.1", "FOO-1.4", "FOO-1.2", "FOO-1.3"]`. A `task_list` call with `{"parent": "FOO-1"}` returns the IDs in that same order — the ordering is observable through the read surface, not just on disk.
     - **Boundary:** `after` naming the first of three siblings — an interior insertion point, the case an append can never produce. The ID is still minted as the next number (`.4`), so position and identity are proven independent.
   - [ ] Confirm fails: `Order = ["FOO-1.1", "FOO-1.2", "FOO-1.3", "FOO-1.4"]` — the handler ignores `after` and `Store.Create` falls through to `AppendToOrder`. As in 2.1, `additionalProperties: false` may surface this as a tool error naming `after` instead; same signal.

2. **Implement (GREEN):**
   - [ ] Add `After string` (`json:"after,omitempty"`) to `taskCreateInput` and pass it as `CreateParams.After`.
   - [ ] Extend the tool's `Description`: `after` names a sibling bar and places the new bar directly after it in the track's order; omitting it appends. Say that `after` is only meaningful together with `parent`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic

## Commit (user)
`feat(mcp): support after placement when creating a bar`