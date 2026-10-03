---
id: BIT-47.7
title: feedback_add records the session's HEAD on a new note
status: done
approved: true
phase: 3
phase_label: records show the commits they were written at
---
## **Verse 3**

BIT-46.17's `AddNote(track, body, head)` (moved to the shared folder in BIT-49.1/49.2) takes a creating commit, and the MCP handler passes `task.Commit{}`. A stubbed repo whose SHA must appear on the note's `.json` forces the handler to pass `sessionHead`. Notes are create-only, so they only ever get the first entry.

## Scope
- `cmd/serve_mcp.go`: `feedbackAddHandler(root string, run git.Runner)` gets `dir` from `sessionDir(root)` and calls `store.AddNote(track, body, sessionHead(ctx, run, dir))`. `runMCPServer` passes `run`.
- Test file: `cmd/serve_mcp_write_test.go` (`TestFeedbackAddHandler`, converted by BIT-46.9; BIT-49.1 moved its feedback paths to the shared folder). It reuses BIT-47.6's `mcpSessionWithGit`, `fakeGit` and `readRecord`.

## TDD cycle

0. **Restructure:** none. Already converted by BIT-46.9.

1. **Write test (RED):**
   - [ ] `TestFeedbackAddHandler/records the session head`
     - **Behavior:** a feedback note shows the commit the work was at when the correction happened.
     - **Setup:** sandbox and registered `dir` with track `testTrackID`. The `fakeGit` from BIT-47.6 (`rev-parse HEAD` → `9f3c2b7e4d1a0c8b6e5f4a3b2c1d0e9f8a7b6c5d`, `symbolic-ref --short -q HEAD` → `worktree-bit-47`). `mcpSessionWithGit(t, dir, fake)`; `feedback_add {track: testTrackID, body: testNoteBody}`.
     - **Assertions:** `readRecord(path)["commits"]` is one entry with that `sha`, `branch == "worktree-bit-47"` and a parseable `at`. The `.md` still holds `testNoteBody` byte for byte.
     - **Boundary:** a new note, HEAD on a worktree branch.
   - [ ] Confirm fails: `commits` is `[]`.

2. **Implement (GREEN):**
   - [ ] Pass `run` into `feedbackAddHandler` and call `sessionHead`.

3. **More tests (RED → GREEN):**
   - [ ] `TestFeedbackAddHandler/no git records no commit`: `mcpSession(t, dir)` → the note is written and `commits == []`. *Boundary:* git errors never fail a write.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit
`feat(bit): feedback_add records the session's HEAD`