---
id: BIT-21.16
title: A wrong-case track no longer overwrites a feedback note
status: todo
phase: 2
phase_label: Recurrence
---
## **Verse 2**

The remaining guarantee: "a new note can never damage one already recorded." Feedback has its own
path builder and its own sequence scan, neither of which routes through `Path`, so the earlier
bars do not reach it. Measured during planning — the wrong-case track made `nextNoteSeq` restart
at 001 and `AddNote` write straight over the existing note with no existence check.

## Scope
- `task/feedback.go` — normalize `track` in `AddNote` so `trackExists`, `nextNoteSeq` and
  `notePath` all agree.
- `cmd/feedback_add_test.go` — the new test.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestFeedbackAddCmd_LowercaseTrackDoesNotOverwriteAnExistingNote`
     - **Behavior:** the note sequence continues from what is on disk, whatever case the track
       was named in, so an existing note is never the write target.
     - **Setup:** `initProject(t, "BIT")`; create track `BIT-1`; `feedback add BIT-1 -d "FIRST"`.
       Then run `feedback add bit-1 -d "SECOND"`.
     - **Assertions:** the printed path is `.bit/feedback/BIT-1-002.md`;
       `.bit/feedback/BIT-1-001.md` still contains exactly `FIRST`; `.bit/feedback/` holds exactly
       two files and neither name contains a lowercase letter.
     - **Boundary:** the second note on a track — sequence 001 → 002 is the first point at which
       a restarted counter collides. Asserting the first note's *contents* is the whole test: the
       destructive case leaves the file present under its original name.
   - [ ] Confirm fails: `.bit/feedback/` holds one file, still named `BIT-1-001.md`, now
     containing `SECOND` — the first note is gone

2. **Implement (GREEN):**
   - [ ] Apply `normalizeID` to `track` at the top of `AddNote`, before the existence check.

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` passes

## User verifies
- [ ] none — deterministic

## Commit (user)
`fix(task): stop a wrong-case track from overwriting a note`