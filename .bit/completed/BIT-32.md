---
id: BIT-32
title: Project counts
status: done
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
- **The rendered column format is not pinned; tests assert it loosely.** A count test collapses runs of whitespace to a single space and checks that the labelled numbers are present and in order, rather than asserting an exact byte string. The format is corrected on the fly under a `## User verifies` check, so pinning it in a test would only make cosmetic tweaks expensive.
- **Buckets are a first-match chain, approval before status: backlog → todo → done.** A track is counted once, in the first bucket it matches. A track that is unapproved *and* already `done` or `doing` is not expected to arise, so where the chain happens to place it is incidental rather than designed — see the open gap in `automation-notes.md`. No bar handles the overlap.
- **`bp status` prints the count table whatever the daemon's state.** The table describes project work, not daemon health, so it renders under `running`, `not running`, and `stopped` alike.
- **A project the loop cannot read is skipped for that tick.** A registered path with no `.bit/` directory, or an unparseable task file, leaves that project's stored counts untouched and the loop moves on to the next project. Skipping keeps one broken project from stalling the tick or zeroing counts that were previously correct — see the open gap in `automation-notes.md`.

## Verses
- [x] Verse 1 — Counts exist in the DB: `projects` gains `backlog`, `todo`, `done`, `completed` columns via a new migration, and the daemon loop populates them for every registered project each tick.
  Touches: `db/migrations/`, `db/queries/projects.sql`, `cmd/serve.go`

- [x] Verse 2 — `bp list` shows counts: each project row renders all four counts (backlog / todo / done / completed) alongside code and path.
  Touches: `cmd/list.go`

- [x] Verse 3 — `bp status` shows per-project counts: the status command adds a project count table (backlog / todo / done) below the daemon state line.
  Touches: `cmd/status.go`