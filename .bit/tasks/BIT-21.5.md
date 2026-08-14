---
id: BIT-21.5
title: Feedback filenames flip without losing note contents
status: done
phase: 1
phase_label: Migration
---
## **Verse 1**

The fifth carrier, and the one that makes this migration urgent rather than tidy: feedback note
filenames embed the track ID (`BIT-19-001.md`). Measured during planning — with the first four
carriers migrated and `feedback/` left alone, `bp feedback add BIT-1` overwrote the contents of
the existing `bit-1-001.md` while leaving its filename intact, destroying the note. Skipping
this step would make the migration itself cause the loss it exists to prevent.

## Scope
- `update/normalize.sh` — rename `.bit/feedback/<track>-NNN.md` so the track portion is uppercase.
- `update/normalize_test.sh` — fixture note with recognisable contents.

## TDD cycle

1. **Write test (RED):**
   - [ ] `test_feedback_note_filenames_are_uppercased_contents_intact`
     - **Behavior:** a note's filename picks up the uppercase track, and the note body survives
       byte for byte.
     - **Setup:** `.bit/feedback/bit-1-001.md` containing the single line `FIRST NOTE`, and
       `.bit/feedback/bit-1-002.md` containing `SECOND NOTE`. Run the script.
     - **Assertions:** `BIT-1-001.md` and `BIT-1-002.md` both exist; their contents are exactly
       `FIRST NOTE` and `SECOND NOTE` respectively; `.bit/feedback/` still holds exactly 2 files;
       the zero-padded sequence suffix is unchanged (`-001`, not `-1`).
     - **Boundary:** two notes on the same track — the case where a rename that collapsed the
       sequence number, or renamed onto an existing name, would destroy one. Note bodies are not
       parsed, so contents are asserted as opaque bytes.
   - [ ] Confirm fails: `.bit/feedback/` still contains `bit-1-001.md`

2. **Implement (GREEN):**
   - [ ] Uppercase the track portion of each `.bit/feedback/*.md` filename, preserving the
     `-NNN` suffix and `.md`. Reuse the temp-name rename helper.

## Claude verifies
- [ ] `bash update/normalize_test.sh` exits 0
- [ ] `shellcheck update/normalize.sh update/normalize_test.sh` reports no errors

## User verifies
- [ ] none — deterministic

## Commit (user)
`feat(update): uppercase feedback note filenames`