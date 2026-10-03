---
id: BIT-50.5
title: task_complete writes the landing commit and branch on the track before filing it
status: done
approved: true
phase: 1
phase_label: complete after push
---
## **Verse 1**

Completion has to record the track's landing commit, but `task_update` can't reach a filed task (`Store.Update` loads from the active list only). So `task_complete` takes `commit` and `branch` and writes them before it moves anything. A test that reads the filed record forces this.

## Scope
- `cmd/serve_mcp.go`:
  - `taskCompleteInput` gains `Commit string \`json:"commit,omitempty"\`` and `Branch string \`json:"branch,omitempty"\``.
  - `taskCompleteHandler`: when either is non-empty, run `store.Update(id, task.Patch{Commit: &c, Branch: &b})` (set only the ones sent), then `store.Complete(id)`. A git-only patch keeps approval (BIT-46.14). With neither sent it only files, which covers the no-git path.
  - If `Complete` then refuses (unfinished bars), the fields stay written on the still-active track. That's harmless, because a retry writes them again.
  - `taskCompleteDescription`: replace the first line ("File a signed-off track…", or BIT-49.26's version) with: file a track and its bars as completed, and when `commit` and `branch` are given, write them on the track first, as its landing commit and the trunk it landed on, the values `task_landing` reports. Keep `testTrackSentence` and the "no override" sentence.
- `Store.Complete` is unchanged (scope).
- Test file: `cmd/serve_mcp_write_test.go`. BIT-46.9 already converted it, so there's no step 0.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTaskCompleteHandler/records commit and branch on the filed track`
     - **Behavior:** a completed track carries the commit it landed at.
     - **Setup:** `seedDoneTrack(t, dir, task.StatusDone)`. Call `task_complete {id: testTrackID, commit: "54abfeb5e1c3a7d2b9f0c4e6a8d1b3f5e7c9a0b2", branch: "main"}`.
     - **Assertions:** `readRecord(t, filepath.Join(pd, "completed", testTrackID+".md"))`, where `pd, _ := store.ProjectDir(testCode)` (BIT-46.1; `readRecord` is BIT-47.6's helper and reads the `.json` beside an `.md` path), gives `commit` equal to that SHA and `branch == "main"`. `testTrackID` is gone from `task_list`.
     - **Boundary:** both fields sent.
   - [ ] Confirm fails: the input schema rejects `commit`, or the filed record's `commit` is `""`.

2. **Implement (GREEN):**
   - [ ] Input fields, the write-then-file handler and the description.

3. **More tests (RED → GREEN):**
   - [ ] `TestTaskCompleteHandler/files a track and its bars` (existing): add the assertion that the filed record's `commit` and `branch` are `""`. *Boundary:* neither field sent, the no-git and CLI-style path.

## Claude verifies
- [ ] `just lint` and `just test` pass (`TestMCPToolDescriptions/carry the domain/task_complete` still finds `testTrackSentence`)

## User verifies
- none, deterministic

## Commit
`feat(bit): task_complete records the landing commit and branch before filing`