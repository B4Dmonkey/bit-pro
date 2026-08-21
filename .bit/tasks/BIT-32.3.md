---
id: BIT-32.3
title: Contradiction splits done out of todo
status: todo
approved: true
phase: 1
phase_label: Counts in the DB
---
## **Verse 1**

Step 2 counts every approved track as `todo`, so an approved track already at `done` lands in the wrong bucket. A table over the three statuses contradicts that and forces the `done` branch into the chain.

## Scope
- `task/counts_test.go` — new: table-driven bucket test
- `task/counts.go` — the `done` branch

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestStoreCounts_Buckets` (table-driven subtests)
     - **Behavior:** each track lands in exactly one bucket, and approval is checked before status — the recorded rule from BIT-32's Decisions.
     - **Setup:** per subtest, `s := task.New(filepath.Join(t.TempDir(), ".bit"))` and one saved track. Cases: `{approved, StatusTodo}`, `{approved, StatusDoing}`, `{approved, StatusDone}`, `{unapproved, StatusTodo}`.
     - **Assertions:** the returned `Counts` equals, in order, `{Todo: 1}`, `{Todo: 1}`, `{Done: 1}`, `{Backlog: 1}` — every other field zero, asserted as a whole-struct comparison so a track double-counted in two buckets fails.
     - **Boundary:** `Status` swept across all three of its values with `Approved` held true. `StatusDone` is the boundary the "not-yet-done" clause in the todo definition bites at; `StatusDoing` is the value none of the four bucket names mention, and it must fall to `todo` rather than nowhere.
   - [ ] Confirm fails: the `{approved, StatusDone}` subtest reports `{Todo: 1}`, want `{Done: 1}`.

2. **Implement (GREEN):**
   - [ ] In `Counts()`, extend the chain to three arms: `!t.Approved` → `Backlog`; else `t.Status == StatusDone` → `Done`; else → `Todo`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(task): split done tracks out of the todo count`