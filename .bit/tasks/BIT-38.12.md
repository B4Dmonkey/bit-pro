---
id: BIT-38.12
title: feedback_add writes a note against a track
status: done
approved: true
phase: 3
phase_label: Run writes
---
## **Verse 3**

Completes bit_do's and bit_feedback's write paths. The status writes bit_do makes already ride on `task_update` from verse 1, so a note against a track is the one remaining write in a run. Forced by a protocol-level test: `feedback_add` is not registered. This is verse 3's only bar.

## Scope
- `cmd/serve_mcp.go` — `feedbackAddTool` const, `feedbackAddInput` / `feedbackAddOutput` types, `feedbackAddHandler`, registration.
- `cmd/serve_mcp_write_test.go` — the test.

## References
- `mcp-notes.md` — "Parity map", the `feedback_add` row: params `track` and `body` (both required), returning `{path}`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeMCPCmd_FeedbackAddWritesANote`
     - **Behavior:** a correction that landed mid-run can be recorded against its track through the protocol, and the caller learns where it went.
     - **Setup:** `dir := t.TempDir()`; `seedTasks(t, dir, &task.Task{ID: "FOO-1", Title: "MCP write surface", Status: task.StatusDoing})`; `s := mcpSession(t, dir)`. Call `feedback_add` twice against `FOO-1`, with multi-paragraph bodies — the first `"## What the plan said\n\nThe handler validates the anchor pair.\n\n## What the work required\n\nThe rule belongs in the store."`, the second any second note.
     - **Assertions:** the first call returns a `path` ending in `feedback/FOO-1-001.md`, the second `feedback/FOO-1-002.md`. Reading the first file off disk gives back the body byte-identically. The sequence increments, so the create-only write never overwrites an existing note.
     - **Boundary:** the existing-note count at 0 and then 1 — the sequence numbering is the thing that must not collide, and one call can never show it.
   - [ ] Confirm fails: `tool "feedback_add" not found`
   - [ ] `TestServeMCPCmd_FeedbackAddRefusesAnUnknownTrack` — `callToolErr` for `{"track": "FOO-9", "body": "..."}` against the same store. `IsError` is true and no file appears under `.bit/feedback/`. No new production code — `Store.AddNote` already refuses a track it cannot find in `tasks/`, `completed/`, or `archive/tasks/`; this pins that the handler surfaces the refusal instead of swallowing it.

2. **Implement (GREEN):**
   - [ ] `const feedbackAddTool = "feedback_add"`.
   - [ ] `type feedbackAddInput struct` — `Track string` (`json:"track"`) and `Body string` (`json:"body"`). Both required: a note with no body records nothing.
   - [ ] `feedbackAddOutput` — one field, `Path string`, tagged as JSON `path`.
   - [ ] `feedbackAddHandler(root string)`: `store.AddNote(in.Track, in.Body)`, wrap as `fmt.Errorf("adding note for %s: %w", in.Track, err)`, return `feedbackAddOutput{Path: path}`.
   - [ ] Register it, with a `Description` saying a note keys to a **track** and cites its bar in the prose, that the write is create-only so it can never damage an existing note, and that a completed or archived track is accepted.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] Whole slice: with the server registered, start a session, ask for a bar's status to be set to `doing` and then for a feedback note recording a correction against that track. Observe: `.bit/feedback/<TRACK>-001.md` exists with the note text, `bp task read <bar>` shows status `doing`, and the transcript shows `mcp__bit__task_update` / `mcp__bit__feedback_add` with no Bash write to `.bit/`.

## Commit (user)
`feat(mcp): add feedback_add for recording a note against a track`