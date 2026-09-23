---
id: BIT-44.2
title: Unknown track contradicts write-anywhere
status: todo
approved: true
phase: 1
phase_label: Analyze
---
## **Verse 1**

Research has to key to a real track. Bar 1 writes for any ID, so a mistyped track silently creates a stray folder. The failing test below is what forces the existence check.

## Scope
- `task/research.go`: `WriteResearch` refuses a track that doesn't exist.
- `cmd/serve_mcp_research_test.go`: tests.

## References
- `task/feedback.go`: `trackExists` checks active, completed, and archived tracks. Reuse it.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeMCPCmd_ResearchWriteRefusesAnUnknownTrack`
     - **Behavior:** writing research against a track that doesn't exist fails and leaves nothing on disk.
     - **Setup:** seed `testTrackID`; call `research_write` with `track: testUnknownTrackID`, `topic: "index"`, and any body.
     - **Assertions:** `result.IsError` is true, and `.bit/research/<testUnknownTrackID>` doesn't exist (`os.Stat` returns `ErrNotExist`).
     - **Boundary:** the track ID isn't in tasks/, completed/, or archive/.
   - [ ] Confirm it fails: `IsError` is false and the folder was created.

2. **Implement (GREEN):**
   - [ ] In `WriteResearch`, after `NormalizeID`, `if !s.trackExists(track) { return "", fmt.Errorf("track %s does not exist", track) }` before `MkdirAll`.

3. **More tests (these pin the Decision; expect them to pass on arrival):**
   - [ ] `TestServeMCPCmd_ResearchWriteAcceptsACompletedTrack`
     - **Behavior:** a finished track can still take research.
     - **Setup:** the track lives under `completed/`. Seed it the way `TestFeedbackAddCmd_AcceptsCompletedTrack` in `cmd/feedback_add_test.go` does. Then write topic `index`.
     - **Assertions:** not an error, and the file exists under `.bit/research/<track>/`.
     - **Boundary:** the track exists only in completed/.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- none (deterministic)

## Commit (user)
`feat(mcp): research_write refuses an unknown track`