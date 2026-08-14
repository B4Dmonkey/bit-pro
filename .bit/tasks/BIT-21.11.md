---
id: BIT-21.11
title: Lowercase track no longer bypasses the unfinished-bars guard
status: todo
phase: 2
phase_label: Recurrence
---
## **Verse 2**

The first of the four symptoms, driven from the outside: `bp task complete` with a lowercase
track must still refuse a track that has an unfinished bar. This is the guarantee `bp instructions`
makes with the words "there is no override", so it is the right place to start.

## Scope
- `task/store.go` — `children` compares a caller-supplied parent against IDs read from disk; add
  the normalization helper and apply it here.
- `cmd/task_complete_test.go` — the new test.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTaskCompleteCmd_LowercaseTrackStillHitsTheGuard`
     - **Behavior:** the unfinished-bars guard is driven by the track's identity, not by the
       spelling the caller happened to use.
     - **Setup:** `initProject(t, "BIT")`; create track `BIT-1`; create bar `BIT-1.1` and leave it
       `todo`. Run `task complete bit-1`.
     - **Assertions:** the returned error is non-nil and its message contains
       `unfinished bars BIT-1.1`; `.bit/tasks/BIT-1.md` still exists; `.bit/completed/` contains
       no entry for this track in either case.
     - **Boundary:** exactly one unfinished bar — the lower bound at which the guard must fire.
       The uppercase spelling of the same command already errors, so the case of the argument is
       the only variable between passing and failing.
   - [ ] Confirm fails: `Execute()` returns nil and the track is filed as
     `.bit/completed/bit-1.md` with its bar left behind — exit 0, no output, the bug as reported

2. **Implement (GREEN):**
   - [ ] Add `normalizeID(id string) string` to `task/` returning `strings.ToUpper(id)`, and apply
     it to `children`'s `parent` argument. Nothing else yet — the write path is the next bar.

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` passes

## User verifies
- [ ] none — deterministic

## Commit (user)
`fix(task): honour the unfinished-bars guard for any ID case`