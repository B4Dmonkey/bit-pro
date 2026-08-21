---
id: BIT-32.1
title: A tick writes counts onto every project row
status: todo
approved: true
phase: 1
phase_label: Counts in the DB
---
## **Verse 1**

The tick gains a per-project write: for each registered project it stores four count columns on the project row. The counts are hardcoded here — the columns and the write path are what this step proves, and Step 2's fixture is what forces them to be real.

## Scope
- `db/migrations/<timestamp>_project_counts.sql` — new migration adding `backlog`, `todo`, `done`, `completed`
- `db/queries/projects.sql` — new `UpdateProjectCounts` query
- `cmd/serve.go` — open the DB once before the loop; on each tick, list projects and write counts
- `cmd/serve_test.go` — new test, plus `HOME`/`XDG_DATA_HOME` isolation on the existing tests

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestServeCmd_WritesProjectCounts`
     - **Behavior:** a tick reaches every registered project row and stores counts there, so `bp list` and `bp status` have something cached to read.
     - **Setup:** `t.Setenv("HOME", t.TempDir())` and `t.Setenv("XDG_DATA_HOME", "")`. Build a project directory with `dir := t.TempDir()`, then `task.New(filepath.Join(dir, ".bit")).Save(&task.Task{ID: "ACME-1", Title: "a track", Status: task.StatusTodo})` — one unapproved track, no bars. Register it with the existing `seedProject` helper as `orm.CreateProjectParams{Path: dir, Code: "ACME"}`. Set `serveTick = 5 * time.Millisecond` with the save/restore `t.Cleanup` the existing serve tests use, and run `runWithContext` with a 50ms timeout on `serveCmdUse`.
     - **Assertions:** re-open the DB and `QueryRow("SELECT backlog, todo, done, completed FROM projects WHERE code = ?", "ACME")` scans `1, 0, 0, 0`.
     - **Boundary:** one registered project and one track — the lower bound above empty. Proves the write happens at all; it says nothing yet about classification.
   - [ ] Confirm fails: `no such column: backlog` from the assertion query — the migration doesn't exist yet.

2. **Implement (GREEN):**
   - [ ] `just db-migrate project_counts`, then fill the generated file: `-- migrate:up` adds four columns, each `ALTER TABLE projects ADD COLUMN <name> INTEGER NOT NULL DEFAULT 0;`; `-- migrate:down` drops the four.
   - [ ] Add to `db/queries/projects.sql`:
         `-- name: UpdateProjectCounts :exec` / `UPDATE projects SET backlog = ?, todo = ?, done = ?, completed = ? WHERE id = ?`
   - [ ] In `cmd/serve.go`'s `RunE`, call `db.Open()` once before the ticker, `defer sqlDB.Close()`, and build `queries := orm.New(sqlDB)`. Log nothing on open — `TestServeCmd_LogsStartAndStop` asserts the log is exactly `started`, `stopped`.
   - [ ] In the `<-ticker.C` branch, after `log.Debug("tick")`, call `queries.ListProjects(ctx)` and for each project call `UpdateProjectCounts` with the hardcoded values `1, 0, 0, 0` and that project's `ID`. A `ListProjects` or update error logs at error level and the loop continues to the next tick.
   - [ ] Add `t.Setenv("HOME", t.TempDir())` and `t.Setenv("XDG_DATA_HOME", "")` to every existing test in `cmd/serve_test.go` that runs the command (`TestServeCmd_ReturnsWhenContextCancelled`, `TestServeCmd_TicksOnlyWhenVerbose`, `TestServeCmd_LogsJSONWhenOutputIsNotATerminal`, `TestServeCmd_LogsStartAndStop`). Without it `bp serve` now opens the developer's real `~/.local/share/bit-pro/bit.db` during the test run.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`
- [ ] `just db-up && just db-down && just db-up` round-trips against the throwaway `db/bit.db` — SQLite only supports `ALTER TABLE ... DROP COLUMN` from 3.35, so the down direction has to be run, not assumed.

## User verifies
- [ ] none — deterministic.

## Commit (user)
`feat(db): store per-project counts on each daemon tick`