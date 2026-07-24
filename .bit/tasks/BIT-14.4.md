---
id: BIT-14.4
title: Focused title renders as an inverted block
status: done
phase: 2
phase_label: Focus
---
## **Verse 2**

Renders a focused title as a reverse-video green block (dark text on green) instead of plain green text, so focus is unmistakable. Because `titledBorder` is called with `active == true` for the focused list pane, the focused board column, and the open modal, this one change makes all three read as focus blocks.

## Scope
- `tui/model.go` — `titledBorder`: in the active branch, style the title span (`" " + title + " "`) with reverse-video green (`Foreground(Color("2")).Reverse(true)`) while the surrounding border characters keep plain green. This means splitting the current single `topStyle.Render(top)` into a green-rendered border and a separately reverse-rendered title span; the title text (the words) is unchanged.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTitledBorder_ActiveTitleInverted`
     - **Behavior:** a focused pane/column/modal renders its title as a reverse-video green block while its border stays green.
     - **Setup:** `got := titledBorder("body", "Tasks (0)", 20, 3, true)`.
     - **Assertions:** `strings.Contains(got, "\x1b[7;32m")` is true (reverse+green wrapping the title); `strings.Contains(got, "\x1b[32m")` is still true (border green). (Probe reference: reverse-green renders `\x1b[7;32m`.)
     - **Boundary:** `active == true` — the focused state; inactive titles are never inverted.
   - [ ] Confirm fails: after Bar 1.1 the active title is plain green (no `7;`), so the `\x1b[7;32m` assertion fails.

2. **Implement (GREEN):**
   - [ ] Style the active title span with `.Reverse(true)` on the green foreground, separately from the border characters.

3. **More tests (RED → GREEN):**
   - [ ] `TestView_ModalTitleInverted`
     - **Behavior:** an opened card modal shows its title as a reverse-video green block (the modal uses the active `titledBorder`).
     - **Setup:** `New([]*task.Task{{ID: "BIT-1", Status: "todo", Title: "T", Body: "b"}})`; WindowSize 80x24; Tab to board; Enter to open the modal; `view := View().Content`.
     - **Assertions:** `strings.Contains(view, "\x1b[7;32m")` is true.
     - **Boundary:** modal open — the composited overlay path, confirming it inherits the active title treatment.
   - [ ] Confirm passes once the active title is inverted.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic (the focus verse's observe-it check lands on its last bar).

## Commit (user)
`feat(tui): focused title renders as an inverted block`
