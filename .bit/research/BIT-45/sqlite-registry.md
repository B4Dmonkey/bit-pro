# The existing sqlite db

**Checked:** Q3 — what is db/, where does it live at runtime, how are migrations applied, how does it fit the sqlc + dbmate plan.

## What it is
A **machine-wide project registry + dispatch queue** for the launchd daemon (not a task store).
- `db/migrations/` (dbmate format, `-- migrate:up/down`): `projects(id, path UNIQUE, code)`, then `queue(id, project_id→projects, target_id, target_typ)`, then counts columns on projects (`backlog, todo, done, completed`), then `UNIQUE(project_id, target_id)` on queue.
- `db/queries/*.sql` → sqlc (`sqlc.yaml`, engine sqlite, package `orm`, out `db/orm/`, **gitignored — must be generated before build**; `just install/test/lint/run` all depend on `db-gen-queries`).
- Queries: CreateProject, ProjectExists, ListProjects, UpdateProjectCounts, GetProjectByPath; EnqueueTask (INSERT OR IGNORE), ListQueueByProject, DeleteQueueRow.
- Users: `bp add` (register), `bp list`, `bp status` (counts), `bp serve daemon` (Tick: refresh counts from each project's `.bit/`, dispatch queue head via `claude.Spawn`), `bp tui` (enqueue, optional — silently off if db/project lookup fails).

## Where it lives at runtime
`db.Open()` (`db/open.go`) → `store.Dir()` (`store/store.go`) = `$XDG_DATA_HOME/bit-pro/bit.db`, fallback `~/.local/share/bit-pro/bit.db`. `daemon.log` sits beside it. Live contents on this machine: projects `BIT` (bit-pro path) and `EX` (tools/example); queue empty. **A client project is not registered** even though it has the bit MCP server — the registry is not a complete list of v1 projects.

`db/bit.db` in the repo is a throwaway dev db for the Justfile's `dbmate` targets (gitignored).

## How migrations are applied
- **Runtime: embedded.** `//go:embed migrations/*.sql`; every `db.Open()` runs dbmate's Go library `CreateAndMigrate()` (`Strict` false, AutoDumpSchema off, log discarded). dbmate v2.35.0.
- Custom dbmate driver `db/driver.go` registered as `sqlite` using pure-Go `modernc.org/sqlite` (no cgo, no sqlite3 binary; DumpSchema deliberately errors).
- **Dev: dbmate CLI** via Justfile (`db-migrate name`, `db-up`, `db-down`, `db-status`) against `db/bit.db`.
- With `Strict=false`, dbmate applies only pending versions from the embedded FS and ignores applied versions it doesn't know — so a newer binary's migrations don't make an older binary error, **but** a schema change that alters columns v1's sqlc queries use would break v1.

## Fit with the operator's plan
The plan (projects table, sqlc + dbmate, under local/share) **already exists in shape** — v2 extends it rather than introducing it. Differences to decide:
- Dir name: code uses `bit-pro`, plan says `bit`. Using `~/.local/share/bit/` for v2 conveniently gives v2 a **separate db** from v1's `bit-pro/bit.db`, avoiding shared-schema breakage during transition (see [coexistence](coexistence.md)). Migrate would copy the registry rows across.
- Layout: `~/.local/share/bit/bit.db` + `~/.local/share/bit/<CODE>/` per project. Codes are uppercased by `task.NormalizeID`; macOS APFS is case-insensitive by default, so `bit/BIT` vs a lowercase dir name collide — pick one casing (operator wrote `local/share/bit/bit`).
- Schema changes likely: `code UNIQUE`, a repos/paths table (path UNIQUE, FK project) for multi-repo, config columns (prefix), migration marker (`migrated_at`, source path), maybe ID counter.
- `store.Dir()` MkdirAll's on every call (side effect on read paths).
