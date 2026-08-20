---
id: BIT-29.2
title: The generated query layer round-trips a project
status: todo
approved: true
phase: 1
phase_label: registry
---
## **Verse 1**

A generated query layer round-trips a project row. Contradicts BIT-29.1, whose only reach into the
database is raw `sqlite_master` SQL in a test — nothing in the package can yet write or read a
`projects` row.

## Scope
- `sqlc.yaml` (new, repo root) — the sqlc config
- `db/queries/projects.sql` (new) — the hand-written queries
- `db/orm/db.go`, `db/orm/models.go`, `db/orm/projects.sql.go` (new, generated) — gitignored, never
  committed
- `.gitignore` — `db/orm/`
- `Justfile` — `db-gen-queries`, and the dependency on it from `install`, `run`, `test`, `lint`
- `install.sh` — `sqlc generate` ahead of the `go build`
- `db/queries_test.go` (new) — the test, in package `db`, importing `db/orm`

The generated files are a build product, not source: `.gitignore` gets `db/orm/` and nothing under it
is committed. That is why the codegen wiring lands in this bar rather than BIT-29.3 — the moment the
output stops being committed, everything that compiles Go has to produce it first, or the very next
`just test` on a clean checkout fails. Both are scope Decisions.

The test lives in package `db`, not in `db/orm/`. Measured: `db/orm/` in `.gitignore` cannot be
undone by a `!db/orm/*_test.go` negation, because git does not descend into an excluded directory —
a test written next to the generated code would itself be untracked. Package `db` imports `db/orm`,
and there is no cycle: `db/orm` imports nothing from `db`.

Verified with `sqlc v1.31.1` against this exact config and a `-- migrate:up` / `-- migrate:down`
migration before planning this bar — `sqlc generate` exits 0 and emits the three files into
`db/orm/` under `package orm`. The scope's Decision that sqlc reads the dbmate format directly
holds. If it errors on the `-- migrate:down` section, that contradicts a scope Decision and goes
back to bit_scope rather than getting worked around with a separate schema file.

## TDD cycle

1. **Write test (RED):** `db/queries_test.go`
   - [ ] `TestProjects_RoundTrip`
     - **Behavior:** the registry is readable and writable through typed Go, which is what `bp add`
       and `bp list` will stand on — a migrated database nothing can query is not a registry.
     - **Setup:** `home := t.TempDir()`; `t.Setenv("HOME", home)`; `t.Setenv("XDG_DATA_HOME", "")`.
       `sqlDB, err := Open()`; `defer sqlDB.Close()`; `q := orm.New(sqlDB)`. Then
       `q.CreateProject(t.Context(), orm.CreateProjectParams{Path: "/tmp/alpha", Code: "ALPHA"})`.
     - **Assertions:** `CreateProject` returns nil. `q.ListProjects(t.Context())` returns a slice of
       length 1 whose element has `Path == "/tmp/alpha"`, `Code == "ALPHA"`, and `ID != 0` — the
       `INTEGER PRIMARY KEY` the queue table will later point at has actually been minted.
     - **Boundary:** row count 0 → 1 — `ListProjects` at the empty-table lower bound on the way in
       and the first row on the way out, and the first `id` SQLite assigns.
   - [ ] Confirm fails: `no required module provides package github.com/B4Dmonkey/bit-pro/db/orm`
         and `FAIL github.com/B4Dmonkey/bit-pro/db [setup failed]` — the verified failure text for an
         in-module package directory that does not exist yet. Not `undefined: New`; the import is
         what breaks first.

2. **Implement (GREEN):**
   - [ ] `sqlc.yaml`: `version: "2"`, one `sql` entry with `engine: "sqlite"`,
         `queries: "./db/queries"`, `schema: "./db/migrations"`, and `gen.go` of
         `package: "orm"`, `out: "db/orm"`.
   - [ ] `db/queries/projects.sql`:
         `-- name: CreateProject :exec` / `INSERT INTO projects (path, code) VALUES (?, ?);` and
         `-- name: ListProjects :many` / `SELECT id, path, code FROM projects;`
         `:exec`, not `:one` with `RETURNING id` — nothing needs the id back yet, and no `ORDER BY`
         — BIT-29.9 contradicts that.
   - [ ] `.gitignore`: add `db/orm/`.
   - [ ] `Justfile`: a `db-gen-queries` recipe running `sqlc generate`, then add it as a dependency of
         `install`, `run`, `test`, and `lint` — `test: db-gen-queries`, `run *ARGS: db-gen-queries`,
         and so on. Verified: a dependency on a variadic recipe parses, and it runs ahead of
         `install`'s shebang body. BIT-29.3 adds the remaining `db-*` targets and the `DATABASE_URL`
         they need; this bar adds only the one the build itself depends on.
   - [ ] `install.sh`: `sqlc generate` before the `go build` line, with a progress `echo` in the
         script's existing style. It does not shell `just`, so the Justfile dependency never reaches
         it — this is the second half of the same Decision, not a duplicate.
   - [ ] `sqlc generate` from the repo root. Nothing to commit under `db/orm/`. Positional `?` params
         are named from the column, so the generated shapes are
         `orm.CreateProjectParams{Path, Code string}` and `orm.Project{ID int64, Path, Code string}`
         — verified, so a mismatch here means the config drifted.

## Claude verifies
- [ ] `just test`
- [ ] `just lint` — the generated files carry sqlc's `Code generated by sqlc. DO NOT EDIT.` header
      and `.golangci.yml` sets `exclusions.generated: lax`, so they should not produce findings. If
      they do, that is the signal, not a reason to add a per-path exclusion.
- [ ] `git status --porcelain` lists nothing under `db/orm/` — the ignore covers all three generated
      files, and `db/queries/projects.sql` and `sqlc.yaml` are still shown as additions.
- [ ] `rm -rf db/orm && just test` exits 0 — the codegen dependency rebuilds what the ignore drops,
      so a fresh clone is one command from green. This is the check the whole gitignore Decision
      rests on.
- [ ] `rm -rf db/orm && ./install.sh` builds — the documented install path, which does not go through
      just.

## User verifies
- [ ] none — deterministic. Verse 1's whole-slice check lands on BIT-29.3.

## Commit (user)
`feat(db): generate the projects query layer with sqlc`