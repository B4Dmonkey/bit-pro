---
id: BIT-10.3
title: Unfinished bars refuse the move
status: done
phase: 1
phase_label: Archive
---
Refuse to relocate a track that still has unfinished bars, so live work is never torn down. Forces the all-done guard ahead of any file move.

**Scope:**
- `task/store.go` — add `UnfinishedBarsError` (`Bars []string`; `Error()` lists them); `Relocate` gathers non-`done` children and returns the error *before moving anything*. `force` is not consulted yet — no test demands it (bar 1.4 does).
- `task/store_test.go` — new test.

**TDD cycle:**

1. **Write test (RED):**
   - [ ] `TestStoreRelocate_RefusesWithUnfinishedBars`
     - **Behavior:** a track holding a non-`done` bar is left untouched, and a typed error names the offender.
     - **Setup:** track `BIT-1` (done), bar `BIT-1.1` (done), bar `BIT-1.2` (todo); `err := s.Relocate("BIT-1", false)`.
     - **Assertions:** `errors.As(err, **UnfinishedBarsError)` true and `.Bars` contains `"BIT-1.2"`; all three files still in `tasks/`; `archive/` empty.
     - **Boundary:** exactly one bar not `done`, `force=false` — the refuse case (contrast 1.2's all-done).
   - [ ] Confirm fails: bar 1.2's impl relocates regardless → files move, no error.

2. **Implement (GREEN):**
   - [ ] Load children; collect ids whose `Status != "done"`; if any exist, return `&UnfinishedBarsError{Bars: …}` before any `relocate` (guard runs first, unconditionally for now).

**Claude verifies:**
- [ ] tests pass (`just test`)
- [ ] linter passes (`just lint`)

**Commit (user):** `feat(task): refuse relocate when bars are unfinished`