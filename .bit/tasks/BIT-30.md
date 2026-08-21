---
id: BIT-30
title: approval survives a forward status move
status: doing
---
## Why

Approval is the signal that a bar has been reviewed and is cleared to run. On the board an
unapproved card is yellow, an approved one is white. Today every status move wipes the
approval flag, so the moment work starts on an approved bar the card turns yellow again —
the board says "nobody has reviewed this" about the exact bar that was just approved and
handed to bit_do. The one piece of state the board exists to show is unreliable during a
normal run, which trains the operator to ignore the colour.

## Summary

Status moves stop counting as edits for the purpose of revoking approval. Moving a task
forward (`todo → doing`, `doing → done`) keeps its approval and its white card. Moving a
task backward to `todo` still revokes approval, because a bar sent back is a bar that wants
re-review before it runs again. Title, description and phase edits keep revoking approval
exactly as they do today.

## Visual aid

```
today                          after
  approve BIT-1   → white        approve BIT-1   → white
  -s doing        → YELLOW  ✗    -s doing        → white  ✓
  -s done         → YELLOW  ✗    -s done         → white  ✓
  -s todo (back)  → yellow       -s todo (back)  → yellow ✓
  -t "new title"  → yellow       -t "new title"  → yellow (unchanged)
```

## Decisions

- **A forward status move preserves approval.** `todo → doing` and `doing → done` are the
  act of doing approved work, not a change to what was approved, so they must not revoke it.
- **A move back to `todo` revokes approval.** Sending a task back means its content or its
  readiness is in question; requiring re-approval before it can run again is the point of
  the flag.
- **Content edits are untouched.** `--title`, `--description`, `--phase` and `--phase-label`
  keep revoking approval on any task, in any column. Edits to a task that is already in
  `todo` and stays there are explicitly out of scope for this change.
- **An unapproved `todo` card stays hidden from the board.** `tui/board.go` already filters
  unapproved todos out of the board's To Do column (they remain in list view), so a
  `done → todo` move shows up as the card leaving the board rather than as a yellow card in
  To Do. That is the existing contract for "not yet cleared to run" and this change does not
  alter it.

## Verses

- [ ] Verse 1 — Approved work stays approved as it runs: an approved task moved `todo → doing`
  or `doing → done` keeps its approval, so the card stays white for the whole run.
  Touches: the update command's revoke condition (`cmd/task_update.go`) — where `--status`
  is currently folded into `anyChanged`.
- [ ] Verse 2 — A task sent back asks to be re-reviewed: moving a task from `doing` or `done`
  to `todo` revokes approval, so it cannot run again until someone approves it.
  Touches: the same revoke condition (`cmd/task_update.go`).