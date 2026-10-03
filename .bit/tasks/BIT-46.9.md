---
id: BIT-46.9
title: MCP tools resolve the project per call through the registry
status: todo
approved: true
phase: 1
phase_label: registered project works from the central store
---
## **Verse 1**

Every MCP handler already builds its store per call (`task.New(bitdir.ForRoot(root))`, `cmd/serve_mcp.go:294,320,360,384,407,423,439,455,472,489`). The new work is the registry lookup per call and resolver errors as tool errors. A session rooted in a subfolder can't find its tasks through `ForRoot`, so that test forces the switch.

## Scope
- `cmd/serve_mcp.go`:
  - unexported `mcpStore(ctx context.Context, root string) (*task.Store, error)`: `dir := root`; if empty, `os.Getwd()` (the opencode fallback); `project.OpenStore(ctx, dir)`.
  - all 10 handlers take their `ctx` (now `_`) and call `mcpStore(ctx, root)`; a resolver error is returned as the handler error, so the client sees `IsError` with the resolver's text.
  - `CLAUDE_PROJECT_DIR` is still read once at startup (`:231`); nothing resolves at startup. `runMCPServer(ctx, root, transport)` keeps its signature.
  - tool descriptions naming `.bit/` (`:76,84,102,109`) stay as they are (BIT-49's sweep).
- `cmd/mcp_harness_test.go`, `cmd/serve_mcp_write_test.go`, `cmd/serve_mcp_research_test.go`, `cmd/serve_mcp_test.go`:
  - `mcpSandbox(t)` in `cmd/mcp_harness_test.go`: an idempotent per-test sandbox helper (sets `HOME` and `XDG_DATA_HOME` to temp dirs once per test), called by `mcpSession` and every seed helper, so no MCP test opens the real registry. **Seam:** BIT-49.3, BIT-49.10, BIT-47.1 and BIT-50.1 call it by this name, and BIT-47.6 keeps the call when it moves `mcpSession`'s body into `mcpSessionWithGit`.
  - `seedConfig` becomes `registerProject(t, dir)`: registers `CanonicalPath(dir)` as `testCode` through `seedProject` (`cmd/list_test.go`), skipping it if `project.Find` already resolves `dir`. Don't name it `seedProject`: that function already exists in package `cmd` and keeps its `(t, orm.CreateProjectParams)` signature.
  - `seedTasks`, `seedEscapedResearch`, `seedOrderedTrack`, `seedDoneTrack` call `registerProject(t, dir)` and write through `project.OpenStore(ctx, dir)` instead of `task.New(filepath.Join(dir, ".bit"))`. Their signatures stay as they are (`seedTasks(t, dir, tasks...)`, `seedDoneTrack(t, dir, lastBarStatus)`).
  - returned-path assertions move to the store dir: `serve_mcp_research_test.go:89,113,132,162,308`, `serve_mcp_write_test.go:523,546,586,591,617,621,705,710`; direct `.bit` peeks (e.g. `:77` `task.New(filepath.Join(dir, ".bit")).Load`) read through `project.OpenStore`.
  - the 36 `mcpSession(t, dir)` sites keep passing `dir`.

## TDD cycle

0. **Convert existing tests (pure restructure, same cases and assertions, still green).** One top-level test per tool handler; each flat test becomes a subtest named by what follows the tool name, in lowercase words:
   - [ ] `cmd/serve_mcp_test.go`: `TestTaskReadHandler/{returns structured fields, returns parent for bar}`; `TestTaskListHandler/{returns every task as fields, parent returns only that tracks bars in order}`; `TestServeMCPCmd_ResolvesWorktreeRootToMainCheckout` → `TestRunMCPServer/resolves worktree root to main checkout`; `TestMCPToolDescriptions_CarryTheDomain` → `TestMCPToolDescriptions/carry the domain`.
   - [ ] `cmd/serve_mcp_write_test.go`: `TestTaskCreateHandler/{mints a track, mints a bar under a track, after places a bar mid track}`; `TestTaskUpdateHandler/{rewrites body and reports revocation, leaves omitted fields alone, refuses an unknown status}`; `TestTaskMoveHandler/{resequences a bar, refuses a bad anchor pair}`; `TestFeedbackAddHandler/{writes a note, refuses an unknown track, refuses a path like track id}`; `TestTaskCompleteHandler/{files a track and its bars, refuses unfinished bars}`; `TestTaskDeleteHandler/{relocates and reserves the id, force overrides unfinished bars}`.
   - [ ] `cmd/serve_mcp_research_test.go`: `TestResearchWriteHandler/{writes a topic, overwrites a topic, refuses an unknown track, accepts a completed track, strips leading dots, refuses a path like topic, refuses a track id that escapes}`; `TestResearchReadHandler/{returns a topic, refuses a missing topic, refuses an unknown track, lists topics, lists nothing for a track with no research, refuses a path like topic, refuses a track id that escapes}`.
   - `cmd/mcp_harness_test.go` holds no tests.

1. **Write test (RED):**
   - [ ] `TestRunMCPServer/resolves a subfolder root through the registry` (subtest, `cmd/serve_mcp_test.go`)
     - **Behavior:** a session started anywhere under a registered project reads and writes that project's central store.
     - **Setup:** sandbox; register `dir` as `BIT`; `os.MkdirAll(dir+"/sub")`; `mcpSession(t, dir+"/sub")`; call `task_create {title: "Track", body: "..."}`, then `task_read {id: "BIT-1"}`.
     - **Assertions:** create returns `id == "BIT-1"`; read returns `title == "Track"`; `$XDG_DATA_HOME/bit/BIT/tasks/BIT-1.md` exists; `dir/sub/.bit` and `dir/.bit` don't.
     - **Boundary:** root one level below the registered path.
   - [ ] Confirm fails: `ForRoot(dir/sub)` → `dir/sub/.bit`, which has no config, so `task_create` errors.

2. **Implement (GREEN):**
   - [ ] `mcpStore` and the 10 handlers.

3. **More tests (RED → GREEN):**
   - [ ] `TestRunMCPServer/unregistered root is a tool error` (subtest): sandbox, unregistered temp dir; `callToolResult(..., task_list, {})` → `IsError == true`, content text contains ``not a bit project; run `bp add` ``. *Boundary:* the error path goes through the MCP result, not a crash.
   - [ ] `TestRunMCPServer/empty root falls back to the working directory` (subtest): register `dir`; `t.Chdir(dir)`; `mcpSession(t, "")`; `task_list` succeeds. *Boundary:* `CLAUDE_PROJECT_DIR` unset (opencode).
   - [ ] `TestRunMCPServer/resolves worktree root to main checkout` (`:97`): unchanged expectation; the worktree path doesn't exist on disk, which exercises the resolver's fallback.
   - [ ] the rewritten seed-based tests all pass with their existing behavior assertions.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `test ! -e ~/.local/share/bit/main.db`
- [ ] `grep -n "bitdir" cmd/serve_mcp.go` finds nothing

## User verifies
- none — deterministic (the verse's end-to-end check is on its last bar)

## Commit (user)
`feat(bit): MCP tools resolve the project per call through the registry`