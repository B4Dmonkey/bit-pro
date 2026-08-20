---
id: BIT-29.1
title: A fresh machine gets a migrated registry
status: done
phase: 1
phase_label: registry
---
## **Verse 1**

`db.Open()` brings an absent registry into existence: creates `bit.db` under `store.Dir()`, applies
the embedded migration, and hands back an open `*sql.DB`. Nothing forces this one — it is the
walking skeleton the rest of the track stands on.

## Scope
- `go.mod` — `github.com/amacneil/dbmate/v2` and `modernc.org/sqlite`
- `db/migrations/<ts>_create_projects.sql` (new) — the schema
- `db/driver.go` (new) — bit-pro's `dbmate.Driver` over `modernc.org/sqlite`
- `db/open.go` (new) — the `embed.FS` and `Open()`
- `db/open_test.go` (new) — the test

sqlc generates into `db/orm` (BIT-29.2), not into this package, so nothing here has to dodge a
generated name. `open.go`/`driver.go` and the entry point `Open` are chosen for what they do — a
`db.New` sitting beside `orm.New` would read as the same constructor twice.

Verified against `dbmate/v2@v2.35.0` and `modernc.org/sqlite@v1.57.0` before planning this bar:

- `dbmate.RegisterDriver(f DriverFunc, scheme string)` is exported from `pkg/dbmate/driver.go`, and
  `DB.Driver()` resolves the driver by `DatabaseURL.Scheme`. Registering under `"sqlite"` is all it
  takes — no fork, no patched copy.
- `DB.FS` is an `fs.FS`, and `readMigrationsDir` runs `path.Clean(dir)` before `fs.ReadDir`. So an
  `embed.FS` declared in `db/` with `//go:embed migrations/*.sql` pairs with
  `MigrationsDir = []string{"migrations"}`.
- `modernc.org/sqlite` registers itself with `database/sql` under the name `"sqlite"`
  (`sqlite.go:50,56`), so `sql.Open("sqlite", path)` is the whole CGO story — the blank import and
  that string are the only coupling dbmate's own driver has to `mattn/go-sqlite3`.
- `dbutil.Transaction` is a three-method interface (`Exec`/`Query`/`QueryRow`), and `dbutil.MustClose`
  is exported. `InsertMigration`/`DeleteMigration` take it.

## References
- `automation-notes.md` (repo root, untracked) — the automation phase's working notes; this track is
  its step 2. Read the measured facts before touching the dbmate wiring.

## TDD cycle

1. **Write test (RED):** `db/open_test.go`
   - [ ] `TestOpen_MigratesAFreshDatabase`
     - **Behavior:** an operator on a machine that has never run bit-pro gets a working registry
       from the first command that needs one — no install step, no migration tool on their PATH.
     - **Setup:** `home := t.TempDir()`; `t.Setenv("HOME", home)`; `t.Setenv("XDG_DATA_HOME", "")`
       (the pair `store/store_test.go` and `cmd/start_test.go` already use to redirect
       `store.Dir()`). Call `sqlDB, err := Open()`; `defer sqlDB.Close()`.
     - **Assertions:** `err` is nil. `os.Stat(filepath.Join(home, ".local", "share", "bit-pro",
       "bit.db"))` succeeds. `sqlDB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND
       name='projects'")` scans `"projects"`. `sqlDB.QueryRow("SELECT count(*) FROM
       schema_migrations")` scans `1`.
     - **Boundary:** applied-migration count at 0 — the lower bound of the range, a database that
       has never been migrated and does not yet exist on disk.
   - [ ] Confirm fails: `undefined: Open` — the package has no such function yet.

2. **Implement (GREEN):**
   - [ ] `go get github.com/amacneil/dbmate/v2@v2.35.0` and `go get modernc.org/sqlite`.
   - [ ] Generate the migration with the installed CLI — `dbmate --migrations-dir db/migrations new
         create_projects` (verified: it needs no `DATABASE_URL` and exits 0, stamping the
         timestamp). Fill it in:
         `-- migrate:up` / `CREATE TABLE projects (id INTEGER PRIMARY KEY, path TEXT NOT NULL
         UNIQUE, code TEXT NOT NULL);` / `-- migrate:down` / `DROP TABLE projects;`
   - [ ] `db/driver.go`: a `driver` struct holding `migrationsTableName`, `databaseURL`, `log`;
         `newDriver(dbmate.DriverConfig) dbmate.Driver`; `func init() {
         dbmate.RegisterDriver(newDriver, "sqlite") }`; and the blank import
         `_ "modernc.org/sqlite"`.
   - [ ] `db/driver.go`: implement all twelve `dbmate.Driver` methods. Model them on
         `pkg/driver/sqlite/sqlite.go`, with three deliberate departures:
         - the file path is `drv.databaseURL.Path` read straight — `Open()` builds the URL itself
           from an absolute `store.Dir()` path, so dbmate's `normalizeSQLiteURL`/`ConnectionString`
           pair has nothing to normalise and drops out;
         - `quoteIdentifier` is local (`` `"` + strings.ReplaceAll(s, `"`, `""`) + `"` `` ) rather
           than `pq.QuoteIdentifier`, so `lib/pq` never enters the module graph for one call;
         - `DumpSchema` returns an error rather than shelling out to a `sqlite3` binary. The scope
           decides nothing in the runtime path may call it, and this is the enforcement.
   - [ ] `db/open.go`: `//go:embed migrations/*.sql` over `var migrationsFS embed.FS`.
   - [ ] `db/open.go`: `Open() (*sql.DB, error)` — `store.Dir()`, `path := filepath.Join(dir,
         "bit.db")`, `mate := dbmate.New(&url.URL{Scheme: "sqlite", Path: path})`, then
         `mate.FS = migrationsFS`, `mate.MigrationsDir = []string{"migrations"}`,
         `mate.AutoDumpSchema = false`, `mate.Log = io.Discard`, `mate.CreateAndMigrate()`, and
         finally `sql.Open("sqlite", path)`.
   - [ ] Both overrides on `mate` are scope Decisions, not tuning, and neither is forced by this
         bar's test — set them here anyway. `AutoDumpSchema` defaults to `true` and `Migrate()` calls
         `DumpSchema()` on success. `Log` defaults to `os.Stdout` and `Migrate()` writes
         `Applying: …` / `Applied: …` to it, which would break "`bp list` … then prints nothing" on
         the very first run on a fresh machine — the one run where there is a migration to apply.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`
- [ ] `CGO_ENABLED=0 go build ./...` exits 0 — the no-CGO claim the scope rests on, checked at the
      bar that introduces the dependency rather than assumed downstream.

## User verifies
- [ ] none — deterministic. Verse 1's whole-slice check lands on BIT-29.3.

## Commit (user)
`feat(db): create and migrate the registry from an embedded migration`