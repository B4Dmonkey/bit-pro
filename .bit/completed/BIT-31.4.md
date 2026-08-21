---
id: BIT-31.4
title: Contradiction forces play-prompt overlay in modeList
status: done
phase: 1
phase_label: play prompt
---
## **Verse 1**

`TestContent_PlayPromptRendersInModeList` contradicts the current `content()`: in modeList with `playPromptOpen=true`, the play prompt is absent from the output. Fixing it requires hoisting the `playPromptOpen` render path out of the `modeBoard` block.

## Scope
- `tui/model_test.go` — add `TestContent_PlayPromptRendersInModeList`
- `tui/model.go` — refactor `content()`: build a canvas per mode first, then apply `playPromptOpen` as a post-step on top of whichever canvas was built

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestContent_PlayPromptRendersInModeList`
     - **Behavior:** `content()` includes the play prompt string when `playPromptOpen` is true and mode is modeList
     - **Setup:** `m := New(nil); m.mode = modeList; m.playPromptOpen = true; m.playPromptTitle = "My Track"`
     - **Assertions:** `strings.Contains(m.content(), "Play My Track? (y / n)")`
     - **Boundary:** `modeList` — the branch `content()` currently exits without ever checking `playPromptOpen`
   - [ ] Confirm fails: `content()` takes the list-mode path which has no `playPromptOpen` branch; the prompt string is absent

2. **Implement (GREEN):**
   - [ ] Refactor `content()` in `tui/model.go`:
     - Declare `var canvas string`
     - In the `modeBoard` branch: build `b := boardView(m)`; apply `if m.modalOpen && !m.playPromptOpen { b = modalView(m, b) }`; set `canvas = b`
     - In the list-mode `else` branch: build the existing `listPane`/`detailPane` join; set `canvas = lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane)`
     - After the mode branch: `if m.playPromptOpen { return lipgloss.JoinVertical(lipgloss.Left, playPromptView(m, canvas), m.help.View(m.helpKeys())) }`
     - Final return: `return lipgloss.JoinVertical(lipgloss.Left, canvas, m.help.View(m.helpKeys()))`

## Claude verifies
- [ ] `go test ./tui/... -run TestContent_PlayPromptRendersInModeList` passes
- [ ] `just test` passes
- [ ] `just lint` passes

## User verifies
- [ ] Whole slice: In `bp tui`, switch to the list view (Tab). Approve the last unapproved bar on a track → "Play [track title]? (y / n)" overlay appears. Press y (or n or Esc) → overlay closes, list view is fully usable. Also confirm the existing board-mode flow still works (Tab back to board, re-approve → overlay appears there too).

## Commit (user)
`fix(tui): render play-prompt overlay in modeList`