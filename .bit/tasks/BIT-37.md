---
id: BIT-37
title: 'MCP read surface: task_list'
status: todo
---
## Why

Every bit skill that reads a plan reaches it through Bash: `bp task list --parent BIT-36`,
then parsing five tab-separated columns. Two of those columns are routinely empty, which is
why the command contract has to warn "count tabs rather than assuming the phase label is the
fourth field" — and a mis-parse is silent, because a bar read as unapproved or unphased looks
exactly like ordinary data. That format exists so a terminal can print a plan; nothing about
it serves a model. `task_read` already returns fields over MCP, so a single task can be read
without a shell. There is no equivalent for reading a track's bars, which is the most common
read in the whole pipeline — 16 `task list` call sites across the skills, more than any other
command.

## Summary

Add `task_list` to the MCP server, mirroring `bp task list`: an optional `parent`, and an
array of task objects — everything `task_read` returns except the body — in the same order
the CLI produces. That closes the read half of the MCP surface; the write tools come next.

## Visual aid

```
today                              after
-----                              -----
skill                              skill
  └─ Bash: bp task list -p BIT-36    └─ task_list{parent: "BIT-36"}
       "BIT-36.1\tdone\t…\t\t"            {tasks: [{id, title, status,
       └─ split on \t, count 5                      approved, phase,
          fields, hope                              phase_label, parent}, …]}
```

## Risks & unknowns

- **Unknown:** The MCP server resolves the store as `$CLAUDE_PROJECT_DIR/.bit`
  (`cmd/serve_mcp.go`), while the CLI resolves it through `bitdir.Canonical`, which cuts a
  `.claude/worktrees/<name>/` path back to the main checkout's `.bit/`. In a worktree session
  the two disagree, so tools would read a different store than `bp` writes. This contradicts
  the "How the server finds `.bit/`" decision in `mcp-notes.md`, which assumed the server was
  already on the BIT-27 cut.
  **Resolve by:** Operator's call — either a verse in this scope puts the server on
  `bitdir.Canonical`, or it stays with the automation phase, which is where worktree sessions
  actually get dispatched.
  **De-risk before planning?** Yes — it decides whether this scope is one verse or two.

## Decisions

- **`task_list` returns a wrapper object, `{tasks: [...]}`, not a bare array.** SEP-2106 and
  the Go SDK both allow a slice as the output type, but an object is what `structuredContent`
  has always been in the base spec; the wrapper costs one key and leaves nothing to discover
  about how Claude Code's client handles a non-object result.
- **The listed shape is `task_read`'s minus `body`.** A list of bodies is never what a caller
  wanted; a caller that needs one calls `task_read`.
- **`parent` is optional and mirrors the CLI exactly.** Absent lists every task; present lists
  that track's direct bars, in the track's explicit order. No new filtering, no new sorting.
- **Domain lives in the tool description, per-tool.** `mcp-notes.md` deferred *where* the
  domain half of `bp instructions` lands until tools were being written — this is that point,
  and the cheapest answer holds: `task_list`'s description carries track vs. bar and the
  dotted-ID rule. Retiring `bp instructions` stays step 5's job.
- **No skill edits in this scope.** Step 2's "done when" is a statement about capability, not
  about call sites; the skills keep using Bash until the migration step.

## Verses

- [ ] Verse 1 — Claude can read a whole plan without a shell: `task_list` over MCP, with an
  optional `parent`, returning the same tasks in the same order `bp task list` prints, as
  structured fields instead of tab columns.
  Touches: `cmd/serve_mcp.go`, `cmd/serve_mcp_test.go` — the tool registration and its
  in-memory-transport test.

## References

- `mcp-notes.md` — the MCP phase's working notes. Todo step 2 ("Read surface") is this scope;
  its Decisions section and parity map are the source for the tool's shape.