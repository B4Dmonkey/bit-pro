---
id: BIT-29
title: 'add and list: project registry in SQLite'
status: doing
---
## Why

The daemon needs to know which projects it's watching, and the operator needs a way to see what's enrolled. Today `bp init` sets up a project locally but nothing registers it anywhere global. Starting the daemon cold, or setting up on a second machine, gives you no list to work from. `bp add` and `bp list` create that global registry — the source of truth the loop will eventually read.

## Summary

SQLite database at `~/.local/share/bit-pro/bit.db` with a `projects` table (id, path, code). `bp add <path>` enrolls a project interactively and sets up the DB if it doesn't exist yet. `bp list` shows all enrolled projects. Dbmate authors the migration files and applies them, both in the dev workflow and inside the binary at runtime; sqlc generates the Go query layer. Migrations are embedded in the binary, so the operator installs nothing.

## Decisions

- **SQLite at `~/.local/share/bit-pro/bit.db`** — same dir as daemon state; one place for all global bit-pro state. Named for the tool, not for one of its tables, because the queue and project counts land in the same file.
- **dbmate authors the migrations and applies them, at both ends.** The `dbmate` CLI drives the dev workflow; dbmate's Go library (`pkg/dbmate`) is what the binary calls at runtime. One tool, one migration file format, one `schema_migrations` ledger shape at both ends — the dev database and an operator's database are migrated by the same engine.
- **sqlc generates the query layer.** Both dbmate and sqlc are dev-time tools for building the binary, never runtime dependencies for the operator.
- **Migrations are embedded in the binary with `embed.FS` and applied via dbmate's `DB.FS` field.** `bp add` calls `CreateAndMigrate` on first run. Verified idempotent — a second run applies nothing and exits clean.
- **Pure-Go SQLite (`modernc.org/sqlite`). No CGO — and dbmate at runtime does not compromise this.** dbmate's bundled sqlite driver is `//go:build cgo` over `mattn/go-sqlite3`, so bit-pro registers its own modernc-backed dbmate driver through the exported `dbmate.RegisterDriver`. Measured: the only CGO coupling in that driver is the blank import and the name passed to `sql.Open` — everything else is plain `database/sql`. A port builds and migrates at `CGO_ENABLED=0`.
- **bit-pro writes its own minimal dbmate driver rather than copying and patching dbmate's.** dbmate's sqlite driver is 291 lines, pulls in `lib/pq` for a single `QuoteIdentifier` call, and its `DumpSchema` shells out to the `sqlite3` binary the decision below forbids. A driver written from scratch over `modernc.org/sqlite` is roughly 100 lines and adds no dependency beyond that one, and the URL-normalisation code drops out because bit-pro builds the database URL itself from `store.Dir()`. Verified end to end: builds at `CGO_ENABLED=0`, `CreateAndMigrate` applies the migration, and a second run applies nothing.
- **`AutoDumpSchema` stays off at runtime.** dbmate's `DumpSchema` shells out to a `sqlite3` binary, which an operator will not have. Nothing in the runtime path may call it.
- **The hand-written dbmate layer is a new top-level `db` package; sqlc generates into `db/orm` (package `orm`) beneath it.** Not `internal/` — the repo has no internal tree and new packages don't go there. Not `store` — that name is taken by the state-dir resolver (`store.Dir()`). `db` is named for the file rather than one of its tables, so the queue and counts land in the same package later. The generated code gets its own directory because it is gitignored (below): one unambiguous `.gitignore` line beats three generated filenames sitting next to hand-written `db/open.go`, and the split also means package `db` no longer has to avoid the names sqlc emits.
- **sqlc reads the dbmate migration files directly — no separate schema file.** Verified against sqlc v1.31.1 with `engine: "sqlite"`, `schema: "./db/migrations"`, `queries: "./db/queries"`, `out: "db/orm"`, `package: "orm"`: sqlc parses the `-- migrate:up` / `-- migrate:down` format without help, exits 0, and emits `db/orm/db.go`, `db/orm/models.go`, and `db/orm/<name>.sql.go` under `package orm`. The relocation does not change the generated shapes — `orm.New`, `orm.Queries`, `orm.Project{ID int64, Path, Code string}`, `orm.CreateProjectParams{Path, Code string}`.
- **The generated query layer is gitignored, not committed** — `.gitignore` gets `db/orm/`. `db/migrations/` and `db/queries/` stay tracked: the hand-written SQL is the source of truth and the Go is a build product. The consequence is accepted deliberately — a fresh clone does not compile until codegen runs, so `sqlc` becomes a build prerequisite alongside the Go toolchain. Measured: a `!db/orm/*_test.go` negation cannot rescue a test file from inside the ignored directory — git does not descend into an excluded directory — so the generated layer's own test lives in package `db` and imports `db/orm`.
- **`install.sh` runs `sqlc generate` before `go build`** rather than delegating to `just install`. `git clone` + `./install.sh` is the documented install path and has to work on a fresh clone; duplicating the one line keeps the script standalone, needing only a Go toolchain and sqlc rather than `just` as well.
- **Every Justfile target that compiles Go depends on the codegen target** — `install`, `run`, `test`, and `lint` all depend on `db-gen-queries`, so a fresh clone builds in one command and no target can fail for a missing generated file. sqlc over two SQL files costs nothing measurable, so there is no reason to be selective.
- **Projects table: `id INTEGER PRIMARY KEY, path TEXT NOT NULL UNIQUE, code TEXT NOT NULL`.** The synthetic `id` is the FK target the queue table will point at, so a project moving on disk doesn't orphan its queued rows; `path` stays unique so enrolling the same directory twice is still one row.
- **`bp add .` resolves `.` to an absolute path before lookup.** Same project registered from two directories is one row, not two.
- **A successful enrollment prints `added <CODE> <path>`** — the stored code and the resolved absolute path, space-separated, in the same code-then-path order `bp list` uses. A test asserts it verbatim.
- **If the path already exists in the DB, `bp add` prints "already added" and exits cleanly.** No re-prompting.
- **If `.bit/config.toml` is present at the path, its `prefix` value is read and offered as the default code.** Accepting requires only Enter.
- **The prompt reads `Project code (BIT): `** — the scope's own vocabulary (the column is `code`), in `bp init`'s parenthesised-default form (`Task ID prefix (BIT): `) so the two prompts still read as the same tool. A test asserts it verbatim, the way `init_test.go` does.
- **With no `.bit/config.toml` to read, the prompt reads `Project code: `** — `bp init`'s no-default form (`Task ID prefix: `), for the same reason the parenthesised form was chosen: the two prompts stay recognisably the same tool. This is the case the init-flow decision below reaches, where no prefix exists yet.
- **If `.bit/` is absent, `bp add` runs the init flow minus `config.toml` creation** — writes `.claude/settings.json` and syncs the plugin — then registers the project.
- **`bp add` is run from inside the project, the same way `bp init` is.** The init flow shells `claude plugin update bit@bit-pro --scope project`, and `--scope project` resolves against the process working directory — neither `claude plugin install` nor `claude plugin update` has a `--cwd`/`--dir` flag. So the sync only lands in the right project when the working directory *is* the project, which `bp add .` guarantees. Running `bp add <some-other-path>` on a project that has no `.bit/` yet is explicitly **out of scope for this track**.
- **The project code is normalised to uppercase, exactly as `bp init` normalises the prefix** (`task.normalizeID`, used by `task/config.go`). A code typed at the `bp add` prompt is stored the same way a prefix is, so the two can never disagree about casing.
- **Project code and task prefix are the same concept, but the terminology stays split for now.** Existing `config.toml` files on disk say `prefix`, so both words remain in the code until those can be migrated. Unifying the naming is explicitly *not* part of this track.
- **`bp list` output: `code\tpath` rows, tab-separated, no headers.** Monospace tab-stop alignment is enough; no library needed.
- **`bp list` orders rows alphabetically by code.** A stable order that does not shift as projects are enrolled, so two runs are comparable.
- **`bp list` creates and migrates the DB when it is absent, then prints nothing.** Same `CreateAndMigrate` path as `bp add`, so a first run on a fresh machine is never an error: zero rows is zero lines of output and exit 0, and anything piping `bp list` needs no special case.
- **Just targets for the dev workflow follow the evus pattern:** `db-migrate <name>`, `db-up`, `db-down`, `db-status`, `db-gen-queries`. They drive the `dbmate` CLI, same as evus.
- **The `db-*` just targets drive a repo-local, gitignored `db/bit.db` — a testing database, not the real registry.** They exist to rehearse migrations; `db-down` against `~/.local/share/bit-pro/bit.db` would delete live enrollments, and later, queued work. No env override: the real database is only ever migrated by the binary through the embedded FS, so the Justfile never reimplements `store.Dir()`'s XDG-or-`HOME` resolution.

## Verses

- [x] Verse 1 — The project registry exists: a dbmate migration authors the schema, sqlc generates the Go query layer, just targets manage the dev workflow. Migration applies cleanly on a fresh machine.
  Touches: `db/migrations/` (new), `db/queries/` (new), `db/orm/` (new, sqlc-generated, gitignored), `Justfile` — where to look to verify.

- [ ] Verse 2 — Operator can register a project: `bp add <path>` creates the DB if absent, prompts for a project code (suggesting the `.bit` prefix when present), inserts the row. Re-running on the same path prints "already added". Running on a path without `.bit/` runs the init flow first.
  Touches: `cmd/add.go` (new), `cmd/root.go` — where to look to verify.

- [ ] Verse 3 — Operator can see all registered projects: `bp list` prints every row as `code\tpath`, tab-aligned, no headers.
  Touches: `cmd/list.go` (new), `cmd/root.go` — where to look to verify.

## References

- `automation-notes.md` (repo root, untracked) — the working notes for the automation phase; this track is its step 2. Holds the ordered todo, the settled decisions, and the measured facts the whole phase rests on. Informs all verses.
- `/Users/appstack/Developer/UniqueDataManagement/clients/engage-voters/evus/justfiles/db.just` — the dbmate + sqlc + just pattern this track's dev workflow follows. Informs Verse 1.