---
id: BIT-23
title: 'Automation rails: human steps, canonical .bit/, runnable checks, approval'
status: todo
---
## Why

Every decision that would let work run unattended currently lives in the user's head or in
skill prose, not in `.bit/`. Nothing on disk answers "has a human blessed this plan?", "is
this step mine or the agent's?", or "what proves this step is done?" — so the loop can only
be driven by a human watching it, one step at a time.

That has a cost today, before any automation exists. A large track queued in another repo
needs its own branch and has manual before/after testing steps in the middle of it. Running
that means remembering which steps the agent must not touch — and any `bp` run from inside a
worktree lands task state on the branch, tangled into the code commits.

Writing that state down makes the manual loop better immediately, and it is the whole
prerequisite for automating the tail end of implementation later. Trust gets built by
watching a recorded pipeline run by hand before anything runs it unattended.

## Summary

Add the vocabulary and the plumbing an automated runner would need, and ship each piece as
something usable by hand. A bar gains a **performer** (agent or human) and a **declared
check** that `bp` can run. A track gains an **approval** flag. Any `bp` invocation, from any
worktree, writes to the canonical `.bit/`. Task state becomes readable as JSON.

No daemon here — that is the `orca` track. This is the manual pipeline with rails installed.

## Visual aid

```
today                          after this track
─────                          ────────────────
track ──▶ bars                 track ──▶ bars
  status: todo|doing|done        status:    todo|doing|done
                                 approved:  bool          ← "I'm done reviewing this"
                                 performer: agent|human   ← who does the work
                                 check:     cmd | human   ← what proves it done, and
                                                            `bp` can run it
```

Why the check has to be runnable by `bp`, not reported by the agent:

```
  dispatch ──▶ claude --bg ──▶ (no stdout contract, session is detached)
                                       │
                                       ▼
                            the model's claim "it passed"   ← not trustworthy
                                       │
  bp check <BAR> ──▶ runs the declared command ──▶ verdict  ← the only trustworthy one
```

## Decisions

- **New fields land in YAML frontmatter, not JSON — for now.** Frontmatter is the cheap path
  and keeps every skill working unchanged. The move to JSON-backed state happens when
  something actually needs it — machine writes at a volume or concurrency frontmatter can't
  take — rather than pre-emptively. The cost of being wrong is migrating three fields, which
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
- **Performer and check are separate axes.** Who does the work (agent or human) is a
  different question from what verifies it (a command or a human sign-off). A human-performed
  step is the mechanism for manual testing bars.
- **Approval does not become a status value.** `status` stays `todo|doing|done`, so the board
  columns and all seven skills keep working unchanged.
- **`bp` runs the declared check; it does not merely record it.** Settled by the choice of
  background sessions as the automation substrate: `--bg` and `-p` are mutually exclusive, so
  a dispatched session returns no structured verdict, and the only alternative source of
  truth is the model's own claim that the bar passed. That claim cannot be trusted, so the
  verdict has to come from `bp` executing the declared command itself. This is what makes the
  check verse load-bearing rather than a convenience.
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
- **One track at a time.** Concurrent tracks, merge queues, and worker supervision are out of
  scope — a correct pipeline first.
- **No daemon in this track.** Everything here is a command a human runs. The daemon is the
  `orca` track and it schedules these same commands.
- **Session identity is not this track's problem.** Recording a dispatched session's name and
  id belongs with the thing that dispatches. It is not usable by hand, which is this track's
  admission test, so it moves to the `orca` track.
- **Only the user pushes.** Nothing added here may push to a remote.

## Verses

- [ ] Verse 1 — A plan can contain steps the agent must not do: a bar records who performs it
  and carries instructions for the human, and bit_do stops and hands off instead of
  implementing it. Unblocks the manual before/after testing on the queued track.
  Touches: `task/task.go` (frontmatter), `cmd/task_create.go`, `cmd/task_update.go`,
  `bit/skills/do` — where to look to verify.
- [ ] Verse 2 — Task state stays put no matter where `bp` runs from: a `bp` invocation inside
  a worktree writes to the canonical `.bit/` in the main checkout rather than the branch's
  copy, so work on a branch never drags task state into its code commits. Needed before
  anything is dispatched, because a background session's working directory is a worktree
  whether or not this project asked for one.
  Touches: `cmd/root.go` (a bit-dir flag / env var), `task/store.go`, `task/config.go` —
  where to look to verify.
- [ ] Verse 3 — A bar states what proves it done, and that proof can be run: the verification
  is recorded on the bar as a command or a human sign-off, and `bp` executes the command and
  reports the verdict, so bit_check follows a written criterion instead of improvising one and
  a later daemon has a source of truth that isn't the model's own say-so.
  Touches: `task/task.go`, `cmd/task_create.go`, `cmd/task_update.go`, a new `cmd/` command,
  `bit/skills/check` — where to look to verify.
- [ ] Verse 4 — Say "I'm done reviewing this" and see what's ready: approval is settable and
  visible in `task list` and the TUI, so "what can be worked on?" is answerable from the
  board instead of from memory.
  Touches: `task/task.go`, a new `cmd/` command, `cmd/task_list.go`, `tui/` — where to look
  to verify.
- [ ] Verse 5 — Query task state from a script: `--json` on read and list, so anything
  outside the skills can consume `.bit/` without parsing markdown.
  Touches: `cmd/task_read.go`, `cmd/task_list.go` — where to look to verify.

## References

- `automation-notes.md` — the working design notes this track came out of, including the
  measured headless facts and the two daemon substrate options. Informs every verse.
- https://github.com/gastownhall/beads — graph-backed issue tracker for agents; the source of
  the readiness-and-claim model. Informs Verse 4.
- https://github.com/gastownhall/gastown — daemon-and-worktree orchestration above a tracker.
  Informs Verse 2, and marks the boundary of what this track deliberately excludes.
- https://steve-yegge.medium.com/welcome-to-gas-town-4f25ee16dd04 — the author's own account
  of the throughput-over-correctness tradeoff, and of acceptance criteria as the thing that
  makes agent work durable. Informs Verse 3 and the one-track-at-a-time decision.
- https://github.com/awslabs/aidlc-workflows — phase gates requiring explicit human approval.
  Informs Verse 4.
- https://code.claude.com/docs/en/headless — headless invocation, structured output, and
  session resume. Background for the `orca` track, not this one.