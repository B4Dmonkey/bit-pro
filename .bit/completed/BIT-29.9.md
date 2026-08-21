---
id: BIT-29.9
title: The registry prints as tab-separated rows ordered by code
status: done
phase: 3
phase_label: list
---
## **Verse 3**

`bp list` prints the registry as `code<TAB>path` rows ordered by code, and prints nothing at all on a
machine that has no registry yet. Contradicts BIT-29.4 through BIT-29.8, which can put projects into
the registry and give the operator no way to see them.

## Scope
- `db/queries/projects.sql` — `ORDER BY code` on `ListProjects`
- `db/orm/projects.sql.go` (generated) — regenerated, not committed
- `cmd/list.go` (new) — the command
- `cmd/root.go` — register it
- `cmd/list_test.go` (new) — the test

Ordering lives in the SQL rather than in a sort in `cmd/list.go`: it is a property of what the
registry hands out, and later readers of the same query get it for free.

## TDD cycle

1. **Write test (RED):** `cmd/list_test.go`
   - [ ] `TestListCmd_PrintsProjectsByCode` (table-driven: `"three projects"` and
         `"no database yet"`)
     - **Behavior:** the operator can see everything the daemon is watching, in an order that does
       not shift as projects are enrolled — so two runs are comparable. And the first run on a fresh
       machine is not an error: nothing enrolled is nothing printed, which is what lets anything
       piping `bp list` skip a special case.
     - **Setup:** both rows: `home := t.TempDir()`; `t.Setenv("HOME", home)`;
       `t.Setenv("XDG_DATA_HOME", "")`. The three-project row seeds the registry directly through
       the `db` and `db/orm` packages — `db.Open()`, then `orm.New(sqlDB).CreateProject` for `/tmp/mid`/`MID`,
       `/tmp/zed`/`ZED`, `/tmp/ace`/`ACE`, **in that order**, so insertion order is deliberately not
       code order. Seeding through `db`/`orm` rather than through `bp add` keeps the fixture from
       depending on verse 2's prompt. The no-database row seeds nothing and must not call
       `db.Open()` — the point is that `bp list` is the first thing to touch the file. Then
       `out, err := run(t, listCmdUse)`.
     - **Assertions:** three-project row —
       `out == "ACE\t/tmp/ace\nMID\t/tmp/mid\nZED\t/tmp/zed\n"`, exact, which pins the tab
       separator, the code-then-path order, the absence of a header, and the alphabetical ordering
       in one comparison. No-database row — `out == ""` and `err` is nil, and
       `os.Stat(filepath.Join(home, ".local", "share", "bit-pro", "bit.db"))` then succeeds.
     - **Boundary:** row count at both ends of the range — 0, the empty table, and 3, where N > 1
       makes ordering observable at all. The three codes are seeded in an order that a missing
       `ORDER BY` (SQLite's rowid order) would print as `MID`, `ZED`, `ACE`, so forgetting it fails
       this row rather than passing by luck.
   - [ ] Confirm fails: `Error: unknown command "list" for "bp"` on both rows. `bp` has a
         `task list`, but no top-level `list`.

2. **Implement (GREEN):**
   - [ ] `db/queries/projects.sql`: add `ORDER BY code` to `ListProjects`, then `sqlc generate`
         (or let `just test` do it). Nothing to commit — `db/orm/` is gitignored.
   - [ ] `cmd/list.go`: `const listCmdUse = "list"` and `newListCmd() *cobra.Command` with
         `Use: listCmdUse`, `Args: cobra.NoArgs`, and a `Short`. In `RunE`: `db.Open()`,
         `defer sqlDB.Close()`, `orm.New(sqlDB).ListProjects(cmd.Context())`, and one
         `fmt.Fprintf(out, "%s\t%s\n", p.Code, p.Path)` per row.
   - [ ] `cmd/root.go`: `rootCmd.AddCommand(newListCmd())`.

## Claude verifies
- [ ] `just test`
- [ ] `just lint`
- [ ] `just install`, then `XDG_DATA_HOME=$(mktemp -d) bp list` prints nothing and exits 0 — the
      built binary, not just the test harness, and against a data directory that has never been
      migrated.

## User verifies
- [ ] Whole slice: `XDG_DATA_HOME=$(mktemp -d) bp list` in a real terminal — no output at all, not
      even a migration line, which is the scope's "creates and migrates the DB when it is absent,
      then prints nothing". Then `bp add .` in this repo and `bp list` against the real registry —
      the rows print, and the tabs line the paths up into a column. That is the capability the verse
      is for: one command that says what the daemon is watching.

## Commit (user)
`feat(list): print the registered projects ordered by code`