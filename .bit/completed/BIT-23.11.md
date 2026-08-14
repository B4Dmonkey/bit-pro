---
id: BIT-23.11
title: Focused list text uses approval color; cursor keeps green
status: done
approved: true
phase: 4
phase_label: TUI approval display
---
## **Verse 1**

The selected-list branch unconditionally applies `selectedStyle` (green + bold) to the row text, overriding the approval color. Split it: the cursor keeps `selectedStyle.Render("▎ ")` unchanged, and the row text applies the existing approval color logic (yellow if unapproved, default otherwise) plus bold.

## Scope
- `tui/delegate.go` — change the `selected` branch in `Render`: replace `main = selectedStyle` with `main = main.Bold(true)` + the existing `!t.Approved` yellow branch; leave `cursor = selectedStyle.Render("▎ ")` untouched; board branch unchanged
- `tui/delegate_test.go` — add two tests capturing new behavior; update `TestDelegate_SelectedUnapprovedItemUsesSelectedStyle` (its assertion that yellow is absent now contradicts the new behavior)

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestDelegate_SelectedUnapprovedItemKeepsYellowText`
     - **Behavior:** a selected unapproved list item renders yellow text — approval color survives focus
     - **Setup:** `list.New([]list.Item{item{t: &task.Task{ID: "BIT-1", Title: "T", Approved: false}}}, delegate{}, 40, 4)`; render index 0 (list default = index 0 selected)
     - **Assertions:** `strings.Contains(got, "33m")` — yellow SGR present in row output
     - **Boundary:** `Approved: false` + selected — the exact case where focus currently stomps approval color
   - [ ] Confirm fails: current code sets `main = selectedStyle` (green+bold), so "33m" is absent

2. **Implement (GREEN):**
   - [ ] In `Render`, replace the list selected branch:
     ```go
     // before
     if selected {
         main = selectedStyle
         if d.board {
             main = selectedBoardStyle
         }
     } else if !t.Approved {
         main = main.Foreground(lipgloss.Color("3"))
     }

     // after
     if selected {
         if d.board {
             main = selectedBoardStyle
         } else {
             main = main.Bold(true)
             if !t.Approved {
                 main = main.Foreground(lipgloss.Color("3"))
             }
         }
     } else if !t.Approved {
         main = main.Foreground(lipgloss.Color("3"))
     }
     ```
   - [ ] Leave `cursor = selectedStyle.Render("▎ ")` unchanged — cursor stays green

3. **More tests (RED → GREEN):**
   - [ ] `TestDelegate_SelectedApprovedItemCursorStaysGreen`
     - **Behavior:** a selected approved list item's row still contains the green cursor SGR
     - **Setup:** `list.New([]list.Item{item{t: &task.Task{ID: "BIT-1", Title: "T", Approved: true}}}, delegate{}, 40, 4)`; render index 0
     - **Assertions:** `strings.Contains(got, "32m")` — green still present (cursor); `!strings.Contains(got, "33m")` — no yellow (approved)
     - **Boundary:** `Approved: true` + selected — approved focused items keep green cursor, no yellow
   - [ ] Update `TestDelegate_SelectedUnapprovedItemUsesSelectedStyle`: remove (or flip) the assertion `strings.Contains(got, "33m")` → should expect NOT to error now that yellow is correct. The test's "32m present" assertion still holds (cursor green). Rename it to `TestDelegate_SelectedUnapprovedItemCursorStaysGreen` to document what it actually proves.

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` clean

## User verifies
- [ ] `bp tui` → list view: focus an unapproved item (yellow when unfocused) — it stays yellow when focused, with a green `▎` on the left; move focus to an approved item — text is not yellow

## Commit (user)
`fix(tui): focused list item keeps approval color; cursor alone stays green`
