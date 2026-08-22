---
id: BIT-33.1
title: Migration + queries drive the orm
status: done
approved: true
phase: 1
phase_label: Queue table exists
---
## **Verse 1**

Compilation failure on missing ORM functions forces the migration, query files, and `sqlc generate` to exist.

## Scope
- `db/migrations/20260821000001_create_queue.sql` — new `queue` table with `id`, `project_id` FK → `projects.id`, `target_id TEXT NOT NULL`, `target_typ TEXT NOT NULL`
- `db/queries/queue.sql` — `EnqueueTask :exec` (INSERT) and `ListQueueByProject :many` (SELECT WHERE project_id ORDER BY id)
- `db/queries/projects.sql` — add `GetProjectByPath :one` (`SELECT id, path, code FROM projects WHERE path = ?`); `ProjectExists` returns a bool and cannot return the `id` needed for queue FK resolution
- `db/orm/` — run `sqlc generate` to regenerate; new files: `queue.sql.go`, updated `models.go`
- `db/queue_test.go` — new integration test file

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestQueue_EnqueueAndList`
     - **Behavior:** inserting a queue row for a project makes it appear in `ListQueueByProject` for that project
     - **Setup:** `t.TempDir()` + `t.Setenv("HOME", home)` + `t.Setenv("XDG_DATA_HOME", "")` + `db.Open()`; seed one project via `orm.New(sqlDB).CreateProject(ctx, orm.CreateProjectParams{Path: "/tmp/proj", Code: "TST"})`; then `GetProjectByPath(ctx, "/tmp/proj")` → `project.ID`; call `EnqueueTask(ctx, orm.EnqueueTaskParams{ProjectID: project.ID, SubjectID: "BIT-33", SubjectKind: "track"})`
     - **Assertions:** `ListQueueByProject(ctx, project.ID)` returns slice of length 1; `rows[0].SubjectID == "BIT-33"`; `rows[0].SubjectKind == "track"`
     - **Boundary:** queue starts at 0 rows — proves a single insert changes the count from 0 to 1
   - [ ] Confirm fails: `undefined: orm.EnqueueTask` (compile error — ORM functions don't exist yet)

   - [ ] `TestQueue_ListEmpty`
     - **Behavior:** `ListQueueByProject` on a project with no queue rows returns an empty slice, not an error
     - **Setup:** same temp-dir pattern; seed project; do not enqueue anything
     - **Assertions:** result is a non-nil empty slice; `err == nil`
     - **Boundary:** 0 queue rows for this project (the lower bound for the list path)

2. **Implement (GREEN):**
   - [ ] Write `db/migrations/20260821000001_create_queue.sql` following existing migration format (`-- migrate:up` / `-- migrate:down`)
   - [ ] Write `db/queries/queue.sql` with `EnqueueTask :exec` and `ListQueueByProject :many`
   - [ ] Add `GetProjectByPath :one` to `db/queries/projects.sql`
   - [ ] Run `sqlc generate` to regenerate `db/orm/`
   - [ ] Run `just install` after code changes

## Claude verifies
- [ ] `go test ./db/...` passes
- [ ] `go build ./...` passes (no compile errors in the new ORM)

## User verifies
- none — deterministic

## Commit (user)
`feat(db): add queue table migration, sqlc queries, and GetProjectByPath`