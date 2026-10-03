---
id: BIT-49.10
title: bp migrate copies a v1 project's active tasks into the central store and registers the folder
status: done
approved: true
phase: 2
phase_label: bp migrate
---
## **Verse 2**

This is the walking skeleton: from a folder with a v1 `.bit/`, `bp migrate` leaves the tasks readable through the v2 store and the folder registered, and the source untouched. It handles `tasks/` only. Later bars bring the other record kinds, git fields, staging and verify, and the refusals, each forced by its own test.

## Scope
- new package `migrate` (`migrate/migrate.go`), with business logic kept out of cobra (cobra-viper):
  - `type Options struct { Dir string }`. `Dir` is the session dir, the CLI's cwd. Later bars add fields.
  - `type Result struct { Code, Path string }`.
  - `func Run(ctx context.Context, q *orm.Queries, opts Options) (Result, error)`:
    1. `src := filepath.Join(opts.Dir, ".bit")`. BIT-49.23 adds the walk-up and worktree handling.
    2. Read `src/config.toml` with `github.com/BurntSushi/toml` into `struct{ Prefix string \`toml:"prefix"\` }`. The raw value is kept for BIT-49.19. `code := prefix`, and BIT-49.22 adds `project.ValidateCode`.
    3. `path, err := project.CanonicalPath(filepath.Dir(src))`.
    4. `data, err := store.Dir()`; `root, err := store.ProjectDir(code)`; `s := task.NewProject(root, code).WithDataRoot(data)`.
    5. For each `os.ReadDir(src/tasks)` entry ending in `.md`: `task.Parse` → `s.Save(t)`. A Parse error is returned for now.
    6. `q.CreateProject(ctx, orm.CreateProjectParams{Path: path, Code: code})`.
  - Every error from these steps is returned, wrapped with what was being done (go: no discarded errors).
  - The source is only ever read, never written.
- new `cmd/migrate.go`: `newMigrateCmd()` with `Use: "migrate"`, `Short: "Copy this folder's v1 .bit/ into the central store and register it"`, and `Args: cobra.NoArgs`. `RunE`: `os.Getwd` → `db.Open` (returns `*sql.DB`; `defer` its `Close`) → `migrate.Run(cmd.Context(), orm.New(sqlDB), migrate.Options{Dir: wd})` → print `migrated <CODE> <path>` (matching `bp add`'s `added <CODE> <path>`). BIT-49.24 adds the `claude.Runner` parameter when the wiring arrives.
- `cmd/root.go`: `rootCmd.AddCommand(newMigrateCmd())`.
- `go.mod`: BIT-46.11 deleted `task/config.go`, its last importer, and its `go mod tidy` dropped `github.com/BurntSushi/toml`. This bar re-adds it as a direct dependency: `go get github.com/BurntSushi/toml@v1.6.0` (the version on v2 @ 6a1d345), then `go mod tidy`.
- Test file touched: new `cmd/migrate_test.go` with a `TestMigrateCmd` function. Each case sandboxes `HOME`/`XDG_DATA_HOME` with BIT-46.9's idempotent per-test sandbox helper `mcpSandbox(t)`, not its own `t.Setenv`, so a later `mcpSession` in the same case (BIT-49.13, BIT-49.14) reuses that sandbox instead of replacing it. It also has a helper `writeV1Store(t, dir, files map[string][]byte)` that writes raw files under `dir/.bit/`. Task fixtures come from `(&task.Task{...}).Bytes()`, so they round-trip exactly (BIT-49.18 checks that).

## TDD cycle

1. **Write test (RED):**
   - [ ] `TestMigrateCmd/copies active tasks and registers the folder`
     - **Behavior:** a v1 project's live tasks become readable from the central store, and its folder resolves as a registered project, with the source left exactly as it was.
     - **Setup:** sandboxed `HOME`/`XDG_DATA_HOME`. `dir := t.TempDir()` with `.bit/config.toml` = `prefix = "BIT"\n`. `tasks/BIT-1.md` is a track (`approved: true`, `order: [BIT-1.2, BIT-1.1]`, a multi-line body with a verse checklist). `BIT-1.1.md` and `BIT-1.2.md` are bars (`phase: 1`, `phase_label: "skeleton"`, one `done`, one `todo`). Hash every source file. `t.Chdir(dir)`; `bp migrate`.
     - **Assertions:**
       - The output's first line is `migrated BIT <canonical dir>` (later bars append the cleanup and wiring lines).
       - `ListProjects` has exactly one row, `{Code: "BIT", Path: canonical dir}`.
       - `project.OpenStore(ctx, dir)` loads `BIT-1` with the same title, `approved`, `order` and body, and `BIT-1.1` with the same `phase`, `phase_label` and status.
       - Every source file hashes the same as before.
     - **Boundary:** the smallest real store: one track, its bars and an `order` list.
   - [ ] Confirm fails: `unknown command "migrate"`.

2. **Implement (GREEN):**
   - [ ] `migrate.Run`, `newMigrateCmd`, the root registration and the toml dependency.

3. **More tests (RED → GREEN):**
   - [ ] `TestMigrateCmd/a migrated project works with bp task commands`: after the migration, `bp task create "Next"` from `dir` prints `BIT-2`. *Boundary:* the registered project resolves, and the next ID follows the copied ones.

## Claude verifies
- [ ] `just lint` and `just test` pass
- [ ] `go mod tidy` leaves no diff
- [ ] `test ! -e ~/.local/share/bit/main.db`

## User verifies
- none, deterministic (the verse's end-to-end check is on its last bar)

## Commit (user)
`feat(bit): bp migrate copies a v1 project's tasks into the central store`