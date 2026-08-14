---
id: BIT-23.4
title: Approved field; bp approve and bp unapprove set and clear it
status: todo
phase: 3
phase_label: Approval
---
## **Verse 3**

Approval is the mandatory gate before any work proceeds. A track approved means its scope is blessed; all bars approved means the plan is blessed and work may start. Without an approve command, the gate exists only in prose. This bar writes the field and exposes the CLI commands.

## Scope
- `task/task.go` — add `Approved bool \`yaml:"approved,omitempty"\`` to `Task`
- `task/store.go` — add `func (s *Store) SetApproved(id string, approved bool) error` (load, set field, save); reused by the command and the TUI write-back in Bar 8
- `cmd/approve.go` — new file; `bp approve <ID>` and `bp unapprove <ID>` commands (both route to `s.SetApproved`)
- `cmd/root.go` — add `rootCmd.AddCommand(newApproveCmd())`

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestApproveCmd_SetsApprovedTrue`
     - **Behavior:** `bp approve BIT-1` sets `Approved = true` on the task
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "Track", "...")`
     - **Assertions:** `mustRun(t, "approve", "BIT-1")`; loaded `BIT-1.Approved == true`
     - **Boundary:** unapproved → approved — the initial state is false; proves the flag is stored
   - [ ] Confirm fails: `bp approve` command does not exist

2. **Implement (GREEN):**
   - [ ] Add `Approved bool \`yaml:"approved,omitempty"\`` to `Task` struct
   - [ ] Add `SetApproved` to Store: load task, set `t.Approved = approved`, call `s.Save(t)`
   - [ ] Create `cmd/approve.go` with two commands: `bp approve <ID>` calls `s.SetApproved(id, true)`; `bp unapprove <ID>` calls `s.SetApproved(id, false)`
   - [ ] Wire both into root

3. **More tests (RED → GREEN):**
   - [ ] `TestApproveCmd_UnapproveClears` (contradiction)
     - **Behavior:** `bp unapprove BIT-1` clears the flag after it was set
     - **Setup:** approve then unapprove
     - **Assertions:** `BIT-1.Approved == false`
     - **Boundary:** approved → unapproved — contradicts "always set true"; forces the bool to be written correctly for both commands
   - [ ] `TestApproveCmd_ErrorsOnUnknownID`
     - **Behavior:** approving a non-existent ID returns an error
     - **Setup:** `initProject(t, "BIT")` with no tasks; run `approve BIT-99`
     - **Assertions:** `run(t, "approve", "BIT-99")` returns non-nil error
     - **Boundary:** unknown ID — proves the load guard

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` clean

## User verifies
- [ ] `bp approve BIT-23 && grep approved .bit/tasks/BIT-23.md` — shows `approved: true` in the frontmatter; `bp unapprove BIT-23 && grep approved .bit/tasks/BIT-23.md` — field absent (omitempty removes it when false)

## Commit (user)
`feat(approve): add bp approve / unapprove commands and Approved field`