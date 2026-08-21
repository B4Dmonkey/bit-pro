---
id: BIT-32.2
title: Contradiction forces the tick to read the ledger
status: done
approved: true
phase: 1
phase_label: Counts in the DB
---
## **Verse 1**

A fixture with one unapproved and two approved tracks can't be satisfied by Step 1's hardcoded `1, 0, 0, 0`, which forces the tick to actually read each project's `.bit/` ledger and split tracks between the backlog and todo buckets.

## Scope
- `task/counts.go` — new: `Counts` struct and `(*Store).Counts()`
- `cmd/serve.go` — the tick calls `Counts()` per project instead of using literals
- `cmd/serve_test.go` — the contradicting test

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeCmd_CountsBacklogAndTodo`
     - **Behavior:** the stored counts are derived from the project's ledger, and only tracks are counted — a bar under a track is a plan step, not work waiting.
     - **Setup:** same isolation and 5ms-tick harness as Step 1. In one project directory's `.bit/`, save four records with `task.New(...).Save`: `ACME-1` (`StatusTodo`, `Approved: false`), `ACME-2` (`StatusTodo`, `Approved: true`), `ACME-3` (`StatusTodo`, `Approved: true`), and `ACME-3.1` (`StatusTodo`, `Approved: true`) — the last is a bar under `ACME-3`. Register the directory as code `ACME`.
     - **Assertions:** the row scans `backlog = 1`, `todo = 2`, `done = 0`, `completed = 0`.
     - **Boundary:** `todo` count > 1 — exercises accumulation across tracks rather than a single hit, which is what kills the literal `1`. `ACME-3.1` is the dotted-ID case that must be excluded, so `todo` is 2 and not 3.
   - [ ] Confirm fails: `todo = 0, want 2` (and `backlog = 1` passes by coincidence of the hardcode) — Step 1 writes literals and never opens `.bit/`.

2. **Implement (GREEN):**
   - [ ] New `task/counts.go`: `type Counts struct { Backlog, Todo, Done, Completed int }`.
   - [ ] `func (s *Store) Counts() (Counts, error)`: call `s.List()`, and for each task skip it when `barParent(t.ID)` reports it is a bar. For the tracks that remain, apply the first-match chain — `!t.Approved` increments `Backlog`, otherwise `Todo`. `Done` and `Completed` stay zero for now; Steps 3 and 4 force them.
   - [ ] In `cmd/serve.go`'s tick, replace the literals with `task.New(filepath.Join(p.Path, ".bit")).Counts()` and pass the result's fields to `UpdateProjectCounts`. An error from `Counts()` is returned/logged for now — Step 5 settles what the loop does with it.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(task): derive project counts from the ledger`