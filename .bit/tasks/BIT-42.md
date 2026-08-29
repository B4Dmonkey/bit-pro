---
id: BIT-42
title: Unapproved bars must never read as queued
status: todo
---
## Why

The board tells the operator something untrue. In the TUI right now, `BIT-41.5` renders
cyan — the queued colour — even though it is **not approved**. The operator reads that as
"approved and about to run," and acts on it: it is the signal they use to decide what is
safe to leave alone. Four already-`done` bars are painted the same way.

Approval is the one gate between a plan and an unattended agent editing a repo. A colour
that says "queued" on a bar the ledger says must not run inverts the meaning of the gate —
the operator stops trusting the board, and the gate stops being a gate. That is the same
class of failure as the dispatch defect chain in START-HERE.md: a state that looked right
and was not, with nothing on screen to distinguish the two.

## Summary

Make the queue agree with the ledger, and make the operator able to see and correct it.
A bar that loses approval, or finishes, leaves the queue at the moment it happens rather
than waiting for a daemon pass that may never come. Trying to queue an unapproved bar is
refused out loud instead of silently ignored. And the queue stops being write-only — the
operator can look at it and clear it.

## Visual aid

```
             what the ledger says            what the queue says       what the board paints
BIT-41.1     done                            queued                    cyan  "about to run"
BIT-41.4     done                            queued                    cyan  "about to run"
BIT-41.5     todo, NOT approved              queued                    cyan  "about to run"
BIT-41.6     todo, approved                  queued                    cyan  "about to run"
                     \                          /
                      `-- these must not disagree --'

today the only thing that reconciles them is daemon/loop.go dropped(), and it is
gated twice over:
    daemon stopped                    -> never runs
    a live session under the project  -> loop.go:81 skips dispatch entirely
                                         (i.e. exactly while you are using the TUI)
```

## Risks & unknowns

- **Unknown:** Where does a transient message live in the TUI? The bottom row currently
  renders `m.help.View(m.helpKeys())` and nothing else competes for it, so swapping that
  row for a message is the obvious home — but it has not been built, and whether a message
  should time out, persist until the next keypress, or persist until the next reload is a
  feel question that only shows up on screen.
  **Resolve by:** Verse 3 builds the thinnest version (replace the help row while a message
  is set, clear it on the next keypress) and the operator judges it live.
  **De-risk before planning?** No — it is a small, reversible piece of rendering, and
  nothing else in the scope depends on the answer.

## Decisions

- **Revoking approval dequeues immediately; readers do not filter defensively.** The queue
  is a statement of intent to run, so it must not outlive the approval that justified it.
  Chosen over having every reader intersect the queue against the ledger, because a
  defensive filter leaves rows that are invisible but still live — re-approving the bar
  would fire a request the operator never re-made.
- **The reconciliation lives in `cmd/`, not in `task/`.** The rule is: after any write that
  leaves a bar unapproved or done, delete its queue row. It is idempotent, so it needs no
  before/after comparison. Putting it at the command layer keeps the file-based store from
  taking a sqlite dependency it does not have today — `cmd/tui.go` already opens the db, and
  the other callers can.
- **Every revocation path is covered, not just `bp unapprove`.** Three things revoke
  approval today: `bp unapprove`, the TUI approve toggle, and the edit-driven revocation in
  `task/store.go:302` that fires when a task's content changes. A fix that only covers the
  explicit command leaves the most common path — editing a bar — still able to strand a
  queue row.
- **Enqueue refuses an unapproved bar out loud.** Silently dropping the keypress is what the
  track path does today, and it is why the mismatch went unnoticed. The operator gets told
  the bar is not approved. Accepted cost: the TUI has no message surface, so one gets built.
- **The daemon's existing `dropped()` check stays.** It is the last line of defence at
  dispatch time and costs nothing. This scope makes it the backstop rather than the only
  mechanism.
- **The cyan/queued colour keeps its meaning.** Cyan means "in the queue." Nothing about the
  palette changes — the fix is that the queue stops containing rows that should not be in it,
  not that the colour learns to lie more precisely.

## Verses

- [ ] Verse 1 — See the queue and clear it: the operator can list what is queued for a
  project and remove a row, so the six stale rows sitting there today can be cleared without
  waiting on a daemon that is stopped. Today the queue can only be added to — there is no
  `bp queue` command and no dequeue key in the TUI.
  Touches: `cmd/`, `db/queries/queue.sql`, `tui/model.go`.

- [ ] Verse 2 — Losing approval takes a bar out of the queue: revoking approval or finishing
  a bar removes its queue row at the moment of the write, through every path that can do it —
  `bp unapprove`, the TUI toggle, and an edit that revokes approval implicitly. After this the
  board cannot drift back into disagreeing with the ledger.
  Touches: `cmd/approve.go`, `cmd/tui.go`, `cmd/task/`, `db/queries/queue.sql`.

- [ ] Verse 3 — Queueing an unapproved bar is refused, visibly: selecting a bar directly and
  pressing enqueue honours the same approval-and-not-done filter the track path already
  applies, and the TUI says why nothing happened instead of swallowing the keypress.
  Touches: `tui/model.go` (`enqueueSelected`, `enqueueableBarIDs`), the TUI's bottom row.

## References

- `START-HERE.md` — the dispatch design session. The failure mode it records, a broken
  install being indistinguishable from a stale one, is the same shape as a queued row being
  indistinguishable from an approved one. Informs the Why.