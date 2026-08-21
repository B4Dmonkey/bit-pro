---
id: BIT-30.3
title: The printed contract states the new approval rule
status: done
approved: true
phase: 2
phase_label: Sent back, re-reviewed
---
## **Verse 2**

The embedded contract's approval gotcha now describes behaviour that no longer exists, and every skill session reads it through `bp instructions` — so it has to be rewritten once both status rules are in. Prose only; it comes last because only now is the whole rule true.

## Scope
- `assets/bit-cli.md` — the gotcha bullet at lines 139–145 ("**Any `task update` to an approved task revokes its approval — including a status change.**"). Rewrite it to the implemented rule: content edits (`--title`, `--description`, `--phase`, `--phase-label`) revoke; `--status` revokes only when the new status is `todo`; forward moves keep approval. Delete the "expect an in-flight or finished bar to read as unapproved" advice and the "set it explicitly with `bp approve` after the status write" workaround — both are now wrong.

## TDD cycle

No new test. `cmd/instructions_test.go:TestInstructionsCmd_PrintsContract` already asserts `bp instructions` prints the embedded file byte-for-byte, so the wiring stays covered and passes for any text change; nothing in the bullet's prose is machine-checkable.

1. **Implement:**
   - [ ] rewrite the bullet in `assets/bit-cli.md`
   - [ ] read it back against `cmd/task_update.go` as implemented — each clause in the bullet maps to a disjunct in the revoke condition, and the condition has no disjunct the bullet omits

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` passes
- [ ] `grep -n "in-flight" assets/bit-cli.md` returns nothing — that phrase occurs only inside the stale advice being removed

## User verifies
- [ ] After `just install` (the contract is embedded at build time, so an un-reinstalled `bp` prints the old text), run `bp instructions` and confirm the approval bullet names `--status` → `todo` as the only status write that revokes, and no longer tells the reader to re-approve after starting work.

## Commit (user)
`docs(cli): state the new approval revoke rule in the contract`