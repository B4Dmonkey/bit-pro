---
id: BIT-17.4
title: A note outlives its track
status: todo
phase: 1
phase_label: A durable record
---
## **Verse 1**

The previous bar's guard looks only in `tasks/`, so it refuses a note about an archived track — and
an archived track is exactly what retro reads. A test demanding that note be accepted contradicts
that guard and forces it to treat `archive/` as a real home for a track. Two guard tests ride along
on the same commit, pinning the other half of the same decision: a note already on disk is
untouched by anything that happens to its track.

## Scope
- `task/feedback.go` — replace the single `os.Stat` with a `trackExists(track string) bool` (or an
  inline two-stat check) covering `s.Path(track)` **and** `s.archivePath(track)`. Both already
  exist in `task/store.go`; this mirrors how `highestReserved` treats the two directories as one
  ID space.
- `cmd/feedback_add_test.go` — the contradicting test plus the two guards.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestFeedbackAddCmd_AcceptsArchivedTrack`
     - **Behavior:** a track's notes stay writable after it is archived, because a finished track
       is the one retro reads — capture must not stop working at the moment the work is filed away.
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "Ship the bit plugin", "...")`;
       `mustRun(t, "task", "update", "BIT-1", "-s", "done")`; `mustRun(t, "task", "archive",
       "BIT-1")`; then `mustRun(t, "feedback", "add", "BIT-1", "-d", note)`.
     - **Assertions:** stdout is `.bit/feedback/BIT-1-001.md\n` and the file contains `note`.
     - **Boundary:** track location — `archive/` rather than `tasks/`, the second of the two
       directories a track can legitimately occupy and the one the previous bar's guard misses.
   - [ ] Confirm fails: `track BIT-1 does not exist`, because `s.Path("BIT-1")` no longer resolves
     after the archive relocated the file.

2. **Implement (GREEN):**
   - [ ] Widen the guard to accept a track found in either directory.

3. **More tests (guards, not a cycle — they pass as soon as the above is green):**
   - [ ] `TestFeedbackAdd_NoteSurvivesTrackRewrite`
     - **Behavior:** rewriting the scope after a correction cannot destroy the note describing that
       correction — the loss the Why is about.
     - **Setup:** create a track, add a note, then `mustRun(t, "task", "update", "BIT-1", "-d",
       "a wholesale rewritten scope body")`.
     - **Assertions:** `.bit/feedback/BIT-1-001.md` still exists with byte-identical content.
     - **Boundary:** the `tasks/` ⇄ `feedback/` directory split — a whole-body `-d` write, the
       widest write the CLI can make to a track, still touches nothing under `feedback/`.
   - [ ] `TestFeedbackAdd_NoteSurvivesTrackArchive`
     - **Behavior:** archiving relocates a track and its bars but leaves its notes in place, so
       notes outlive the track's active life.
     - **Setup:** create a track and a bar, add a note, mark both `done`, `task archive BIT-1`.
     - **Assertions:** `.bit/tasks/BIT-1.md` is gone and `.bit/archive/BIT-1.md` exists, while
       `.bit/feedback/BIT-1-001.md` is still there with identical content.
     - **Boundary:** the archive path's blast radius — `Relocate` globs `tasks/` only, and this
       fixes that as an invariant rather than a coincidence a later refactor is free to break.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`
- [ ] `just install`, so the `bp` on PATH has `feedback add` for the manual check below

## User verifies
- [ ] Whole slice: record a real note about something that actually went wrong in this cycle —
      `bp feedback add BIT-17 -d '<the real correction>'` — then open the file it names and read it
      cold. You should be able to tell from the file alone which track and bar it happened at and
      what the plan failed to specify. If you cannot, that is the gap Verse 2's skill exists to
      close, and it is worth knowing now rather than after.

## Commit (user)
`feat(feedback): keep notes writable and intact across a track's whole life`