---
id: BIT-23.6
title: Editing an approved task revokes its approval
status: done
approved: true
phase: 3
phase_label: Approval
---
## **Verse 3**

Approval revocation is a safety property, not a preference. Bars are approved in a batch before running unattended; if an edit could leave a previously-approved bar in place unchanged, unreviewed work could execute. Any `bp task update` that changes a field on an approved task clears the approval.

## Scope
- `cmd/task_update.go` — after loading the task and before saving: if `t.Approved && anyFieldChanged`, set `t.Approved = false`

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTaskUpdateCmd_RevokesApprovalOnTitleChange`
     - **Behavior:** updating the title of an approved task clears its Approved flag
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "Old title", "...")`; `mustRun(t, "approve", "BIT-1")`; `mustRun(t, "task", "update", "BIT-1", "--title", "New title")`
     - **Assertions:** loaded `BIT-1.Approved == false`
     - **Boundary:** a title change on an approved task — the core revocation case; proves update does not preserve approval
   - [ ] Confirm fails: current `task update` does not touch the Approved field; it remains true after the edit

2. **Implement (GREEN):**
   - [ ] In `newTaskUpdateCmd().RunE`, after all `Changed` checks, compute `anyChanged := cmd.Flags().Changed("title") || cmd.Flags().Changed("description") || cmd.Flags().Changed("status") || cmd.Flags().Changed("phase") || cmd.Flags().Changed("phase-label")`
   - [ ] If `t.Approved && anyChanged`, set `t.Approved = false`

3. **More tests (RED → GREEN):**
   - [ ] `TestTaskUpdateCmd_NoOpPreservesApproval` (contradiction)
     - **Behavior:** calling `task update` with no flags leaves Approved unchanged
     - **Setup:** approve BIT-1; `mustRun(t, "task", "update", "BIT-1")` (no flags)
     - **Assertions:** `BIT-1.Approved == true`
     - **Boundary:** no-op update — lower bound; contradicts clearing Approved on every update call; proves the guard is `anyChanged`
   - [ ] `TestTaskUpdateCmd_RevokesApprovalOnBodyChange`
     - **Behavior:** changing the description of an approved task clears approval
     - **Setup:** approve BIT-1; update with `--description "new body"`
     - **Assertions:** `BIT-1.Approved == false`
     - **Boundary:** description change — confirms revocation is not title-specific

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` clean

## User verifies
- [ ] none — deterministic; the automated tests cover all cases

## Commit (user)
`feat(task update): revoke approval when any field changes`