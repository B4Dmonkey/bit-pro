---
id: BIT-46.1
title: A fresh open creates bit/main.db with only a projects table
status: done
phase: 1
phase_label: registered project works from the central store
---
## **Verse 1**

The registry moves from v1's `~/.local/share/bit-pro/bit.db` (4 migrations, non-UNIQUE `code`) to `~/.local/share/bit/main.db` with one initial migration. The existing path/migration-count assertions in `db/open_test.go` and `store/store_test.go` can't pass against the old names, so they force it. Precondition: BIT-45 has landed (no queue queries, `ListProjects` without counts, `GetProjectByPath` and `UpdateProjectCounts` gone).

## Scope
- `store/store.go` — `Dir()` joins `"bit"` instead of `"bit-pro"`. Add `ProjectDir(code string) (string, error)` = `pathologize.Join(dir, code)` (the code comes from a db row, so the path is built with `Join`, the project's rule for paths from data). `ProjectDir` does not create the dir: the task store creates subdirs lazily on write (`task/store.go` `Save` → `MkdirAll`).
- `db/open.go` — file name `main.db`.
- `db/migrations/` — delete all four files; add one `<timestamp>_create_projects.sql` (`just db-migrate create_projects`):
  `projects(id INTEGER PRIMARY KEY, code TEXT NOT NULL UNIQUE, path TEXT NOT NULL UNIQUE)`. No `COLLATE NOCASE`: case-insensitive matching is done in Go (track Decision).
- `db/queries/projects.sql` — exactly: `CreateProject :exec` (path, code), `ProjectExists :one` (kept for now; `cmd/add.go:40` still calls it; removed in the `bp add` bar), `ListProjects :many` → `SELECT id, code, path FROM projects ORDER BY code`. `ListProjects` returns every row; later callers (resolver, add, BIT-49 migrate) filter in Go, so there is no `GetProjectByPath`/`GetProjectByCode` (an SQL `=` would be case-sensitive, against the Decision).
- `cmd/list.go` — only if the generated row type's field order changes; output stays `code\tpath`.
- `cmd/list_test.go` — `TestListCmd` (converted in BIT-45.4) stats `<home>/.local/share/bit-pro/bit.db` for every seeded row (`:59`); change it to `<home>/.local/share/bit/main.db`. Same cases otherwise.
- delete stale `db/orm/*.go` for removed queries by hand before `sqlc generate` (gitignored; generate never deletes).

## TDD cycle

0. **Convert existing tests (pure restructure, same cases and assertions, still green; skip a file BIT-45 already converted):**
   - [ ] `db/open_test.go`: `TestOpen_MigratesAFreshDatabase` → `TestOpen/migrates a fresh database` (subtest).
   - [ ] `store/store_test.go`: `TestDir_FollowsXDGDataHome` → `TestDir`, its two existing rows kept as table rows.
   - [ ] `db/queries_test.go`: `TestProjects_RoundTrip` → `TestCreateProject/round trips through list projects` (subtest).

1. **Write test (RED):**
   - [ ] `TestOpen/migrates a fresh database` (subtest; rewrite it in `db/open_test.go`)
     - **Behavior:** a first open on a clean machine creates the registry at the v2 location with the v2 schema and nothing from v1.
     - **Setup:** `t.Setenv("HOME", t.TempDir())`, `t.Setenv("XDG_DATA_HOME", "")`; `Open()`.
     - **Assertions:** `<home>/.local/share/bit/main.db` exists; `<home>/.local/share/bit-pro` does not; `sqlite_master` tables are exactly `{projects, schema_migrations}`; `schema_migrations` count == 1.
     - **Boundary:** migration count == 1 (down from 4) — proves the old set is gone, not appended to.
   - [ ] `TestDir` (`store/store_test.go`, table) — want `.../bit` in both rows (XDG set / unset).
   - [ ] `TestListCmd/three projects` (`cmd/list_test.go`) — the db stat wants `.local/share/bit/main.db`.
   - [ ] Confirm fails: path `bit-pro/bit.db` / count 4; `TestListCmd/three projects` can't stat `bit/main.db`.

2. **Implement (GREEN):**
   - [ ] `store.Dir` → `bit`; `db.Open` → `main.db`; replace migrations; rewrite `projects.sql`; regenerate.

3. **More tests (RED → GREEN):**
   - [ ] `TestCreateProject/refuses a duplicate code` (table row in `TestCreateProject`, `db/queries_test.go`)
     - **Behavior:** two projects can't share a code.
     - **Setup:** sandboxed HOME; `CreateProject{Path:"/tmp/alpha", Code:"ALPHA"}` then `{Path:"/tmp/beta", Code:"ALPHA"}`.
     - **Assertions:** second call returns a non-nil error; `ListProjects` has 1 row.
     - **Boundary:** duplicate `code`, distinct `path` — the constraint v1 lacked.
   - [ ] `TestCreateProject/refuses a duplicate path` (table row, same table) — same path, different codes → error, 1 row.
   - [ ] `TestProjectDir` (`store/store_test.go`, new): sandboxed `XDG_DATA_HOME`; `ProjectDir("BIT")` == `<data>/bit/BIT`, and the dir isn't created. Codes are validated before they reach the db (`bp add`, BIT-49's migrate), so there is no traversal-shaped-code case.

## Claude verifies
- [ ] `just lint` and `just test` pass (both run `sqlc generate` first)
- [ ] `test ! -e ~/.local/share/bit/main.db` after the run — no test touched the real store
- [ ] delete any dev sandbox stores (the initial migration changed)

## User verifies
- none — deterministic

## Commit (user)
`feat(bit): fresh v2 registry at bit/main.db with one initial migration`