---
id: BIT-29
title: 'add and list: project registry in SQLite'
status: todo
---
## Why

The daemon needs to know which projects it's watching, and the operator needs a way to see what's enrolled. Today `bp init` sets up a project locally but nothing registers it anywhere global. Starting the daemon cold, or setting up on a second machine, gives you no list to work from. `bp add` and `bp list` create that global registry — the source of truth the loop will eventually read.

## Summary

SQLite database at `~/.local/share/bit-pro/state.db` with a `projects` table (path, code). `bp add <path>` enrolls a project interactively and sets up the DB if it doesn't exist yet. `bp list` shows all enrolled projects. Dbmate authors migration files; sqlc generates the Go query layer; the binary applies migrations at runtime without requiring dbmate on the user's machine.

## Decisions

- **SQLite at `~/.local/share/bit-pro/state.db`** — same dir as daemon state; one place for all global bit-pro state.
- **dbmate for authoring migrations, sqlc for query generation.** Both are dev tools used to build the binary, not runtime dependencies for the operator.
- **Migrations are embedded in the binary and applied at runtime using `golang-migrate`.** `bp add` runs `db up` on first call (idempotent). The operator does not need dbmate installed.
- **Pure-Go SQLite driver (`modernc.org/sqlite`).** No CGO — binary stays self-contained.
- **Projects table: `path TEXT PRIMARY KEY, code TEXT NOT NULL`.** Absolute path is the natural unique key; code is the human identifier.
- **`bp add .` resolves `.` to an absolute path before lookup.** Same project registered from two directories is one row, not two.
- **If the path already exists in the DB, `bp add` prints "already added" and exits cleanly.** No re-prompting.
- **If `.bit/config.toml` is present at the path, its `prefix` value is read and offered as the default code.** Accepting requires only Enter.
- **If `.bit/` is absent, `bp add` runs the init flow minus `config.toml` creation** — writes `.claude/settings.json` and syncs the plugin — then registers the project.
- **`bp list` output: `code\tpath` rows, tab-separated, no headers.** Monospace tab-stop alignment is enough; no library needed.
- **Just targets for dev workflow follow the evus pattern:** `db-migrate <name>`, `db-up`, `db-down`, `db-status`, `db-gen-queries`.

## Verses

- [ ] Verse 1 — The project registry exists: dbmate migration authors the schema, sqlc generates the Go query layer, just targets manage the dev workflow. Migration applies cleanly on a fresh machine.
  Touches: `db/migrations/` (new), `db/queries/` (new), `internal/store/` (new, generated), `Justfile` — where to look to verify.

- [ ] Verse 2 — Operator can register a project: `bp add <path>` creates the DB if absent, prompts for a project code (suggesting the `.bit` prefix when present), inserts the row. Re-running on the same path prints "already added". Running on a path without `.bit/` runs the init flow first.
  Touches: `cmd/add.go` (new), `cmd/root.go` — where to look to verify.

- [ ] Verse 3 — Operator can see all registered projects: `bp list` prints every row as `code\tpath`, tab-aligned, no headers.
  Touches: `cmd/list.go` (new), `cmd/root.go` — where to look to verify.

## References

- `/Users/appstack/Developer/UniqueDataManagement/clients/engage-voters/evus/justfiles/db.just` — the dbmate + sqlc + just pattern this track's dev workflow follows. Informs Verse 1.