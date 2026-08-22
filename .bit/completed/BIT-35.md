---
id: BIT-35
title: MCP server skeleton — task_read over stdio
status: done
---
## Why
Claude reaches `bp` today through Bash — one of thousands of things a shell can invoke.
The consequence is visible: Claude reaches for `mv`, `cat`, and `sed` against `.bit/`
instead of the command that owns the file format, and the command contract spends half
its length teaching shell technique rather than the domain. A typed MCP surface makes the
right tools appear in Claude's tool list without explanation, and makes the wrong path
closeable by denial rules rather than by documentation. This first verse proves the
transport: a live stdio server that Claude Code can wire in and Claude can query without
a shell.

## Summary
Add `bp serve mcp` as a sibling of `bp serve daemon` (which BIT-36 creates). The
subcommand runs a stdio MCP server exposing one read-only tool: `task_read`. The server
speaks JSON-RPC over stdin/stdout, initialises on the `initialize` handshake, and returns
a task's fields as structured JSON rather than the tab-column text the CLI formats for a
terminal. `task_read` is the only tool; the write surface comes in a later scope.

## Decisions

- **`bp serve mcp` is the command — added as a sibling of `serve daemon` after BIT-36
  lands.** BIT-36 turns `serve` into a parent (`newServeCmd()` returns a group,
  `newServeDaemonCmd()` holds the loop body). This scope adds `newServeMCPCmd()` in a new
  `cmd/serve_mcp.go`, consistent with how every other command is organised.
- **Use `github.com/modelcontextprotocol/go-sdk/mcp` v1.7.0.** The generic `AddTool`
  derives JSON Schema from Go structs, so each tool is two structs (input, output) rather
  than a hand-written schema that drifts. v1.7.0 ships a no-breaking-API-changes guarantee
  and requires Go 1.25+; `go.mod` is on 1.26.5. First new direct dependency since the
  daemon work.
- **The server resolves `.bit/` from `CLAUDE_PROJECT_DIR`, not cwd.** The env var is set
  by Claude Code to the stable project root at server startup. `bitdir.Resolve()` runs in
  `PersistentPreRunE` for all commands and is harmless here (no stdout, pure path math),
  but the MCP server calls `task.New(root)` with the env-var path directly rather than
  `bitdir.Current()`.
- **No project/root parameter on any tool call.** One server process per session; it
  resolves the root once at startup.
- **Nothing writes to stdout.** The stdio transport owns stdout; any stray `fmt.Println`
  in a shared code path corrupts the protocol stream. The plan step that wires the server
  must confirm no shared init path touches stdout — `bitdir.Resolve()` does not (it is
  pure path math with no I/O), and `PersistentPreRunE` does nothing else today.
- **`task_read` returns structured fields, not formatted text.** Fields: `id`, `title`,
  `status`, `approved`, `phase`, `phase_label`, `parent`, `body`. The header line and
  `--body` flag are terminal affordances that disappear on a structured return.
- **The confirmation prompt does not cross over; `--force` does.** (Relevant for later
  `task_delete`; recorded here so bit_plan doesn't relitigate it.)

## Verses

- [x] Verse 1 — Operator can wire `bp serve mcp` into Claude Code and Claude can call
  `task_read` on a real task:
  `bp serve mcp` starts a stdio MCP server. Typing an `initialize` frame by hand gets a
  response. Claude Code lists `task_read` in the tool panel for a project wired to it.
  A `task_read` call on a real track returns its structured fields including body.
  Touches: `cmd/serve_mcp.go` (new — `newServeMCPCmd()` added to `newServeCmd()` in
  `cmd/serve.go`), `task/store.go` (`Load` is the read path the tool calls into),
  `go.mod` / `go.sum` (new `github.com/modelcontextprotocol/go-sdk` dep).

## References

- `mcp-notes.md` — working notes for the full MCP phase; this scope covers step 1
  (Server skeleton) only. Decisions section records the settled answers from that doc.
- BIT-36 — the `serve` → `serve daemon` rename this scope depends on.