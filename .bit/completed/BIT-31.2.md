---
id: BIT-31.2
title: Contradiction forces all-approved check before firing play prompt
status: done
approved: true
phase: 1
phase_label: play prompt
---
## **Verse 1**

The hardcoded `playPromptOpen = true` from the previous bar can't satisfy the case where another bar is still unapproved. Contradiction forces the real all-approved check. After this bar, `handleReloaded` reads the post-reload task list, finds the parent track's bars, and fires only when ≥1 bar exists and none are unapproved.

## Scope
- `tui/model.go` — replace the hardcoded `m.playPromptOpen = true` in `handleReloaded` with: compute `parentID = strings.SplitN(m.pendingApprovalID, ".", 2)[0]`; collect bars of that parent from `msg.tasks`; fire only when `len(bars) >= 1 && allApproved(bars)`; add package-level helpers `barChildrenOf(parentID string, tasks []*task.Task) []*task.Task` and `allApproved(tasks []*task.Task) bool`
- `tui/model_test.go` — add `TestUpdate_PartialApprovalSkipsPlayPrompt`, `TestUpdate_ZeroBarTrackSkipsPlayPrompt`, `TestUpdate_ReapprovalRefiresPlayPrompt`

## TDD cycle

1. **Write test (RED — contradiction):**
   - [ ] `TestUpdate_PartialApprovalSkipsPlayPrompt`
     - **Behavior:** When a bar is approved but a sibling bar on the same track remains unapproved, `playPromptOpen` stays false
     - **Setup:** `New([]*task.Task{{ID: ttid1}, {ID: ttid1_1, Approved: false}, {ID: ttid1_2, Approved: false}}).WithApprove(...)` — Tab to list; move selection to ttid1_1; space; then `reloadedMsg{tasks: []*task.Task{{ID: ttid1}, {ID: ttid1_1, Approved: true}, {ID: ttid1_2, Approved: false}}}`
     - **Assertions:** `updated.(model).playPromptOpen == false`
     - **Boundary:** `ttid1_2` count = 1 unapproved sibling — the minimal "not ready" state; hardcode returns true here, breaking it
   - [ ] Confirm fails: the hardcode sets `playPromptOpen = true` regardless, so the assertion fails

2. **Implement (GREEN):**
   - [ ] Add `func barChildrenOf(parentID string, tasks []*task.Task) []*task.Task` — returns tasks whose ID starts with `parentID + "."`
   - [ ] Add `func allApproved(tasks []*task.Task) bool` — returns true when every task in the slice has `Approved == true`
   - [ ] Replace hardcode in `handleReloaded` with: `parentID := strings.SplitN(m.pendingApprovalID, ".", 2)[0]; bars := barChildrenOf(parentID, msg.tasks); if len(bars) >= 1 && allApproved(bars) { m.playPromptOpen = true }; m.pendingApprovalID = ""`

3. **More tests (RED → GREEN):**
   - [ ] `TestUpdate_ZeroBarTrackSkipsPlayPrompt`
     - **Behavior:** Approving a track ID (no dot) does not fire the play prompt — the `isBar` guard in `handleApprove` prevents `pendingApprovalID` from being set
     - **Setup:** `New([]*task.Task{{ID: ttid1, Approved: false}}).WithApprove(...)` — Tab; space; `reloadedMsg{tasks: []*task.Task{{ID: ttid1, Approved: true}}}`
     - **Assertions:** `updated.(model).playPromptOpen == false`
     - **Boundary:** `ttid1` = "BIT-1" contains no "." → `isBar` returns false → `pendingApprovalID` never set — the lower bound of the isBar check
   - [ ] `TestUpdate_ReapprovalRefiresPlayPrompt`
     - **Behavior:** After unapproving and re-approving the last bar, `playPromptOpen` becomes true again — the condition is stateless
     - **Setup:** Start with track + 1 approved bar; unapprove the bar (space + reload showing unapproved); then re-approve (space + reload showing approved again)
     - **Assertions:** `updated.(model).playPromptOpen == true` on the second reload
     - **Boundary:** Exercises the "unapproved then approved" sequence — proves no toggle-state memory in the condition

## Claude verifies
- [ ] `go test ./tui/... -run TestUpdate_Partial` passes
- [ ] `go test ./tui/... -run TestUpdate_ZeroBar` passes
- [ ] `go test ./tui/... -run TestUpdate_Reapproval` passes
- [ ] `just test` passes
- [ ] `just lint` passes

## User verifies
- [ ] none — deterministic

## Commit (user)
`feat(tui): check all bars approved before firing play prompt`