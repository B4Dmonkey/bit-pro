---
id: BIT-14.5
title: Unfocused titles gain a framed look
status: done
phase: 2
phase_label: Focus
---
## **Verse 2**

Frames an unfocused title inline in its border as `| Title |`, a calm, colorless "not focused" signal that pairs with the inverted focused block. Applies uniformly through the shared `titledBorder`, so both the unfocused list pane and the unfocused board columns get the frame.

## Scope
- `tui/model.go` — `titledBorder`: in the inactive branch, build the title segment as `border.Top + "| " + title + " |"` so the top line reads `╭─| Title |─…╮` instead of the current `╭─ Title ─…╮`. No color change; the title words are unchanged.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTitledBorder_InactiveTitleFramed`
     - **Behavior:** an unfocused pane/column frames its title inline as `| Title |`.
     - **Setup:** `got := titledBorder("body", "Doing (0)", 20, 3, false)`.
     - **Assertions:** `strings.Contains(got, "| Doing (0) |")` is true.
     - **Boundary:** `active == false` — the unfocused state; focused titles are inverted, never framed.
   - [ ] Confirm fails: the inactive title currently renders `─ Doing (0) ─` (spaces, no pipes).

2. **Implement (GREEN):**
   - [ ] Frame the inactive title segment with `| ` and ` |`.

3. **More tests (RED → GREEN):**
   - [ ] `TestTitledBorder_ActiveTitleNotFramed`
     - **Behavior:** a focused title uses the inverted block, not the pipe frame — the two states stay visually distinct.
     - **Setup:** `got := titledBorder("body", "Tasks (0)", 20, 3, true)`.
     - **Assertions:** `strings.Contains(got, "| Tasks (0) |")` is false.
     - **Boundary:** `active == true` — confirms the frame is exclusive to the inactive branch.
   - [ ] Confirm passes (the frame is only added in the inactive branch).

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(tui): unfocused titles gain a framed look`
