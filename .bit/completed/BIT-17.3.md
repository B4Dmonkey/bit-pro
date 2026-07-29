---
id: BIT-17.3
title: A note refuses a track that does not exist
status: done
phase: 1
phase_label: A durable record
---
## **Verse 1**

A note whose track ID is a typo is a note retro can never find, because retro locates notes by
walking tracks. Nothing currently stops one, so a test that demands a refusal forces the same
existence guard `NextChildID` already applies to a parent.

## Scope
- `task/feedback.go` — before writing, `os.Stat(s.Path(track))` and wrap the failure the way
  `NextChildID` does: `fmt.Errorf("track %s does not exist: %w", track, err)`. The guard runs
  before `MkdirAll`, so a refused note leaves no `.bit/feedback/` behind either.
- `cmd/feedback_add_test.go` — add the unknown-track test.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestFeedbackAddCmd_ErrorsOnUnknownTrack`
     - **Behavior:** a note can only key to a track that exists, so a mistyped ID fails loudly
       instead of producing evidence nothing will ever read.
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "Ship the bit plugin", "...")` so one
       real track exists; then `run(t, "feedback", "add", "BIT-99", "-d", "...")`.
     - **Assertions:** the returned error is non-nil; `os.Stat(".bit/feedback/BIT-99-001.md")`
       gives `fs.ErrNotExist`; and `os.Stat(".bit/feedback")` also gives `fs.ErrNotExist`, proving
       the guard ran before the directory was created.
     - **Boundary:** track existence, exercised in its false state — the true state is what
       Bars 1 and 2 already cover, so between them both sides of the condition are pinned.
   - [ ] Confirm fails: the command exits 0 and `.bit/feedback/BIT-99-001.md` exists.

2. **Implement (GREEN):**
   - [ ] Add the `os.Stat` guard at the top of `AddNote`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(feedback): refuse a note keyed to a track that does not exist`