---
id: BIT-31.3
title: y / n / Esc dismiss the play prompt; board input blocked while open
status: todo
phase: 1
phase_label: play prompt
---
## **Verse 1**

Key routing and rendering for the play prompt. When `playPromptOpen` is true, y/n/Esc dismiss it and all other board input is blocked. The model also gains `playPromptTitle string` (set at fire-time from the parent track's title) so the overlay can render "Play [track title]? (y / n)".

## Scope
- `tui/model.go` — add `playPromptTitle string` to `model` struct; set it in `handleReloaded` when firing the prompt (look up parent track from `msg.tasks`); add `func (m model) handlePlayPrompt(msg tea.KeyPressMsg) (tea.Model, tea.Cmd)` — handles y/n/Esc → clear `playPromptOpen`, ctrl+c → quit, anything else → no-op; in `handleKeyPress`, add `if m.playPromptOpen { return m.handlePlayPrompt(msg) }` as the first branch (before the `modalOpen` check)
- `tui/board.go` — add `func playPromptView(m model, board string) string` rendering "Play [title]? (y / n)" as a centred lipgloss overlay (same compositor pattern as `modalView`); in `content()`, add `if m.playPromptOpen { return playPromptView(m, board) }` before the existing `modalOpen` branch
- `tui/model_test.go` — add `TestUpdate_PlayPromptDismissedByY`, `TestUpdate_PlayPromptDismissedByN`, `TestUpdate_PlayPromptDismissedByEsc`, `TestUpdate_BoardInputBlockedDuringPlayPrompt`

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestUpdate_PlayPromptDismissedByY`
     - **Behavior:** Pressing y while `playPromptOpen` is true clears it
     - **Setup:** `m := New(nil); m.playPromptOpen = true` (same-package white-box); send `tea.KeyPressMsg{Code: 'y'}`
     - **Assertions:** `updated.(model).playPromptOpen == false`
     - **Boundary:** 'y' key — one of the three dismissal keys; the lower bound (dismissal by first key)
   - [ ] Confirm fails: `handleKeyPress` does not yet branch on `playPromptOpen` — the key falls through to board/list routing, leaving `playPromptOpen` unchanged

2. **Implement (GREEN):**
   - [ ] Add `playPromptTitle string` to `model` struct (after `playPromptOpen`)
   - [ ] In `handleReloaded`, when setting `m.playPromptOpen = true`: look up the parent track via `barChildrenOf`'s complement — find the task in `msg.tasks` where `task.ID == parentID` and set `m.playPromptTitle = track.Title` (if found; default to `parentID` if not)
   - [ ] Add `func (m model) handlePlayPrompt(msg tea.KeyPressMsg) (tea.Model, tea.Cmd)` with a switch on `msg.String()`: `"y", "n", keyEsc` → `m.playPromptOpen = false; return m, nil`; `keyCtrlC` → `return m, tea.Quit`; default → `return m, nil`
   - [ ] In `handleKeyPress`, prepend: `if m.playPromptOpen { return m.handlePlayPrompt(msg) }`
   - [ ] Add `func playPromptView(m model, board string) string` — renders a small centred box with `titledBorder("Play "+m.playPromptTitle+"? (y / n)", ...)` using `lipgloss.NewCompositor` on top of `board`
   - [ ] In `content()`, add `if m.mode == modeBoard && m.playPromptOpen { return playPromptView(m, board) }` before the `modalOpen` check

3. **More tests (RED → GREEN):**
   - [ ] `TestUpdate_PlayPromptDismissedByN`
     - **Behavior:** Pressing n clears `playPromptOpen`
     - **Setup/Assertions:** same as Y test but `Code: 'n'`
     - **Boundary:** 'n' — second dismissal key
   - [ ] `TestUpdate_PlayPromptDismissedByEsc`
     - **Behavior:** Pressing Esc clears `playPromptOpen`
     - **Setup/Assertions:** same as Y test but `Code: tea.KeyEscape`
     - **Boundary:** Esc — third dismissal key; the upper bound of the dismissal set
   - [ ] `TestUpdate_BoardInputBlockedDuringPlayPrompt`
     - **Behavior:** While `playPromptOpen` is true, space (which would otherwise trigger approval) is a no-op — the approve callback is never called
     - **Setup:** `m := New([]*task.Task{{ID: ttid1_1, Status: task.StatusDoing, Approved: false}}).WithApprove(func(_, _ string, _ bool) error { called = true; return nil }); m.playPromptOpen = true`; send `tea.KeyPressMsg{Code: ' '}`
     - **Assertions:** `called == false`; `updated.(model).playPromptOpen == true`
     - **Boundary:** space key while prompt open — verifies that only the three designated keys interact with the prompt

## Claude verifies
- [ ] `go test ./tui/... -run TestUpdate_PlayPrompt` passes
- [ ] `go test ./tui/... -run TestUpdate_BoardInput` passes
- [ ] `just test` passes
- [ ] `just lint` passes

## User verifies
- [ ] Whole slice: Approve the last unapproved bar on a track in `bp tui` → a "Play [track title]? (y / n)" overlay appears centred on the board; press y → overlay closes; try again, press n → overlay closes; try again, press Esc → overlay closes. The board is fully usable after each dismissal.

## Commit (user)
`feat(tui): route y/n/Esc to dismiss play prompt; render overlay`