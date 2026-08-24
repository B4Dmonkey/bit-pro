---
id: BIT-38.8
title: 'Contradiction: a parent forces dotted-ID minting'
status: done
approved: true
phase: 2
phase_label: Plan writes
---
## **Verse 2**

Bar 1.3's handler mints from the config prefix and ignores everything but title and body, so it cannot produce a dotted ID or carry a phase tag. This test cannot pass under that reading — it is the contradiction that forces `parent` and the phase tags into the tool.

## Scope
- `cmd/serve_mcp.go` — `taskCreateInput` gains `parent`, `phase`, `phase_label`; the handler forwards them.
- `cmd/serve_mcp_write_test.go` — the contradicting test.

## References
- `mcp-notes.md` — "Parity map", the `task_create` row.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeMCPCmd_TaskCreateMintsABarUnderATrack`
     - **Behavior:** Claude can lay out a plan step through the protocol — a bar under a named track, tagged with the verse it serves, appended to the track's order.
     - **Setup:** `dir := t.TempDir()`; `task.New(filepath.Join(dir, ".bit")).SaveConfig(&task.Config{Prefix: "FOO"})`; `seedTasks(t, dir, &task.Task{ID: "FOO-1", Title: "MCP write surface", Status: task.StatusTodo})`; `s := mcpSession(t, dir)`. Call `task_create` twice: first `{"title": "Store.Create owns ID minting", "parent": "FOO-1", "phase": 1, "phase_label": "Scope writes", "body": "## **Verse 1**\n\nstep detail"}`, then `{"title": "task_create mints a track", "parent": "FOO-1", "phase": 1, "phase_label": "Scope writes"}`.
     - **Assertions:** the two calls return `{"id": "FOO-1.1"}` and `{"id": "FOO-1.2"}`. `Load("FOO-1.1")` has `Phase` 1, `PhaseLabel` `Scope writes`, `Status` `todo`, and the body sent. `Load("FOO-1.2").Body` is `""`. A `task_list` call with `{"parent": "FOO-1"}` returns the two bars in creation order with `phase_label` populated.
     - **Boundary:** `parent` at its non-empty state — the branch that selects `NextChildID` over `NextID`. Two creations rather than one puts the existing-children count at 1 for the second, which is what distinguishes real child minting from an always-`.1` return. `phase` at 1, its lowest meaningful value (0 means untagged).
   - [ ] Confirm fails: the first call returns `{"id": "FOO-2"}` — the handler ignores `parent` and mints from the prefix — and `Load("FOO-1.1")` fails with `fs.ErrNotExist`. Note the schema has `additionalProperties: false`, so an unknown key is rejected outright; if the call comes back as a tool error complaining about `parent`, that is the same signal.

2. **Implement (GREEN):**
   - [ ] Add `Parent string` (`json:"parent,omitempty"`), `Phase int` (`json:"phase,omitempty"`), `PhaseLabel string` (`json:"phase_label,omitempty"`) to `taskCreateInput`.
   - [ ] Forward them in the handler: `task.CreateParams{Title: in.Title, Body: in.Body, Parent: in.Parent, Phase: in.Phase, PhaseLabel: in.PhaseLabel}`. No ordering code here — `Store.Create` already calls `AppendToOrder` when a parent is set and no anchor is given.
   - [ ] Extend the tool's `Description`: a `parent` mints a dotted bar ID under that track and appends it to the track's order; `phase`/`phase_label` tag the verse the bar serves.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic

## Commit (user)
`feat(mcp): create bars under a track with verse tags`