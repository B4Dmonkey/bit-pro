---
id: BIT-43.9
title: Contradiction forces approval and rollup onto task_update
status: done
approved: true
phase: 4
phase_label: Domain on descriptions
---
## **Verse 4**

Two pieces of domain have nowhere to live on the current surface, and both belong to
`task_update` because it is the tool a caller reaches for when either one bites.

**Approval:** `taskUpdateDescription` (`cmd/serve_mcp.go:45`) says only *"a change to what was
reviewed revokes it"*. It never says **which** changes, and it never says that writing `todo`
revokes while a forward move to `doing` or `done` does not — the rule `bit:do` depends on for a
whole run.

**Rollup:** nothing on the surface says the server does not cascade status from a bar to its
track. A caller that sets the last bar `done` and expects the track to follow is silently wrong,
and `bit:do` is exactly that caller.

**What forces this bar:** the single-row table from the previous bar can't express either fact.
Adding `task_update` rows whose `want` substrings the current description does not contain is a
contradiction the existing constant cannot satisfy.

## Scope
- `cmd/serve_mcp.go` — expand `taskUpdateDescription` (line 45)
- `cmd/serve_mcp_test.go` — add rows to `TestMCPToolDescriptions_CarryTheDomain`

## References
- `assets/bit-cli.md` — its "Gotchas" section (approval revocation) and "Rollup is skill logic,
  run through the CLI" are the two passages being relocated. Read them before they retire in the
  next bar.
- Verify the revocation rule against the implementation rather than the prose: the fields that
  revoke are the ones `taskUpdateInput` carries (`cmd/serve_mcp.go`) as applied by `task.Patch`.

## TDD cycle

1. **Write test (RED):**
   - [ ] Extend `TestMCPToolDescriptions_CarryTheDomain` with two `task_update` rows
     - **Behavior:** a caller that only reads the tool list learns both rules it can otherwise
       only discover by being wrong — that a `todo` write revokes approval while a forward status
       move preserves it, and that the parent track's status is the caller's job, not the server's.
     - **Setup:** unchanged from the previous bar — `mcpSession(t, t.TempDir())` then `ListTools`.
       Two new rows for `taskUpdateTool`.
     - **Assertions:** the approval row wants substrings covering the revoking field set, the
       literal `todo`, and that a forward move keeps approval. The rollup row wants a phrase
       stating the parent is not cascaded to (e.g. `"does not cascade"`) and that the caller sets
       the track's status itself. Choose the exact substrings when writing the constant and assert
       those — the point is that the concept is present and stays present, not any one wording.
     - **Boundary:** `status` is the field whose *direction* changes the outcome — `todo` (the
       backward edge) revokes, `doing`/`done` (forward) do not. Both sides of that boundary have
       to appear in the text, since a description carrying only one half is worse than none.
   - [ ] Confirm fails: the two new subtests report missing substrings while the `task_read` row
         from the previous bar still passes. If the `task_read` row also fails, the previous bar
         regressed — fix that first.

2. **Implement (GREEN):**
   - [ ] Expand `taskUpdateDescription` to name the fields whose change revokes approval, to state
         that writing `todo` revokes while `doing`/`done` preserve it, and to give the reason —
         a replan must not quietly alter a blessed bar, and a bar pulled back for rework must be
         re-reviewed before it runs again.
   - [ ] Add the rollup sentence: setting a bar's status leaves its track untouched; a caller that
         wants the track to reflect its bars sets the track's status in a separate call. Note that
         the `status` enum makes a misspelling impossible, so rollup can only break by omission.

**Do not restate the status-spelling warning as a caution.** The enum already rejects a typo; the
description should say the constraint is enforced, not warn about a failure that can no longer
happen.

## Claude verifies
- [ ] `just test` passes — including the `task_read` row from the previous bar
- [ ] `just lint` passes
- [ ] `just install`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(mcp): task_update's description carries approval and rollup rules`