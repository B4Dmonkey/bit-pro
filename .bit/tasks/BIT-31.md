---
id: BIT-31
title: TUI — play prompt
status: todo
---
## Why
The approval keypress is already where the operator signals intent — "this bar is ready." When that gesture completes the last unapproved bar on a track, the moment to ask "do you want to run this track?" is right now, not later. Without a prompt here, the operator has to remember to navigate somewhere else to enqueue, which breaks the flow at the exact point where it should feel seamless.

## Summary
Approving the last unapproved bar on a track immediately opens a modal in the TUI asking whether to play it. Both "yes" and "no" dismiss the modal and return to the board. "Yes" is a no-op for now; the actual enqueueing is automation step 5.

## Decisions
- **Trigger condition is a pure state check after reload.** After any bar approval, if the task just approved is a bar and its parent track now has ≥1 bar with none unapproved, the prompt fires. The check reads the post-reload task list — it does not track the sequence of keystrokes.
- **Re-approving re-fires the prompt.** The condition is stateless: "is the track in the ready state right now?" Unapproving then re-approving a bar puts the track back into the ready state, so the prompt appears again. Simplest consistent rule.
- **Zero-bar tracks don't get a prompt.** No bars means planning hasn't happened yet; "≥1 bar" is part of the condition.
- **"Yes" is inert for now.** Queuing is automation step 5; this scope ships the modal and the keys, nothing more.

## Verses
- [ ] Verse 1 — Operator sees the play prompt on the last approval: approving the last unapproved bar on a track opens a modal reading "Play [track title]? (y / n)". Pressing y, n, or Esc closes it; y does nothing else yet.
  Touches: `tui/model.go` (`handleApprove`, `handleReloaded`, modal management), `cmd/tui.go` (may need to pass a children-check callback) — the post-reload hook and modal flag live here.