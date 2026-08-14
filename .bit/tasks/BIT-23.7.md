---
id: BIT-23.7
title: Unapproved items render yellow in the TUI
status: todo
phase: 4
phase_label: TUI approval display
---
## **Verse 4**

The approval state is now in the model, but the TUI renders all items identically regardless of it. Unapproved items need a visible warning signal — yellow — so the board communicates "needs a look" at a glance without the user reading each task's details.

## Scope
- `tui/delegate.go` — add `unapprovedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))` (terminal yellow); in `Render`, when the item is not selected and `!t.Approved`, apply `unapprovedStyle` to the rendered row text instead of `main`
- `tui/model.go` — add `x.Approved == y.Approved` to the `sameTasks` equality check so a remote approval change triggers a reload

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestDelegate_UnapprovedItemRendersYellow`
     - **Behavior:** an unapproved, unselected item renders with the yellow SGR color code
     - **Setup:** `l := list.New([]list.Item{item{t: &task.Task{ID: "BIT-1", Title: "Track", Approved: false}}}, delegate{}, 40, 4)`; select a different index so this item is unselected
     - **Assertions:** rendered string contains `33m` (terminal yellow SGR code)
     - **Boundary:** `Approved == false` on an unselected item — the unapproved case; proves the yellow style is applied
   - [ ] Confirm fails: current `Render` always uses `main` (bold or plain); no yellow code present
   - [ ] `TestDelegate_ApprovedItemDoesNotRenderYellow` (contradiction)
     - **Behavior:** an approved, unselected item does not render yellow
     - **Setup:** same but `Approved: true`
     - **Assertions:** rendered string does not contain `33m`
     - **Boundary:** `Approved == true` — contradicts always applying yellow; forces the conditional check

2. **Implement (GREEN):**
   - [ ] Add `unapprovedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))` to delegate vars in `tui/delegate.go`
   - [ ] In `Render`, after computing `main` style: `if !selected && !t.Approved { main = unapprovedStyle }`
   - [ ] In `tui/model.go`'s `sameTasks`, add `x.Approved == y.Approved` to the equality conjunction

3. **More tests (RED → GREEN):**
   - [ ] `TestDelegate_SelectedUnapprovedItemUsesSelectedStyle`
     - **Behavior:** a selected item that happens to be unapproved still uses the selected (green) style, not yellow
     - **Setup:** selected index 0; `Approved: false`
     - **Assertions:** contains `32m` (green); does not contain `33m` (yellow)
     - **Boundary:** selected wins over unapproved — proves the selection check has higher precedence

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` clean

## User verifies
- [ ] `bp tui` — unapproved items appear in yellow in both the list view and board view; items that have been approved via `bp approve` appear in white

## Commit (user)
`feat(tui): render unapproved items in yellow`