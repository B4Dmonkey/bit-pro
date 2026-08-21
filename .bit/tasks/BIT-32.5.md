---
id: BIT-32.5
title: An unreadable project is skipped, not fatal
status: todo
approved: true
phase: 1
phase_label: Counts in the DB
---
## **Verse 1**

Every step so far assumes each registered project's ledger is readable. A project whose task file won't parse currently takes the whole tick down with it, or writes zeros over counts that were previously correct. BIT-32's Decisions settle this as *skip the project*, so a two-project fixture where the broken one is visited first forces the loop to carry on.

## Scope
- `cmd/serve.go` — the tick logs and continues past a project it can't read
- `task/counts.go` — a missing `.bit/` root is an error, not an empty ledger
- `cmd/serve_test.go` — the contradicting test

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeCmd_SkipsAProjectItCannotRead` (table-driven subtests)
     - **Behavior:** one unreadable project neither aborts the tick nor destroys its own stored counts. This is what keeps a project that was `bp add`ed and later moved or corrupted from silently zeroing a number the operator reads as fact.
     - **Setup:** two projects registered, named so `ListProjects`' `ORDER BY code` visits the broken one first — `AAA` (broken) and `ZZZ` (healthy). `ZZZ`'s `.bit/` holds one approved `StatusTodo` track. Immediately after seeding, write non-zero counts onto `AAA`'s row with a direct `UPDATE projects SET backlog = 9 WHERE code = 'AAA'`, so "untouched" is distinguishable from "coincidentally zero". Then the 5ms-tick / 50ms-timeout harness. Two subtests for how `AAA` is broken:
       1. *unparseable task file* — `os.WriteFile(<AAA>/.bit/tasks/AAA-1.md, []byte("not frontmatter"), 0o644)`.
       2. *no `.bit/` directory* — the registered path exists but is empty.
     - **Assertions:** in both subtests, `ZZZ` scans `todo = 1` (the loop reached it) and `AAA` still scans `backlog = 9` (its counts were left alone).
     - **Boundary:** the failing project sorts *first* — the lower bound of iteration order, and the only position that proves the error doesn't abort the projects behind it. `backlog = 9` is the non-zero pre-existing value that separates "skipped" from "recomputed as empty".
   - [ ] Confirm fails: subtest 1 — `AAA` scans `backlog = 0`, want `9` (the tick writes zeros or returns early on the parse error). Subtest 2 — same, because `filepath.Glob` on a missing `tasks/` yields no matches and no error, so a missing ledger currently reads as an empty one.

2. **Implement (GREEN):**
   - [ ] In `Counts()`, `os.Stat` the store root first and return a wrapped error when it doesn't exist. Without this, subtest 2 can't be told apart from a project whose ledger is genuinely empty — and "no ledger" and "empty ledger" are different facts, only one of which should overwrite a stored count.
   - [ ] In `cmd/serve.go`'s tick, when `Counts()` errors for a project, log at warn with the project's code and path and `continue` to the next project rather than returning.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(serve): skip a project whose ledger cannot be read`