---
id: BIT-37.2
title: task_list returns every task as structured fields
status: todo
approved: true
phase: 1
phase_label: Read surface
---
## **Verse 1**

There is no `task_list` tool, so a track's bars can only be read through Bash and five
tab-separated columns. This bar adds the tool with no `parent`, returning every task as fields.

## Scope
- `cmd/serve_mcp.go` — the `task_list` const, input/output types, handler, and registration;
  plus extracting the parent-from-ID derivation `taskReadHandler` does inline so both handlers
  share it.
- `cmd/serve_mcp_test.go` — the new test.

## References
- `mcp-notes.md` — the Parity map row for `task list` → `task_list` (params and return shape)
  and the "Structured returns, not text" decision.
- `assets/bit-cli.md:13-22` — the settled wording for the track-vs-bar and dotted-ID domain the
  tool description has to carry. Paraphrase it; don't reinvent it.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeMCPCmd_TaskListReturnsEveryTaskAsFields`
     - **Behavior:** a caller reads a whole plan as structured objects in one call — every field
       `task_read` returns except the body — with no `parent` given.
     - **Setup:** store at `<t.TempDir()>/.bit` holding a track
       `&task.Task{ID: "FOO-1", Title: "mcp test track", Status: task.StatusTodo,
       Approved: true, Order: []string{"FOO-1.1"}}` and a bar
       `&task.Task{ID: "FOO-1.1", Title: "a bar", Status: task.StatusDoing, Phase: 2,
       PhaseLabel: "Read surface"}`. Call `task_list` with `Arguments: map[string]any{}`.
     - **Assertions:** `result.IsError` is false; `structuredContent` unmarshals to an object
       whose `tasks` key holds 2 entries in the order `FOO-1`, `FOO-1.1`. Entry 0: `id FOO-1`,
       `title mcp test track`, `status todo`, `approved true`, `phase 0`, `phase_label ""`,
       `parent ""`. Entry 1: `id FOO-1.1`, `title a bar`, `status doing`, `approved false`,
       `phase 2`, `phase_label Read surface`, `parent FOO-1`. Neither entry carries a `body`
       key.
     - **Boundary:** `parent` omitted — the absent end of the optional input, which is also
       what proves `omitempty` kept it out of the schema's `required`. The store holds one of
       each kind of task, so the track case (no dot, zero phase, no parent) and the bar case
       (dotted, phased, parented) are both exercised instead of one standing in for both.
   - [ ] Confirm fails: `session.CallTool` returns an error — the server registers no tool named
     `task_list`.

2. **Implement (GREEN):**
   - [ ] `taskListTool = "task_list"` const.
   - [ ] `taskListInput` with `Parent string` tagged `json:"parent,omitempty"` — `omitempty` is
     what keeps `parent` out of the inferred schema's `required` list.
   - [ ] `taskSummary` — `taskReadOutput`'s fields minus `Body` — and
     `taskListOutput` with `Tasks []taskSummary` tagged `json:"tasks"`, which is the wrapper
     object the scope decided on.
   - [ ] extract `parentOf(id string) string` from `taskReadHandler`'s inline
     `strings.LastIndex`, and call it from both handlers.
   - [ ] `taskListHandler(root)`: build the store from the resolver added in the previous bar,
     call `store.List()`, map each task into a `taskSummary`. There is no lower layer to force
     here — `task.Store.List` already exists and already sorts by the track's order — so the
     mapping *is* the minimum.
   - [ ] register it with `mcp.AddTool`, the description carrying the domain: what a track is,
     what a bar is, that a bar's ID is dotted, and that `parent` lists one track's direct bars.

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` passes — `funlen`/`cyclop` are enabled, so if the handler outgrows a mapping
  loop the mapping moves to its own function rather than the handler growing

## User verifies
- [ ] none — deterministic; the verse's live check belongs on its last bar

## Commit (user)
`feat(mcp): add task_list returning every task as structured fields`