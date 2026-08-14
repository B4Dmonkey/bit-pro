---
id: BIT-21.12
title: Contradiction forces uppercase filenames on the write path
status: done
phase: 2
phase_label: Recurrence
---
## **Verse 2**

Second symptom, and a direct contradiction of the previous bar: normalizing the *comparison* made
the guard fire, but the moment a track legitimately completes, the write path still builds its
destination from the raw argument. A lowercase `complete` on a finished track therefore files
`completed/bit-1.md`. Only normalizing the path construction can satisfy both this test and the
last one.

## Scope
- `task/store.go` — `Path`, `completedPath`, `archivePath`, and the destination built inside
  `relocateInto`.
- `cmd/task_complete_test.go` — the new test.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestTaskCompleteCmd_LowercaseTrackFilesUppercaseFilenames`
     - **Behavior:** a completed track and its bars land under their canonical uppercase
       filenames regardless of the case used to complete them.
     - **Setup:** `initProject(t, "BIT")`; create `BIT-1` and bars `BIT-1.1`, `BIT-1.2`; set both
       bars and the track to `done`. Run `task complete bit-1`.
     - **Assertions:** `.bit/completed/BIT-1.md`, `.bit/completed/BIT-1.1.md` and
       `.bit/completed/BIT-1.2.md` all exist; reading `.bit/completed/` back yields exactly three
       entries and none of their names contains a lowercase letter; `.bit/tasks/` is empty.
     - **Boundary:** a track plus two bars — the relocation walks children and the track through
       two different call paths (`relocateInto` per kid, then the track), so more than one bar is
       what distinguishes normalizing the loop from normalizing only the final move.
   - [ ] Confirm fails: the files land as `.bit/completed/bit-1.md` and friends — lowercase names,
     which on this filesystem also means a later uppercase write would silently overwrite them

2. **Implement (GREEN):**
   - [ ] Apply `normalizeID` in `Path`, `completedPath` and `archivePath`, and to the destination
     `pathologize.Join(dir, id+".md")` inside `relocateInto`. Normalize before joining, so
     `pathologize`'s sanitisation still runs on the final segment.

## Claude verifies
- [ ] `just test` passes
- [ ] `just lint` passes

## User verifies
- [ ] none — deterministic

## Commit (user)
`fix(task): build task paths from the canonical ID case`