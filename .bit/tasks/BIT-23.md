---
id: BIT-23
title: 'Automation rails: human steps, canonical .bit/, runnable checks, approval'
status: doing
---
## Why

Every decision that would let work run unattended currently lives in the user's head or in
skill prose, not in `.bit/`. Nothing on disk answers "has a human blessed this plan?", or
"what proves this step is done?" — so the loop can only be driven by a human watching it,
one step at a time.

That has a cost today, before any automation exists. A large track queued in another repo
needs its own branch and has before/after testing steps in the middle of it. Running
that means remembering which steps are ready to work on — and any `bp` run from inside a
worktree lands task state on the branch, tangled into the code commits.

Writing that state down makes the manual loop better immediately, and it is the whole
prerequisite for automating the tail end of implementation later. Trust gets built by
watching a recorded pipeline run by hand before anything runs it unattended.

## Summary

Add the vocabulary and the plumbing an automated runner would need, and ship each piece as
something usable by hand. A track and its bars gain an **approval** flag — a mandatory gate:
work may not proceed until each record is approved. Any `bp` invocation, from any worktree,
writes to the canonical `.bit/`. The board shows only what is approved and ready.

No daemon here — that is the `orca` track. This is the manual pipeline with rails installed.

## Visual aid

```
today                          after this track
─────                          ────────────────
track ──▶ bars                 track ──▶ bars
  status: todo|doing|done        status:   todo|doing|done
                                 approved: bool      ← must be set before work proceeds
```

## Decisions

- **New fields land in YAML frontmatter, not JSON — for now.** Frontmatter is the cheap path
  and keeps every skill working unchanged. The move to JSON-backed state happens when
  something actually needs it — machine writes at a volume or concurrency frontmatter can't
  take — rather than pre-emptively. The cost of being wrong is migrating two fields, which
  is small enough to accept in exchange for not blocking this track on a format decision.
- **Editing an approved track or bar revokes its approval.** Settled as a safety requirement
  rather than a preference, because the dispatch design approves a batch of bars up front and
  then runs them unattended: without revocation you could approve eight bars, have a replan
  edit bar six, and have unreviewed work execute with nobody watching. Since the replanner is
  an agent and agents cannot approve, revocation is what forces the track to park and come
  back to a human. This means a flag plus change detection, not a bare flag.
- **One approval flag, not "refined" plus "approved".** Approval means "I've finished
  looking at this." Iteration happens before it; there is no intermediate state.
- **Approval is per-record, and the two gates fall out of that.** A track approved means the
  scope is blessed and planning may proceed; all bars approved means the plan is blessed and
  work may proceed. No special-casing beyond one flag per record.
- **Approval does not become a status value.** `status` stays `todo|doing|done`, so the board
  columns and all seven skills keep working unchanged.
- **`bp` does not create worktrees — Claude Code already does, and `bp` must survive that.**
  Measured 2026-08-13 on Claude Code 2.1.231, not assumed: a `--bg` session that edits files
  in a git repo auto-creates its own worktree with **no `-w` flag given**, at
  `<repo>/.claude/worktrees/<random-slug>` on branch `worktree-<slug>`, marks it `locked`,
  and sets the session's `cwd` to it. So a worktree is not something this project opts into —
  it is imposed, and every `bp` call an agent makes inside a dispatched session lands in the
  branch's checked-out `.bit/` copy unless something prevents it. Building a competing
  `bp worktree` command would duplicate a mechanism that already exists and fight it.
- **Worktree location and cleanup are Claude Code's, not ours.** Also measured: `-w <name>`
  makes the path deterministic (`.claude/worktrees/<name>`, branch `worktree-<name>`), so a
  caller that wants a predictable branch names it at dispatch. `claude stop <id>` **retains**
  the worktree; `claude rm <id>` removes the worktree *and* its branch. That is a complete
  lifecycle, so this project neither places nor reaps worktrees. This retires the earlier open
  question about where they live and who removes them.
- **No daemon in this track.** Everything here is a command a human runs. The daemon is the
  `orca` track and it schedules these same commands.
- **Session identity is not this track's problem.** Recording a dispatched session's name and
  id belongs with the thing that dispatches. It is not usable by hand, which is this track's
  admission test, so it moves to the `orca` track.
- **Only the user pushes.** Nothing added here may push to a remote.
- **Space toggles approval in the TUI.** Space is unbound in both the list view and the board
  view; it reads as "stamp it" and is the standard toggle key in terminal UIs. No other key
  was a better fit.
- **Unapproved items appear in yellow; approved items stay white.** Yellow is a visible
  warning-level signal without being alarming — it communicates "needs a look" at a glance.
  Approved items keep the current color so approval is the zero-friction state.
- **The board's todo column shows only approved items.** Unapproved todos are hidden from the
  kanban view — the board answers "what is ready to work on," not "what exists." They remain
  visible in the list view.

## Verses

- [x] Verse 1 — Task state stays put no matter where `bp` runs from: a `bp` invocation inside
  a worktree writes to the canonical `.bit/` in the main checkout rather than the branch's
  copy, so work on a branch never drags task state into its code commits. Needed before
  anything is dispatched, because a background session's working directory is a worktree
  whether or not this project asked for one.
  Touches: `cmd/root.go` (a bit-dir flag / env var), `task/store.go`, `task/config.go` —
  where to look to verify.
- [x] Verse 3 — Say "I'm done reviewing this" and see what's ready via the CLI: a new
  `bp approve` / `bp unapprove` command sets the flag on any record, and `task list` surfaces
  it, so "what can be worked on?" is answerable without opening the TUI.
  Touches: `task/task.go`, a new `cmd/approve.go`, `cmd/task_list.go` — where to look to verify.
- [x] Verse 4 — The TUI shows approval state and lets you toggle it without leaving the board:
  unapproved items appear in yellow, approved items in the current white, and pressing space
  toggles the flag on the focused item.
  Touches: `tui/` — where to look to verify.
- [ ] Verse 5 — The board's todo column shows only approved items: unapproved todos are
  filtered out of the kanban view so the board answers "what is ready to work on" rather than
  "what exists." They remain visible in the list view.
  Touches: `tui/` — where to look to verify.
- [ ] Verse 6 — bit:do knows how to handle the approved field: skill-creator updates the bit:do
  skill so it reads `approved` from each bar and gates on it before starting work. Planning
  this verse means invoking the skill-creator skill.
  Touches: `bit/skills/do`, skill-creator — where to look to verify.

## References

- `automation-notes.md` — the working design notes this track came out of, including the
  measured headless facts and the two daemon substrate options. Informs every verse.
- https://github.com/gastownhall/beads — graph-backed issue tracker for agents; the source of
  the readiness-and-claim model. Informs Verse 3.
- https://github.com/gastownhall/gastown — daemon-and-worktree orchestration above a tracker.
  Informs Verse 1, and marks the boundary of what this track deliberately excludes.
- https://github.com/awslabs/aidlc-workflows — phase gates requiring explicit human approval.
  Informs Verse 3.
- https://code.claude.com/docs/en/headless — headless invocation, structured output, and
  session resume. Background for the `orca` track, not this one.