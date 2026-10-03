---
id: BIT-49.13
title: migrate carries research topics across as records under their tracks
status: todo
phase: 2
phase_label: bp migrate
---
## **Verse 2**

v1 research is `research/<TRACK>/<topic>.md`, with no frontmatter. A test that reads a topic back through `research_read` after the migration forces the copy through `WriteResearch`.

## Scope
- `migrate/migrate.go`: for each dir in `src/research/` and each `*.md` in it, read it into `raw` and call `s.WriteResearch(trackDir, strings.TrimSuffix(name, ".md"), string(raw), task.Commit{})`. It runs after the tasks, so tracks in `completed/` or `archive/` resolve.
- No `task` change: `WriteResearch` already takes a caller commit (BIT-46.16).
- Test file touched: `cmd/migrate_test.go`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestMigrateCmd/carries research topics`
     - **Behavior:** a track's research notes stay readable through the tools after the migration.
     - **Setup:** sandbox. The v1 store has `tasks/BIT-49.md` and `completed/BIT-44.md`. `research/BIT-49/index.md` holds a body with a markdown table and `[x](x.md)` links. `research/BIT-49/claim-audit-2026-10-02.md` and `research/BIT-44/index.md` sit beside it. Then `bp migrate`, then an MCP session at `dir` in the same sandbox (BIT-49.10's helper): `research_read {track: "BIT-49"}` and `research_read {track: "BIT-49", topic: "index"}`.
     - **Assertions:** topics == `["claim-audit-2026-10-02", "index"]`. The body is byte-equal to the source. `<data>/bit/BIT/research/BIT-44/index.json` has `"project": "BIT"` and `"track": "BIT-44"`.
     - **Boundary:** research on a completed track, and a topic name with digits and dashes.
   - [ ] Confirm fails: `research_read` lists no topics.

2. **Implement (GREEN):**
   - [ ] The research loop.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): migrate carries research topics across`