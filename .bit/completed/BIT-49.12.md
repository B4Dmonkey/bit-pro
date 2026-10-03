---
id: BIT-49.12
title: migrate moves feedback notes into the shared folder with their numbers kept
status: done
approved: true
phase: 2
phase_label: bp migrate
---
## **Verse 2**

v1 notes are `feedback/<TRACK>-NNN.md`, with no frontmatter. `AddNote` would renumber them from 1, so a test that expects `BIT-19-002` to still be `002`, and the next note to be `003`, forces a write that keeps the number.

## Scope
- `task/feedback.go`: `func (s *Store) ImportNote(track string, seq int, body string, head Commit) (string, error)`. It resolves the track (`resolveTrack`; an archived track is accepted, `trackExists`), then calls BIT-49.2's `writeNote` exclusively. A clash is an error, never a retry. **Seam:** the only seq-preserving writer.
- `migrate/migrate.go`: after the tasks, for each `src/feedback/*.md` matching `^(.+)-(\d+)\.md$`, call `ImportNote(track, seq, string(raw), task.Commit{})`. Tasks are written first, so every note's track exists in the store. A note whose track isn't in the source fails `resolveTrack`, and that error stops the migration. From BIT-49.16 on, it stops with nothing left behind.
- Test file touched: `cmd/migrate_test.go`.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestMigrateCmd/moves feedback notes with their numbers`
     - **Behavior:** a project's feedback history comes across intact, and the next note continues the numbering.
     - **Setup:** sandbox. The v1 store has `archive/tasks/BIT-19.md` and `tasks/BIT-44.md`. `feedback/BIT-19-001.md` and `BIT-19-002.md` hold realistic bodies (`## What the plan said … ## What happened`). `feedback/BIT-44-005.md` stands alone, with no 001–004. Then `bp migrate`, then `bp feedback add BIT-19 -d "next"`.
     - **Assertions:**
       - `<data>/bit/feedback/BIT-19-001.md`, `BIT-19-002.md` and `BIT-44-005.md` are byte-equal to the sources, and each `.json` has `"project": "BIT"` and the right `track` and `seq`.
       - The new note's path ends in `BIT-19-003.md`.
       - `<data>/bit/BIT/feedback` doesn't exist.
     - **Boundary:** a note on an archived track, and a gap in the numbers (005 alone).
   - [ ] Confirm fails: no `<data>/bit/feedback/BIT-19-001.md`.

2. **Implement (GREEN):**
   - [ ] `ImportNote` and the feedback loop.

3. **More tests (RED → GREEN):**
   - [ ] `TestMigrateCmd/a note on a missing track stops the migration`: `feedback/BIT-9-001.md` with no `BIT-9` task anywhere → non-nil error that names `BIT-9`. *Boundary:* an orphan note.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): migrate moves feedback notes into the shared folder`