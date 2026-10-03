---
id: BIT-47.1
title: task_read and task_list return a task's commit and branch
status: todo
phase: 1
phase_label: a bar lands as a real commit with its hash recorded
---
## **Verse 1**

BIT-46.14 stores `commit` and `branch` on every track and bar, but no MCP tool returns them (`taskReadOutput`, `taskSummary` in `cmd/serve_mcp.go`). Without that, nobody can see that a bar holds its commit (scope Decision "`task_read` and `task_list` return `commit` and `branch`"). A seeded task with git fields that must come back, next to one without them that must come back as `""`, rules out a hardcoded value.

## Scope
- `cmd/serve_mcp.go` (anchor on names, BIT-46.9 rewrote the handlers):
  - `taskReadOutput` and `taskSummary` gain `Commit string \`json:"commit"\`` and `Branch string \`json:"branch"\``. No `omitempty`: an empty value is returned as `""`, the same as `parent`.
  - `taskReadHandler` and `taskListHandler` copy `t.Commit` / `t.Branch` (BIT-46.14's `Task` fields).
  - `taskReadDescription`: the field list "id, title, status, approved, phase, phase_label and parent" becomes "id, title, status, approved, phase, phase_label, parent, commit and branch". Keep `testTrackSentence` and `testBarIDExample` intact.
- Test constants, in the `const` block of `cmd/serve_mcp_test.go` beside `testParentKey`: `testCommitKey = "commit"`, `testBranchKey = "branch"`, `testBarSHA = "6a1d3459c0ffee000000000000000000000000ab"`. BIT-47.2 reuses them.
- Test file: `cmd/serve_mcp_test.go`, already converted by BIT-46.9 (`TestTaskReadHandler`, `TestTaskListHandler`). Seeds go through BIT-46.9's seed helpers (`seedTasks(t, dir, tasks...)`, which calls `registerProject` and `mcpSandbox`), which save through `project.OpenStore`, and `Save` writes `Branch`/`Commit` as given (BIT-46.14).
- **Seam for BIT-50:** `task_read` and `task_list` return `commit` and `branch` on every task.

## TDD cycle

0. **Restructure:** none. `cmd/serve_mcp_test.go` was already converted by BIT-46.9.

1. **Write test (RED):**
   - [ ] `TestTaskReadHandler/returns commit and branch` (a table subtest with two rows)
     - **Behavior:** an agent reading a bar sees which commit it became and on which branch.
     - **Setup:** sandbox and registered `dir` (BIT-46.9's `mcpSandbox` and `registerProject`, both called by `seedTasks`). Seed two bars under `testTrackID`: `testBarID` with `Status: done, Commit: testBarSHA, Branch: "v2"`, and `testSecondBarID` with no git fields. Call `task_read` on each.
     - **Assertions:** row `committed bar`: `commit == testBarSHA`, `branch == "v2"`. Row `uncommitted bar`: the keys `commit` and `branch` are present and equal `""`.
     - **Boundary:** git fields set vs empty. An empty value is returned, not omitted.
   - [ ] Confirm fails: the decoded map has no `commit` key.

2. **Implement (GREEN):**
   - [ ] Add the two fields to `taskReadOutput` and copy them in `taskReadHandler`.

3. **More tests (RED → GREEN):**
   - [ ] `TestTaskListHandler/returns commit and branch`
     - **Behavior:** a listing of a track's bars carries each bar's commit, so a caller can see every bar's hash in one call.
     - **Setup:** the same two seeded bars; `task_list {parent: testTrackID}`.
     - **Assertions:** the first entry has `commit == testBarSHA` and `branch == "v2"`. The second has both keys equal to `""`.
     - **Boundary:** mixed list, one bar with git fields and one without.
   - [ ] Add the fields to `taskSummary` and copy them in `taskListHandler`. Update `taskReadDescription`.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `grep -n 'phase_label, parent, commit and branch' cmd/serve_mcp.go` finds the read description

## User verifies
- none, deterministic

## Commit
`feat(bit): task_read and task_list return commit and branch`