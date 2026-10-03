---
id: BIT-49.3
title: feedback_list returns only the current project's notes
status: todo
approved: true
phase: 1
phase_label: feedback and retro are shared across projects
---
## **Verse 1**

retro needs to list notes without reading files. The shared folder holds every project's notes, so a test with two projects' notes in it can't pass unless the list filters on the record's `project`. That's the confidentiality line.

## Scope
- `task/feedback.go`: `func (s *Store) ListNotes(track string) ([]string, error)`.
  - It reads every `feedback/*.json` under the data root, unmarshals `noteRecord`, and keeps `rec.Project == s.code`.
  - When `track != ""`, it resolves the track through `resolveTrack`, so an unknown or path-like track errors, and keeps `rec.Track == track`.
  - It returns IDs sorted by track (`compareIDs`), then by seq.
  - A missing `feedback/` folder gives an empty list, not an error.
  - **Seam:** BIT-49.16's verify reuses it.
- `cmd/serve_mcp.go`:
  - const `feedbackListTool = "feedback_list"`.
  - `feedbackListInput{Track string \`json:"track,omitempty"\`}` and `feedbackListOutput{Notes []string \`json:"notes"\`}` (always `[]`, never `null`).
  - `feedbackListHandler(root)` uses `mcpStore(ctx, root)` (BIT-46.9) and is registered after `feedback_add`.
  - `feedbackListDescription`: lists the note IDs of the current project, optionally for one track. It says that notes from other projects are never listed, and it names no path (written without `.bit/` so Verse 3 doesn't revisit it).
- `cmd/testconst_test.go`: `testOwnProjectOnly = "only the current project's notes"`, which the description must contain.
- Test files touched: new `cmd/serve_mcp_feedback_test.go` (created in shape here), and `cmd/serve_mcp_test.go` (the tool-description table), already converted by BIT-46.9.

## TDD cycle

0. **Restructure:** none. `cmd/serve_mcp_test.go` was converted by BIT-46.9 (`TestRunMCPServer`, `TestTaskReadHandler`, `TestTaskListHandler`, and the description table as `TestMCPToolDescriptions/carry the domain`).

1. **Write test (RED):**
   - [ ] `TestFeedbackListHandler/lists only this project's notes` (`cmd/serve_mcp_feedback_test.go`)
     - **Behavior:** retro in one project sees its own notes and never another project's, even though they share a folder.
     - **Setup:** MCP sandbox (`mcpSandbox(t)`, BIT-46.9). `dir` registered as `BIT` (`seedProject(t, orm.CreateProjectParams{Path: <canonical dir>, Code: "BIT"})`; not `registerProject`, which registers `testCode`, `FOO`) with tracks `BIT-1` and `BIT-2`. `other` registered as `EX` (`seedProject(t, orm.CreateProjectParams{Path: <canonical other>, Code: "EX"})`, then `project.OpenStore(ctx, other)`) with track `EX-1`. Through `feedback_add`: two notes on `BIT-1`, one on `BIT-2` from a `dir` session, and one on `EX-1` from an `other` session. Then `feedback_list {}` from the `dir` session.
     - **Assertions:** `notes == ["BIT-1-001", "BIT-1-002", "BIT-2-001"]`.
     - **Boundary:** a second project's note sits in the shared folder.
   - [ ] Confirm fails: `feedback_list` isn't a registered tool.

2. **Implement (GREEN):**
   - [ ] `ListNotes`, plus the tool's const, types, description, handler and `mcp.AddTool`.

3. **More tests (RED → GREEN):**
   - [ ] `TestFeedbackListHandler/filters to one track`: the same seed with `{track: "bit-1"}` → `["BIT-1-001", "BIT-1-002"]`. *Boundary:* a lowercase track ID is normalized.
   - [ ] `TestFeedbackListHandler/a project with no notes lists none`: a fresh registered project → `notes == []`. *Boundary:* no `feedback/` folder yet.
   - [ ] `TestFeedbackListHandler/refuses an unknown track`: `{track: "BIT-9"}` → `IsError`. *Boundary:* the track filter goes through `resolveTrack`.
   - [ ] A row in `TestMCPToolDescriptions/carry the domain`: `feedback_list` wants `testOwnProjectOnly`.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] the `feedback_list` description contains no `.bit`

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): feedback_list lists the current project's notes`