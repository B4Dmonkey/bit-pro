---
id: BIT-49.6
title: retro_list shows every project's proposals with each one's project
status: todo
phase: 1
phase_label: feedback and retro are shared across projects
---
## **Verse 1**

learn reads every project's proposals, and retro checks them for duplicates. Seeding proposals from two projects forces a listing that, unlike `feedback_list`, doesn't filter, and that reports each record's `project`.

## Scope
- `task/retro.go`:
  - `type Proposal struct { Name string \`json:"name"\`; Project string \`json:"project"\` }`.
  - `func (s *Store) ListRetro() ([]Proposal, error)`. It reads every `retro/*.json` and returns name and project, sorted by name. A missing folder gives an empty list.
  - **Seam:** BIT-49.16's verify reuses it.
- `cmd/serve_mcp.go`: const `retroListTool = "retro_list"`; empty input; `retroListOutput{Proposals []task.Proposal \`json:"proposals"\`}` (`[]`, never `null`). The description says it lists every project's proposals, each with its project, so learn sees them all and retro can avoid re-proposing a pattern. It names no path.
- `cmd/testconst_test.go`: `testEveryProject = "every project's proposals"`.
- Test files touched: `cmd/serve_mcp_retro_test.go` (created in shape by BIT-49.5) and `cmd/serve_mcp_test.go` (already converted by BIT-46.9).

## TDD cycle

0. **Restructure:** none. Both files are already in shape (BIT-49.5, BIT-46.9).

1. **Write test (RED):**
   - [ ] `TestRetroListHandler/lists every project's proposals` (`cmd/serve_mcp_retro_test.go`)
     - **Behavior:** a session in any project sees all proposals, labelled by project.
     - **Setup:** sandbox; `BIT` at `dir` and `EX` at `other`. `retro_write {name: "BIT-49-proposals"}` from `dir`, and `retro_write {name: "album-proposals"}` from `other`. Then `retro_list {}` from `dir`.
     - **Assertions:** `proposals == [{name: "BIT-49-proposals", project: "BIT"}, {name: "EX-album-proposals", project: "EX"}]`.
     - **Boundary:** another project's record is included, the opposite of `feedback_list`.
   - [ ] Confirm fails: the tool isn't registered.

2. **Implement (GREEN):**
   - [ ] `Proposal`, `ListRetro` and the tool.

3. **More tests (RED → GREEN):**
   - [ ] `TestRetroListHandler/no proposals lists none`: a fresh sandbox with one registered project (the tool resolves a project through `mcpStore`) and no proposals → `proposals == []`. *Boundary:* no `retro/` folder.
   - [ ] A row in `TestMCPToolDescriptions/carry the domain`: `retro_list` wants `testEveryProject`.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): retro_list lists every project's proposals`