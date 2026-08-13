---
id: BIT-23
title: 'Automation rails: approval, human steps, declared checks, worktrees'
status: todo
---
## Why

Every decision that would let work run unattended currently lives in the user's head or in
skill prose, not in `.bit/`. Nothing on disk answers "has a human blessed this plan?", "is
this step mine or the agent's?", or "what proves this step is done?" — so the loop can only
be driven by a human watching it, one step at a time.

That has a cost today, before any automation exists. A large track queued in another repo
needs its own branch and has manual before/after testing steps in the middle of it. Running
that means hand-rolled git plumbing plus remembering which steps the agent must not touch —
and if the agent runs `bp` from inside a worktree, task state lands on the branch tangled
into the code commits.

Writing that state down makes the manual loop better immediately, and it is the whole
prerequisite for automating the tail end of implementation later. Trust gets built by
watching a recorded pipeline run by hand before anything runs it unattended.

## Summary

Add the vocabulary and the plumbing an automated runner would need, and ship each piece as
something usable by hand. A bar gains a **performer** (agent or human) and a **declared
check**; a track gains an **approval** flag and a **worktree**. A `bp` invocation inside a
worktree writes to the canonical `.bit/`. Task state becomes readable as JSON.

No daemon here. This is the manual pipeline with rails installed.

## Visual aid

```
today                          after this track
─────                          ────────────────
track ──▶ bars                 track ──▶ bars
  status: todo|doing|done        status:    todo|doing|done
                                 approved:  bool          ← "I'm done reviewing this"
                                 performer: agent|human   ← who does the work
                                 check:     cmd | human   ← what proves it done
                                 worktree:  path/branch   ← where the work happens
```

## Risks & unknowns

- **Unknown:** Do these new fields land in YAML frontmatter now, given that a move to
  JSON-backed state is being considered for the automation phase after this one?
  **Resolve by:** A call from the user. Frontmatter is the cheap path and keeps every skill
  working unchanged; the cost is migrating four new fields later instead of one format.
  **De-risk before planning?** Yes — it decides the shape of four of the five verses.

- **Unknown:** Does editing an approved track or bar revoke its approval?
  **Resolve by:** A call from the user. Revoking is the safe answer once anything runs
  unattended (otherwise you can approve, then edit, then have unreviewed work executed), but
  it may be irritating while a human is still the only executor.
  **De-risk before planning?** Yes — it's the difference between a flag and a flag plus
  change detection.

- **Unknown:** Where do worktrees live, and who removes them?
  **Resolve by:** A call from the user. Sibling directory, a path under the repo, or a
  central location; and whether `bp` cleans up on completion or leaves it to the user.
  **De-risk before planning?** Yes — Verse 2 can't be planned without it.

- **Unknown:** Does `bp` ever *run* a declared check in this track, or only record it?
  **Resolve by:** A call from the user. Recording only is the smaller slice and still
  gives bit_check something repeatable to follow; running it is what a later daemon needs.
  **De-risk before planning?** Yes — it's the whole scope of Verse 4.

## Decisions

- **One approval flag, not "refined" plus "approved".** Approval means "I've finished
  looking at this." Iteration happens before it; there is no intermediate state.
- **Approval is per-record, and the two gates fall out of that.** A track approved means the
  scope is blessed and planning may proceed; all bars approved means the plan is blessed and
  work may proceed. No special-casing beyond one flag per record.
- **Performer and check are separate axes.** Who does the work (agent or human) is a
  different question from what verifies it (a command or a human sign-off). A human-performed
  step is the mechanism for manual testing bars.
- **Approval does not become a status value.** `status` stays `todo|doing|done`, so the board
  columns and all seven skills keep working unchanged.
- **One track at a time.** Concurrent tracks, merge queues, and worker supervision are out of
  scope — a correct pipeline first.
- **No daemon in this track.** Everything here is a command a human runs. The daemon is the
  next scope and it schedules these same commands.
- **Only the user pushes.** Nothing added here may push to a remote.

## Verses

- [ ] Verse 1 — A plan can contain steps the agent must not do: a bar records who performs it
  and carries instructions for the human, and bit_do stops and hands off instead of
  implementing it. Unblocks the manual before/after testing on the queued track.
  Touches: `task/task.go` (frontmatter), `cmd/task_create.go`, `cmd/task_update.go`,
  `bit/skills/do` — where to look to verify.
- [ ] Verse 2 — Work a track on its own branch with one command: `bp` derives a branch name,
  creates the worktree, and records both on the track; a `bp` run from inside that worktree
  writes to the canonical `.bit/` in the main checkout rather than the branch's copy.
  Touches: a new `cmd/` command, `cmd/root.go` (a bit-dir flag / env var), `task/store.go`
  — where to look to verify.
- [ ] Verse 3 — Say "I'm done reviewing this" and see what's ready: approval is settable and
  visible in `task list` and the TUI, so "what can be worked on?" is answerable from the
  board instead of from memory.
  Touches: `task/task.go`, a new `cmd/` command, `cmd/task_list.go`, `tui/` — where to look
  to verify.
- [ ] Verse 4 — A bar states what proves it done: the verification is recorded on the bar as
  a command or a human sign-off, so bit_check follows a written criterion instead of
  improvising one per step.
  Touches: `task/task.go`, `cmd/task_create.go`, `cmd/task_update.go`, `bit/skills/check`
  — where to look to verify.
- [ ] Verse 5 — Query task state from a script: `--json` on read and list, so anything
  outside the skills can consume `.bit/` without parsing markdown.
  Touches: `cmd/task_read.go`, `cmd/task_list.go` — where to look to verify.

## References

- https://github.com/gastownhall/beads — graph-backed issue tracker for agents; the source of
  the readiness-and-claim model. Informs Verse 3.
- https://github.com/gastownhall/gastown — daemon-and-worktree orchestration above a tracker.
  Informs Verse 2, and marks the boundary of what this track deliberately excludes.
- https://steve-yegge.medium.com/welcome-to-gas-town-4f25ee16dd04 — the author's own account
  of the throughput-over-correctness tradeoff, and of acceptance criteria as the thing that
  makes agent work durable. Informs Verse 4 and the one-track-at-a-time decision.
- https://github.com/awslabs/aidlc-workflows — phase gates requiring explicit human approval.
  Informs Verse 3.
- https://code.claude.com/docs/en/headless — headless invocation, structured output, and
  session resume. Background for the next scope, not this one.