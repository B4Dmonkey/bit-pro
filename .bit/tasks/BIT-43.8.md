---
id: BIT-43.8
title: task_read's description teaches track vs bar
status: todo
phase: 4
phase_label: Domain on descriptions
---
## **Verse 4**

`task_read` is the only tool registered with an inline literal description — `"Read a task by ID"`
at `cmd/serve_mcp.go:186` — while the other seven carry named constants that already teach
track-vs-bar. It is the zero case, so it is where the domain-enrichment test starts.

Unlike Verses 1–3 this is Go, so the normal TDD cycle applies.

## Scope
- `cmd/serve_mcp.go` — add a `taskReadDescription` constant beside the others and pass it at the
  `mcp.AddTool` call for `taskReadTool`
- `cmd/serve_mcp_test.go` — new test (the harness in `cmd/mcp_harness_test.go` already gives
  `mcpSession`)

## References
- `assets/bit-cli.md` — "The two kinds of task" is the prose being relocated. It retires in this
  verse's last bar, so read it now while it still exists.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestMCPToolDescriptions_CarryTheDomain` — table-driven, one subtest per tool, each row a
         `{tool string; want []string}` pair asserted with `strings.Contains`
     - **Behavior:** a caller learns the domain from the tool list alone, with no separate
       instructions command to fetch. The registered description for `task_read` explains what a
       track and a bar are, the way the other tools' descriptions already do.
     - **Setup:** `s := mcpSession(t, t.TempDir())`, then
       `res, err := s.ListTools(t.Context(), nil)` (signature confirmed at
       go-sdk v1.7.0 `mcp/client.go:1257` — returns `*ListToolsResult`, whose `Tools` is
       `[]*mcp.Tool` with `Name` and `Description`). Index the tools by `Name`. No task fixtures
       are needed; this reads registration, not data.
     - **Assertions:** the `task_read` row wants `"A track is a top-level task"` and `"BIT-7.3"` —
       the exact phrasing `taskListDescription`, `taskCreateDescription` and
       `taskCompleteDescription` already share, so the enriched description matches its neighbours
       rather than inventing a third way to say it. Fail with the tool name and the missing
       substring.
     - **Boundary:** `task_read` sits at zero domain sentences while the other seven sit at one or
       more — the lower bound of the description-enrichment set, and the only tool whose
       `mcp.AddTool` call passes a literal instead of a constant.
   - [ ] Confirm fails: the registered description is exactly `"Read a task by ID"`, so both
         `strings.Contains` assertions report the substring missing. **Not** a nil-map or
         tool-not-found panic — if it fails that way the indexing is wrong, not the description.

2. **Implement (GREEN):**
   - [ ] Add `const taskReadDescription` to the block in `cmd/serve_mcp.go`, opening with the
         shared sentence *"A track is a top-level task — one whole scope — and its ID has no dot,
         as in BIT-7. A bar is a child of a track — one plan step — and its ID is dotted, as in
         BIT-7.3."*
   - [ ] State what the tool returns, since `--body` has no analogue: the result carries `body`
         alongside `id, title, status, approved, phase, phase_label, parent`, so reading a body and
         reading a summary are the same call.
   - [ ] Replace `Description: "Read a task by ID"` at the `taskReadTool` `mcp.AddTool` call with
         `taskReadDescription`.

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` passes
- [ ] `just install` — the locally installed `bp` is the binary the MCP server runs from, so a
      description change is invisible until it is reinstalled

## User verifies
- [ ] none — deterministic. Verse 4's end-to-end check lands on its last bar.

## Commit (user)
`feat(mcp): task_read's description teaches track vs bar`