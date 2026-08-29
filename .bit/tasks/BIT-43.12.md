---
id: BIT-43.12
title: A full cycle runs on tools alone
status: todo
phase: 5
phase_label: Full cycle
---
## **Verse 5**

A grep proves no skill *mentions* `bp`; it does not prove a skill *works*. This bar is the one
that settles it — a full scope → plan → do pass in the fixture project, driven entirely by tools.

This is the track's acceptance test and it can only be run by a human, because what is being
verified is a Claude session's behaviour. **A "no" here is a real result**, not a failed bar: it
means a skill lost something in translation, and the fix routes back through the verse that owns
that skill.

No RED step — the falsifiable observation below replaces it.

**Yes looks like:** a complete cycle produces a track, its bars, and a done bar with a commit, and
every `.bit/` write in the transcript is an `mcp__bit__*` call.
**No looks like:** a skill stalls, asks for something the tool surface cannot give, or falls back
to a `bp` Bash call. Record which skill and which step — that names the verse to reopen.

## Scope
- `tools/example` — the fixture project, reset and run. Nothing in `bit-pro` changes unless the
  cycle finds a defect.

## References
- `tools/example/RUNME.md` — the fixture's guided mode. Note its standing rule that the *guide*
  runs no commands and touches no `bp`; this bar is a real pipeline run, not a guided one, so
  drive it directly rather than pasting RUNME.
- `tools/example/reset.sh` — `./reset.sh blank` gives a `bp init`'d project with the code skeleton
  and hooks but **no track**, which is exactly the starting state a full scope → plan → do pass
  needs. (`planned` would skip the first two skills.)

## Method
- [ ] Confirm everything is pushed. The plugin installs from the default branch, so an unpushed
      skill edit reaches no project — including the fixture.
- [ ] In `tools/example`: `./reset.sh blank`
- [ ] Refresh the plugin and confirm the server: `claude plugin marketplace update bit-pro`,
      `claude plugin update bit@bit-pro --scope project`, then `claude mcp get bit` shows the
      server. If it is missing, `bp init` registers it (`claude/sync.go` wires
      `claude mcp add bit -- bp serve mcp`).
- [ ] Run `/bit:scope` on the fixture's small change, then `/bit:plan` on the track it produces,
      approve a bar with `bp approve <bar>` (still the CLI — deliberately), then `/bit:do` it.
- [ ] Watch the transcript for any `bp task`, `bp feedback`, or `bp instructions` Bash call.

## Claude verifies
- [ ] In `tools/example`, `bp task list` shows the track and its bars, and the executed bar's
      status is `done`. (Read-side only — the operator CLI is still the right tool for looking.)
- [ ] `git log` in the fixture shows the bar's commit

## User verifies
- [ ] **Whole slice, and the track's acceptance test:** the cycle above completes, and no step of
      it shows a `bp task`/`bp feedback`/`bp instructions` Bash call in the transcript. This is the
      claim the whole track exists to make — skills reach `bp` through a typed surface, not a
      shell.
- [ ] The approval gate still stops an unapproved bar, and `bp approve` is still what clears it.

## Report back
- [ ] If the cycle ran clean, the track is ready for sign-off — tell the user, and leave flipping
      the track to `done` to them.
- [ ] If it did not, take the specific failure to `bit:scope`: name the skill and the step, and
      let the verse that owns that skill be reopened rather than patching it from here.
- [ ] Either way, note that step 6 of `mcp-notes.md` — deny rules for shell writes against
      `.bit/` — is now unblocked. It is explicitly **not** this track's work; the scope's
      Decisions say so.

## Commit (user)
`chore(bit): BIT-43 verified — a full cycle runs on MCP tools alone`