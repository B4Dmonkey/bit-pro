---
id: BIT-38
title: MCP write surface
status: doing
approved: true
---
## Why

Claude reaches `.bit/` through Bash today, which means every write is one of a thousand things
that can be typed into a shell — and the visible consequence is Claude reaching for `mv`, `cat`,
and `sed` against files whose format the CLI owns. The read half of that problem is already
solved: `task_read` and `task_list` are typed tools with schemas. The write half is still shell,
so the operations that can actually corrupt a task file — or silently leave a bar approved after
its body changed — are precisely the ones with no schema in front of them. And because every
skill's writes still go through Bash, the deny rules that close the shell path can never be
turned on: nothing downstream in the MCP phase can move until this lands.

## Summary

Add the six write tools to `bp serve mcp` — `task_create`, `task_update`, `task_move`,
`task_complete`, `task_delete`, `feedback_add` — mirroring the CLI's nouns rather than
redesigning them. The shell affordances go (`$( )` ID capture, `-d "$(cat body.md)"`, the `-y`
confirmation prompt); every domain rule stays, above all approval revocation. Three of those rules
live in the Cobra layer rather than the task package today, so this scope moves them down: one
implementation, two callers, no second copy to drift.

## Visual aid

```
today                                     after this scope

cmd/task/create.go                        cmd/task/create.go ─┐
  ID minting, order insertion             cmd/task/update.go ─┤
cmd/task/update.go                        cmd/task/move.go ───┤ (thin callers)
  approval revocation                     cmd/serve_mcp.go ───┤
cmd/task/move.go                                              │
  exactly one anchor                                          v
        │                                            task.Store
        v                                              Create()  <- minting + ordering
   task.Store                                          Update()  <- approval revocation
     (persistence only)                                Move()    <- exactly one anchor
                                                       Complete/Relocate/AddNote
cmd/serve_mcp.go ──> task.Store (reads only)
```

## Decisions

- **The write rules move into the `task` package; both callers stay thin.** `cmd/task/create.go`
  owns ID minting and order insertion, `cmd/task/update.go:53` owns approval revocation, and
  `cmd/task/move.go:22` owns "exactly one of `--before` or `--after`" — none of the three is in
  the store. A tool handler calling the store directly would have to re-implement all three,
  which is the exact drift this phase exists to kill.
- **`Store.Move` takes the anchor pair, so the one-anchor rule moves down with the other two.**
  Today the signature is `Move(id, anchor string, before bool)`, which cannot see whether the
  caller supplied both anchors or neither — so the rule has to live above it, and the CLI is where
  it lives. Changing it to take `before` and `after` separately lets the store own the rule and
  leaves both `cmd/task/move.go` and the `task_move` handler thin. Rejected alternatives: a
  `oneOf` in the tool's input schema (the protocol would enforce it, but the CLI keeps its own
  copy), and repeating the check in the handler (two copies of a one-liner). Cost: an existing
  store signature changes, so `task/store_test.go` and `cmd/task/move_test.go` are updated with
  it.
- **Approval revocation is preserved verbatim.** A change to title, body, phase, or phase-label
  revokes approval; a status write of `todo` revokes it; a forward status move keeps it. This is
  load-bearing for the automation phase — an approved bar has to stay approved for a whole run.
- **`task_update` returns `{id, approved}`.** Revocation is silent in the CLI today: a skill has
  to know the rule and infer it fired. Returning the flag makes it observable, which matters most
  where it is most dangerous — a replan touching approved bars.
- **An omitted param means "leave unchanged"; a sent param means "set".** This is the schema's
  version of Cobra's `Changed()` check, and it is what stops a phase tag being zeroed or a body
  being blanked by an update that never mentioned them.
- **`status` is a schema enum: `todo | doing | done`.** The gotcha that `-s doen` succeeds and
  breaks rollup forever is a stringly-typed-flag accident, not a domain rule. No state machine is
  introduced — status stays a plain field.
- **Rollup stays skill logic.** The CLI does not cascade a bar's status up to its track and
  neither does the MCP. bit_do owns that rule and keeps owning it.
- **`task_delete` drops `-y` and keeps `force`.** The prompt is a terminal affordance that
  Claude Code's own permission prompt already covers; the unfinished-bars override is a domain
  rule.
- **`task_create` carries `after`.** The flag exists in code and not in the contract, so a bar can
  be inserted mid-track at create time and no skill knows it. Putting it in the schema fixes the
  drift by construction, since the schema *is* the documentation.
- **The CLI keeps every one of these commands.** Deleting the Claude-only ones is step 7 of the
  MCP phase, is optional, and is out of scope here.
- **Verification never requires a skill edit.** Two paths exist today. The Go tests in
  `cmd/serve_mcp_test.go` stand up the real server over `mcp.NewInMemoryTransports()` and call
  the tool as a client, so a verse is proven end to end through the protocol and its generated
  schema, not just at the handler. For a by-hand check, `claude mcp add bit -- bp serve mcp`
  registers the server in this repo — that is step 4's own command run early, undone with
  `claude mcp remove bit` — after which a plain-language request in a session picks the tool out
  of the tool list, because a tool needs no skill to tell Claude it exists. The server is
  registered nowhere today, so the by-hand path needs that one command first.
- **Skills are not touched.** They keep shelling out until the migration step, which sits behind
  a version tag. This scope is purely additive, and its undo is not using the tools.

## Verses

- [ ] Verse 1 — Claude can author and refine a scope without a shell: `task_create` mints a track
  with a multi-line body, `task_update` rewrites it in place, and the update result says whether
  approval survived the edit. The riskiest rule in the phase moves into the task package first, so
  a wrong reading of it fails on the first verse rather than the last.
  Touches: `cmd/serve_mcp.go`, `task/store.go`, `cmd/task/create.go`, `cmd/task/update.go`.
- [ ] Verse 2 — Claude can lay out a whole plan without a shell: bars created under a track with
  phase tags and `after` placement, plus `task_move` to resequence one. This is the step-3 bar in
  `mcp-notes.md` — a scope → plan cycle running start to finish with Bash never touching `.bit/`.
  Touches: `cmd/serve_mcp.go`, `task/store.go` (`Move`, `InsertAfter`).
- [ ] Verse 3 — Claude can run a bar and record a correction without a shell: `feedback_add`
  writes a note against a track; the status writes bit_do makes already ride on `task_update`. The
  bit_do and bit_feedback write paths are now fully covered.
  Touches: `cmd/serve_mcp.go`, `task/feedback.go`.
- [ ] Verse 4 — Claude can close a track out: `task_complete` files a signed-off track and its
  bars under `.bit/completed/`, and `task_delete` removes one, with `force` for a track that still
  has unfinished bars. With this the pipeline has no remaining write that needs a shell.
  Touches: `cmd/serve_mcp.go`, `task/store.go` (`Complete`, `Relocate`).

## References

- `mcp-notes.md` — the phase's working notes. Step 3 is this scope; its Decisions section is the
  source for the tool shape, the parity map, and the scope boundary against the automation phase.