---
id: BIT-32.7
title: bp status renders three counts in every daemon state
status: done
approved: true
phase: 3
phase_label: bp status
---
## **Verse 3**

`bp status` reports the daemon and nothing else, and returns early once it has printed a pid. A fixture asserting a project table under all three daemon states forces the table to render independently of daemon health, and to omit `completed` — the one column BIT-32 keeps off this command.

## Scope
- `cmd/status.go` — print a project count table below the state line
- `cmd/status_test.go` — the new test, plus `HOME`/`XDG_DATA_HOME` isolation on the existing three

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestStatusCmd_ShowsProjectCounts` (table-driven subtests)
     - **Behavior:** the count table describes project work, not daemon health, so it appears whatever `launchctl` says — including when nothing is running and there is no daemon to attribute the numbers to.
     - **Setup:** `HOME`/`XDG_DATA_HOME` isolation. Seed two projects and write distinct counts with `UpdateProjectCounts` — `ACE` gets `2, 1, 4, 7`, `MID` gets `0, 3, 12, 2`. Three subtests, each supplying the `lc` fake the existing tests already use: `launchctlDict` for *running*, `nothingLoaded` for *not running*, and a `disabledStore` entry for *stopped*.
     - **Assertions:** in every subtest the normalized output equals the state line followed by `ACE backlog:2 todo:1 done:4 MID backlog:0 todo:3 done:12` — the state line being `running (pid 4242)`, `not running`, or `stopped` respectively. Also assert the output does **not** contain `completed:`.
     - **Boundary:** the daemon state swept across all three of its values, with the same project fixture each time — `running` is the case that currently `return`s before anything else can print, so it is the one that proves the early return is gone. The absent `completed:` is the omitted-column bound; `MID`'s `backlog: 0` renders rather than blanks.
   - [ ] Confirm fails: the *running* subtest's output is `running (pid 4242)` and nothing more.

2. **Implement (GREEN):**
   - [ ] Add `t.Setenv("HOME", t.TempDir())` and `t.Setenv("XDG_DATA_HOME", "")` to `TestStatusCmd_ReportsNotRunning`, `TestStatusCmd_ReportsWhatLaunchctlSays`, and `TestStatusCmd_ReportsStoppedFromTheDisabledStore`. `bp status` now opens the database, and without isolation those tests would read the developer's real registry and print its projects into the assertion.
   - [ ] In `cmd/status.go`, drop the early `return nil` in the `StateRunning` branch so both paths fall through to the same table.
   - [ ] After the state line, `db.Open()`, `defer Close`, `ListProjects`, and print one row per project as `"  %s\tbacklog:%d\ttodo:%d\tdone:%d\n"`. Emit nothing at all — not even a separator — when there are no projects, so the three existing tests' exact expected output stays correct against an empty registry.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`
- [ ] `just install` — the User-verify below runs the real binary.

## User verifies
- [ ] Whole slice: run `bp stop` then `bp status`, and confirm `stopped` is followed by one row per registered project showing three counts and no `completed:` column. Then `bp start && bp status` and confirm the same table appears under `running (pid …)`.

## Commit (user)
`feat(status): show per-project counts below the daemon state`