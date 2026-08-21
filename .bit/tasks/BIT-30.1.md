---
id: BIT-30.1
title: A forward status move keeps its approval
status: doing
phase: 1
phase_label: Approved work stays approved
---
## **Verse 1**

Takes `--status` out of the revoke condition so an approved task keeps its flag while it moves through the columns. This is the entry point for the change — the first test is what forces it.

## Scope
- `cmd/task_update.go` — the `anyChanged` expression (lines 44–52): drop the `--status` disjunct
- `cmd/task_update_test.go` — new test

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTaskUpdateCmd_ForwardStatusMovePreservesApproval` — table-driven, subtests `todo → doing` and `doing → done`
     - **Behavior:** an approved task that advances a column is still approved afterwards, so the board keeps reporting it as reviewed for the whole run.
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "Old title", "...")` (lands in `todo`); for the `doing → done` case first `mustRun(t, taskCmdUse, updateCmd, trackID, "-s", "doing")`; then `mustRun(t, "approve", trackID)`; then the move under test — `mustRun(t, taskCmdUse, updateCmd, trackID, "-s", <to>)`.
     - **Assertions:** `task.New(".bit").Load(trackID)` returns `Approved == true` and `Status == <to>` (`"doing"` / `"done"`).
     - **Boundary:** the `--status` value at each forward edge of the three-state column range — `todo → doing` and `doing → done`, the two writes a normal run performs. `todo` as a target is Verse 2's case and is not exercised here.
   - [ ] Confirm fails: both subtests report `Approved = false`. `cmd.Flags().Changed("status")` is still one of the `anyChanged` disjuncts, so the status write wipes the approve that preceded it.

2. **Implement (GREEN):**
   - [ ] delete the `cmd.Flags().Changed("status") ||` line from the `anyChanged` expression in `cmd/task_update.go`, leaving `title`, `description`, `phase` and `phase-label`. Nothing else — `-s todo` losing its revoke as a result is Verse 2's problem, not a bug to pre-empt here.

## Claude verifies
- [ ] `just test` passes — in particular the three existing approval tests (`RevokesApprovalOnTitleChange`, `RevokesApprovalOnBodyChange`, `NoOpPreservesApproval`) stay green
- [ ] `just lint` passes

## User verifies
- [ ] Whole slice: after `just install`, `bp approve <id>` a `todo` task, then `bp task update <id> -s doing` and `-s done`. Open `bp tui`, press `→` for the board, and confirm the card is white in Doing and in Done — it never turns yellow mid-run.

## Commit (user)
`feat(task): keep approval through a forward status move`