---
id: BIT-47.2
title: task_update records a bar's commit and branch and keeps its approval
status: todo
approved: true
phase: 1
phase_label: a bar lands as a real commit with its hash recorded
---
## **Verse 1**

`bit:commit` records a bar's hash with one `task_update {id, commit, branch, status: done}`. Today the input schema rejects `commit` (the SDK schema has `additionalProperties: false`). BIT-46.14 already gives `task.Patch` `Commit`/`Branch *string` and keeps them out of `Update`'s revocation check, so this bar only exposes them over MCP. An approved bar that must stay approved after the call, and must read back the hash, forces the mapping.

## Scope
- `cmd/serve_mcp.go`:
  - `taskUpdateInput` gains `Commit *string \`json:"commit,omitempty"\`` and `Branch *string \`json:"branch,omitempty"\``. `taskUpdateSchema` derives them automatically (only `status` gets an enum).
  - `taskUpdateHandler` maps both into `task.Patch{..., Commit: in.Commit, Branch: in.Branch}`.
  - `taskUpdateDescription`: after the sentence ending "the act of doing work that was already approved.", add: "Sending commit or branch keeps approval too: they record the full SHA and the branch a bar's approved work was committed as, not what was reviewed. An empty branch is written as empty, which is how a commit on a detached HEAD is recorded." Keep every guarded phrase (`testRevokingFields`, `testTodoRevokes`, `testForwardKeepsApproval`, `testNoCascade`, `testCallerRollsUp`) word for word.
- Test constants: `testGitKeepsApproval = "Sending commit or branch keeps approval"`, in the `const` block of `cmd/serve_mcp_test.go` beside `testForwardKeepsApproval`. Reuse BIT-47.1's `testCommitKey`, `testBranchKey` and `testBarSHA`.
- Test files: `cmd/serve_mcp_write_test.go` (`TestTaskUpdateHandler`) and `cmd/serve_mcp_test.go` (`TestMCPToolDescriptions/carry the domain`), both converted by BIT-46.9.
- No CLI flags (scope Decision). `cmd/task/update.go` is unchanged.
- **Seam for BIT-50:** `task_update {id, commit?, branch?}`. `commit` is a full 40-character SHA by convention (not validated), a sent value overwrites, an omitted one is left alone, and neither revokes approval. BIT-50 repoints bars through it before `task_complete`.

## TDD cycle

0. **Restructure:** none. Both files were already converted by BIT-46.9.

1. **Write test (RED):**
   - [ ] `TestTaskUpdateHandler/records commit and branch without revoking approval`
     - **Behavior:** closing a bar records the commit it became, and doing so never costs the bar its approval.
     - **Setup:** sandbox; seed track `testTrackID` and bar `testBarID` with `Status: doing, Approved: true`. Call `task_update {id: testBarID, commit: testBarSHA, branch: "v2", status: "done"}`, then `task_read {id: testBarID}`.
     - **Assertions:** the update returns `approved == true`. The read returns `status == "done"`, `commit == testBarSHA`, `branch == "v2"`, `approved == true`, and title, body and phase unchanged.
     - **Boundary:** a git-only patch plus a forward status move, on an approved bar.
   - [ ] Confirm fails: the call is refused for the unexpected property `commit`.

2. **Implement (GREEN):**
   - [ ] The two input fields and the `Patch` mapping.

3. **More tests (RED → GREEN):**
   - [ ] Row `commit and branch only leave the rest alone` in `TestTaskUpdateHandler/leaves omitted fields alone`: args `{id, commit: testBarSHA, branch: "v2"}`; want the seed with `Commit` and `Branch` set and every other field unchanged (the table compares whole `task.Task` values). *Boundary:* git fields alone, with no status.
   - [ ] Subtest `TestTaskUpdateHandler/an empty branch overwrites the stored branch`: seed `testBarID` with `Commit: "aaaa…"` (40 chars), `Branch: "v2"`; call `{id, commit: testBarSHA, branch: ""}`. The stored task has `Commit == testBarSHA` and `Branch == ""`. *Boundary:* a sent empty string writes `""`, unlike an omitted field. This is a follow-up commit made on a detached HEAD.
   - [ ] Row `task_update git keeps approval` in `TestMCPToolDescriptions/carry the domain`: `taskUpdateTool` wants `testGitKeepsApproval`. Add the description sentence.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic (the verse's end-to-end check is on BIT-47.4)

## Commit
`feat(bit): task_update records commit and branch without revoking approval`