---
id: BIT-17.2
title: A second note contradicts the hardcoded sequence
status: todo
phase: 1
phase_label: A durable record
---
## **Verse 1**

A second note under the same track cannot share `001` with the first, so the hardcoded sequence
dies here — a scan for the highest reserved number replaces it. This is the bar that makes capture
create-only rather than one overwritable slot.

## Scope
- `task/feedback.go` — replace the hardcoded seq with a scan. `highestSuffix(dir, glob string, re
  *regexp.Regexp) (int, error)` already exists unexported in `task/store.go` and is exactly this
  scan; call it with `s.feedbackDir()`, glob `track + "-*.md"`, and regex `^` +
  `regexp.QuoteMeta(track)` + `-(\d+)\.md$`, then write `highest+1` formatted `%s-%03d.md`.
- `cmd/feedback_add_test.go` — add the second-note test.

Scan `feedbackDir()` only — not `tasks/`, not `archive/`. Notes are never relocated, so unlike
`highestReserved` there is no second directory holding reserved note numbers. A missing
`feedback/` makes `filepath.Glob` return no matches and no error, so the first note still works.

The glob is per-track by construction: `BIT-1-*.md` cannot match `BIT-11-001.md`, and the regex
double-guards it. That is what keeps each track's notes numbered from 001 instead of sharing one
global counter.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestFeedbackAddCmd_SecondNoteGetsNextSequence`
     - **Behavior:** capture that fires twice in one cycle keeps both records — the write path
       cannot damage a note already on disk, which matters because capture fires right after a run
       went wrong.
     - **Setup:** `initProject(t, "BIT")`; `createTask(t, "Ship the bit plugin", "...")`; two
       `mustRun(t, "feedback", "add", "BIT-1", "-d", …)` calls with distinct bodies `first` and
       `second`.
     - **Assertions:** stdout of the second call is `.bit/feedback/BIT-1-002.md\n`;
       `os.ReadFile(".bit/feedback/BIT-1-002.md")` returns `second`; and
       `os.ReadFile(".bit/feedback/BIT-1-001.md")` still returns `first`, unchanged.
     - **Boundary:** note count 1 → 2 — the first value a hardcoded sequence cannot serve, and the
       point where create-only either holds or silently clobbers.
   - [ ] Confirm fails: `.bit/feedback/BIT-1-002.md` does not exist and `BIT-1-001.md` contains
     `second` — the overwrite the hardcoded sequence causes. A failure reporting the *first* file
     missing instead means the note landed somewhere else entirely.

2. **Implement (GREEN):**
   - [ ] Swap the hardcoded 1 for `highestSuffix` + 1.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(feedback): number notes so a second cannot overwrite the first`