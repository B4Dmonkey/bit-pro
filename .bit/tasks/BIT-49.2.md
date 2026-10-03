---
id: BIT-49.2
title: A new note takes the next free number instead of overwriting one another session wrote
status: todo
approved: true
phase: 1
phase_label: feedback and retro are shared across projects
---
## **Verse 1**

In a shared folder, two worktrees of one project can both pick the same next number, and `os.WriteFile` (`task/feedback.go:74`) would let the second note overwrite the first. This test reproduces the race deterministically: another writer's `.md` is already sitting at the number `nextNoteSeq` picks. It forces an exclusive create plus a retry.

## Scope
- `task/feedback.go`:
  - unexported `writeNote(track string, seq int, body string, head Commit) (string, error)`. It creates `<TRACK>-NNN.md` with `os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, fileMode)`, then writes the `.json` as BIT-46.17 does. A clash returns an error that wraps `fs.ErrExist`. **Seam:** BIT-49.12's `ImportNote` exports this.
  - `AddNote`: `seq := nextNoteSeq(track)`, then call `writeNote` in a loop. On `errors.Is(err, fs.ErrExist)`, take `seq+1` and try again. Any other error returns. The loop ends because each retry moves to a higher number.
  - `.md` is still written first (BIT-46's format rule). A stray `.md` left by a crash only skips its number.
- Test file touched: `task/feedback_test.go`. BIT-46.17 created it in shape (`TestStoreAddNote`), so there's no conversion step.

## TDD cycle

0. **Restructure:** none. `task/feedback_test.go` was created in shape by BIT-46.17.

1. **Write test (RED):**
   - [ ] `TestStoreAddNote/skips a number another writer claimed` (`task/feedback_test.go`)
     - **Behavior:** a note never replaces a note another session is writing at the same moment.
     - **Setup:** shared-root store `BIT` with track `BIT-49` and one note added (`BIT-49-001`). Then hand-write `feedback/BIT-49-002.md` with `other session` and no `.json` (another writer mid-write). Call `AddNote("BIT-49", "mine", Commit{})`.
     - **Assertions:** the returned path ends in `BIT-49-003.md` and holds `mine`. `BIT-49-002.md` still holds `other session`. `BIT-49-003.json` has `"seq": 3`.
     - **Boundary:** the number `nextNoteSeq` picks (2, because records count only `.json`) is already taken on disk.
   - [ ] Confirm fails: `BIT-49-002.md` is overwritten with `mine`.

2. **Implement (GREEN):**
   - [ ] `writeNote` with `O_EXCL`, plus the retry loop in `AddNote`.

3. **More tests (RED → GREEN):**
   - [ ] The existing `TestStoreAddNote` cases (BIT-46.17's) still pass: first note seq 1, second note seq 2, and a stray `BIT-46-003.md` doesn't affect seq 2. *Boundary:* no clash, no retry.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`fix(bit): a feedback note never overwrites another in the shared folder`