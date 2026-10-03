---
id: BIT-49.11
title: migrate carries completed/ and archive/ across as they are, so their IDs stay reserved
status: done
approved: true
phase: 2
phase_label: bp migrate
---
## **Verse 2**

`Save` only writes to `tasks/`. A v1 store with finished and deleted work must land in `completed/` and `archive/tasks/`, with statuses, split trees and partial `order` lists unchanged. Otherwise old IDs get re-minted. A test whose next minted ID depends on those dirs forces a save into a given place.

## Scope
- `task/store.go`:
  - `type Place int` with `const ( Active Place = iota; Completed; Archived )`, and an unexported `(s *Store) placeDir(p Place) string` mapping to `tasksDir`/`completedDir`/`archiveTasksDir`.
  - `func (s *Store) SaveTo(p Place, t *Task) error` holds the body of today's `Save` (stamping, `.md` then `.json`) aimed at `placeDir(p)`. `Save(t)` becomes `SaveTo(Active, t)`. Nothing else changes: `order` is written verbatim and never repaired, and the status isn't checked against the place.
  - **Seam:** BIT-50 can file into `Completed` with it if needed.
- `migrate/migrate.go`: one loop over `[]struct{ dir string; place task.Place }{{"tasks", Active}, {"completed", Completed}, {filepath.Join("archive", "tasks"), Archived}}`. A missing dir is skipped.
- Test file touched: `cmd/migrate_test.go`, already in shape.

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestMigrateCmd/keeps completed and archived records where they were`
     - **Behavior:** finished and deleted work comes across in its own place, so its IDs stay reserved and the active list stays clean.
     - **Setup:** sandbox. The v1 store mirrors bit-pro's real shapes:
       - `tasks/BIT-40.md` (todo).
       - `completed/BIT-39.md`: a track with status `doing` and an `order` that lists `BIT-39.1` but omits `BIT-39.13`.
       - `completed/BIT-39.1.md` (done).
       - `archive/tasks/BIT-39.13.md` (todo): a bar whose track is in `completed/`.
       - `completed/BIT-10.md` with `order: [BIT-10.1]` while `completed/BIT-10.1.md` and `completed/BIT-10.9.md` both exist.
       Then `bp migrate`.
     - **Assertions:**
       - `<data>/bit/BIT/completed/BIT-39.json` has `"status": "doing"` and `"order": ["BIT-39.1"]`.
       - `<data>/bit/BIT/archive/tasks/BIT-39.13.json` exists and `completed/BIT-39.13.json` doesn't.
       - `completed/BIT-10.json` has `"order": ["BIT-10.1"]`, and `BIT-10.9.json` exists.
       - `bp task list` shows only `BIT-40`.
     - **Boundary:** a split tree, a partial `order`, and a non-`done` record in `completed/`.
   - [ ] Confirm fails: `completed/BIT-39.json` doesn't exist (only `tasks/` is copied).

2. **Implement (GREEN):**
   - [ ] `Place`, `SaveTo`, `Save` delegating to it, and the place loop in `migrate.Run`.

3. **More tests (RED → GREEN):**
   - [ ] `TestMigrateCmd/an archived track's id is never re-minted`: a store with `tasks/BIT-2.md` plus `archive/tasks/BIT-7.md` → after the migration, `bp task create "T"` prints `BIT-8`. *Boundary:* the highest reserved ID lives only in `archive/`.

## Claude verifies
- [ ] `just lint` and `just test` pass

## User verifies
- none, deterministic

## Commit (user)
`feat(bit): migrate carries completed and archived records across unchanged`