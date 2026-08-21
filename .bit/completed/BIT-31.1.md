---
id: BIT-31.1
title: Hardcoded play-prompt fires on any bar approval
status: done
approved: true
phase: 1
phase_label: play prompt
---
## **Verse 1**

Adds the `playPromptOpen` and `pendingApprovalID` fields to the model and wires the approval paths to set them. `handleReloaded` fires `playPromptOpen = true` whenever `pendingApprovalID` is set — hardcoded, not yet checking sibling approval state. The first test proves the flag goes true after approving a bar.

## Scope
- `tui/model.go` — add `playPromptOpen bool` and `pendingApprovalID string` to `model` struct; in `handleApprove`, set `pendingApprovalID = t.ID` when `!t.Approved && isBar(t.ID)`; in `handleReloaded`, after `m.setTasks(msg.tasks)`, if `m.pendingApprovalID != ""` set `m.playPromptOpen = true` and clear `m.pendingApprovalID`
- `tui/board.go` — in `updateBoard` space handler, set `m.pendingApprovalID = t.ID` when `!t.Approved && isBar(t.ID)` (before the existing `m.approve(...)` call)
- `tui/model_test.go` — add `TestUpdate_BarApprovalSetsPlayPromptOpen`

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestUpdate_BarApprovalSetsPlayPromptOpen`
     - **Behavior:** Approving the only bar on a track (in list mode) sets `playPromptOpen` to true after the subsequent reload
     - **Setup:** `New([]*task.Task{{ID: ttid1}, {ID: ttid1_1, Approved: false}}).WithApprove(func(_, _ string, _ bool) error { return nil })` — Tab to enter list mode (default is board); select the second item (ttid1_1); press space; then send `reloadedMsg{tasks: []*task.Task{{ID: ttid1}, {ID: ttid1_1, Approved: true}}}`
     - **Assertions:** `updated.(model).playPromptOpen == true`
     - **Boundary:** `ttid1_1` = "BIT-1.1" contains "." → `isBar` returns true; `Approved: false` before the press → toggling to approved; exactly 1 bar on the track — the minimal case that should fire
   - [ ] Confirm fails: `model` struct has no `playPromptOpen` field — compile error

2. **Implement (GREEN):**
   - [ ] Add `playPromptOpen bool` and `pendingApprovalID string` to `model` struct (after `modalOpen`)
   - [ ] In `handleApprove`: before `_ = m.approve(t.ID, !t.Approved)`, add `if !t.Approved && isBar(t.ID) { m.pendingApprovalID = t.ID }`
   - [ ] In `updateBoard` space handler: before `_ = m.approve(t.ID, !t.Approved)`, add `if !t.Approved && isBar(t.ID) { m.pendingApprovalID = t.ID }`
   - [ ] In `handleReloaded`, after `m.setTasks(msg.tasks)` and before `return m, tick()`: add `if m.pendingApprovalID != "" { m.playPromptOpen = true; m.pendingApprovalID = "" }`

## Claude verifies
- [ ] `go test ./tui/... -run TestUpdate_BarApprovalSetsPlayPromptOpen` passes
- [ ] `just test` passes
- [ ] `just lint` passes

## User verifies
- [ ] none — deterministic

## Commit (user)
`feat(tui): set playPromptOpen flag on bar approval reload (hardcoded)`