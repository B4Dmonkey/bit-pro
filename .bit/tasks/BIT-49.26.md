---
id: BIT-49.26
title: MCP tool descriptions and bp task complete's help no longer name .bit/
status: todo
approved: true
phase: 3
phase_label: no .bit/ in sight
---
## **Verse 3**

Agents read tool descriptions on every session, and four of them still point at `.bit/` paths that v2 doesn't have. This bar only changes wording, so per the operator rule it adds no new tests. The existing description table already guards the domain sentences, and they must survive the edit.

## Scope
Re-grep first: `grep -n '\.bit' cmd/serve_mcp.go cmd/task/complete.go`.
- `cmd/serve_mcp.go`:
  - `:76` (`task_complete`): "filing it and its bars under `.bit/completed/`" → "filing it and its bars as completed". Keep `testTrackSentence`. BIT-50 rewrites the semantics later.
  - `:84` (`task_delete`): drop the `.bit/archive` path and say "moves it to the archive".
  - `:102` and `:109` (`research_write`/`research_read`): "kept per track under .bit/research/<track>/, one file per topic" → "kept per track in the store, one record per topic".
- `cmd/task/complete.go:12` `Short`: `"Complete a task, filing it and its bars as completed"`.
- No test file is touched. If a description test fails, the edit dropped a guarded sentence, so restore the sentence rather than changing the test.

## Change checklist
- [ ] Edit the four descriptions and the `Short`.
- [ ] `grep -n '\.bit' cmd/serve_mcp.go cmd/task/complete.go` finds nothing.

## Claude verifies
- [ ] `just lint` and `just test` pass (the description table is unchanged and green)

## User verifies
- none, deterministic

## Commit (user)
`docs(bit): MCP descriptions and complete help drop .bit/ paths`