---
id: BIT-25
title: 'Dispatch on approval: approve a bar and Claude starts working it'
status: todo
---
## Why

`bp` has two jobs in the automated pipeline, and neither exists yet. It shows the operator
what is being worked on, and it hands reviewed work to Claude. It does not do the work, and
it does not manage the machinery Claude uses to do it — creating a branch, building a
worktree, tearing either down — because Claude Code already owns all of that.

Today the hand-off is entirely manual and entirely undocumented. Starting a bar means opening
a terminal, remembering the invocation, remembering which flags keep task state out of the
branch, remembering to namespace the session so it can be found again, and getting the bar's
ID right by hand. Every one of those is a step that can be silently fumbled, and the
consequence of fumbling the task-state flag in particular is corrupted `.bit/` on a branch.

And once something is running there is nowhere to look. The board shows what the plan says,
not what is actually happening right now. The operator has to leave the tool they were just
reviewing in and go ask a different tool what it is doing.

Closing both gaps is small, because the hard parts belong to someone else: Claude Code owns
the session and the worktree, and `BIT-23` owns knowing whether a bar is eligible. What is
left is the hand-off itself.

## Summary

Approving a bar dispatches it. `bp` builds the invocation — the right worktree name, the
right session name, the flag that keeps task state in the canonical `.bit/` — spawns a
background session, and records which session belongs to which bar. The board then shows what
is actually running alongside what is planned.

An `orca` package owns background session dispatch and tracking. There is **no daemon**: the
operator is the loop, and approving the next bar is what advances the track.

## Visual aid

```
the loop, with the operator in it                    what bp adds
─────────────────────────────────                    ────────────
  review a bar on the board                             ─
  approve it                          ──▶  bp builds the invocation and spawns
                                            claude --bg -n bit/BIT-23.4 \
                                                       -w bit-BIT-23-4  \
                                                       BIT_DIR=<canonical .bit/>
  watch it work                       ──▶  board shows live session state
  it finishes                                           ─
  run the bar's declared check        ──▶  bp check   (BIT-23 Verse 3)
  approve the next bar                ──▶  dispatches  ← the loop closes here,
                                                          not in a daemon
```

Nothing watches for completion and advances on its own. That is the whole reason no daemon
is needed, and it is a deliberate limit rather than an omission.

## Risks & unknowns

- **Unknown:** What prompt does a dispatched session receive?
  **Resolve by:** A call from the user. `/bit_do <BAR>` is the obvious candidate since the
  skill already knows how to read a bar and execute it, but it is a real choice — the
  alternative is injecting the bar's body directly so the session does not have to re-derive
  context it was already given.
  **De-risk before planning?** Yes — it decides whether dispatch depends on the `bit` plugin
  being installed in the target repo, which is a meaningfully different constraint.

- **Unknown:** Does approving a bar dispatch it unconditionally, or does the operator get a
  confirmation first?
  **Resolve by:** A call from the user. Unconditional matches "approve and it goes" and is
  fewer keystrokes; a confirmation is the cheap guard against approving on the wrong row in a
  list UI, which is an easy mistake to make and an expensive one to undo once a session has
  started editing.
  **De-risk before planning?** No — it is a small change either way and does not reshape a
  verse.

## Decisions

- **No daemon.** Neither of `bp`'s jobs requires one. Displaying what is running is answered
  on demand by `claude agents --json`, and dispatching is synchronous with the operator's
  approval. A daemon would only be needed to advance a track *unattended* — dispatching bar 2
  when bar 1 finishes and passes, with nobody present — and that is explicitly not wanted.
  This retires the earlier `bp start` / `bp stop` shape entirely.
- **The operator is the loop.** Approving the next bar is what advances the track. Nothing
  polls, nothing watches for completion, nothing auto-commits. This is what keeps the design
  small, and it matches the stated position that automation is earned incrementally rather
  than assumed.
- **`bp` reviews, dispatches, and displays — it never does the work or the setup.** Creating
  a branch, building a worktree, and tearing both down belong to Claude Code, which already
  does them: measured 2026-08-13, a `--bg` editing session builds its own worktree with no
  flag asked for, and `claude rm <id>` removes the worktree and its branch. Anything in this
  package that starts looking like setup or teardown means the boundary has been crossed.
- **Approving a *bar* dispatches; approving a *track* does not.** `BIT-23` gives the two
  approvals different meanings — a track approved means the scope is blessed and planning may
  proceed, a bar approved means the plan is blessed and work may proceed — so only the second
  is a hand-off to Claude.
- **Dispatch passes a deterministic worktree name and session name.** `-w bit-<BAR>` and
  `-n bit/<BAR>`, both measured to work. The deterministic worktree is what makes the branch
  predictable; the namespaced session is what lets `bp` find its own sessions among every
  other session on the machine.
- **Dispatch passes the canonical `.bit/` location.** A dispatched editing session's working
  directory is a worktree whether or not this project asked for one, so without this the
  session's `bp` writes land on the branch. This is why `BIT-23` Verse 2 is a hard
  prerequisite rather than a nicety.
- **A new `orca` package owns this.** Background session dispatch and tracking are a distinct
  concern from task storage (`task/`) and command wiring (`cmd/`), and the commands stay thin
  wrappers over it, as the existing commands are over `task/`.
- **Session-to-bar mapping is machine state and never enters `.bit/`.** Which session id is
  working which bar is local to one machine and meaningless to anyone who checks the repo
  out, so it does not belong in git-tracked project state.
- **`BIT-23` lands first.** Approval is the trigger, the performer field is what stops a
  human-only bar from being handed to an agent, and the canonical-`.bit/` escape is what
  keeps a dispatched session from corrupting task state. All three are `BIT-23` verses.
- **Only the user pushes.** A dispatched session must be denied push at the tool layer — a
  deny rule is evaluated before the permission mode, so this is enforceable rather than
  merely requested.

## Verses

- [ ] Verse 1 — Approve a bar and Claude starts working it: one action spawns a correctly
  configured background session, so the operator never hand-assembles the invocation and
  cannot forget the flag that keeps task state off the branch. A bar whose performer is human
  is handed back to the operator instead of dispatched.
  Touches: a new `orca/` package, a new `cmd/` command, `cmd/root.go` — where to look to
  verify.
- [ ] Verse 2 — See what Claude is actually doing from the board: the TUI shows live session
  state next to the planned work, so "is this running, blocked, or finished?" is answerable
  without leaving the tool. A session waiting on a permission prompt is visible as waiting,
  which is the only channel telling the operator something is parked.
  Touches: `orca/`, `tui/model.go`, `tui/delegate.go`, `tui/board.go` — where to look to
  verify.

## References

- `automation-notes.md` — the design notes this track came out of. "Daemon substrate: two
  options" records what was measured about background sessions, including the `blocked` /
  `waitingFor` states Verse 2 surfaces. Informs both verses.
- `BIT-23` — the rails this dispatcher reads: approval, performer, and the canonical `.bit/`
  escape. A prerequisite, not a reference to consult.
- https://github.com/gastownhall/gastown — a Go daemon dispatching work to agents in
  worktrees. Prior art for the shape this track deliberately does *not* take; most of its
  complexity comes from concurrency and unattended operation, both excluded here.
- https://code.claude.com/docs/en/headless — headless invocation, permissions, and session
  management.