---
id: BIT-25
title: 'Dispatch on approval: approving the last bar starts the track''s work'
status: todo
---
## Why

Approval is already the signal that work has been reviewed and may start — a human reads the
bar in the TUI and blesses it. Nothing acts on that signal. Starting the work is still a
person in a terminal hand-assembling an invocation: remember the flags, name the worktree,
namespace the session so it can be found again, and type the bar ID correctly.

Every one of those is silently fumbleable, and two of them are expensive. Name the worktree
per bar instead of per track and each bar restarts from `main`, losing the previous bar's
work. Type the wrong bar ID and a reviewed plan runs out of order. The operator pays this tax
once per bar, on every track, forever.

The gap is small because the hard parts belong to someone else. Claude Code owns the session;
`BIT-23` owns knowing whether a bar is eligible; `BIT-27` already guarantees a session running
inside a worktree writes to the canonical `.bit/`. What is missing is the hand-off itself —
the step between "I approved this" and "Claude is working it".

## Summary

Approving the last unapproved bar on a track starts the work. `bp` creates the track's
worktree if it does not exist, spawns a background Claude session in it prompted `/bit_do
<BAR>` for the first bar that is not `done`, and stops caring. It records nothing about the
session, never asks how it went, and re-reads `.bit/` from disk the next time it is invoked.

The worktree is named once, per track, at scope time, and that one string is the worktree
name, the branch name, and the session name — so a row in `claude agents` is directly
recognisable. Signing the track off reclaims both.

There is **no daemon** and no `bp dispatch` command. Approval is the only trigger.

## Visual aid

```
  scope a track ──▶ worktree name chosen and stored on the record
                                    │
  plan it ──▶ bars written, unapproved
                                    │
  review each bar ──▶ approve them one by one
                                    │
                    last one flips ─┴──▶ bp creates <name> worktree if absent
                                          spawns claude --bg -n <name> in it,
                                          prompted /bit_do <first not-done bar>
                                    │
                    already a live session for this track? ──▶ refuse, say so
                                    │
                          bp exits. It is done with this bar.
                                    │
  ... work happens; the operator watches it in `claude` ...
                                    │
  every bar done, PR merged ──▶ bp task complete ──▶ worktree + branch reaped
```

Nothing polls, nothing watches for completion, nothing advances bar 1 → bar 2. Advancing is a
later track (a `Stop` hook); this one only closes the gap between approval and a running
session.

## Decisions

- **No daemon, and dispatch is fire-and-forget.** Once a session is spawned, `bp` is done with
  that bar — it never learns how the session went. The only signal that matters is a bar
  reaching `done` in `.bit/`, which `bp` reads off disk like any other state. This retires the
  earlier `bp start` / `bp stop` shape and the idea of `bp` tracking sessions.
- **There is no `bp dispatch` subcommand.** Dispatch lives inside the approval path, so the
  TUI — where approval actually happens — fires it too. `bp approve` ships auto-denied for
  Claude, and **Claude never approves**; that stays true, which is what keeps an agent from
  dispatching itself.
- **The trigger is every bar on the track being approved**, and the bar dispatched is the
  first one whose status is not `done`. Eligibility is re-read from `.bit/` at the moment of
  dispatch, never captured as a snapshot at kickoff.
- **`bp` creates the worktree.** This reverses "`bp` never does setup", and the reason is
  naming: `-w` forces a `worktree-` prefix onto the branch, while a pre-created worktree
  handed to a session as its cwd keeps its exact branch name and is not re-isolated (measured
  2026-08-13).
- **The worktree is named per track, never per bar.** Two sessions given the same worktree
  report the same cwd, which is what lets bar 2 build on bar 1. Per-bar naming — the previous
  `-w bit-<BAR>` decision — would restart every bar from `main`.
- **One identifier everywhere.** Worktree name, branch name, and session name (`-n`) are the
  same string.
- **The name is decided at scope time and stored on the track record.** bit_scope asks for it
  and proposes a readable short one; bit_plan and bit_do carry it down. The default is
  `<track-id>-<short name>` (e.g. `bit-25-dispatch`); a name the user gives explicitly is used
  verbatim, with **no length limit and no mechanical truncation** — a name from Jira is
  whatever Jira says. The skill owns the name, not `bp` config.
- **Dispatch passes no `.bit/` location.** `BIT-27` made `bp` derive the canonical `.bit/` by
  cutting the path at `.claude/worktrees/`, and removed `--bit-dir` and `BIT_DIR` outright.
  The previous "dispatch passes `BIT_DIR`" decision is void.
- **A dispatched session is prompted `/bit_do <BAR>`.** The `bit` plugin is installed in every
  target repo, so that dependency is free. This settles the earlier open question and retires
  the alternative of injecting the bar's body inline.
- **The first pass spawns with no agent override.** A dedicated `bit:bot-dev` — different tool
  permissions, instructions written for no human continuously present — is deliberate follow-up
  work, not part of this track.
- **The session registry is the lock.** "Is this track already live?" is answered by filtering
  `claude agents --json` on the session-name prefix and `state ∈ {working, blocked}`. A
  finished session lingers as `state: done`, so presence is not liveness. No lockfile, no
  claim field, and nothing about sessions is written to `.bit/` — that mapping is local to one
  machine and meaningless to anyone who checks the repo out.
- **`bp` reaps the worktree on `bp task complete`,** not `claude rm`. A `bp`-created worktree
  is not locked, so a plain `git worktree remove` plus a branch delete works. Teardown lives
  in the command rather than in agent prose, because prose that must be remembered silently
  does not run.
- **No `performer` field.** Status already encodes it: bit_do leaves a bar `doing` and stops
  whenever it has `## User verifies` items, and only marks it `done` when there is nothing for
  a human to judge. A chain that advances on `done` therefore parks at exactly the bars
  needing a person, with no field to store and nothing to parse.
- **"Only the user pushes" is reversed.** bit_do commits and pushes; the tool-layer push deny
  rule comes out. The safety mechanism that replaces the push gate is the permission prompt —
  a background session parks on one and the operator answers it — so no permission mode may be
  chosen that suppresses prompts. Opening the PR stays the operator's.
- **Two changes are ordered before this track.** bit_do committing and pushing (without it a
  dispatched bar cannot close out on its own), and the approval-revocation fix in
  `cmd/task_update.go` — today `--status` revokes approval, so the moment bar 1 goes `doing`
  the track is no longer all-approved and the trigger can never re-arm.
- **Deliberately out of scope.** Advancing bar → bar on a `Stop` hook; `blocked_by` and
  readiness (dispatch gains an `AND ready` clause when that lands, in that track); the
  `bit:bot-dev` agent; and showing live session state in the TUI — `claude` already shows it,
  and `bp`'s display job is the board out of `.bit/`, not a second session viewer.

## Risks & unknowns

- **Unknown:** Does a background session spawned by `bp` survive `bp` exiting — including when
  the spawner is the TUI process rather than a shell?
  **Resolve by:** Verse 2, which proves it by building. Answer is yes if the session appears in
  `claude agents` with `state: working` and keeps working after `bp` (or the TUI) has exited;
  no if it dies with its parent or never registers.
  **De-risk before planning?** No — fire-and-forget is the whole shape of the verse, so this is
  the first thing verse 2 exercises. If it fails, the fallback is spawning detached rather than
  redesigning the track.

## Verses

- [ ] Verse 1 — A track carries the name its work will land under: bit_scope asks for and
  proposes a short readable name, `bp` stores it on the track record and defaults it to
  `<track-id>-<short name>`, and bit_plan and bit_do carry it down. The operator can see, before
  any work starts, exactly which branch a track's bars will build on.
  Touches: `task/` frontmatter, `cmd/task_create.go`, `cmd/task_update.go`, the scope/plan/do
  skills under `bit/skills/`, `assets/bit-cli.md` — where to look to verify.
- [ ] Verse 2 — Approving the last bar starts the work: `bp` creates the track's worktree if it
  is absent and spawns a background session in it, prompted `/bit_do <BAR>` for the first bar
  that is not `done`, then exits. The operator never hand-assembles an invocation and cannot
  fumble the bar ID or the worktree name.
  Touches: `cmd/approve.go`, the Claude Code integration package (`claude/`), `task/` — where to
  look to verify.
- [ ] Verse 3 — A track that is already working cannot be dispatched twice: `bp` checks the
  session registry for a live session on this track's name and refuses with a reason instead of
  spawning a second one over the top of a bar in progress. Re-approving after an edit becomes
  safe rather than destructive.
  Touches: the Claude Code integration package (`claude/`), `cmd/approve.go` — where to look to
  verify.
- [ ] Verse 4 — Signing a track off reclaims its worktree: `bp task complete` removes the
  worktree and deletes its branch alongside filing the track under `.bit/completed/`, so a
  finished track leaves nothing behind for the operator to clean up by hand.
  Touches: `cmd/task_complete.go`, the Claude Code integration package (`claude/`) — where to
  look to verify.

## References

- `automation-notes.md` — the running design notes this track is cut from. Its "Measured facts"
  section records what was observed about background sessions, worktree imposition, branch
  prefixes, and teardown; its "Settled decisions" section is the source for every decision
  above. Informs all four verses.
- `BIT-23` — the approval rails this dispatcher hangs off: the `approved` flag, `bp approve` /
  `bp unapprove`, and the TUI toggle. A shipped prerequisite, not a reference to consult.
- `BIT-27` — canonical `.bit/` derivation from a worktree path, and the removal of `--bit-dir` /
  `BIT_DIR`. Shipped; it is why verse 2 passes no task-state location.
- <https://code.claude.com/docs/en/headless> — headless invocation, permission modes, and
  session management. Informs verses 2 and 3.
- <https://github.com/gastownhall/gastown> — prior art for the shape this track deliberately
  does *not* take. Nearly all of its complexity follows from concurrency and unattended
  operation, both excluded here.