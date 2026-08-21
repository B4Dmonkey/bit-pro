---
id: BIT-30.2
title: A task sent back to todo needs re-approval
status: done
approved: true
phase: 2
phase_label: Sent back, re-reviewed
---
## **Verse 2**

Restores the revoke for a send-back only. The contradiction is within one table: `-s todo` must revoke while `-s doing` must not, so neither "status always revokes" (before Verse 1) nor "status never revokes" (after Verse 1) can satisfy it — the condition has to key on the target value.

## Scope
- `cmd/task_update.go` — add a `--status`-to-`todo` disjunct alongside `anyChanged`
- `cmd/task_update_test.go` — new test

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTaskUpdateCmd_StatusToTodoRevokesApproval` — table-driven, subtests `doing → todo` (revokes), `done → todo` (revokes), `todo → todo` (revokes), `done → doing` (preserves)
     - **Behavior:** a task written to `todo` goes back behind the approval gate, while a write to any other column leaves approval alone — so the flag means "cleared to run", not "has been touched".
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "Old title", "...")`; move to the case's starting status with `mustRun(t, taskCmdUse, updateCmd, trackID, "-s", <from>)` (skip when `<from>` is `todo`); then `mustRun(t, "approve", trackID)` — the approve must come after the setup status write, since that write would otherwise clear it; then the move under test, `-s <to>`.
     - **Assertions:** the reloaded task's `Approved` equals the case's `want` (`false`, `false`, `false`, `true`) and `Status == <to>`.
     - **Boundary:** the `--status` value at the one column that gates running (`todo`) against one that doesn't (`doing`), including the degenerate `todo → todo` where the value doesn't change — revocation keys on the target value, not on movement.
   - [ ] Confirm fails: the three `→ todo` subtests report `Approved = true` (after Verse 1 no status write revokes anything); `done → doing` already passes and must stay passing.

2. **Implement (GREEN):**
   - [ ] in `cmd/task_update.go`, add `sentBack := cmd.Flags().Changed("status") && status == task.StatusTodo` (the `task` package is already imported) and revoke on `t.Approved && (anyChanged || sentBack)`. Read the flag variable `status`, not `t.Status`, so the condition doesn't depend on where the assignment sits.

## Claude verifies
- [ ] `just test` passes, including Verse 1's `ForwardStatusMovePreservesApproval`
- [ ] `just lint` passes

## User verifies
- [ ] Whole slice: after `just install`, `bp approve <id>` a task and `bp task update <id> -s done` — in `bp tui`'s board (`→`) the card is white in Done. Then `bp task update <id> -s todo` and reload: the card leaves the board entirely rather than reappearing yellow in To Do, and `bp task list` shows it as `todo` with an empty approved column.

## Commit (user)
`feat(task): revoke approval when a task is sent back to todo`