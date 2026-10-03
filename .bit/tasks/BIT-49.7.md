---
id: BIT-49.7
title: retro_read returns any project's proposal by name
status: todo
approved: true
phase: 1
phase_label: feedback and retro are shared across projects
---
## **Verse 1**

learn opens the proposals `retro_list` shows. A test that reads a record another project wrote forces a read with no project filter. Proposals are generalized before they're written, so they carry no confidentiality line.

## Scope
- `task/retro.go`: `func (s *Store) ReadRetro(name string) (Proposal, string, error)`. It validates `name` like `WriteRetro` but applies no prefix: the caller passes the name exactly as `retro_list` shows it. It reads `retro/<name>.json` (via `pathologize.Join`), reads the body from `pathologize.Join(dir, content)` as BIT-46.12 and BIT-46.15 do, with no separate check on `content`, and returns the record's name and project plus the body. A missing record errors. **Seam:** BIT-49.16's verify reuses it.
- `cmd/serve_mcp.go`: const `retroReadTool = "retro_read"`; `retroReadInput{Name string}`; `retroReadOutput{Name, Project, Body string}`. The description says it reads one proposals record by the name `retro_list` shows, from any project. It names no path and reuses `testEveryProject`'s phrase.
- Test files touched: `cmd/serve_mcp_retro_test.go` (created in shape by BIT-49.5) and `cmd/serve_mcp_test.go` (already converted by BIT-46.9).

## TDD cycle

0. **Restructure:** none. Both files are already in shape (BIT-49.5, BIT-46.9).

1. **Write test (RED):**
   - [ ] `TestRetroReadHandler/reads another project's proposal` (`cmd/serve_mcp_retro_test.go`)
     - **Behavior:** learn, running in bit-pro, opens a proposal that retro wrote in another project.
     - **Setup:** sandbox; `EX` at `other` writes `retro_write {name: "album-proposals", body: "## Proposal 1\n..."}`. From a `BIT` session at `dir`, call `retro_read {name: "EX-album-proposals"}`.
     - **Assertions:** `name == "EX-album-proposals"`, `project == "EX"`, and `body` equals the written body.
     - **Boundary:** the reader's project differs from the record's.
   - [ ] Confirm fails: the tool isn't registered.

2. **Implement (GREEN):**
   - [ ] `ReadRetro` and the tool.

3. **More tests (RED → GREEN):**
   - [ ] `TestRetroReadHandler/refuses an unknown or path-like name` (table): `"BIT-none-proposals"` and `"../x"` → `IsError`. *Boundary:* a missing record, and untrusted input.
   - [ ] A row in `TestMCPToolDescriptions/carry the domain`: `retro_read` wants `testEveryProject`.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): retro_read reads any project's proposal`