---
id: BIT-49.4
title: feedback_read returns a note's body and refuses another project's note
status: todo
approved: true
phase: 1
phase_label: feedback and retro are shared across projects
---
## **Verse 1**

retro reads each listed note. Reading by ID would bypass `feedback_list`'s filter, so a second test reads another project's ID and expects a refusal. That forces the check on the record's `project`.

## Scope
- `task/feedback.go`:
  - `var ErrOtherProject = errors.New("note belongs to another project")`.
  - `func (s *Store) ReadNote(id string) (string, error)`. It rejects a path-like ID (`..`, `/`, `\`, the same as `resolveTrack`) and normalizes the ID. It reads `feedback/<ID>.json` (path built with `pathologize.Join`). If `rec.Project != s.code`, it returns `fmt.Errorf("%s: %w", id, ErrOtherProject)`. It then reads the `.md` body from `pathologize.Join(dir, content)`, as BIT-46.12 and BIT-46.15 do, with no separate check on `content`.
  - **Seam:** BIT-49.16's verify reuses it.
- `cmd/serve_mcp.go`: const `feedbackReadTool = "feedback_read"`; `feedbackReadInput{ID string \`json:"id"\`}`, `feedbackReadOutput{Body string \`json:"body"\`}`; handler; registration. The description says it reads one note of the current project by the ID `feedback_list` returns, and that another project's note is refused. It contains `testOwnProjectOnly`'s phrase and names no path.
- Test files touched: `cmd/serve_mcp_feedback_test.go` (created in shape by BIT-49.3) and `cmd/serve_mcp_test.go` (already converted by BIT-46.9).

## TDD cycle

0. **Restructure:** none. Both files are already in shape (BIT-49.3, BIT-46.9).

1. **Write test (RED):**
   - [ ] `TestFeedbackReadHandler/returns the note body` (`cmd/serve_mcp_feedback_test.go`)
     - **Behavior:** retro reads a note's verbatim text through the tool surface.
     - **Setup:** sandbox; `BIT` with track `BIT-1`; `feedback_add {track: "BIT-1", body: "## What the plan said\n\n> quoted exchange\n"}`; then `feedback_read {id: "BIT-1-001"}`.
     - **Assertions:** `body` equals the added body exactly.
     - **Boundary:** a note of the session's own project.
   - [ ] Confirm fails: the tool isn't registered.

2. **Implement (GREEN):**
   - [ ] `ReadNote`, plus the tool's const, types, description, handler and registration.

3. **More tests (RED → GREEN):**
   - [ ] `TestFeedbackReadHandler/refuses another project's note`: BIT-49.3's two-project seed; `feedback_read {id: "EX-1-001"}` from the `BIT` session → `IsError`, and the text contains `another project`. *Boundary:* an ID that exists in the shared folder but carries another project's code.
   - [ ] `TestFeedbackReadHandler/refuses an unknown or path-like id` (table): `BIT-1-099` and `../BIT-1-001` → `IsError`. *Boundary:* a missing record, and untrusted input.
   - [ ] A row in `TestMCPToolDescriptions/carry the domain`: `feedback_read` wants `testOwnProjectOnly`.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): feedback_read reads one of the current project's notes`