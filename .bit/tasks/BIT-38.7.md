---
id: BIT-38.7
title: The status enum refuses a typo the CLI accepts
status: todo
approved: true
phase: 1
phase_label: Scope writes
---
## **Verse 1**

`bp task update -s doen` succeeds and stores `doen`, which breaks rollup forever — a verse never checks off and a track never signals ready. That is a stringly-typed-flag accident, and the schema fixes it without introducing the state machine the CLI deliberately lacks. Forced by a test asserting the tool refuses a status the CLI accepts. Completes verse 1.

## Scope
- `cmd/serve_mcp.go` — `task_update`'s `InputSchema` is supplied explicitly rather than inferred, so `status` can carry an enum.
- `cmd/serve_mcp_write_test.go` — the test.
- `go.mod` — `github.com/google/jsonschema-go` moves from indirect to direct.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeMCPCmd_TaskUpdateRefusesAnUnknownStatus`
     - **Behavior:** the protocol rejects a misspelled status before the handler runs, so a typo can never be written to a task file.
     - **Setup:** `dir := t.TempDir()`; `seedTasks(t, dir, &task.Task{ID: "FOO-1", Title: "MCP write surface", Status: task.StatusTodo})`; `s := mcpSession(t, dir)`; `callToolErr(t, s, taskUpdateTool, map[string]any{"id": "FOO-1", "status": "doen"})`.
     - **Assertions:** the call result has `IsError` true (a validation failure comes back as a tool error, not a transport error). `Load("FOO-1").Status` is still `todo` — the write never happened.
     - **Boundary:** `status` just outside its valid set. Pair it with subtests sending each of `todo`, `doing`, `done` and asserting the call succeeds, so the enum is proven to admit the whole valid range rather than merely reject one string.
   - [ ] Confirm fails: the `"doen"` case returns `IsError` false and `Load("FOO-1").Status` is `"doen"` — the inferred schema types `status` as a plain nullable string, so anything passes.

2. **Implement (GREEN):**
   - [ ] Build the schema explicitly: `schema, err := jsonschema.For[taskUpdateInput](nil)`, then `schema.Properties["status"].Enum = []any{task.StatusTodo, task.StatusDoing, task.StatusDone}`, then pass it as `InputSchema` on the `mcp.Tool` for `task_update`. `mcp.Tool.InputSchema` is `any` and the SDK uses a supplied `*jsonschema.Schema` instead of reflecting over the input type, so the rest of the schema stays derived from the struct — only `status` is hand-tightened.
   - [ ] `runMCPServer` returns an error today only from `s.Run`; `jsonschema.For` can fail, so thread that error out rather than dropping it — `AddTool` panics on a bad schema and a server that panics at registration is worse than one that refuses to start.
   - [ ] `go mod tidy` so `github.com/google/jsonschema-go` is recorded as a direct dependency. It is already in the build via the MCP SDK, so no new module is downloaded.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`
- [ ] `go mod tidy` leaves `go.mod` unchanged on a second run
- [ ] `grep -n "jsonschema-go" go.mod` shows it without the `// indirect` marker

## User verifies
- [ ] Whole slice: `just install`, then `claude mcp add bit -- bp serve mcp` (this is step 4's own command, run early — undo with `claude mcp remove bit`). Start a new session in this repo and ask, in plain language, for a new scope track to be created with a couple of paragraphs of body, then for that body to be rewritten. Observe: the transcript shows `mcp__bit__task_create` and `mcp__bit__task_update` rather than a Bash `bp task` call, `bp task read <id> --body` prints the rewritten body intact, and the `task_update` result carries an `approved` field. Then `bp task delete <id> -y` to clean up.

## Commit (user)
`feat(mcp): constrain task_update status to the three valid values`