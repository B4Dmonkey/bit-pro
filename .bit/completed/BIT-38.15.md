---
id: BIT-38.15
title: 'Contradiction: force overrides unfinished bars'
status: done
approved: true
phase: 4
phase_label: Close out
---
## **Verse 4**

Bar 4.2's handler passes `false` for `force` unconditionally, so a track with an unfinished bar can never be deleted through the tool. This test cannot pass under that reading — it is the contradiction that forces the param through. "A track with unfinished bars needs an override" is a domain rule, unlike the confirmation prompt, which is why this one crosses over. Completes verse 4, and with it the last pipeline write that needed a shell.

## Scope
- `cmd/serve_mcp.go` — `taskDeleteInput` gains `force`; the handler forwards it.
- `cmd/serve_mcp_write_test.go` — the contradicting test.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeMCPCmd_TaskDeleteForceOverridesUnfinishedBars` (table-driven)
     - **Behavior:** an abandoned track can be dropped along with its unfinished bars when the caller says so explicitly, and refused when it does not.
     - **Setup:** `dir := t.TempDir()`; per case, `seedTasks` track `FOO-1` (`Order: []string{"FOO-1.1", "FOO-1.2"}`) plus bars `FOO-1.1` at `task.StatusDone` and `FOO-1.2` at `task.StatusDoing`; `s := mcpSession(t, dir)`. Cases: (a) `{"id": "FOO-1", "force": true}`; (b) `{"id": "FOO-1", "force": false}`; (c) `{"id": "FOO-1"}` — force omitted.
     - **Assertions:** (a) returns `{}` with `IsError` false, and all three of `FOO-1.md`, `FOO-1.1.md`, `FOO-1.2.md` are under `.bit/archive/tasks/` with none left under `.bit/tasks/`. (b) and (c) both come back `IsError` true with all three files still under `.bit/tasks/` and nothing archived — an omitted `force` must behave exactly like an explicit `false`, never like `true`.
     - **Boundary:** `force` in all three of its wire states (true, false, absent), against an unfinished-bar count of 1 — the lowest count that triggers the rule at all.
   - [ ] Confirm fails: case (a) returns `IsError` true with the files still in `.bit/tasks/` — the handler hardcodes `false`, so `relocateTree` returns `*UnfinishedBarsError` regardless of what was sent. Cases (b) and (c) already pass, which is what makes (a) the contradiction.

2. **Implement (GREEN):**
   - [ ] Add `Force bool` (`json:"force,omitempty"`) to `taskDeleteInput`. A plain bool, not a pointer: `false` and absent must mean the same thing, and that is what a value type says.
   - [ ] Pass it through: `store.Relocate(in.ID, in.Force)`.
   - [ ] Extend the tool's `Description`: `force` deletes a track that still has unfinished bars; without it such a track is refused.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`
- [ ] `grep -rn "mcp.AddTool" cmd/serve_mcp.go | wc -l` prints 8 — every tool in the parity map is registered

## User verifies
- [ ] Whole slice: with the server registered, start a session and take a throwaway track all the way out — ask for a scope track plus two bars, mark both bars done, then ask for the track to be completed. Observe `bp task list` no longer shows it and `.bit/completed/` holds all three files. Then, on a second throwaway track with one bar left at `todo`, ask for it to be deleted: observe it is refused, then ask again saying to force it, and observe all its files land in `.bit/archive/tasks/`. The whole sequence should show `mcp__bit__*` calls with no Bash command touching `.bit/`. Finish with `claude mcp remove bit` unless you want the server registered before step 4 lands.

## Commit (user)
`feat(mcp): support force when deleting a track with unfinished bars`