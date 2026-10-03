# MCP and CLI write surface: where `commit`/`branch`/`commits` hook in (Q3)

v1 code on branch `v2`; BIT-46 rewrites the store (JSON metadata under `~/.local/share/bit/<CODE>/`), so these are today's seams, not final ones.

## Input structs, `cmd/serve_mcp.go`
- `taskUpdateInput` `:151-158`: `ID string`; pointer fields `Title, Body, Status *string`, `Phase *int`, `PhaseLabel *string` (omitempty, nil = unchanged). Natural place for `Commit *string`, `Branch *string`.
- Schema built by `taskUpdateSchema` `:277-286` via `jsonschema.For[taskUpdateInput]`, adding a status enum. New fields come for free.
- Handler `taskUpdateHandler` `:378-399` maps input to `task.Patch` (`task/store.go:260-266`) and calls `store.Update`. Patch needs matching fields.
- **Approval trap:** `Store.Update` `task/store.go:298-303` revokes approval when `contentChanged` (title/body/phase/phase_label) or status → todo. Commit/branch must not count as content change, or recording a hash would un-approve a bar. Keep them out of `contentChanged`.
- `taskCompleteInput` `:160-162`: only `ID`. Handler `:417-431` calls `store.Complete(id)` → `relocateTree` (`task/store.go:108-112`). Adding `commit`/`branch` here is mostly BIT-50's (track landing commit); stub says `task_complete` accepts them in this track.
- `feedbackAddInput` `:169-172`: `Track, Body`. Handler `:449-464` → `store.AddNote`.
- `researchWriteInput` `:178-182`: `Track, Topic, Body`. Handler `:466-481` → `store.WriteResearch`.
- `retro_write` does not exist yet (BIT-49).
- For `commits` on research/feedback, the stub says capture is automatic (HEAD at write time), so no new input field is needed; the handler supplies it (see [capture](capture.md)).
- Every handler builds `task.New(bitdir.ForRoot(root))` where `root = os.Getenv("CLAUDE_PROJECT_DIR")` (`:231`, passed into `runMCPServer` `:238`). `root` is the session dir; `ForRoot` cuts a worktree path back to the main checkout's `.bit` (`bitdir/bitdir.go:49-55`). So the session dir is available in the handler but **lost** once inside `Store` — git lookups must happen in the handler (or Store gets it passed in), not derived from `s.root`.

## CLI equivalents
- `cmd/task/update.go:9-53`: flags `--title/-t`, `--description/-d`, `--status/-s`, `--phase`, `--phase-label`, each mapped to Patch only when `Changed` (`:23-41`). `--commit` / `--branch` would follow the same pattern. Store from `bitdir.Current()` (`:19`); session dir = `os.Getwd()`.
- `cmd/task/complete.go:15`: `Complete(args[0])`, no flags.
- `cmd/feedback_add.go:19,29`: `AddNote(args[0], description)`, flag `--description/-d`.
- No CLI for research writes (MCP-only).
- Whether the CLI should take the inputs: the stub says "MCP write tools"; CLI parity is not decided. bp CLI and MCP share the Store, so adding to Patch makes CLI flags cheap. **Open for scope.**

## Tests
- `cmd/serve_mcp_write_test.go`, `cmd/serve_mcp_research_test.go`, `cmd/mcp_harness_test.go` — in-memory MCP harness to extend; `cmd/task/update_test.go`, `cmd/feedback_add_test.go` for CLI.
