---
id: BIT-32.6
title: bp list renders the four counts
status: done
approved: true
phase: 2
phase_label: bp list
---
## **Verse 2**

`bp list` prints code and path only, so the counts the tick now stores are invisible. Two projects with different counts in one fixture forces the four columns to come from each row rather than from a literal.

## Scope
- `db/queries/projects.sql` — widen `ListProjects`' SELECT to the four count columns
- `cmd/list.go` — render the counts
- `cmd/list_test.go` — the new test, and the existing exact-output test updated
- `cmd/cmd_test.go` — new shared whitespace-normalizing helper

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestListCmd_ShowsProjectCounts`
     - **Behavior:** every project row carries its four cached counts, so an operator can see from `bp list` alone which project has work waiting.
     - **Setup:** `HOME`/`XDG_DATA_HOME` isolation as in the existing list test. Seed two projects with `seedProject`, then write distinct counts onto each with `orm.New(sqlDB).UpdateProjectCounts` — `ACE` gets `2, 1, 4, 7` and `MID` gets `0, 3, 12, 2`. Run `bp list`.
     - **Assertions:** the output, passed through the new normalizer, equals `"ACE /tmp/ace backlog:2 todo:1 done:4 completed:7 MID /tmp/mid backlog:0 todo:3 done:12 completed:2"` — comparing on the normalized string, not the raw bytes, per BIT-32's decision that the column format isn't pinned.
     - **Boundary:** two rows with *different* counts — a single row could be satisfied by hardcoded literals, two cannot. `MID`'s `backlog: 0` is the zero bound (a project with nothing waiting still renders the column rather than blanking it), and `done: 12` is a two-digit value, which is where a fixed-width format assumption would show up.
   - [ ] Confirm fails: the output has no `backlog:` in it at all. Note that `orm.Project` already *has* the four fields — the migration in Verse 1 generated them — so this is an assertion failure, not a compile error.

2. **Implement (GREEN):**
   - [ ] Add `normalizeSpaces` to `cmd/cmd_test.go`: collapse every run of whitespace (including newlines) to a single space and trim the result. Shared, because the status bar needs it too.
   - [ ] In `db/queries/projects.sql`, widen `ListProjects` to `SELECT id, path, code, backlog, todo, done, completed FROM projects ORDER BY code`. `cmd/list.go` is the only caller.
   - [ ] In `cmd/list.go`, change the `Fprintf` to `"%s\t%s\tbacklog:%d\ttodo:%d\tdone:%d\tcompleted:%d\n"`. Tabs keep it consistent with what the command already emits; the format is provisional and the User-verify below is where it gets adjusted.
   - [ ] Update `TestListCmd_PrintsProjectsByCode`'s `want` for the new columns (all counts zero, since it never writes any) and compare it through `normalizeSpaces` as well.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`
- [ ] `just install` — the User-verify below runs the real binary.

## User verifies
- [ ] Whole slice: with the daemon running (or `bp serve` in a terminal long enough for one tick), run `bp list`. Pick one project and check its four numbers against `bp task list` inside it — of the rows whose ID has no dot: unapproved ones are `backlog`, approved-and-not-`done` are `todo`, `done` ones are `done`, and `completed` is what's under that project's `.bit/completed/`. Then say whether the columns are readable at your terminal width — the format is deliberately not pinned in a test, so this is where it gets changed.

## Commit (user)
`feat(list): render per-project counts`