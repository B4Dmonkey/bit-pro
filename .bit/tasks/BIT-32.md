---
id: BIT-32
title: Project counts
status: todo
---
## Why
`bp list` and `bp status` show a project path but no signal of health — an operator can't tell at a glance whether a project has work waiting, work approved and ready to run, work already finished, or work archived. Adding counts to both commands gives the context needed to decide where to direct the daemon without opening the TUI.

## Summary
Add `backlog`, `todo`, `done`, and `completed` integer columns to the `projects` table. The daemon loop computes and writes these counts each tick. `bp list` renders all four; `bp status` renders three (backlog / todo / done — archived work is omitted as it is no longer active).

## Visual aid
```
$ bp list
example   /Users/josiah/Developer/example   backlog:2  todo:1  done:4  completed:7
acme      /Users/josiah/Developer/acme       backlog:0  todo:3  done:12  completed:2

$ bp status
running (pid 4821)

  example   backlog:2  todo:1  done:4
  acme      backlog:0  todo:3  done:12
```

## Decisions
- **Read from SQLite, accept lag.** `bp list` reads the cached counts from the database; the daemon loop is the sole writer. Counts go stale between ticks and while the daemon is stopped — this is accepted.
- **backlog = unapproved tracks; todo = approved, not-yet-done tracks; done = tracks with status `done`; completed = tracks archived via `bp task complete` (filed under `.bit/completed/`).** Four distinct states.
- **`bp list` shows all four counts; `bp status` shows three (backlog / todo / done).** Archived work is active-project context on list but not relevant to daemon health on status.
- **Migration adds four NOT NULL INTEGER columns defaulting to 0.** Existing rows stay valid without a data migration.
- **Efficiency of the refresh is a planning-time question.** Do not pre-optimise the implementation here.

## Verses
- [ ] Verse 1 — Counts exist in the DB: `projects` gains `backlog`, `todo`, `done`, `completed` columns via a new migration, and the daemon loop populates them for every registered project each tick.
  Touches: `db/migrations/`, `db/queries/projects.sql`, `cmd/serve.go`

- [ ] Verse 2 — `bp list` shows counts: each project row renders all four counts (backlog / todo / done / completed) alongside code and path.
  Touches: `cmd/list.go`

- [ ] Verse 3 — `bp status` shows per-project counts: the status command adds a project count table (backlog / todo / done) below the daemon state line.
  Touches: `cmd/status.go`