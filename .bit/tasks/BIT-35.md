---
id: BIT-35
title: MCP server skeleton — task_read over stdio
status: todo
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
Add `bp serve mcp` as a subcommand under the existing `serve` parent (which the operator
is renaming from a leaf to a parent — `bp serve` → `bp serve daemon` — before this
scope runs). The subcommand runs a stdio MCP server exposing one read-only tool:
`task_read`. The server speaks JSON-RPC over stdin/stdout, initialises on the
`initialize` handshake, and returns a task's fields as structured JSON rather than the
tab-column text the CLI formats for a terminal. `task_read` is the only tool; the write
surface comes in a later scope.

## Decisions

- **`bp serve mcp` is the command — no temporary top-level `bp mcp`.** The operator is
  doing the `serve` → `serve daemon` rename before this scope runs, so `serve` arrives as
  a parent and `mcp` slots in as a sibling of `daemon`. No migration needed later.
- **Use `github.com/modelcontextprotocol/go-sdk/mcp` v1.7.0.** The generic `AddTool`
  derives JSON Schema from Go structs, so each tool is two structs (input, output) rather
  than a hand-written schema that drifts. v1.7.0 ships a no-breaking-API-changes guarantee
  and requires Go 1.25+; `go.mod` is on 1.26.5.
- **The server resolves `.bit/` from `CLAUDE_PROJECT_DIR`, not cwd.** The env var is set
  by Claude Code to the stable project root at server startup. `bitdir.Resolve()` runs in
  `PersistentPreRunE` for all commands and is harmless here (no stdout, pure path math),
  but the MCP server maintains its own store reference rooted at `CLAUDE_PROJECT_DIR`.
- **No project/root parameter on any tool call.** One server process per session; it
  resolves the root once at startup.
- **Nothing writes to stdout.** The stdio transport owns stdout; any stray `fmt.Println`
  in a shared code path corrupts the protocol stream. The mcp command guards against this
  by construction — the plan step that wires the server must audit shared init paths.
- **`task_read` returns structured fields, not formatted text.** Fields: `id`, `title`,
  `status`, `approved`, `phase`, `phase_label`, `parent`, `body`. The header line and
  `--body` flag are terminal affordances that disappear on a structured return.
- **The confirmation prompt does not cross over; `--force` does.** (Relevant for later
  `task_delete`; recorded here so bit_plan doesn't relitigate it.)

## Verses

- [ ] Verse 1 — Operator can wire `bp serve mcp` into Claude Code and Claude can call
  `task_read` on a real task:
  `bp serve mcp` starts a stdio MCP server. Typing an `initialize` frame by hand gets a
  response. Claude Code lists `task_read` in the tool panel for a project wired to it.
  A `task_read` call on a real track returns its structured fields including body.
  Touches: `cmd/serve.go` (the new `mcp` subcommand hangs off the `serve` parent the
  rename creates), `task/store.go` (Load path the tool calls into).

## References

- `mcp-notes.md` — working notes for the full MCP phase; this scope covers step 1
  (Server skeleton) only. Decisions section records the settled answers from that doc.