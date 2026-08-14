---
id: BIT-23.5
title: bp task list shows approved state
status: done
approved: true
phase: 3
phase_label: Approval
---
## **Verse 3**

`bp task list` is the first place to answer "what can be worked on?" without opening the TUI. Adding approval state to its output makes that question answerable at a glance from the terminal.

## Scope
- `cmd/task_list.go` — add a 5th tab-separated field between `title` and `phase`: `"approved"` when `t.Approved == true`, else `""`. Update the format string accordingly.
- `cmd/task_list_test.go` — update all tests that compare exact output strings: add the new empty field to each expected string.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTaskListCmd_ShowsApprovedMarker`
     - **Behavior:** `bp task list` shows `approved` in the output for an approved task
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "Track", "...")`; `mustRun(t, "approve", "BIT-1")`
     - **Assertions:** `mustRun(t, "task", "list")` output contains `BIT-1\ttodo\tTrack\tapproved\t`
     - **Boundary:** approved task — the positive state; proves the field appears in output
   - [ ] Confirm fails: list output has 4 fields (current format); no `approved` column
   - [ ] `TestTaskListCmd_UnapprovedShowsEmptyField` (contradiction)
     - **Behavior:** `bp task list` for an unapproved task has an empty approval column
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "Track", "...")`
     - **Assertions:** output contains `BIT-1\ttodo\tTrack\t\t`
     - **Boundary:** unapproved (default) — lower bound; contradicts hardcoding "approved" always; proves the field is conditional

2. **Implement (GREEN):**
   - [ ] In `newTaskListCmd()`, compute `approved := ""` then `if t.Approved { approved = "approved" }`
   - [ ] Change format string from `"%s\t%s\t%s\t%s\n"` to `"%s\t%s\t%s\t%s\t%s\n"` with args `t.ID, t.Status, t.Title, approved, phase`
   - [ ] Update existing test expectations in `task_list_test.go` to include the new empty `\t` for the approval column (e.g. `"BIT-1\ttodo\tFirst\t\t"`)

## Claude verifies
- [ ] `just test` passes (including updated existing list tests)
- [ ] `just lint` clean

## User verifies
- [ ] `bp approve BIT-23 && bp task list | grep BIT-23` — shows `BIT-23\ttodo\t...\tapproved\t` in the output

## Commit (user)
`feat(task list): show approved state as a column in bp task list`