---
id: BIT-29.3
title: The dev workflow rehearses migrations on a throwaway database
status: todo
approved: true
phase: 1
phase_label: registry
---
## **Verse 1**

Just targets drive the `dbmate` CLI against a throwaway repo-local database, so a migration can be
rehearsed and rolled back without touching the operator's real registry. Not test-driven — a
`Justfile` recipe has no seam a Go test can reach, so the checks below are the cycle.

## Scope
- `Justfile` — `db-migrate`, `db-up`, `db-down`, `db-status` and the `DATABASE_URL` they export.
  `db-gen-queries` already exists — BIT-29.2 added it, because the build depends on it.
- `.gitignore` — `db/bit.db`

Two properties matter more than the recipes themselves, and both are scope Decisions. The targets
point at a repo-local `db/bit.db`, never `~/.local/share/bit-pro/bit.db` — `db-down` against the
real file would delete live enrollments, and later, queued work. And there is deliberately no env
override: the real database is only ever migrated by the binary through the embedded FS, so the
`Justfile` must not reimplement `store.Dir()`'s XDG-or-`HOME` resolution.

## References
- `/Users/appstack/Developer/UniqueDataManagement/clients/engage-voters/evus/justfiles/db.just` —
  the pattern to follow. Take the target names and the `--no-dump-schema` /
  `--migrations-dir "{{MIGRATIONS_DIR}}"` shape; drop the `docker compose` indirection (dbmate is on
  PATH here) and the postgres URL assembly.

## Implementation
- [ ] `Justfile`: `MIGRATIONS_DIR := justfile_directory() / "db" / "migrations"` and
      `export DATABASE_URL := "sqlite:" + justfile_directory() / "db" / "bit.db"`.
- [ ] `Justfile`: `db-migrate name:` → `dbmate --migrations-dir "{{MIGRATIONS_DIR}}" new {{name}}`.
- [ ] `Justfile`: `db-up`, `db-down`, `db-status` → the matching `dbmate --no-dump-schema
      --migrations-dir "{{MIGRATIONS_DIR}}"` subcommand. `--no-dump-schema` on `up`/`down` for the
      same reason `AutoDumpSchema` is off in the binary: dbmate's sqlite `DumpSchema` shells out to
      a `sqlite3` binary.
- [ ] `.gitignore`: add `db/bit.db`.

## Claude verifies
- [ ] `just db-up` exits 0 and reports the `create_projects` migration applied.
- [ ] `just db-status` lists that migration as applied.
- [ ] `just db-down` exits 0, then `just db-up` again — the migration's `down` section is real and
      the pair round-trips.
- [ ] `git status --porcelain` lists neither `db/bit.db` nor anything under `db/orm/` — the
      rehearsal database and the generated layer are both build products.
- [ ] `just test` and `just lint`.

## User verifies
- [ ] Whole slice: with the daemon's own database in place, run `just db-down` and then
      `ls -l ~/.local/share/bit-pro/bit.db` — the file's size and mtime are unchanged. The
      rehearsal database is the one that moved. That is the declutter this verse is actually for: a
      migration can be developed and rolled back all day without risk to enrollments.
- [ ] `just db-migrate scratch` creates a new timestamped file under `db/migrations/`. Delete it
      afterwards — it is a smoke test of the authoring path, not a migration.

## Commit (user)
`chore(db): add the dbmate and sqlc dev workflow targets`