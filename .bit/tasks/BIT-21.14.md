---
id: BIT-21.14
title: A hand-edited lowercase id is normalized on read
status: todo
phase: 2
phase_label: Recurrence
---
## **Verse 2**

Everything so far normalizes what the *caller* types. This step covers the other direction the
scope decided on — normalize on read — because a hand-edited task file puts a lowercase ID back
into circulation without any command being mistyped. The test uses a correctly-cased command
against a corrupted file, so the previous bars' fixes cannot mask the failure.

## Scope
- `task/task.go` — normalize `ID` in `Parse`, so every task reaches the rest of the code
  canonical regardless of what its frontmatter says.
- `cmd/task_complete_test.go`, `cmd/task_update_test.go` — the new tests.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTaskCompleteCmd_HandEditedLowercaseIDStillHitsTheGuard`
     - **Behavior:** a task's identity comes from its normalized ID, so corrupt frontmatter
       cannot detach a bar from its track.
     - **Setup:** `initProject(t, "BIT")`; write `.bit/tasks/BIT-1.md` and `.bit/tasks/BIT-1.1.md`
       directly with `os.WriteFile`, giving them frontmatter `id: bit-1` and `id: bit-1.1` with
       the bar `todo` — correct filenames, corrupt contents. Run `task complete BIT-1`, uppercase.
     - **Assertions:** error is non-nil and names `unfinished bars BIT-1.1` — reported in the
       canonical case, not the case stored in the file; nothing is relocated.
     - **Boundary:** the argument is already correct, so the only lowercase input is the file's
       own frontmatter — this isolates read normalization from every fix in the earlier bars.
   - [ ] `TestTaskUpdateCmd_RewritesACorruptIDToCanonicalCase`
     - **Behavior:** the next write repairs the corrupt field rather than preserving it.
     - **Setup:** the same hand-written `.bit/tasks/BIT-1.md` with `id: bit-1`. Run
       `task update BIT-1 -s doing`.
     - **Assertions:** `.bit/tasks/BIT-1.md` now contains `id: BIT-1` and `status: doing`; no
       second file was created.
     - **Boundary:** a read followed by a write of the same task — proves normalization survives
       the round trip instead of being applied only in memory.
   - [ ] Confirm fails: the first test returns nil (the bar's parent parses as `bit-1`, matching
     nothing) and the second leaves `id: bit-1` in place

2. **Implement (GREEN):**
   - [ ] In `Parse`, apply `normalizeID` to `t.ID` after unmarshalling. `Save` already writes
     `t.ID`, so the repair falls out of the next write with no change there.

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` passes

## User verifies
- [ ] none — deterministic

## Commit (user)
`fix(task): normalize task IDs read from disk`